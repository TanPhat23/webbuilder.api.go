package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"my-go-app/internal/models"
)

// imageHasArg reports whether any recorded statement was sent with the given bind value.
func imageHasArg(f *fakeDB, want any) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, args := range f.args {
		for _, a := range args {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

var imageBoom = errors.New("boom")

func imageRow(id, user string) []driver.Value {
	return []driver.Value{id, "https://img/" + id, user}
}

var imageCols = []string{"ImageId", "ImageLink", "UserId"}

func TestImageRepositoryCreateImage(t *testing.T) {
	ctx := context.Background()
	name := "logo"
	t.Run("inserts and returns the image", func(t *testing.T) {
		db, f := newFakeDB(t)
		got, err := (&ImageRepository{db: db}).CreateImage(ctx, models.Image{ImageId: "i1", UserId: "u1", ImageLink: "l", ImageName: &name})
		if err != nil {
			t.Fatal(err)
		}
		if got.ImageId != "i1" || got.UserId != "u1" || got.ImageName == nil || *got.ImageName != "logo" {
			t.Fatalf("unexpected image %+v", got)
		}
		if !f.sawQuery("INSERT INTO", "Image") || !imageHasArg(f, "i1") {
			t.Fatalf("missing insert: %v", f.queries)
		}
	})
	t.Run("wraps database errors", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT", imageBoom)
		got, err := (&ImageRepository{db: db}).CreateImage(ctx, models.Image{ImageId: "i1"})
		if got != nil || !errors.Is(err, imageBoom) || !strings.HasPrefix(err.Error(), "failed to create image") {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}

func TestImageRepositoryGetImagesByUserID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", imageCols, imageRow("a", "u1"), imageRow("b", "u1"))
	got, err := (&ImageRepository{db: db}).GetImagesByUserID(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ImageId != "a" || got[1].ImageLink != "https://img/b" {
		t.Fatalf("unexpected images %+v", got)
	}
	if !f.sawQuery("DeletedAt", "IS NULL", "ORDER BY", "DESC") || !imageHasArg(f, "u1") {
		t.Fatalf("query should filter by user, skip deleted and order newest first: %v", f.queries)
	}

	db, f = newFakeDB(t)
	f.onErr("SELECT", imageBoom)
	if got, err := (&ImageRepository{db: db}).GetImagesByUserID(ctx, "u1"); got != nil || !errors.Is(err, imageBoom) {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestImageRepositoryGetImageByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", imageCols, imageRow("a", "u1"))
	got, err := (&ImageRepository{db: db}).GetImageByID(ctx, "a", "u1")
	if err != nil || got.ImageId != "a" || got.UserId != "u1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !imageHasArg(f, "a") || !imageHasArg(f, "u1") {
		t.Fatalf("image and user must both scope the query: %v", f.args)
	}

	db, _ = newFakeDB(t)
	if got, err := (&ImageRepository{db: db}).GetImageByID(ctx, "a", "u1"); got != nil || !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("missing row: got %+v, %v", got, err)
	}

	db, f = newFakeDB(t)
	f.onErr("SELECT", imageBoom)
	if _, err := (&ImageRepository{db: db}).GetImageByID(ctx, "a", "u1"); !errors.Is(err, imageBoom) || errors.Is(err, ErrImageNotFound) {
		t.Fatalf("database error must not become not-found: %v", err)
	}
}

func TestImageRepositoryDeleteImage(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"deleted", 1, nil, nil},
		{"nothing matched", 0, nil, ErrImageNotFound},
		{"database error", 0, imageBoom, imageBoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("DELETE", tc.dbErr)
			} else {
				f.onExec("DELETE", tc.affected)
			}
			err := (&ImageRepository{db: db}).DeleteImage(ctx, "a", "u1")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if !f.sawQuery("DELETE FROM", "Image") || !imageHasArg(f, "u1") {
				t.Fatalf("unexpected statements: %v", f.queries)
			}
		})
	}
}

func TestImageRepositorySoftDeleteImage(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"deleted", 1, nil, nil},
		{"already deleted or not owned", 0, nil, ErrImageNotFound},
		{"database error", 0, imageBoom, imageBoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			if tc.dbErr != nil {
				f.onErr("UPDATE", tc.dbErr)
			} else {
				f.onExec("UPDATE", tc.affected)
			}
			err := (&ImageRepository{db: db}).SoftDeleteImage(ctx, "a", "u1")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if !f.sawQuery("UPDATE", "DeletedAt", "IS NULL") {
				t.Fatalf("soft delete must only touch live rows: %v", f.queries)
			}
		})
	}
}

func TestImageRepositoryGetAllImages(t *testing.T) {
	ctx := context.Background()
	t.Run("no paging", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("SELECT", imageCols, imageRow("a", "u1"))
		got, err := (&ImageRepository{db: db}).GetAllImages(ctx, 0, 0)
		if err != nil || len(got) != 1 || got[0].ImageId != "a" {
			t.Fatalf("got %+v, %v", got, err)
		}
		q := f.lastQuery()
		if strings.Contains(q, "LIMIT") || strings.Contains(q, "OFFSET") {
			t.Fatalf("unexpected paging: %s", q)
		}
	})
	t.Run("limit and offset", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := (&ImageRepository{db: db}).GetAllImages(ctx, 5, 10); err != nil {
			t.Fatal(err)
		}
		q := f.lastQuery()
		if !strings.Contains(q, "LIMIT") || !strings.Contains(q, "OFFSET") {
			t.Fatalf("expected paging: %s", q)
		}
		if !imageHasArg(f, int64(5)) && !strings.Contains(q, "LIMIT 5") {
			t.Fatalf("limit value not applied: %s %v", q, f.args)
		}
	})
	t.Run("database error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", imageBoom)
		got, err := (&ImageRepository{db: db}).GetAllImages(ctx, 0, 0)
		if got != nil || !errors.Is(err, imageBoom) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}
