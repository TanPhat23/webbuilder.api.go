package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type stubPubCTRepo struct {
	repositories.ContentTypeRepositoryInterface
	ct *models.ContentType
}

func (s *stubPubCTRepo) GetContentTypeByProjectAndSlugOrID(_ context.Context, projectID, slugOrID string) (*models.ContentType, error) {
	if s.ct == nil || s.ct.ProjectId == nil || *s.ct.ProjectId != projectID || (s.ct.Id != slugOrID && s.ct.Slug != slugOrID) {
		return nil, repositories.ErrContentTypeNotFound
	}
	return s.ct, nil
}

type stubPubItemRepo struct {
	repositories.ContentItemRepositoryInterface
	created []*models.ContentItem
	items   []models.ContentItem
}

func (s *stubPubItemRepo) GetPublicContentItems(_ context.Context, _ string, _ int, _, _ string) ([]models.ContentItem, error) {
	return s.items, nil
}

func (s *stubPubItemRepo) CreateContentItem(_ context.Context, ci *models.ContentItem, fvs []models.ContentFieldValue) (*models.ContentItem, error) {
	ci.FieldValues = fvs
	s.created = append(s.created, ci)
	return ci, nil
}

func TestPublicCollection_SubmitValidationHoneypotAndRateLimit(t *testing.T) {
	projID := "proj_1"
	ct := &models.ContentType{
		Id:               "ct_1",
		ProjectId:        &projID,
		Name:             "Catering Inquiries",
		Slug:             "catering-inquiries",
		PublicRead:       true,
		PublicCreate:     true,
		DefaultPublished: false,
		RateLimitPerMin:  2,
		Fields: []models.ContentField{
			{Id: "f1", Name: "name", Type: "text", Required: true},
			{Id: "f2", Name: "email", Type: "email", Required: true},
			{Id: "f3", Name: "guests", Type: "number", Required: false},
		},
	}

	itemRepo := &stubPubItemRepo{}
	svc := services.NewPublicCollectionService(&stubPubCTRepo{ct: ct}, itemRepo)
	h := NewPublicCollectionHandler(svc)

	app := fiber.New()
	app.Get("/api/v1/public/projects/:projectId/collections/:collection/items", h.ListPublicItems)
	app.Post("/api/v1/public/projects/:projectId/collections/:collection/items", h.SubmitPublicItem)

	post := func(body string) (*httptest.ResponseRecorder, int, map[string]any) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/public/projects/proj_1/collections/catering-inquiries/items", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		var parsed map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&parsed)
		return nil, resp.StatusCode, parsed
	}

	// 1. Malformed JSON -> 400
	_, code, _ := post(`["not-an-object"]`)
	if code != fiber.StatusBadRequest {
		t.Fatalf("expected 400 for array body, got %d", code)
	}

	// 2. Honeypot -> 201 without creating item
	_, code, _ = post(`{"name":"Bot","email":"bot@example.com","_hp":"spam"}`)
	if code != fiber.StatusCreated {
		t.Fatalf("expected 201 for honeypot, got %d", code)
	}
	if len(itemRepo.created) != 0 {
		t.Fatalf("expected 0 items created for honeypot, got %d", len(itemRepo.created))
	}

	// 3. Invalid email -> 400 with fields.email
	_, code, body := post(`{"name":"Maya","email":"not-an-email"}`)
	if code != fiber.StatusBadRequest {
		t.Fatalf("expected 400 for invalid email, got %d", code)
	}
	fields, ok := body["fields"].(map[string]any)
	if !ok || fields["email"] == nil {
		t.Fatalf("expected fields.email error, got %v", body)
	}

	// 4. Third request in 1 minute exceeds RateLimitPerMin=2 -> 429
	_, code, _ = post(`{"name":"Maya","email":"maya@example.com","guests":40,"ignored":"x"}`)
	if code != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429 on rate limit, got %d", code)
	}

	// 5. Fresh service with higher limit -> valid submission persists item with Published=false
	ct.RateLimitPerMin = 10
	svc2 := services.NewPublicCollectionService(&stubPubCTRepo{ct: ct}, itemRepo)
	h2 := NewPublicCollectionHandler(svc2)
	app2 := fiber.New()
	app2.Post("/api/v1/public/projects/:projectId/collections/:collection/items", h2.SubmitPublicItem)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/public/projects/proj_1/collections/catering-inquiries/items", bytes.NewBufferString(`{"name":"Maya Lin","email":"maya@example.com","guests":45,"unknown":"drop"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app2.Test(req)
	if err != nil {
		t.Fatalf("app2.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 for valid submission, got %d", resp.StatusCode)
	}
	if len(itemRepo.created) != 1 || itemRepo.created[0].Title != "Maya Lin" || itemRepo.created[0].Published {
		t.Fatalf("unexpected created item: %+v", itemRepo.created)
	}
}
