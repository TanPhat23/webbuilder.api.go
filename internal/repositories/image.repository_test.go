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
	db, f := newFakeDB(t)
	got, err := (&ImageRepository{db: db}).CreateImage(ctx, models.Image{ImageId: "i1", UserId: "u1", ImageLink: "l", ImageName: &name})
	if err != nil {
		t.Fatal(err)
	}
	if got.ImageId != "i1" || got.UserId != "u1" {
		t.Fatalf("unexpected image %+v", got)
	}
	if got.ImageName == nil || *got.ImageName != "logo" {
		t.Fatalf("unexpected image name %+v", got.ImageName)
	}
	if !f.sawQuery("INSERT INTO", "Image") || !imageHasArg(f, "i1") {
		t.Fatalf("missing insert: %v", f.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("INSERT", imageBoom)
	gotErr, err := (&ImageRepository{db: dbErr}).CreateImage(ctx, models.Image{ImageId: "i1"})
	if gotErr != nil || !errors.Is(err, imageBoom) || !strings.HasPrefix(err.Error(), "failed to create image") {
		t.Fatalf("got %+v, %v", gotErr, err)
	}
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

	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", imageBoom)
	if got, err := (&ImageRepository{db: db2}).GetImagesByUserID(ctx, "u1"); got != nil || !errors.Is(err, imageBoom) {
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

	dbMiss, fMiss := newFakeDB(t)
	if fMiss == nil {
		t.Fatal("missing fake db")
	}
	if got, err := (&ImageRepository{db: dbMiss}).GetImageByID(ctx, "a", "u1"); got != nil || !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("missing row: got %+v, %v", got, err)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", imageBoom)
	if _, err := (&ImageRepository{db: dbErr}).GetImageByID(ctx, "a", "u1"); !errors.Is(err, imageBoom) || errors.Is(err, ErrImageNotFound) {
		t.Fatalf("database error must not become not-found: %v", err)
	}
}

func TestImageRepositoryDeleteImage(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"deleted", 1, nil, nil},
		{"nothing matched", 0, nil, ErrImageNotFound},
		{"database error", 0, imageBoom, imageBoom},
	} {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("DELETE", tc.dbErr)
		} else {
			f.onExec("DELETE", tc.affected)
		}
		err := (&ImageRepository{db: db}).DeleteImage(ctx, "a", "u1")
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s: want %v, got %v", tc.name, tc.wantErr, err)
		}
		if !f.sawQuery("DELETE FROM", "Image") || !imageHasArg(f, "u1") {
			t.Fatalf("%s: unexpected statements: %v", tc.name, f.queries)
		}
	}
}

func TestImageRepositorySoftDeleteImage(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		affected int64
		dbErr    error
		wantErr  error
	}{
		{"deleted", 1, nil, nil},
		{"already deleted or not owned", 0, nil, ErrImageNotFound},
		{"database error", 0, imageBoom, imageBoom},
	} {
		db, f := newFakeDB(t)
		if tc.dbErr != nil {
			f.onErr("UPDATE", tc.dbErr)
		} else {
			f.onExec("UPDATE", tc.affected)
		}
		err := (&ImageRepository{db: db}).SoftDeleteImage(ctx, "a", "u1")
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s: want %v, got %v", tc.name, tc.wantErr, err)
		}
		if !f.sawQuery("UPDATE", "DeletedAt", "IS NULL") {
			t.Fatalf("%s: soft delete must only touch live rows: %v", tc.name, f.queries)
		}
	}
}

func TestImageRepositoryGetAllImages(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("SELECT", imageCols, imageRow("a", "u1"))
	got, err := (&ImageRepository{db: db}).GetAllImages(ctx, 0, 0)
	if err != nil || len(got) != 1 || got[0].ImageId != "a" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if q := f.lastQuery(); strings.Contains(q, "LIMIT") || strings.Contains(q, "OFFSET") {
		t.Fatalf("unexpected paging: %s", q)
	}

	dbPage, fPage := newFakeDB(t)
	if _, err := (&ImageRepository{db: dbPage}).GetAllImages(ctx, 5, 10); err != nil {
		t.Fatal(err)
	}
	q := fPage.lastQuery()
	if !strings.Contains(q, "LIMIT") || !strings.Contains(q, "OFFSET") {
		t.Fatalf("expected paging: %s", q)
	}
	if !imageHasArg(fPage, int64(5)) && !strings.Contains(q, "LIMIT 5") {
		t.Fatalf("limit value not applied: %s %v", q, fPage.args)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", imageBoom)
	if got, err := (&ImageRepository{db: dbErr}).GetAllImages(ctx, 0, 0); got != nil || !errors.Is(err, imageBoom) {
		t.Fatalf("got %+v, %v", got, err)
	}
}
