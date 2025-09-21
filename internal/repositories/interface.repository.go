package repositories

import "my-go-app/internal/models"

type ElementRepositoryInterface interface {
	GetElements(projectID string) ([]models.EditorElement, error)
	CreateElement(elements []models.EditorElement, projectID string) error
	InsertElementAfter(projectID string, previousElementID string, element models.EditorElement) error
}
type ProjectRepositoryInterface interface {
	GetProjects() ([]models.Project, error)
	GetProjectByID(projectID string, userId string) (*models.Project, error)
	GetProjectsByUserID(userID string) ([]models.Project, error)
	GetProjectPages(projectID string, userId string) ([]models.Page, error)
}

type RepositoriesInterface struct {
	ElementRepository ElementRepositoryInterface
	ProjectRepository ProjectRepositoryInterface
}
