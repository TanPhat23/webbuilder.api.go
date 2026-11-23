package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
)

type snapshotHandlerService struct {
	services.SnapshotServiceInterface
	err, saveErr error

	saved               *models.Snapshot
	savedProject        string
	gotProject, gotUser string
	gotID               string
}

func (s *snapshotHandlerService) SaveSnapshot(_ context.Context, projectID string, snap *models.Snapshot) error {
	s.savedProject, s.saved = projectID, snap
	return s.saveErr
}
func (s *snapshotHandlerService) GetSnapshotsByProjectID(_ context.Context, projectID, userID string) ([]models.Snapshot, error) {
	s.gotProject, s.gotUser = projectID, userID
	if s.err != nil {
		return nil, s.err
	}
	return []models.Snapshot{{Id: "s1", ProjectId: projectID, Name: "v1"}}, nil
}
func (s *snapshotHandlerService) GetSnapshotByID(_ context.Context, id string) (*models.Snapshot, error) {
	s.gotID = id
	if s.err != nil {
		return nil, s.err
	}
	return &models.Snapshot{Id: id, Name: "v1"}, nil
}
func (s *snapshotHandlerService) DeleteSnapshotWithAccess(_ context.Context, id, projectID, userID string) error {
	s.gotID, s.gotProject, s.gotUser = id, projectID, userID
	return s.err
}

type snapshotHandlerElements struct {
	repositories.ElementRepositoryInterface
	err      error
	replaced []models.EditorElement
	project  string
	calls    int
}

func (r *snapshotHandlerElements) ReplaceElements(_ context.Context, projectID string, els []models.EditorElement) error {
	r.calls++
	r.project, r.replaced = projectID, els
	return r.err
}

func snapshotHandlerApp(user string, svc *snapshotHandlerService, els *snapshotHandlerElements) *fiber.App {
	h := NewSnapshotHandler(svc, els)
	return newKitApp(user, func(app *fiber.App) {
		app.Post("/projects/:projectid/snapshots", h.SaveSnapshot)
		app.Get("/projects/:projectid/snapshots", h.GetSnapshots)
		app.Get("/snapshots/:snapshotid", h.GetSnapshotByID)
		app.Delete("/projects/:projectid/snapshots/:snapshotid", h.DeleteSnapshot)
		app.Post("/noproject", h.SaveSnapshot)
		app.Get("/noproject", h.GetSnapshots)
		app.Get("/nosnapshot", h.GetSnapshotByID)
		app.Delete("/nosnapshot/:projectid", h.DeleteSnapshot)
	})
}

const snapshotHandlerBody = `{"id":"snap-1","name":"First","elements":[{"id":"e1","type":"Text","name":"a"},{"id":"e2","type":"Button"}]}`

func TestSaveSnapshotStoresSnapshotAndSyncsElements(t *testing.T) {
	svc, els := &snapshotHandlerService{}, &snapshotHandlerElements{}
	kitExpect(t, kitSend(t, snapshotHandlerApp("", svc, els), "POST", "/projects/p1/snapshots", snapshotHandlerBody), 201, "Snapshot saved successfully")
	s := svc.saved
	if svc.savedProject != "p1" || s == nil || s.Id != "snap-1" || s.ProjectId != "p1" || s.Name != "First" || s.Type != "working" || s.Timestamp == 0 {
		t.Fatalf("saved %+v for %q", s, svc.savedProject)
	}
	var stored []map[string]any
	if err := json.Unmarshal(s.Elements, &stored); err != nil || len(stored) != 2 || stored[0]["id"] != "e1" || stored[1]["type"] != "Button" {
		t.Fatalf("stored elements = %s (%v)", s.Elements, err)
	}
	if els.project != "p1" || len(els.replaced) != 2 {
		t.Fatalf("synced %d elements for %q", len(els.replaced), els.project)
	}
}

func TestSaveSnapshotHonoursTypeAndTimestamp(t *testing.T) {
	svc, els := &snapshotHandlerService{}, &snapshotHandlerElements{}
	body := `{"id":"s","name":"n","type":"published","timestamp":1700000000123,"elements":[{"id":"e1","type":"Text"}]}`
	kitExpect(t, kitSend(t, snapshotHandlerApp("", svc, els), "POST", "/projects/p1/snapshots", body), 201)
	if svc.saved.Type != "published" || svc.saved.Timestamp != 1700000000123 {
		t.Fatalf("type=%q timestamp=%d", svc.saved.Type, svc.saved.Timestamp)
	}
}

func TestSaveSnapshotGeneratesIDWithoutValidator(t *testing.T) {
	// The production validator requires an id, so generation only applies where
	// no struct validator is installed.
	svc, els := &snapshotHandlerService{}, &snapshotHandlerElements{}
	app := fiber.New()
	app.Post("/projects/:projectid/snapshots", NewSnapshotHandler(svc, els).SaveSnapshot)
	kitExpect(t, kitSend(t, app, "POST", "/projects/p1/snapshots", `{"name":"n","elements":[{"id":"e1","type":"Text"}]}`), 201)
	if _, err := uuid.Parse(svc.saved.Id); err != nil {
		t.Fatalf("generated id %q is not a uuid", svc.saved.Id)
	}
}

func TestSaveSnapshotRejections(t *testing.T) {
	cases := []struct {
		name, path, body string
		want             int
		expect           string
	}{
		{"missing project", "/noproject", snapshotHandlerBody, 400, "projectid is required"},
		{"bad json", "/projects/p1/snapshots", `{`, 0, ""},
		{"no elements", "/projects/p1/snapshots", `{"id":"s","name":"n","elements":[]}`, 0, ""},
		{"no name", "/projects/p1/snapshots", `{"id":"s","elements":[{"id":"e"}]}`, 0, ""},
		{"no id", "/projects/p1/snapshots", `{"name":"n","elements":[{"id":"e"}]}`, 0, ""},
		{"element is not an object", "/projects/p1/snapshots", `{"id":"s","name":"n","elements":["oops"]}`, 400, "Invalid element structure"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, els := &snapshotHandlerService{}, &snapshotHandlerElements{}
			r := kitSend(t, snapshotHandlerApp("", svc, els), "POST", tc.path, tc.body)
			if tc.expect == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.expect)
			}
			if svc.saved != nil || els.calls != 0 {
				t.Fatal("nothing may be saved for a rejected request")
			}
		})
	}
}

func TestSaveSnapshotFailures(t *testing.T) {
	t.Run("snapshot store fails and elements are left alone", func(t *testing.T) {
		svc, els := &snapshotHandlerService{saveErr: errInternal}, &snapshotHandlerElements{}
		kitExpect(t, kitSend(t, snapshotHandlerApp("", svc, els), "POST", "/projects/p1/snapshots", snapshotHandlerBody), 500, "Failed to save and sync snapshot")
		if els.calls != 0 {
			t.Fatal("elements were replaced although the snapshot was not stored")
		}
	})
	t.Run("element sync fails", func(t *testing.T) {
		svc, els := &snapshotHandlerService{}, &snapshotHandlerElements{err: errors.New("replace failed")}
		kitExpect(t, kitSend(t, snapshotHandlerApp("", svc, els), "POST", "/projects/p1/snapshots", snapshotHandlerBody), 500, "Failed to save and sync snapshot")
	})
}

func TestSnapshotReadAndDeleteHandlers(t *testing.T) {
	cases := []struct {
		name, user, method, path string
		err                      error
		want                     int
		expect                   string
	}{
		{"list ok", "u1", "GET", "/projects/p1/snapshots", nil, 200, "v1"},
		{"list anonymous", "", "GET", "/projects/p1/snapshots", nil, 401, ""},
		{"list missing project", "u1", "GET", "/noproject", nil, 400, "projectid is required"},
		{"list not found", "u1", "GET", "/projects/p1/snapshots", gorm.ErrRecordNotFound, 404, "Project not found"},
		{"list failure", "u1", "GET", "/projects/p1/snapshots", errInternal, 500, "Failed to get snapshots"},
		{"get ok", "", "GET", "/snapshots/s9", nil, 200, `"id":"s9"`},
		{"get missing id", "", "GET", "/nosnapshot", nil, 400, "snapshotid is required"},
		{"get not found", "", "GET", "/snapshots/s9", gorm.ErrRecordNotFound, 404, "Snapshot not found"},
		{"get failure", "", "GET", "/snapshots/s9", errInternal, 500, "Failed to get snapshot"},
		{"delete ok", "u1", "DELETE", "/projects/p1/snapshots/s9", nil, 200, "Snapshot deleted successfully"},
		{"delete anonymous", "", "DELETE", "/projects/p1/snapshots/s9", nil, 401, ""},
		{"delete missing snapshot id", "u1", "DELETE", "/nosnapshot/p1", nil, 400, "snapshotid is required"},
		{"delete not found", "u1", "DELETE", "/projects/p1/snapshots/s9", gorm.ErrRecordNotFound, 404, "Snapshot not found"},
		{"delete failure", "u1", "DELETE", "/projects/p1/snapshots/s9", errInternal, 500, "Failed to delete snapshot"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &snapshotHandlerService{err: tc.err}
			r := kitSend(t, snapshotHandlerApp(tc.user, svc, &snapshotHandlerElements{}), tc.method, tc.path, "")
			if tc.expect == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.expect)
			}
			if tc.want == 200 && tc.method == "DELETE" && (svc.gotID != "s9" || svc.gotProject != "p1" || svc.gotUser != "u1") {
				t.Fatalf("delete got snapshot=%q project=%q user=%q", svc.gotID, svc.gotProject, svc.gotUser)
			}
			if tc.name == "list ok" && (svc.gotProject != "p1" || svc.gotUser != "u1") {
				t.Fatalf("list got project=%q user=%q", svc.gotProject, svc.gotUser)
			}
		})
	}
}
