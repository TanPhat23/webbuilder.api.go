package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ElementCommentService struct {
	elementCommentRepo repositories.ElementCommentRepositoryInterface
	elementRepo        repositories.ElementRepositoryInterface
	projectRepo        repositories.ProjectRepositoryInterface
}

func NewElementCommentService(
	elementCommentRepo repositories.ElementCommentRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *ElementCommentService {
	return &ElementCommentService{
		elementCommentRepo: elementCommentRepo,
		elementRepo:        elementRepo,
		projectRepo:        projectRepo,
	}
}

func (s *ElementCommentService) resolveElementProjectID(ctx context.Context, elementID string) (string, error) {
	element, err := s.elementRepo.GetByID(ctx, elementID)
	if err != nil {
		if errors.Is(err, repositories.ErrElementNotFound) {
			return "", fmt.Errorf("%w: element does not exist", authz.ErrNotFound)
		}
		return "", fmt.Errorf("element not found: %w", err)
	}
	if element == nil {
		return "", fmt.Errorf("%w: element does not exist", authz.ErrNotFound)
	}
	if element.Page == nil || element.Page.ProjectId == "" {
		return "", fmt.Errorf("%w: element project could not be resolved", authz.ErrNotFound)
	}
	return element.Page.ProjectId, nil
}

func (s *ElementCommentService) Create(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
	if comment == nil {
		return nil, errors.New("comment cannot be nil")
	}
	if comment.ElementId == "" {
		return nil, errors.New("elementId is required")
	}
	if comment.AuthorId == "" {
		return nil, errors.New("authorId is required")
	}
	if comment.Content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, comment.ElementId)
		if err != nil {
			return nil, err
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, comment.AuthorId, authz.PermCommentCreate); err != nil {
			return nil, err
		}
	} else {
		element, err := s.elementRepo.GetByID(ctx, comment.ElementId)
		if err != nil {
			return nil, fmt.Errorf("element not found: %w", err)
		}
		if element == nil {
			return nil, errors.New("element does not exist")
		}
	}

	return s.elementCommentRepo.Create(ctx, comment)
}

func (s *ElementCommentService) GetByID(ctx context.Context, id string) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}

	return s.elementCommentRepo.GetByID(ctx, id)
}

func (s *ElementCommentService) GetByIDForUser(ctx context.Context, id, userID string) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}

	comment, err := s.elementCommentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, fmt.Errorf("%w: comment not found", authz.ErrNotFound)
	}
	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, comment.ElementId)
		if err != nil {
			return nil, err
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCommentView); err != nil {
			return nil, err
		}
	}
	return comment, nil
}

func (s *ElementCommentService) List(ctx context.Context, elementID string, filter models.ElementCommentFilter) ([]models.ElementComment, error) {
	if elementID == "" {
		return nil, errors.New("elementId is required")
	}

	element, err := s.elementRepo.GetByID(ctx, elementID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify element: %w", err)
	}
	if element == nil {
		return nil, errors.New("element does not exist")
	}

	return s.elementCommentRepo.List(ctx, elementID, &filter)
}

func (s *ElementCommentService) ListForUser(ctx context.Context, elementID, userID string, filter models.ElementCommentFilter) ([]models.ElementComment, error) {
	if elementID == "" {
		return nil, errors.New("elementId is required")
	}
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, elementID)
		if err != nil {
			if errors.Is(err, authz.ErrNotFound) {
				// Elements created over realtime may not have flushed to Postgres yet.
				return []models.ElementComment{}, nil
			}
			return nil, err
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCommentView); err != nil {
			return nil, err
		}
	}
	return s.elementCommentRepo.List(ctx, elementID, &filter)
}

func (s *ElementCommentService) ListByAuthor(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error) {
	if authorID == "" {
		return nil, errors.New("authorId is required")
	}
	if limit < 0 {
		return nil, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}

	return s.elementCommentRepo.ListByAuthor(ctx, authorID, limit, offset)
}

func (s *ElementCommentService) ListByProject(ctx context.Context, projectID, userID string, limit int, offset int) ([]models.ElementComment, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if limit < 0 {
		return nil, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	return s.elementCommentRepo.ListByProject(ctx, projectID, limit, offset)
}

func (s *ElementCommentService) Update(ctx context.Context, id string, userID string, updates map[string]any) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	comment, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comment: %w", err)
	}
	if comment == nil {
		return nil, errors.New("comment does not exist")
	}
	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, comment.ElementId)
		if err != nil {
			return nil, err
		}
		perm := authz.PermCommentEditAll
		if comment.AuthorId == userID {
			perm = authz.PermCommentEditOwn
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, perm); err != nil {
			return nil, err
		}
		if hasResolvedUpdate(updates) {
			if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCommentResolve); err != nil {
				return nil, err
			}
		}
	} else if comment.AuthorId != userID {
		return nil, fmt.Errorf("%w: user is not the comment author", authz.ErrForbidden)
	}

	return s.elementCommentRepo.Update(ctx, id, updates)
}

func hasResolvedUpdate(updates map[string]any) bool {
	if _, ok := updates["resolved"]; ok {
		return true
	}
	_, ok := updates["Resolved"]
	return ok
}

func (s *ElementCommentService) Delete(ctx context.Context, id string, userID string) error {
	if id == "" {
		return errors.New("comment id is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	comment, err := s.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve comment: %w", err)
	}
	if comment == nil {
		return errors.New("comment does not exist")
	}
	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, comment.ElementId)
		if err != nil {
			return err
		}
		perm := authz.PermCommentDeleteAll
		if comment.AuthorId == userID {
			perm = authz.PermCommentDeleteOwn
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, perm); err != nil {
			return err
		}
	} else if comment.AuthorId != userID {
		return fmt.Errorf("%w: user is not the comment author", authz.ErrForbidden)
	}

	return s.elementCommentRepo.Delete(ctx, id)
}

func (s *ElementCommentService) ToggleResolved(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("comment id is required")
	}

	_, err := s.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve comment: %w", err)
	}

	_, err = s.elementCommentRepo.ToggleResolved(ctx, id)
	return err
}

func (s *ElementCommentService) ToggleResolvedForUser(ctx context.Context, id, userID string) error {
	if id == "" {
		return errors.New("comment id is required")
	}
	if userID == "" {
		return authz.ErrUnauthorized
	}

	comment, err := s.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve comment: %w", err)
	}
	if comment == nil {
		return fmt.Errorf("%w: comment does not exist", authz.ErrNotFound)
	}
	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, comment.ElementId)
		if err != nil {
			return err
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCommentResolve); err != nil {
			return err
		}
	}

	_, err = s.elementCommentRepo.ToggleResolved(ctx, id)
	return err
}

func (s *ElementCommentService) Count(ctx context.Context, elementID string) (int64, error) {
	if elementID == "" {
		return 0, errors.New("elementId is required")
	}

	element, err := s.elementRepo.GetByID(ctx, elementID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify element: %w", err)
	}
	if element == nil {
		return 0, errors.New("element does not exist")
	}

	return s.elementCommentRepo.Count(ctx, elementID)
}

func (s *ElementCommentService) CountByProject(ctx context.Context, projectID, userID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return 0, fmt.Errorf("failed to verify project: %w", err)
	}

	return s.elementCommentRepo.CountByProject(ctx, projectID)
}

func (s *ElementCommentService) DeleteByElement(ctx context.Context, elementID string) error {
	if elementID == "" {
		return errors.New("elementId is required")
	}

	element, err := s.elementRepo.GetByID(ctx, elementID)
	if err != nil {
		return fmt.Errorf("failed to verify element: %w", err)
	}
	if element == nil {
		return errors.New("element does not exist")
	}

	return s.elementCommentRepo.DeleteByElement(ctx, elementID)
}
