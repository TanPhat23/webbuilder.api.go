package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/auth"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type previewProjectServiceStub struct {
	services.ProjectServiceInterface
	project *models.Project
	pages   []models.Page
	err     error
}

func (s *previewProjectServiceStub) GetWithAccess(_ context.Context, projectID, userID string) (*models.Project, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.project != nil && s.project.ID == projectID {
		return s.project, nil
	}
	return nil, repositories.ErrProjectNotFound
}

func (s *previewProjectServiceStub) ListPages(_ context.Context, projectID, userID string) ([]models.Page, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.pages, nil
}

type previewElementServiceStub struct {
	services.ElementServiceInterface
	elements []models.EditorElement
}

func (s *previewElementServiceStub) List(_ context.Context, _ string, _ ...string) ([]models.EditorElement, error) {
	return s.elements, nil
}

type previewContentTypeRepoStub struct {
	repositories.ContentTypeRepositoryInterface
	ct  *models.ContentType
	err error
}

func (s *previewContentTypeRepoStub) GetByID(_ context.Context, _ string) (*models.ContentType, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.ct, nil
}

type previewContentItemRepoStub struct {
	repositories.ContentItemRepositoryInterface
	items []models.ContentItem
	item  *models.ContentItem
	err   error
}

func (s *previewContentItemRepoStub) ListByContentType(_ context.Context, _ string) ([]models.ContentItem, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func (s *previewContentItemRepoStub) GetBySlug(_ context.Context, _, _ string) (*models.ContentItem, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.item, nil
}

func newTestPreviewApp(t *testing.T, rdb *redis.Client, secret string, proj *previewProjectServiceStub, elems *previewElementServiceStub, ctRepo *previewContentTypeRepoStub, ciRepo *previewContentItemRepoStub) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewPreviewTicketHandler(rdb, proj, elems, ctRepo, ciRepo).WithInternalSecret(secret)

	priv := app.Group("/api/v1", func(c fiber.Ctx) error {
		auth.SetFiber(c, auth.Identity{UserID: "user-1", AuthMethod: "clerk"})
		return c.Next()
	})
	priv.Post("/projects/:projectid/preview-ticket", h.Mint)

	pub := app.Group("/api/v1")
	pub.Post("/preview/exchange", h.Exchange)

	read := pub.Group("/preview/:projectid", h.RequirePreviewToken)
	read.Get("/project", h.GetProject)
	read.Get("/pages", h.ListPages)
	read.Get("/elements", h.ListElements)
	read.Get("/cms/:contentTypeId", h.ListCMSItems)
	read.Get("/cms/:contentTypeId/item/*", h.GetCMSItemBySlug)

	return app
}

func TestPreviewTicket_MintExchangeSingleUseAndScopedReads(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	projID := "proj-1"
	ctProjID := projID
	pageID := "page-1"
	summaryVal := "Hello Draft"
	app := newTestPreviewApp(
		t,
		rdb,
		"super-secret-internal-value",
		&previewProjectServiceStub{
			project: &models.Project{ID: projID, Name: "Draft Site", Published: false},
			pages:   []models.Page{{Id: pageID, ProjectId: projID, Name: "Home", Slug: "", Type: "sp"}},
		},
		&previewElementServiceStub{
			elements: []models.EditorElement{&models.Element{Id: "el-1", Type: "Text", PageId: &pageID}},
		},
		&previewContentTypeRepoStub{
			ct: &models.ContentType{Id: "ct-1", Name: "Posts", ProjectId: &ctProjID, PublicRead: false},
		},
		&previewContentItemRepoStub{
			items: []models.ContentItem{
				{
					Id:            "item-1",
					ContentTypeId: "ct-1",
					Slug:          "guides/intro",
					Title:         "Draft Intro",
					Published:     false,
					FieldValues: []models.ContentFieldValue{
						{Value: &summaryVal, Field: &models.ContentField{Name: "summary"}},
					},
				},
			},
			item: &models.ContentItem{
				Id:            "item-1",
				ContentTypeId: "ct-1",
				Slug:          "guides/intro",
				Title:         "Draft Intro",
				Published:     false,
				FieldValues: []models.ContentFieldValue{
					{Value: &summaryVal, Field: &models.ContentField{Name: "summary"}},
				},
			},
		},
	)

	// 1. Mint ticket
	mintRes, err := app.Test(httptest.NewRequest("POST", "/api/v1/projects/proj-1/preview-ticket", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer mintRes.Body.Close()
	if mintRes.StatusCode != fiber.StatusCreated {
		t.Fatalf("mint status = %d, want 201", mintRes.StatusCode)
	}
	var mintOut struct {
		Ticket    string `json:"ticket"`
		ExpiresIn int    `json:"expiresIn"`
	}
	if err := json.NewDecoder(mintRes.Body).Decode(&mintOut); err != nil {
		t.Fatal(err)
	}
	if mintOut.Ticket == "" || mintOut.ExpiresIn != 60 {
		t.Fatalf("unexpected mint response: %+v", mintOut)
	}

	// Verify only SHA-256 digest is stored in Redis with 60s TTL.
	sum := sha256.Sum256([]byte(mintOut.Ticket))
	ticketKey := previewTicketPrefix + hex.EncodeToString(sum[:])
	if !mr.Exists(ticketKey) {
		t.Fatalf("expected hashed ticket key in redis")
	}
	if ttl := mr.TTL(ticketKey); ttl <= 0 || ttl > 60*time.Second {
		t.Fatalf("unexpected ticket ttl: %v", ttl)
	}

	// 2. Wrong internal secret is rejected without consuming ticket.
	badSecretReq := httptest.NewRequest("POST", "/api/v1/preview/exchange", bytes.NewBufferString(`{"ticket":"`+mintOut.Ticket+`"}`))
	badSecretReq.Header.Set("Content-Type", "application/json")
	badSecretReq.Header.Set("X-Preview-Internal-Secret", "wrong-secret")
	badSecretRes, err := app.Test(badSecretReq)
	if err != nil {
		t.Fatal(err)
	}
	if badSecretRes.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("bad secret status = %d, want 401", badSecretRes.StatusCode)
	}

	// 3. Valid exchange consumes ticket and returns 15m session token.
	exchReq := httptest.NewRequest("POST", "/api/v1/preview/exchange", bytes.NewBufferString(`{"ticket":"`+mintOut.Ticket+`"}`))
	exchReq.Header.Set("Content-Type", "application/json")
	exchReq.Header.Set("X-Preview-Internal-Secret", "super-secret-internal-value")
	exchRes, err := app.Test(exchReq)
	if err != nil {
		t.Fatal(err)
	}
	defer exchRes.Body.Close()
	if exchRes.StatusCode != fiber.StatusOK {
		t.Fatalf("exchange status = %d, want 200", exchRes.StatusCode)
	}
	var exchOut struct {
		Token     string `json:"token"`
		ProjectID string `json:"projectId"`
		ExpiresIn int    `json:"expiresIn"`
	}
	if err := json.NewDecoder(exchRes.Body).Decode(&exchOut); err != nil {
		t.Fatal(err)
	}
	if exchOut.Token == "" || exchOut.ProjectID != projID || exchOut.ExpiresIn != 900 {
		t.Fatalf("unexpected exchange output: %+v", exchOut)
	}

	// 4. Ticket replay MUST fail (single-use GETDEL).
	replayReq := httptest.NewRequest("POST", "/api/v1/preview/exchange", bytes.NewBufferString(`{"ticket":"`+mintOut.Ticket+`"}`))
	replayReq.Header.Set("Content-Type", "application/json")
	replayReq.Header.Set("X-Preview-Internal-Secret", "super-secret-internal-value")
	replayRes, err := app.Test(replayReq)
	if err != nil {
		t.Fatal(err)
	}
	if replayRes.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("replay status = %d, want 401", replayRes.StatusCode)
	}

	// 5. Cross-project read MUST be forbidden (403).
	crossReq := httptest.NewRequest("GET", "/api/v1/preview/other-project/project", nil)
	crossReq.Header.Set("X-Preview-Token", exchOut.Token)
	crossRes, err := app.Test(crossReq)
	if err != nil {
		t.Fatal(err)
	}
	if crossRes.StatusCode != fiber.StatusForbidden {
		t.Fatalf("cross-project status = %d, want 403", crossRes.StatusCode)
	}

	// 6. Scoped project, pages, elements, and draft CMS reads succeed.
	for _, path := range []string{
		"/api/v1/preview/proj-1/project",
		"/api/v1/preview/proj-1/pages",
		"/api/v1/preview/proj-1/elements",
		"/api/v1/preview/proj-1/cms/ct-1?limit=5&sortBy=title&sortOrder=asc",
		"/api/v1/preview/proj-1/cms/ct-1/item/guides/intro",
	} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-Preview-Token", exchOut.Token)
		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("GET %s status = %d, want 200 (body=%s)", path, res.StatusCode, string(body))
		}
		if path == "/api/v1/preview/proj-1/cms/ct-1/item/guides/intro" {
			var flattened map[string]any
			if err := json.Unmarshal(body, &flattened); err != nil {
				t.Fatalf("unmarshal flattened cms item: %v", err)
			}
			if flattened["summary"] != "Hello Draft" || flattened["slug"] != "guides/intro" {
				t.Fatalf("unexpected flattened cms item: %+v", flattened)
			}
		}
	}
}

func TestPreviewTicket_RejectsCrossProjectContentType(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	otherProj := "proj-other"
	app := newTestPreviewApp(
		t,
		rdb,
		"sec",
		&previewProjectServiceStub{project: &models.Project{ID: "proj-1", Name: "Draft"}},
		&previewElementServiceStub{},
		&previewContentTypeRepoStub{ct: &models.ContentType{Id: "ct-foreign", ProjectId: &otherProj}},
		&previewContentItemRepoStub{},
	)

	// Insert a valid preview session directly for proj-1.
	token := "valid-preview-token"
	payload, _ := json.Marshal(previewSessionPayload{UserID: "user-1", ProjectID: "proj-1"})
	if err := rdb.Set(context.Background(), previewSessionPrefix+hashTokenHex(token), payload, time.Minute).Err(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/v1/preview/proj-1/cms/ct-foreign/item/secret", nil)
	req.Header.Set("X-Preview-Token", token)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("foreign content type status = %d, want 404", res.StatusCode)
	}
}

func TestPreviewTicket_FailsClosedWithoutSecretOrRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appNoSecret := newTestPreviewApp(t, rdb, "", nil, nil, nil, nil)
	req := httptest.NewRequest("POST", "/api/v1/preview/exchange", bytes.NewBufferString(`{"ticket":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := appNoSecret.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("missing secret status = %d, want 503", res.StatusCode)
	}

	appNoRedis := newTestPreviewApp(t, nil, "sec", nil, nil, nil, nil)
	mintRes, err := appNoRedis.Test(httptest.NewRequest("POST", "/api/v1/projects/proj-1/preview-ticket", nil))
	if err != nil {
		t.Fatal(err)
	}
	if mintRes.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("missing redis status = %d, want 503", mintRes.StatusCode)
	}
}
