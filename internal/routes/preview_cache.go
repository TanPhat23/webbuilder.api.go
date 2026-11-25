package routes

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
)

const (
	previewCacheTTL     = 30 * time.Second
	previewCacheTimeout = 150 * time.Millisecond
)

func cachePublicPreview(client *redis.Client, projects repositories.ProjectRepositoryInterface, resource string) fiber.Handler {
	if client != nil {
		client = client.WithTimeout(previewCacheTimeout)
	}
	return func(c fiber.Ctx) error {
		if c.Query("previewCache") != "1" {
			return c.Next()
		}
		return serveCachedPreview(c, client, projects, resource)
	}
}

func serveCachedPreview(c fiber.Ctx, client *redis.Client, projects repositories.ProjectRepositoryInterface, resource string) error {
	projectID, err := utils.RequireParam(c, "projectid")
	if err != nil {
		return err
	}
	project, err := projects.GetPublicByID(c.RequestCtx(), projectID)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to verify publication")
	}
	if project == nil || !project.Published {
		return fiber.NewError(fiber.StatusNotFound, "Project not found")
	}
	c.Locals(handlers.PublishedProjectVerifiedKey, true)
	if client == nil || project.AuditFields == nil {
		return c.Next()
	}

	// Publishing stamps UpdatedAt, so a republish cannot reuse a stale key.
	revision := strconv.FormatInt(project.UpdatedAt.UnixNano(), 10)
	key := "webbuilder:preview:v1:" + resource + ":" + projectID + ":" + revision
	cached, cacheErr := readPreviewCacheEntry(c.Context(), client, key)
	if cacheErr == nil {
		c.Type("json")
		return c.Status(fiber.StatusOK).Send(cached)
	}
	loadErr := c.Next()
	if loadErr == nil && c.Response().StatusCode() == fiber.StatusOK && errors.Is(cacheErr, redis.Nil) {
		writePreviewCacheEntry(c.Context(), client, key, c.Response().Body())
	}
	return loadErr
}

func readPreviewCacheEntry(parent context.Context, client *redis.Client, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, previewCacheTimeout)
	defer cancel()
	return client.Get(ctx, key).Bytes()
}

func writePreviewCacheEntry(parent context.Context, client *redis.Client, key string, body []byte) {
	ctx, cancel := context.WithTimeout(parent, previewCacheTimeout)
	defer cancel()
	_ = client.SetNX(ctx, key, body, previewCacheTTL).Err()
}
