package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrCustomElementNotFound      = errors.New("custom element not found")
	ErrCustomElementUnauthorized  = errors.New("unauthorized to access custom element")
	ErrCustomElementAlreadyExists = errors.New("custom element with this name already exists")
)

type CustomElementRepository struct {
	db *gorm.DB
}

func isCustomElementNameConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" &&
		pgErr.ConstraintName == "CustomElement_UserId_Name_key"
}

func NewCustomElementRepository(db *gorm.DB) CustomElementRepositoryInterface {
	return &CustomElementRepository{
		db: db,
	}
}

func (r *CustomElementRepository) GetCustomElements(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	query := gorm.G[models.CustomElement](r.db).Where(`"UserId" = ?`, userID)
	if isPublic != nil {
		query = query.Where(`"IsPublic" = ?`, *isPublic)
	}

	customElements, err := query.Order("\"CreatedAt\" DESC").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom elements: %w", err)
	}

	return customElements, nil
}

func (r *CustomElementRepository) GetCustomElementByID(ctx context.Context, id string, userID string) (*models.CustomElement, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	customElement, err := gorm.G[models.CustomElement](r.db).
		Where("\"Id\" = ? AND (\"UserId\" = ? OR \"IsPublic\" = ?)", id, userID, true).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomElementNotFound
		}
		return nil, fmt.Errorf("failed to get custom element: %w", err)
	}

	return &customElement, nil
}

func (r *CustomElementRepository) CreateCustomElement(ctx context.Context, customElement *models.CustomElement) (*models.CustomElement, error) {
	if customElement == nil {
		return nil, errors.New("customElement is required")
	}
	if customElement.UserId == "" {
		return nil, errors.New("userId is required")
	}
	if customElement.Name == "" {
		return nil, errors.New("name is required")
	}

	if err := gorm.G[models.CustomElement](r.db).Create(ctx, customElement); err != nil {
		if isCustomElementNameConflict(err) {
			return nil, ErrCustomElementAlreadyExists
		}
		return nil, fmt.Errorf("failed to create custom element: %w", err)
	}

	return customElement, nil
}

func (r *CustomElementRepository) UpdateCustomElement(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	customElement, err := gorm.G[models.CustomElement](r.db).
		Where("\"Id\" = ? AND \"UserId\" = ?", id, userID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomElementUnauthorized
		}
		return nil, fmt.Errorf("failed to find custom element: %w", err)
	}

	result := r.db.WithContext(ctx).
		Model(&customElement).
		Where(`"UserId" = ?`, userID).
		Updates(updates)

	if result.Error != nil {
		if isCustomElementNameConflict(result.Error) {
			return nil, ErrCustomElementAlreadyExists
		}
		return nil, fmt.Errorf("failed to update custom element: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, ErrCustomElementUnauthorized
	}

	err = r.db.WithContext(ctx).
		Where(`"Id" = ? AND "UserId" = ?`, id, userID).
		First(&customElement).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom element: %w", err)
	}

	return &customElement, nil
}

func (r *CustomElementRepository) DeleteCustomElement(ctx context.Context, id string, userID string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	rowsAffected, err := gorm.G[models.CustomElement](r.db).
		Where("\"Id\" = ? AND \"UserId\" = ?", id, userID).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete custom element: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCustomElementUnauthorized
	}

	return nil
}

func (r *CustomElementRepository) GetPublicCustomElements(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error) {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	query := gorm.G[models.CustomElement](r.db).
		Where("\"IsPublic\" = ?", true)

	if category != nil && *category != "" {
		query = query.Where("\"Category\" = ?", *category)
	}

	customElements, err := query.
		Order("\"CreatedAt\" DESC").
		Limit(limit).
		Offset(offset).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get public custom elements: %w", err)
	}

	return customElements, nil
}

func (r *CustomElementRepository) DuplicateCustomElement(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	if newName == "" {
		return nil, errors.New("newName is required")
	}

	original, err := gorm.G[models.CustomElement](r.db).
		Where("\"Id\" = ? AND (\"UserId\" = ? OR \"IsPublic\" = ?)", id, userID, true).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomElementNotFound
		}
		return nil, fmt.Errorf("failed to find custom element: %w", err)
	}

	duplicate := models.CustomElement{
		Name:         newName,
		Description:  original.Description,
		Category:     original.Category,
		Icon:         original.Icon,
		Thumbnail:    original.Thumbnail,
		Structure:    original.Structure,
		DefaultProps: original.DefaultProps,
		Tags:         original.Tags,
		UserId:       userID,
		IsPublic:     false,
		Version:      original.Version,
	}

	err = gorm.G[models.CustomElement](r.db).Create(ctx, &duplicate)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate custom element: %w", err)
	}

	return &duplicate, nil
}
