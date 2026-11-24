package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type contentItemSvcStub struct {
	services.ContentItemServiceInterface
	err error

	items  []models.ContentItem
	bySlug *models.ContentItem

	created      *models.ContentItem
	createdVals  []models.ContentFieldValue
	updatedID    string
	updates      map[string]any
	updatedVals  []models.ContentFieldValue
	deleted      string
	publicArgs   []any
	slugArgs     []string
}

func (s *contentItemSvcStub) GetContentItemsByContentType(_ context.Context, typeID string) ([]models.ContentItem, error) {
	return []models.ContentItem{{Id: "i1", ContentTypeId: typeID}}, s.err
}
func (s *contentItemSvcStub) GetContentItemByID(_ context.Context, id string) (*models.ContentItem, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.ContentItem{Id: id}, nil
}
func (s *contentItemSvcStub) GetContentItemBySlug(_ context.Context, typeID, slug string) (*models.ContentItem, error) {
	s.slugArgs = []string{typeID, slug}
	if s.err != nil {
		return nil, s.err
	}
	return s.bySlug, nil
}
func (s *contentItemSvcStub) GetPublicContentItems(_ context.Context, typeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error) {
	s.publicArgs = []any{typeID, limit, sortBy, sortOrder}
	return s.items, s.err
}
func (s *contentItemSvcStub) CreateContentItem(_ context.Context, item *models.ContentItem, vals []models.ContentFieldValue) (*models.ContentItem, error) {
	s.created, s.createdVals = item, vals
	if s.err != nil {
		return nil, s.err
	}
	item.Id = "i-new"
	return item, nil
}
func (s *contentItemSvcStub) UpdateContentItem(_ context.Context, id string, updates map[string]any, vals []models.ContentFieldValue) (*models.ContentItem, error) {
	s.updatedID, s.updates, s.updatedVals = id, updates, vals
	if s.err != nil {
		return nil, s.err
	}
	return &models.ContentItem{Id: id}, nil
}
func (s *contentItemSvcStub) DeleteContentItem(_ context.Context, id string) error {
	s.deleted = id
	return s.err
}

func (s *contentItemSvcStub) GetContentItemsByContentTypeForUser(_ context.Context, typeID, _ string) ([]models.ContentItem, error) {
	return s.GetContentItemsByContentType(context.Background(), typeID)
}
func (s *contentItemSvcStub) GetContentItemByIDForUser(_ context.Context, id, _ string) (*models.ContentItem, error) {
	return s.GetContentItemByID(context.Background(), id)
}
func (s *contentItemSvcStub) CreateContentItemForUser(_ context.Context, item *models.ContentItem, vals []models.ContentFieldValue, _ string) (*models.ContentItem, error) {
	return s.CreateContentItem(context.Background(), item, vals)
}
func (s *contentItemSvcStub) UpdateContentItemForUser(_ context.Context, id, _ string, updates map[string]any, vals []models.ContentFieldValue) (*models.ContentItem, error) {
	return s.UpdateContentItem(context.Background(), id, updates, vals)
}
func (s *contentItemSvcStub) DeleteContentItemForUser(_ context.Context, id, _ string) error {
	return s.DeleteContentItem(context.Background(), id)
}

func contentItemKit(svc *contentItemSvcStub) *fiber.App {
	h := NewContentItemHandler(svc)
	return newKitApp("u1", func(app *fiber.App) {
		app.Get("/public/items", h.GetPublicContentItems)
		app.Get("/public/:contentTypeId/:slug", h.GetPublicContentItemBySlug)
		app.Get("/types/:contentTypeId/items", h.GetContentItemsByContentType)
		app.Post("/types/:contentTypeId/items", h.CreateContentItem)
		app.Get("/items/:itemId", h.GetContentItemByID)
		app.Patch("/items/:itemId", h.UpdateContentItem)
		app.Delete("/items/:itemId", h.DeleteContentItem)
	})
}

func TestContentItemHandlerStatuses(t *testing.T) {
	cases := []struct {
		name         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"list ok", "GET", "/types/t1/items", "", nil, 200, "t1"},
		{"list internal", "GET", "/types/t1/items", "", errInternal, 500, "Failed to retrieve content items"},
		{"get ok", "GET", "/items/i1", "", nil, 200, "i1"},
		{"get missing", "GET", "/items/i1", "", gorm.ErrRecordNotFound, 404, "Content item not found"},
		{"get internal", "GET", "/items/i1", "", errInternal, 500, "Failed to retrieve content item"},
		{"create ok", "POST", "/types/t1/items", `{"slug":"a","title":"A"}`, nil, 201, "i-new"},
		{"create bad json", "POST", "/types/t1/items", `{`, nil, 0, ""},
		{"create internal", "POST", "/types/t1/items", `{"slug":"a","title":"A"}`, errInternal, 500, "Failed to create content item"},
		{"update ok", "PATCH", "/items/i1", `{"title":"B"}`, nil, 200, "i1"},
		{"update field values only", "PATCH", "/items/i1", `{"fieldValues":[{"fieldId":"f1","value":"x"}]}`, nil, 200, "i1"},
		{"update nothing", "PATCH", "/items/i1", `{}`, nil, 400, "No valid fields to update"},
		{"update empty field values", "PATCH", "/items/i1", `{"fieldValues":[]}`, nil, 400, "No valid fields to update"},
		{"update bad json", "PATCH", "/items/i1", `[`, nil, 0, ""},
		{"update missing", "PATCH", "/items/i1", `{"title":"B"}`, gorm.ErrRecordNotFound, 404, "Content item not found"},
		{"update internal", "PATCH", "/items/i1", `{"title":"B"}`, errInternal, 500, "Failed to update content item"},
		{"delete ok", "DELETE", "/items/i1", "", nil, 204, ""},
		{"delete missing", "DELETE", "/items/i1", "", gorm.ErrRecordNotFound, 404, "Content item not found"},
		{"delete internal", "DELETE", "/items/i1", "", errInternal, 500, "Failed to delete content item"},
		{"public list ok", "GET", "/public/items", "", nil, 200, "[]"},
		{"public list internal", "GET", "/public/items", "", errInternal, 500, "Failed to retrieve content items"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := kitSend(t, contentItemKit(&contentItemSvcStub{err: tc.err}), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestContentItemCreateSplitsFieldValuesAndForcesParent(t *testing.T) {
	svc := &contentItemSvcStub{}
	body := `{"slug":"hello","title":"Hello","contentTypeId":"spoofed","fieldValues":[{"fieldId":"f1","value":"v"}]}`
	kitExpect(t, kitSend(t, contentItemKit(svc), "POST", "/types/real/items", body), 201)
	if svc.created.ContentTypeId != "real" || svc.created.Slug != "hello" {
		t.Fatalf("created %+v", svc.created)
	}
	if svc.created.FieldValues != nil {
		t.Fatal("field values must be passed separately, not embedded in the item")
	}
	if len(svc.createdVals) != 1 || svc.createdVals[0].FieldId != "f1" || *svc.createdVals[0].Value != "v" {
		t.Fatalf("field values %+v", svc.createdVals)
	}
}

func TestContentItemUpdateForwardsOnlyAllowedColumns(t *testing.T) {
	svc := &contentItemSvcStub{}
	kitExpect(t, kitSend(t, contentItemKit(svc), "PATCH", "/items/i9", `{"published":true,"slug":"s","id":"hijack","contentTypeId":"x"}`), 200)
	if svc.updatedID != "i9" || len(svc.updates) != 2 || svc.updates["Published"] != true || svc.updates["Slug"] != "s" {
		t.Fatalf("id=%q updates=%v", svc.updatedID, svc.updates)
	}
}

func TestPublicContentItemsQueryNormalisation(t *testing.T) {
	cases := []struct {
		query     string
		wantLimit int
		wantSort  string
		wantOrder string
	}{
		{"", 10, "CreatedAt", "desc"},
		{"?limit=25&sortBy=title&sortOrder=asc", 25, "Title", "asc"},
		{"?sortBy=updatedAt", 10, "UpdatedAt", "desc"},
		{"?sortBy=password;drop", 10, "CreatedAt", "desc"},
		{"?limit=0", 10, "CreatedAt", "desc"},
		{"?limit=-4", 10, "CreatedAt", "desc"},
		{"?limit=abc", 10, "CreatedAt", "desc"},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			svc := &contentItemSvcStub{}
			kitExpect(t, kitSend(t, contentItemKit(svc), "GET", "/public/items"+tc.query+"", ""), 200)
			got := svc.publicArgs
			if got[1] != tc.wantLimit || got[2] != tc.wantSort || got[3] != tc.wantOrder {
				t.Fatalf("service args = %v", got)
			}
		})
	}
	svc := &contentItemSvcStub{}
	kitSend(t, contentItemKit(svc), "GET", "/public/items?contentTypeId=ct7", "")
	if svc.publicArgs[0] != "ct7" {
		t.Fatalf("content type filter = %v", svc.publicArgs[0])
	}
}

func TestPublicContentItemsAreFlattened(t *testing.T) {
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	val := "Hello body"
	svc := &contentItemSvcStub{items: []models.ContentItem{{
		Id: "i1", ContentTypeId: "t1", Slug: "hello", Title: "Hello", Published: true,
		AuditFields: models.AuditFields{CreatedAt: created},
		FieldValues: []models.ContentFieldValue{
			{FieldId: "f1", Field: &models.ContentField{Name: "body"}, Value: &val},
			{FieldId: "f2", Field: &models.ContentField{Name: ""}, Value: &val}, // unnamed fields are dropped
		},
	}}}
	r := kitSend(t, contentItemKit(svc), "GET", "/public/items", "")
	kitExpect(t, r, 200)
	var out []map[string]any
	r.decode(t, &out)
	if len(out) != 1 {
		t.Fatalf("got %d items", len(out))
	}
	item := out[0]
	if item["body"] != "Hello body" || item["slug"] != "hello" || item["title"] != "Hello" || item["published"] != true || item["contentTypeId"] != "t1" {
		t.Fatalf("flattened item = %v", item)
	}
	if _, ok := item["fieldValues"]; ok {
		t.Fatal("raw fieldValues should be replaced by flattened fields")
	}
	if len(item) != 8+1 { // 8 base keys + "body"
		t.Fatalf("unexpected keys in %v", item)
	}
}

func TestPublicContentItemBySlug(t *testing.T) {
	val := "v"
	published := &models.ContentItem{Id: "i1", Slug: "s", Published: true, FieldValues: []models.ContentFieldValue{{Field: &models.ContentField{Name: "extra"}, Value: &val}}}
	draft := &models.ContentItem{Id: "i2", Slug: "s", Published: false}

	t.Run("published item", func(t *testing.T) {
		svc := &contentItemSvcStub{bySlug: published}
		r := kitSend(t, contentItemKit(svc), "GET", "/public/blog/s", "")
		kitExpect(t, r, 200, `"extra":"v"`)
		if len(svc.slugArgs) != 2 || svc.slugArgs[0] != "blog" || svc.slugArgs[1] != "s" {
			t.Fatalf("service args %v", svc.slugArgs)
		}
	})
	t.Run("draft is hidden", func(t *testing.T) {
		kitExpect(t, kitSend(t, contentItemKit(&contentItemSvcStub{bySlug: draft}), "GET", "/public/blog/s", ""), 404, "Content item not found")
	})
	t.Run("missing", func(t *testing.T) {
		kitExpect(t, kitSend(t, contentItemKit(&contentItemSvcStub{err: gorm.ErrRecordNotFound}), "GET", "/public/blog/s", ""), 404, "Content item not found")
	})
	t.Run("internal", func(t *testing.T) {
		kitExpect(t, kitSend(t, contentItemKit(&contentItemSvcStub{err: errInternal}), "GET", "/public/blog/s", ""), 500, "Failed to retrieve content item")
	})
}
