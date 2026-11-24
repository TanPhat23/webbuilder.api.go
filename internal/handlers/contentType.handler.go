package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

var contentTypeAllowedCols = map[string]string{
	"name":        "Name",
	"description": "Description",
}

type ContentTypeHandler struct {
	contentTypeService services.ContentTypeServiceInterface
}

func NewContentTypeHandler(contentTypeService services.ContentTypeServiceInterface) *ContentTypeHandler {
	return &ContentTypeHandler{
		contentTypeService: contentTypeService,
	}
}

func (h *ContentTypeHandler) GetContentTypes(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	projectID := c.Params("projectid")
	if projectID == "" {
		projectID = c.Query("projectId")
	}

	contentTypes, err := h.contentTypeService.GetContentTypesForUser(c.RequestCtx(), userID, projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to retrieve content types")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentTypes)
}

func (h *ContentTypeHandler) GetContentTypeByID(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	contentType, err := h.contentTypeService.GetContentTypeByIDForUser(c.RequestCtx(), id, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to retrieve content type")
	}

	return utils.SendJSON(c, fiber.StatusOK, contentType)
}

func (h *ContentTypeHandler) CreateContentType(c fiber.Ctx) error {
	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}
	req := new(dto.CreateContentTypeRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	projectID := req.ProjectID
	if paramProjectID := c.Params("projectid"); paramProjectID != "" {
		projectID = &paramProjectID
	} else if queryProjectID := c.Query("projectId"); queryProjectID != "" && (projectID == nil || *projectID == "") {
		projectID = &queryProjectID
	}

	publicRead := true
	if req.PublicRead != nil {
		publicRead = *req.PublicRead
	}
	publicCreate := false
	if req.PublicCreate != nil {
		publicCreate = *req.PublicCreate
	}
	defaultPublished := false
	if req.DefaultPublished != nil {
		defaultPublished = *req.DefaultPublished
	}
	rateLimit := 30
	if req.RateLimitPerMin != nil && *req.RateLimitPerMin > 0 {
		rateLimit = *req.RateLimitPerMin
	}
	slug := ""
	if req.Slug != nil {
		slug = *req.Slug
	}

	contentType := &models.ContentType{
		ProjectId:        projectID,
		Name:             req.Name,
		Slug:             slug,
		Description:      req.Description,
		PublicRead:       publicRead,
		PublicCreate:     publicCreate,
		DefaultPublished: defaultPublished,
		RateLimitPerMin:  rateLimit,
	}

	created, err := h.contentTypeService.CreateContentTypeForUser(c.RequestCtx(), contentType, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to create content type")
	}

	return utils.SendJSON(c, fiber.StatusCreated, created)
}

func (h *ContentTypeHandler) UpdateContentType(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	req := new(dto.UpdateContentTypeRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if req.Name == nil && req.Slug == nil && req.Description == nil && req.PublicRead == nil && req.PublicCreate == nil && req.DefaultPublished == nil && req.RateLimitPerMin == nil {
		return utils.SendJSON(c, fiber.StatusBadRequest, fiber.Map{"error": "At least one field must be updated"})
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Slug != nil {
		updates["slug"] = models.SlugifyCollectionName(*req.Slug, id)
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.PublicRead != nil {
		updates["publicRead"] = *req.PublicRead
	}
	if req.PublicCreate != nil {
		updates["publicCreate"] = *req.PublicCreate
	}
	if req.DefaultPublished != nil {
		updates["defaultPublished"] = *req.DefaultPublished
	}
	if req.RateLimitPerMin != nil {
		updates["rateLimitPerMin"] = *req.RateLimitPerMin
	}

	updated, err := h.contentTypeService.UpdateContentTypeFieldsForUser(c.RequestCtx(), id, userID, updates)
	if err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to update content type")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

func (h *ContentTypeHandler) DeleteContentType(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "id")
	if err != nil {
		return err
	}
	id := ids[0]

	if err := h.contentTypeService.DeleteContentTypeForUser(c.RequestCtx(), id, userID); err != nil {
		return utils.HandleRepoError(c, err, "Content type not found", "Failed to delete content type")
	}

	return utils.SendNoContent(c)
}
