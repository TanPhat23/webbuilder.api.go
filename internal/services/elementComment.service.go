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
	element, err := s.elementRepo.GetElementByID(ctx, elementID)
	if err != nil {
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

func (s *ElementCommentService) CreateElementComment(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
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
		element, err := s.elementRepo.GetElementByID(ctx, comment.ElementId)
		if err != nil {
			return nil, fmt.Errorf("element not found: %w", err)
		}
		if element == nil {
			return nil, errors.New("element does not exist")
		}
	}

	return s.elementCommentRepo.CreateElementComment(ctx, comment)
}

func (s *ElementCommentService) GetElementCommentByID(ctx context.Context, id string) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}

	return s.elementCommentRepo.GetElementCommentByID(ctx, id)
}

func (s *ElementCommentService) GetElementCommentByIDForUser(ctx context.Context, id, userID string) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}

	comment, err := s.elementCommentRepo.GetElementCommentByID(ctx, id)
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

func (s *ElementCommentService) GetElementComments(ctx context.Context, elementID string, filter models.ElementCommentFilter) ([]models.ElementComment, error) {
	if elementID == "" {
		return nil, errors.New("elementId is required")
	}

	element, err := s.elementRepo.GetElementByID(ctx, elementID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify element: %w", err)
	}
	if element == nil {
		return nil, errors.New("element does not exist")
	}

	return s.elementCommentRepo.GetElementComments(ctx, elementID, &filter)
}

func (s *ElementCommentService) GetElementCommentsForUser(ctx context.Context, elementID, userID string, filter models.ElementCommentFilter) ([]models.ElementComment, error) {
	if elementID == "" {
		return nil, errors.New("elementId is required")
	}
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	if s.projectRepo != nil {
		projectID, err := s.resolveElementProjectID(ctx, elementID)
		if err != nil {
			return nil, err
		}
		if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermCommentView); err != nil {
			return nil, err
		}
	}
	return s.elementCommentRepo.GetElementComments(ctx, elementID, &filter)
}

func (s *ElementCommentService) GetElementCommentsByAuthorID(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error) {
	if authorID == "" {
		return nil, errors.New("authorId is required")
	}
	if limit < 0 {
		return nil, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}

	return s.elementCommentRepo.GetElementCommentsByAuthorID(ctx, authorID, limit, offset)
}

func (s *ElementCommentService) GetElementCommentsByProjectID(ctx context.Context, projectID, userID string, limit int, offset int) ([]models.ElementComment, error) {
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

	return s.elementCommentRepo.GetElementCommentsByProjectID(ctx, projectID, limit, offset)
}

func (s *ElementCommentService) UpdateElementComment(ctx context.Context, id string, userID string, updates map[string]any) (*models.ElementComment, error) {
	if id == "" {
		return nil, errors.New("comment id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	comment, err := s.GetElementCommentByID(ctx, id)
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
	} else if comment.AuthorId != userID {
		return nil, fmt.Errorf("%w: user is not the comment author", authz.ErrForbidden)
	}

	return s.elementCommentRepo.UpdateElementComment(ctx, id, updates)
}

func (s *ElementCommentService) DeleteElementComment(ctx context.Context, id string, userID string) error {
	if id == "" {
		return errors.New("comment id is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	comment, err := s.GetElementCommentByID(ctx, id)
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

	return s.elementCommentRepo.DeleteElementComment(ctx, id)
}

func (s *ElementCommentService) ToggleResolvedStatus(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("comment id is required")
	}

	_, err := s.GetElementCommentByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve comment: %w", err)
	}

	_, err = s.elementCommentRepo.ToggleResolvedStatus(ctx, id)
	return err
}

func (s *ElementCommentService) ToggleResolvedStatusForUser(ctx context.Context, id, userID string) error {
	if id == "" {
		return errors.New("comment id is required")
	}
	if userID == "" {
		return authz.ErrUnauthorized
	}

	comment, err := s.GetElementCommentByID(ctx, id)
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

	_, err = s.elementCommentRepo.ToggleResolvedStatus(ctx, id)
	return err
}

func (s *ElementCommentService) CountElementComments(ctx context.Context, elementID string) (int64, error) {
	if elementID == "" {
		return 0, errors.New("elementId is required")
	}

	element, err := s.elementRepo.GetElementByID(ctx, elementID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify element: %w", err)
	}
	if element == nil {
		return 0, errors.New("element does not exist")
	}

	return s.elementCommentRepo.CountElementComments(ctx, elementID)
}

func (s *ElementCommentService) CountElementCommentsByProjectID(ctx context.Context, projectID, userID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectId is required")
	}

	if err := ensureProjectVisible(ctx, s.projectRepo, projectID, userID); err != nil {
		return 0, fmt.Errorf("failed to verify project: %w", err)
	}

	return s.elementCommentRepo.CountElementCommentsByProjectID(ctx, projectID)
}

func (s *ElementCommentService) DeleteElementCommentsByElementID(ctx context.Context, elementID string) error {
	if elementID == "" {
		return errors.New("elementId is required")
	}

	element, err := s.elementRepo.GetElementByID(ctx, elementID)
	if err != nil {
		return fmt.Errorf("failed to verify element: %w", err)
	}
	if element == nil {
		return errors.New("element does not exist")
	}

	return s.elementCommentRepo.DeleteElementCommentsByElementID(ctx, elementID)
}