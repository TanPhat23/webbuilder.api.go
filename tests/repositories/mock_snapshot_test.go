package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockSnapshotRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockSnapshotRepository{}
	_, err := repo.GetSnapshotByID(context.Background(), "snap-1")
	if !errors.Is(err, repositories.ErrSnapshotNotFound) {
		t.Errorf("GetSnapshotByID default: want ErrSnapshotNotFound, got %v", err)
	}
}

func TestMockSnapshotRepository_SaveSnapshotFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, projectID string, snap *models.Snapshot) error {
			called = true
			snap.ProjectId = projectID
			return nil
		},
	}

	snap := &models.Snapshot{Name: "v1", Type: "version"}
	if err := repo.SaveSnapshot(context.Background(), "p1", snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("SaveSnapshotFn was not called")
	}
	if snap.ProjectId != "p1" {
		t.Errorf("ProjectId: got %q, want %q", snap.ProjectId, "p1")
	}
}

func TestMockSnapshotRepository_GetSnapshotsByProjectIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockSnapshotRepository{}
	snaps, err := repo.GetSnapshotsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected empty slice, got %d", len(snaps))
	}
}

func TestMockSnapshotRepository_GetSnapshotByIDFnReturnsSnapshot(t *testing.T) {
	want := &models.Snapshot{Id: "snap-1", ProjectId: "p1", Name: "Before deploy"}
	repo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, id string) (*models.Snapshot, error) {
			if id == want.Id {
				return want, nil
			}
			return nil, repositories.ErrSnapshotNotFound
		},
	}

	got, err := repo.GetSnapshotByID(context.Background(), "snap-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetSnapshotByID(context.Background(), "snap-missing")
	if !errors.Is(err, repositories.ErrSnapshotNotFound) {
		t.Errorf("missing snap: want ErrSnapshotNotFound, got %v", err)
	}
}

func TestMockSnapshotRepository_DeleteSnapshotDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockSnapshotRepository{}
	if err := repo.DeleteSnapshot(context.Background(), "snap-1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockSnapshotRepository_DeleteSnapshotFnCalled(t *testing.T) {
	var captured string
	repo := &testutil.MockSnapshotRepository{
		DeleteSnapshotFn: func(_ context.Context, id string) error {
			captured = id
			return nil
		},
	}

	if err := repo.DeleteSnapshot(context.Background(), "snap-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != "snap-42" {
		t.Errorf("snapshotID: got %q, want %q", captured, "snap-42")
	}
}

func TestMockSnapshotRepository_GetSnapshotsByProjectIDFnFilters(t *testing.T) {
	all := []models.Snapshot{
		{Id: "s1", ProjectId: "p1"},
		{Id: "s2", ProjectId: "p2"},
		{Id: "s3", ProjectId: "p1"},
	}
	repo := &testutil.MockSnapshotRepository{
		GetSnapshotsByProjectIDFn: func(_ context.Context, projectID string) ([]models.Snapshot, error) {
			var out []models.Snapshot
			for _, s := range all {
				if s.ProjectId == projectID {
					out = append(out, s)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetSnapshotsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 snapshots for p1, got %d", len(got))
	}
	for _, s := range got {
		if s.ProjectId != "p1" {
			t.Errorf("unexpected snapshot project %q", s.ProjectId)
		}
	}
}