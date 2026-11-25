package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ContentItemService struct {
	contentItemRepo repositories.ContentItemRepositoryInterface
	contentTypeRepo repositories.ContentTypeRepositoryInterface
	projectRepo     repositories.ProjectRepositoryInterface
	emitter         ItemEventEmitter
}

func NewContentItemService(contentItemRepo repositories.ContentItemRepositoryInterface) *ContentItemService {
	return &ContentItemService{
		contentItemRepo: contentItemRepo,
	}
}

func (s *ContentItemService) WithEventEmitter(emitter ItemEventEmitter) *ContentItemService {
	s.emitter = emitter
	return s
}

func (s *ContentItemService) emitForItem(ctx context.Context, event string, item *models.ContentItem) {
	if s.emitter == nil || s.contentTypeRepo == nil || item == nil {
		return
	}
	ct, err := s.contentTypeRepo.GetByID(ctx, item.ContentTypeId)
	if err != nil || ct == nil || ct.ProjectId == nil || *ct.ProjectId == "" {
		return
	}
	s.emitter.EmitItemEvent(ctx, *ct.ProjectId, ct, event, item)
}

// WithAuth configures project-scoped RBAC checks via contentTypeRepo and projectRepo.
func (s *ContentItemService) WithAuth(
	contentTypeRepo repositories.ContentTypeRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *ContentItemService {
	s.contentTypeRepo = contentTypeRepo
	s.projectRepo = projectRepo
	return s
}

func (s *ContentItemService) authorizeContentType(ctx context.Context, contentTypeID, userID string, perm authz.Permission) error {
	if userID == "" {
		return authz.ErrUnauthorized
	}
	if s.contentTypeRepo == nil || s.projectRepo == nil {
		return nil
	}
	ct, err := s.contentTypeRepo.GetByID(ctx, contentTypeID)
	if err != nil {
		return err
	}
	if ct == nil {
		return repositories.ErrContentTypeNotFound
	}
	if ct.ProjectId == nil || *ct.ProjectId == "" {
		return fmt.Errorf("%w: content type is not scoped to a project", authz.ErrForbidden)
	}
	_, err = authorizeProjectPerm(ctx, s.projectRepo, *ct.ProjectId, userID, perm)
	return err
}

func (s *ContentItemService) ListByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}

	return s.contentItemRepo.ListByContentType(ctx, contentTypeID)
}

func (s *ContentItemService) ListByContentTypeForUser(ctx context.Context, contentTypeID, userID string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if err := s.authorizeContentType(ctx, contentTypeID, userID, authz.PermCMSView); err != nil {
		return nil, err
	}
	return s.contentItemRepo.ListByContentType(ctx, contentTypeID)
}

func (s *ContentItemService) GetByID(ctx context.Context, id string) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("content item id is required")
	}

	return s.contentItemRepo.GetByID(ctx, id)
}

func (s *ContentItemService) GetByIDForUser(ctx context.Context, id, userID string) (*models.ContentItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil || item == nil {
		return item, err
	}
	if err := s.authorizeContentType(ctx, item.ContentTypeId, userID, authz.PermCMSView); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ContentItemService) GetBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	if s.contentTypeRepo != nil && s.projectRepo != nil {
		ct, err := s.contentTypeRepo.GetByID(ctx, contentTypeID)
		if err != nil {
			if errors.Is(err, repositories.ErrContentTypeNotFound) {
				return nil, repositories.ErrContentItemNotFound
			}
			return nil, err
		}
		if ct == nil || !ct.PublicRead || ct.ProjectId == nil || *ct.ProjectId == "" {
			return nil, repositories.ErrContentItemNotFound
		}
		proj, err := s.projectRepo.GetPublicByID(ctx, *ct.ProjectId)
		if err != nil {
			if errors.Is(err, repositories.ErrProjectNotFound) {
				return nil, repositories.ErrContentItemNotFound
			}
			return nil, err
		}
		if proj == nil || !proj.Published || (proj.AuditFields != nil && proj.DeletedAt != nil) {
			return nil, repositories.ErrContentItemNotFound
		}
	}

	return s.contentItemRepo.GetBySlug(ctx, contentTypeID, slug)
}

func (s *ContentItemService) ListPublic(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error) {
	if contentTypeID == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	if sortBy == "" {
		return nil, errors.New("sortBy is required")
	}
	if sortOrder == "" {
		return nil, errors.New("sortOrder is required")
	}

	return s.contentItemRepo.ListPublic(ctx, contentTypeID, limit, sortBy, sortOrder)
}

func (s *ContentItemService) Create(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if item == nil {
		return nil, errors.New("content item cannot be nil")
	}
	if item.ContentTypeId == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if item.Slug == "" {
		return nil, errors.New("slug is required")
	}
	if item.Title == "" {
		return nil, errors.New("title is required")
	}

	if fieldValues == nil {
		fieldValues = []models.ContentFieldValue{}
	}
	item.FieldValues = nil

	created, err := s.contentItemRepo.Create(ctx, item, fieldValues)
	if err != nil {
		return nil, err
	}
	s.emitForItem(ctx, models.WebhookEventItemCreated, created)

	return created, nil
}

func (s *ContentItemService) CreateForUser(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue, userID string) (*models.ContentItem, error) {
	if item == nil {
		return nil, errors.New("content item cannot be nil")
	}
	if err := s.authorizeContentType(ctx, item.ContentTypeId, userID, authz.PermCMSEdit); err != nil {
		return nil, err
	}
	if item.Published {
		if err := s.authorizeContentType(ctx, item.ContentTypeId, userID, authz.PermCMSPublish); err != nil {
			return nil, err
		}
	}
	return s.Create(ctx, item, fieldValues)
}

func (s *ContentItemService) Update(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if id == "" {
		return nil, errors.New("content item id is required")
	}
	if len(updates) == 0 && len(fieldValues) == 0 {
		return nil, errors.New("no updates provided")
	}

	if updates == nil {
		updates = map[string]any{}
	}
	if fieldValues == nil {
		fieldValues = []models.ContentFieldValue{}
	}

	if slug, ok := updates["Slug"].(string); ok && slug == "" {
		return nil, errors.New("slug cannot be empty")
	}
	if title, ok := updates["Title"].(string); ok && title == "" {
		return nil, errors.New("title cannot be empty")
	}

	updated, err := s.contentItemRepo.Update(ctx, id, updates, fieldValues)
	if err != nil {
		return nil, err
	}
	s.emitForItem(ctx, models.WebhookEventItemUpdated, updated)

	return updated, nil
}

func (s *ContentItemService) UpdateForUser(ctx context.Context, id, userID string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, repositories.ErrContentItemNotFound
	}
	if err := s.authorizeContentType(ctx, current.ContentTypeId, userID, authz.PermCMSEdit); err != nil {
		return nil, err
	}
	if _, hasPublished := updates["Published"]; hasPublished {
		if err := s.authorizeContentType(ctx, current.ContentTypeId, userID, authz.PermCMSPublish); err != nil {
			return nil, err
		}
	}
	return s.Update(ctx, id, updates, fieldValues)
}

func (s *ContentItemService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("content item id is required")
	}

	return s.contentItemRepo.Delete(ctx, id)
}

func (s *ContentItemService) DeleteForUser(ctx context.Context, id, userID string) error {
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return repositories.ErrContentItemNotFound
	}
	if err := s.authorizeContentType(ctx, current.ContentTypeId, userID, authz.PermCMSEdit); err != nil {
		return err
	}
	if err := s.Delete(ctx, id); err != nil {
		return err
	}
	s.emitForItem(ctx, models.WebhookEventItemDeleted, current)
	return nil
}

func (s *ContentItemService) NormalizePublicSort(sortBy, sortOrder string) (string, string, error) {
	if sortBy == "" {
		return "", "", errors.New("sortBy is required")
	}
	if sortOrder == "" {
		return "", "", errors.New("sortOrder is required")
	}

	validSortOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if !validSortOrders[sortOrder] {
		return "", "", fmt.Errorf("invalid sortOrder: %s", sortOrder)
	}

	return sortBy, sortOrder, nil
}
