package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type memWebhookRepo struct {
	mu         sync.Mutex
	webhooks   map[string]models.ProjectWebhook
	deliveries map[string]models.WebhookDelivery
}

func newMemWebhookRepo() *memWebhookRepo {
	return &memWebhookRepo{
		webhooks:   make(map[string]models.ProjectWebhook),
		deliveries: make(map[string]models.WebhookDelivery),
	}
}

func (r *memWebhookRepo) Create(_ context.Context, wh *models.ProjectWebhook) (*models.ProjectWebhook, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.webhooks[wh.Id] = *wh
	return wh, nil
}
func (r *memWebhookRepo) ListByProject(_ context.Context, projectID string) ([]models.ProjectWebhook, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []models.ProjectWebhook
	for _, wh := range r.webhooks {
		if wh.ProjectId == projectID {
			out = append(out, wh)
		}
	}
	return out, nil
}
func (r *memWebhookRepo) GetByID(_ context.Context, id string) (*models.ProjectWebhook, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	wh, ok := r.webhooks[id]
	if !ok {
		return nil, repositories.ErrWebhookNotFound
	}
	return &wh, nil
}
func (r *memWebhookRepo) Update(_ context.Context, id string, _ map[string]any) (*models.ProjectWebhook, error) {
	return r.GetByID(context.Background(), id)
}
func (r *memWebhookRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.webhooks, id)
	return nil
}
func (r *memWebhookRepo) FindMatchingForEvent(_ context.Context, projectID, contentTypeID, event string) ([]models.ProjectWebhook, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []models.ProjectWebhook
	for _, wh := range r.webhooks {
		if wh.ProjectId == projectID && wh.Enabled && (wh.ContentTypeId == nil || *wh.ContentTypeId == contentTypeID) && wh.SubscribesTo(event) {
			out = append(out, wh)
		}
	}
	return out, nil
}
func (r *memWebhookRepo) CreateDelivery(_ context.Context, d *models.WebhookDelivery) (*models.WebhookDelivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deliveries[d.Id] = *d
	return d, nil
}
func (r *memWebhookRepo) GetDeliveryByID(_ context.Context, id string) (*models.WebhookDelivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d := r.deliveries[id]
	return &d, nil
}
func (r *memWebhookRepo) UpdateDelivery(_ context.Context, id string, updates map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d := r.deliveries[id]
	if v, ok := updates["Status"].(string); ok {
		d.Status = v
	}
	if v, ok := updates["HttpStatus"].(int); ok {
		d.HttpStatus = &v
	} else if v, ok := updates["HttpStatus"].(*int); ok {
		d.HttpStatus = v
	}
	if v, ok := updates["Attempts"].(int); ok {
		d.Attempts = v
	}
	if v, ok := updates["NextRetryAt"].(time.Time); ok {
		d.NextRetryAt = &v
	} else if updates["NextRetryAt"] == nil {
		d.NextRetryAt = nil
	}
	r.deliveries[id] = d
	return nil
}
func (r *memWebhookRepo) ListDeliveriesByWebhook(_ context.Context, webhookID string, _ int) ([]models.WebhookDelivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []models.WebhookDelivery
	for _, d := range r.deliveries {
		if d.WebhookId == webhookID {
			out = append(out, d)
		}
	}
	return out, nil
}
func (r *memWebhookRepo) ListPendingRetries(_ context.Context, _ time.Time, _ int) ([]models.WebhookDelivery, error) {
	return nil, nil
}

func TestWebhook_SSRFValidation(t *testing.T) {
	blocked := []string{
		"http://127.0.0.1:8080/hook",
		"http://localhost/hook",
		"http://169.254.169.254/latest/meta-data",
		"http://10.0.0.1/hook",
		"http://100.64.0.1/hook",
		"http://0.0.0.0/hook",
		"http://198.18.0.1/hook",
		"http://[64:ff9b::1]/hook",
		"http://[fc00::1]/hook",
		"http://[fe80::1]/hook",
		"http://[::1]/hook",
		"http://[::ffff:127.0.0.1]/hook",
		"file:///etc/passwd",
	}
	for _, u := range blocked {
		if err := ValidateWebhookURL(u, false); err == nil {
			t.Errorf("expected ValidateWebhookURL(%q, false) to fail", u)
		}
	}
	if err := ValidateWebhookURL("http://127.0.0.1:8099/hook", true); err != nil {
		t.Fatalf("expected local URL allowed when allowLocal=true, got %v", err)
	}
	if err := ValidateWebhookURL("https://hooks.example.com/wb", false); err != nil {
		t.Fatalf("expected public https URL allowed, got %v", err)
	}
}

func TestWebhook_NoRedirectsAndEnvAndFullQueue(t *testing.T) {
	// 1. HTTP redirects must not be followed
	targetCalled := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer redirector.Close()

	repo := newMemWebhookRepo()
	_, _ = repo.Create(context.Background(), &models.ProjectWebhook{
		Id: "wh_r", ProjectId: "p1", Url: redirector.URL, Secret: "sec",
		Events: models.StringSlice{models.WebhookEventItemCreated}, Enabled: true,
	})
	_, _ = repo.CreateDelivery(context.Background(), &models.WebhookDelivery{
		Id: "del_r", WebhookId: "wh_r", ProjectId: "p1", Event: models.WebhookEventItemCreated,
		Payload: []byte(`{}`), Status: models.WebhookDeliveryStatusPending,
	})
	svc := NewWebhookServiceWithOptions(repo, nil, nil, true, 2*time.Second)
	_ = svc.DeliverByID(context.Background(), "del_r")
	if targetCalled {
		t.Fatal("webhook HTTP client must not follow redirects")
	}
	delAfter, _ := repo.GetDeliveryByID(context.Background(), "del_r")
	if delAfter.HttpStatus == nil || *delAfter.HttpStatus != http.StatusFound || delAfter.Status != models.WebhookDeliveryStatusFailed {
		t.Fatalf("expected failed 302 delivery, got %+v", delAfter)
	}

	// 2. ALLOW_LOCAL_WEBHOOKS=true is refused in production
	t.Setenv("ALLOW_LOCAL_WEBHOOKS", "true")
	for _, env := range []string{"production", "prod"} {
		t.Setenv("ENV", env)
		if err := ValidateWebhookEnv(); err == nil {
			t.Fatalf("expected ValidateWebhookEnv error for ENV=%s", env)
		}
		prodSvc := NewWebhookService(repo, nil, nil)
		if prodSvc.allowLocal {
			t.Fatalf("expected NewWebhookService to disable allowLocal in ENV=%s", env)
		}
	}

	// 3. Full queue schedules NextRetryAt for DB retry poller without blocking or spawning goroutines
	svcQ := NewWebhookServiceWithOptions(repo, nil, nil, true, time.Second)
	for i := 0; i < cap(svcQ.queue); i++ {
		svcQ.queue <- "filled"
	}
	ct := &models.ContentType{Id: "ct1", Slug: "posts", Name: "Posts"}
	item := &models.ContentItem{Id: "it1", ContentTypeId: "ct1", Slug: "s", Title: "T"}
	svcQ.EmitItemEvent(context.Background(), "p1", ct, models.WebhookEventItemCreated, item)
	deliveries, _ := repo.ListDeliveriesByWebhook(context.Background(), "wh_r", 10)
	foundScheduled := false
	for _, d := range deliveries {
		if d.Id != "del_r" && d.Status == models.WebhookDeliveryStatusPending && d.NextRetryAt != nil {
			foundScheduled = true
		}
	}
	if !foundScheduled {
		t.Fatalf("expected full-queue delivery to be scheduled with NextRetryAt, got %+v", deliveries)
	}
}

func TestWebhook_DeliverySigningAndRetries(t *testing.T) {
	var gotEvent, gotSig, gotTS string
	var gotBody []byte
	statusToReturn := http.StatusOK

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEvent = r.Header.Get("X-WebBuilder-Event")
		gotSig = r.Header.Get("X-WebBuilder-Signature")
		gotTS = r.Header.Get("X-WebBuilder-Timestamp")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(statusToReturn)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	repo := newMemWebhookRepo()
	wh := &models.ProjectWebhook{
		Id:        "wh_1",
		ProjectId: "proj_1",
		Name:      "Test Hook",
		Url:       srv.URL,
		Secret:    "whsec_supersecret",
		Events:    models.StringSlice{models.WebhookEventItemCreated},
		Enabled:   true,
	}
	_, _ = repo.Create(context.Background(), wh)

	svc := NewWebhookServiceWithOptions(repo, nil, nil, true, 2*time.Second)

	// 1. 200 OK -> succeeded + valid HMAC
	d1 := &models.WebhookDelivery{
		Id:        "del_1",
		WebhookId: "wh_1",
		ProjectId: "proj_1",
		Event:     models.WebhookEventItemCreated,
		Payload:   []byte(`{"hello":"world"}`),
		Status:    models.WebhookDeliveryStatusPending,
	}
	_, _ = repo.CreateDelivery(context.Background(), d1)
	if err := svc.DeliverByID(context.Background(), "del_1"); err != nil {
		t.Fatalf("DeliverByID failed: %v", err)
	}
	after1, _ := repo.GetDeliveryByID(context.Background(), "del_1")
	if after1.Status != models.WebhookDeliveryStatusSucceeded || after1.HttpStatus == nil || *after1.HttpStatus != 200 {
		t.Fatalf("expected succeeded/200, got %+v", after1)
	}
	tsNum, _ := strconv.ParseInt(gotTS, 10, 64)
	wantSig := SignWebhookPayload("whsec_supersecret", tsNum, gotBody)
	if gotEvent != models.WebhookEventItemCreated || gotSig != wantSig {
		t.Fatalf("signature mismatch: got %q want %q", gotSig, wantSig)
	}

	// 2. 500 -> pending with retry scheduled, then failed on 3rd attempt
	statusToReturn = http.StatusInternalServerError
	d2 := &models.WebhookDelivery{
		Id:        "del_2",
		WebhookId: "wh_1",
		ProjectId: "proj_1",
		Event:     models.WebhookEventItemCreated,
		Payload:   []byte(`{"retry":true}`),
		Status:    models.WebhookDeliveryStatusPending,
	}
	_, _ = repo.CreateDelivery(context.Background(), d2)
	_ = svc.DeliverByID(context.Background(), "del_2")
	after2, _ := repo.GetDeliveryByID(context.Background(), "del_2")
	if after2.Status != models.WebhookDeliveryStatusPending || after2.Attempts != 1 || after2.NextRetryAt == nil {
		t.Fatalf("expected pending retry after 500, got %+v", after2)
	}
	_ = svc.DeliverByID(context.Background(), "del_2")
	_ = svc.DeliverByID(context.Background(), "del_2")
	afterFinal, _ := repo.GetDeliveryByID(context.Background(), "del_2")
	if afterFinal.Status != models.WebhookDeliveryStatusFailed || afterFinal.Attempts != 3 {
		t.Fatalf("expected failed after 3 attempts, got %+v", afterFinal)
	}

	// 3. 400 -> failed immediately (no retry)
	statusToReturn = http.StatusBadRequest
	d3 := &models.WebhookDelivery{
		Id:        "del_3",
		WebhookId: "wh_1",
		ProjectId: "proj_1",
		Event:     models.WebhookEventItemCreated,
		Payload:   []byte(`{"bad":true}`),
		Status:    models.WebhookDeliveryStatusPending,
	}
	_, _ = repo.CreateDelivery(context.Background(), d3)
	_ = svc.DeliverByID(context.Background(), "del_3")
	after3, _ := repo.GetDeliveryByID(context.Background(), "del_3")
	if after3.Status != models.WebhookDeliveryStatusFailed || after3.Attempts != 1 {
		t.Fatalf("expected immediate failure on 400, got %+v", after3)
	}
}
