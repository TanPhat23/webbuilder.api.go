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

		projectID, err := utils.ValidateRequiredParam(c, "projectid")
		if err != nil {
			return err
		}
		project, err := projects.GetPublicProjectByID(c.RequestCtx(), projectID)
		if err != nil {
			return utils.HandleRepoError(c, err, "Project not found", "Failed to verify publication")
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
		cacheCtx, cancel := context.WithTimeout(c.Context(), previewCacheTimeout)
		cached, err := client.Get(cacheCtx, key).Bytes()
		cancel()
		if err == nil {
			c.Type("json")
			return c.Status(fiber.StatusOK).Send(cached)
		}

		loadErr := c.Next()
		if loadErr == nil && c.Response().StatusCode() == fiber.StatusOK && errors.Is(err, redis.Nil) {
			cacheCtx, cancel := context.WithTimeout(c.Context(), previewCacheTimeout)
			defer cancel()
			_ = client.SetNX(cacheCtx, key, c.Response().Body(), previewCacheTTL).Err()
		}
		return loadErr
	}
}
