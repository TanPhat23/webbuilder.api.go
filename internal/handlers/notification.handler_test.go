package handlers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type notificationHandlerStub struct {
	repositories.NotificationRepositoryInterface
	err error

	gotUser, gotID string
	gotFilter      models.NotificationFilter
	gotRead        bool
	markedAll      string
	created        *models.Notification
}

func (s *notificationHandlerStub) ListForUser(_ context.Context, userID string, f models.NotificationFilter) (*models.NotificationList, error) {
	s.gotUser, s.gotFilter = userID, f
	if s.err != nil {
		return nil, s.err
	}
	return &models.NotificationList{
		Notifications: []models.Notification{{Id: "n1", UserId: userID, Title: "Hello"}},
		Stats:         models.NotificationStats{Total: 7, Unread: 3},
	}, nil
}

func (s *notificationHandlerStub) SetRead(_ context.Context, userID, id string, read bool) (*models.Notification, error) {
	s.gotUser, s.gotID, s.gotRead = userID, id, read
	if s.err != nil {
		return nil, s.err
	}
	return &models.Notification{Id: id, UserId: userID, Read: read}, nil
}

func (s *notificationHandlerStub) MarkAllRead(_ context.Context, userID string) error {
	s.markedAll = userID
	return s.err
}

func (s *notificationHandlerStub) Create(_ context.Context, n *models.Notification) error {
	s.created = n
	return s.err
}

func notificationHandlerApp(user string, repo *notificationHandlerStub) *fiber.App {
	h := NewNotificationHandler(repo)
	return newKitApp(user, func(app *fiber.App) {
		app.Get("/notifications", h.GetNotifications)
		app.Patch("/notifications/read-all", h.MarkAllRead)
		app.Post("/notifications/profile-update", h.CreateProfileUpdate)
		app.Patch("/notifications/:id", h.SetNotificationRead)
		app.Patch("/noid", h.SetNotificationRead)
	})
}

func TestGetNotifications(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		r := kitSend(t, notificationHandlerApp("u1", repo), "GET", "/notifications", "")
		kitExpect(t, r, 200, "Hello", `"total":7`, `"unread":3`)
		if repo.gotUser != "u1" || repo.gotFilter != (models.NotificationFilter{Filter: "all", Limit: 50}) {
			t.Fatalf("repo got user=%q filter=%+v", repo.gotUser, repo.gotFilter)
		}
	})
	t.Run("query is forwarded", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", repo), "GET", "/notifications?filter=unread&search=bill&limit=5", ""), 200)
		if want := (models.NotificationFilter{Filter: "unread", Search: "bill", Limit: 5}); repo.gotFilter != want {
			t.Fatalf("filter = %+v, want %+v", repo.gotFilter, want)
		}
	})
	t.Run("unparsable limit becomes zero", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", repo), "GET", "/notifications?limit=abc", ""), 200)
		if repo.gotFilter.Limit != 0 {
			t.Fatalf("limit = %d, want 0 (repository applies its own cap)", repo.gotFilter.Limit)
		}
	})
	t.Run("anonymous", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		kitExpect(t, kitSend(t, notificationHandlerApp("", repo), "GET", "/notifications", ""), 401)
		if repo.gotUser != "" {
			t.Fatal("repository queried without a user")
		}
	})
	t.Run("failure", func(t *testing.T) {
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", &notificationHandlerStub{err: errInternal}), "GET", "/notifications", ""), 500, "Failed to fetch notifications")
	})
}

func TestSetNotificationRead(t *testing.T) {
	cases := []struct {
		name, user, path, body string
		err                    error
		want                   int
		expect                 []string
		wantRead               bool
	}{
		{"defaults to read", "u1", "/notifications/n5", "", nil, 200, []string{`"id":"n5"`, `"read":true`}, true},
		{"explicit unread", "u1", "/notifications/n5", `{"read":false}`, nil, 200, []string{`"read":false`}, false},
		{"empty object defaults to read", "u1", "/notifications/n5", `{}`, nil, 200, []string{`"read":true`}, true},
		{"anonymous", "", "/notifications/n5", "", nil, 401, nil, false},
		{"missing id", "u1", "/noid", "", nil, 400, []string{"id is required"}, false},
		{"bad json", "u1", "/notifications/n5", `{"read":`, nil, 0, nil, false},
		{"not found", "u1", "/notifications/n5", "", repositories.ErrNotificationNotFound, 404, []string{"Notification not found"}, true},
		{"wrapped not found", "u1", "/notifications/n5", "", fmt.Errorf("set read: %w", repositories.ErrNotificationNotFound), 404, []string{"Notification not found"}, true},
		{"failure", "u1", "/notifications/n5", "", errInternal, 500, []string{"Failed to update notification"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &notificationHandlerStub{err: tc.err}
			kitExpect(t, kitSend(t, notificationHandlerApp(tc.user, repo), "PATCH", tc.path, tc.body), tc.want, tc.expect...)
			if tc.want == 200 || tc.err != nil {
				if repo.gotUser != "u1" || repo.gotID != "n5" || repo.gotRead != tc.wantRead {
					t.Fatalf("repo got user=%q id=%q read=%v", repo.gotUser, repo.gotID, repo.gotRead)
				}
			} else if repo.gotID != "" {
				t.Fatalf("repository reached for a rejected request (id=%q)", repo.gotID)
			}
		})
	}
}

func TestMarkAllRead(t *testing.T) {
	repo := &notificationHandlerStub{}
	kitExpect(t, kitSend(t, notificationHandlerApp("u1", repo), "PATCH", "/notifications/read-all", ""), 200, `"success":true`)
	if repo.markedAll != "u1" {
		t.Fatalf("marked all for %q", repo.markedAll)
	}
	repo = &notificationHandlerStub{}
	kitExpect(t, kitSend(t, notificationHandlerApp("", repo), "PATCH", "/notifications/read-all", ""), 401)
	if repo.markedAll != "" {
		t.Fatal("notifications were changed without a user")
	}
	kitExpect(t, kitSend(t, notificationHandlerApp("u1", &notificationHandlerStub{err: errInternal}), "PATCH", "/notifications/read-all", ""), 500, "Failed to mark notifications as read")
}

func TestCreateProfileUpdate(t *testing.T) {
	t.Run("default description", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", repo), "POST", "/notifications/profile-update", ""), 201, `"success":true`)
		want := models.Notification{UserId: "u1", Type: "settings", Title: "Profile Updated", Description: "You updated your profile information."}
		if repo.created == nil || *repo.created != want {
			t.Fatalf("created %+v, want %+v", repo.created, want)
		}
	})
	t.Run("custom description", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", repo), "POST", "/notifications/profile-update", `{"description":"Changed avatar"}`), 201)
		if repo.created == nil || repo.created.Description != "Changed avatar" {
			t.Fatalf("created %+v", repo.created)
		}
	})
	t.Run("empty description falls back to default", func(t *testing.T) {
		repo := &notificationHandlerStub{}
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", repo), "POST", "/notifications/profile-update", `{"description":""}`), 201)
		if repo.created == nil || repo.created.Description != "You updated your profile information." {
			t.Fatalf("created %+v", repo.created)
		}
	})
	t.Run("rejections", func(t *testing.T) {
		for name, tc := range map[string]struct{ user, body string }{
			"anonymous":        {"", ""},
			"bad json":         {"u1", `{"description":`},
			"description long": {"u1", `{"description":"` + strings.Repeat("x", 2001) + `"}`},
		} {
			repo := &notificationHandlerStub{}
			r := kitSend(t, notificationHandlerApp(tc.user, repo), "POST", "/notifications/profile-update", tc.body)
			if name == "anonymous" {
				kitExpect(t, r, 401)
			} else {
				kitExpect(t, r, 0)
			}
			if repo.created != nil {
				t.Fatalf("%s: notification was created", name)
			}
		}
	})
	t.Run("failure", func(t *testing.T) {
		kitExpect(t, kitSend(t, notificationHandlerApp("u1", &notificationHandlerStub{err: errInternal}), "POST", "/notifications/profile-update", ""), 500, "Failed to create notification")
	})
}
