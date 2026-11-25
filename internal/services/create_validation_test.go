package services

import (
	"context"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type createProjectRepo struct {
	repositories.ProjectRepositoryInterface
	calls int
}

func (r *createProjectRepo) Create(_ context.Context, _ *models.Project) error {
	r.calls++
	return nil
}

type createFieldRepo struct {
	repositories.ContentFieldRepositoryInterface
	calls int
}

func (r *createFieldRepo) Create(_ context.Context, f *models.ContentField) (*models.ContentField, error) {
	r.calls++
	return f, nil
}

type createCommentRepo struct {
	repositories.ElementCommentRepositoryInterface
	calls int
}

func (r *createCommentRepo) Create(_ context.Context, c *models.ElementComment) (*models.ElementComment, error) {
	r.calls++
	return c, nil
}

type existingElementRepo struct {
	repositories.ElementRepositoryInterface
}

func (existingElementRepo) GetByID(_ context.Context, id string) (*models.Element, error) {
	return &models.Element{Id: id}, nil
}

func TestCreateProjectValidationBeforeRepository(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *models.Project
		valid bool
	}{
		{"nil", nil, false}, {"missing name", &models.Project{OwnerId: "owner"}, false}, {"missing owner", &models.Project{Name: "name"}, false}, {"valid", &models.Project{Name: "name", OwnerId: "owner"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &createProjectRepo{}
			result, err := NewProjectService(repo, nil, nil).Create(context.Background(), tc.value)
			if (err == nil) != tc.valid || (result != nil) != tc.valid || (repo.calls == 1) != tc.valid {
				t.Fatalf("result=%v error=%v calls=%d", result, err, repo.calls)
			}
		})
	}
}

func TestCreateContentFieldValidationBeforeRepository(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *models.ContentField
		valid bool
	}{
		{"nil", nil, false}, {"missing type ID", &models.ContentField{Name: "name", Type: "text"}, false}, {"blank type ID", &models.ContentField{ContentTypeId: " ", Name: "name", Type: "text"}, false}, {"blank name", &models.ContentField{ContentTypeId: "type", Name: " ", Type: "text"}, false}, {"blank type", &models.ContentField{ContentTypeId: "type", Name: "name", Type: " "}, false}, {"valid", &models.ContentField{ContentTypeId: "type", Name: "name", Type: "text"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &createFieldRepo{}
			result, err := NewContentFieldService(repo).Create(context.Background(), tc.value)
			if (err == nil) != tc.valid || (result != nil) != tc.valid || (repo.calls == 1) != tc.valid {
				t.Fatalf("result=%v error=%v calls=%d", result, err, repo.calls)
			}
		})
	}
}

func TestCreateElementCommentValidationBeforeRepository(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *models.ElementComment
		valid bool
	}{
		{"nil", nil, false}, {"missing element", &models.ElementComment{Content: "hello", AuthorId: "author"}, false}, {"missing author", &models.ElementComment{Content: "hello", ElementId: "element"}, false}, {"missing content", &models.ElementComment{ElementId: "element", AuthorId: "author"}, false}, {"valid", &models.ElementComment{Content: "hello", ElementId: "element", AuthorId: "author"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &createCommentRepo{}
			result, err := NewElementCommentService(repo, existingElementRepo{}, nil).Create(context.Background(), tc.value)
			if (err == nil) != tc.valid || (result != nil) != tc.valid || (repo.calls == 1) != tc.valid {
				t.Fatalf("result=%v error=%v calls=%d", result, err, repo.calls)
			}
		})
	}
}
