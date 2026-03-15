package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

func TestMockCommentRepository_DefaultGetByIDReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	_, err := repo.GetCommentByID(context.Background(), "c1")
	if !errors.Is(err, repositories.ErrCommentNotFound) {
		t.Errorf("want ErrCommentNotFound, got %v", err)
	}
}

func TestMockCommentRepository_DefaultUpdateReturnsSentinel(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	_, err := repo.UpdateComment(context.Background(), "c1", "u1", map[string]any{"Content": "hi"})
	if !errors.Is(err, repositories.ErrCommentUnauthorized) {
		t.Errorf("want ErrCommentUnauthorized, got %v", err)
	}
}

func TestMockCommentRepository_CreateDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	input := models.Comment{Content: "hello", ItemId: "item-1"}
	got, err := repo.CreateComment(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != input.Content {
		t.Errorf("Content: got %q, want %q", got.Content, input.Content)
	}
}

func TestMockCommentRepository_CreateFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockCommentRepository{
		CreateCommentFn: func(_ context.Context, c models.Comment) (*models.Comment, error) {
			called = true
			c.Id = "c-generated"
			return &c, nil
		},
	}
	got, err := repo.CreateComment(context.Background(), models.Comment{Content: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreateCommentFn was not called")
	}
	if got.Id != "c-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "c-generated")
	}
}

func TestMockCommentRepository_GetCommentsDefaultsToEmpty(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	comments, total, err := repo.GetComments(context.Background(), models.CommentFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 0 || total != 0 {
		t.Errorf("expected empty result, got %d comments total=%d", len(comments), total)
	}
}

func TestMockCommentRepository_GetCommentsFnFilters(t *testing.T) {
	all := []models.Comment{
		{Id: "c1", ItemId: "item-1"},
		{Id: "c2", ItemId: "item-2"},
		{Id: "c3", ItemId: "item-1"},
	}
	repo := &testutil.MockCommentRepository{
		GetCommentsFn: func(_ context.Context, filter models.CommentFilter) ([]models.Comment, int64, error) {
			var out []models.Comment
			for _, c := range all {
				if filter.ItemId == "" || c.ItemId == filter.ItemId {
					out = append(out, c)
				}
			}
			return out, int64(len(out)), nil
		},
	}

	comments, total, err := repo.GetComments(context.Background(), models.CommentFilter{ItemId: "item-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 2 || total != 2 {
		t.Errorf("expected 2 comments for item-1, got %d (total=%d)", len(comments), total)
	}
}

func TestMockCommentRepository_UpdateFnReturnsUpdated(t *testing.T) {
	want := &models.Comment{Id: "c1", Content: "updated"}
	repo := &testutil.MockCommentRepository{
		UpdateCommentFn: func(_ context.Context, id, _ string, _ map[string]any) (*models.Comment, error) {
			if id == "c1" {
				return want, nil
			}
			return nil, repositories.ErrCommentNotFound
		},
	}

	got, err := repo.UpdateComment(context.Background(), "c1", "u1", map[string]any{"Content": "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != want.Content {
		t.Errorf("Content: got %q, want %q", got.Content, want.Content)
	}

	_, err = repo.UpdateComment(context.Background(), "c-missing", "u1", map[string]any{})
	if !errors.Is(err, repositories.ErrCommentNotFound) {
		t.Errorf("missing comment: want ErrCommentNotFound, got %v", err)
	}
}

func TestMockCommentRepository_DeleteDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	if err := repo.DeleteComment(context.Background(), "c1", "u1"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockCommentRepository_DeleteFnCalled(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCommentRepository{
		DeleteCommentFn: func(_ context.Context, id, _ string) error {
			capturedID = id
			return nil
		},
	}
	if err := repo.DeleteComment(context.Background(), "c-42", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "c-42" {
		t.Errorf("id: got %q, want %q", capturedID, "c-42")
	}
}

func TestMockCommentRepository_CreateReactionDefaultReturnsInput(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	input := models.CommentReaction{CommentId: "c1", UserId: "u1", Type: "like"}
	got, err := repo.CreateReaction(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Type != input.Type {
		t.Errorf("Type: got %q, want %q", got.Type, input.Type)
	}
}

func TestMockCommentRepository_CreateReactionFnCalled(t *testing.T) {
	called := false
	repo := &testutil.MockCommentRepository{
		CreateReactionFn: func(_ context.Context, r models.CommentReaction) (*models.CommentReaction, error) {
			called = true
			r.Id = "reaction-generated"
			return &r, nil
		},
	}
	got, err := repo.CreateReaction(context.Background(), models.CommentReaction{CommentId: "c1", UserId: "u1", Type: "heart"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreateReactionFn was not called")
	}
	if got.Id != "reaction-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "reaction-generated")
	}
}

func TestMockCommentRepository_DeleteReactionDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	if err := repo.DeleteReaction(context.Background(), "c1", "u1", "like"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockCommentRepository_GetReactionsByCommentIDDefaultsToEmpty(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	reactions, err := repo.GetReactionsByCommentID(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reactions) != 0 {
		t.Errorf("expected empty slice, got %d", len(reactions))
	}
}

func TestMockCommentRepository_GetReactionsByCommentIDFnFilters(t *testing.T) {
	all := []models.CommentReaction{
		{Id: "r1", CommentId: "c1", Type: "like"},
		{Id: "r2", CommentId: "c1", Type: "heart"},
		{Id: "r3", CommentId: "c2", Type: "like"},
	}
	repo := &testutil.MockCommentRepository{
		GetReactionsByCommentIDFn: func(_ context.Context, commentID string) ([]models.CommentReaction, error) {
			var out []models.CommentReaction
			for _, r := range all {
				if r.CommentId == commentID {
					out = append(out, r)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetReactionsByCommentID(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 reactions for c1, got %d", len(got))
	}
}

func TestMockCommentRepository_ReactionSummaryDefaultsToEmpty(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	summaries, err := repo.GetReactionSummary(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected empty slice, got %d", len(summaries))
	}
}

func TestMockCommentRepository_ReactionSummaryFnReturnsData(t *testing.T) {
	want := []models.ReactionSummary{{Type: "like", Count: 5}, {Type: "heart", Count: 2}}
	repo := &testutil.MockCommentRepository{
		GetReactionSummaryFn: func(_ context.Context, _ string) ([]models.ReactionSummary, error) {
			return want, nil
		},
	}
	got, err := repo.GetReactionSummary(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("expected %d summaries, got %d", len(want), len(got))
	}
	if got[0].Count != 5 {
		t.Errorf("Count: got %d, want 5", got[0].Count)
	}
}

func TestMockCommentRepository_GetCommentCountByItemIDDefaultsToZero(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	count, err := repo.GetCommentCountByItemID(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestMockCommentRepository_GetCommentCountByItemIDFnReturnsCount(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetCommentCountByItemIDFn: func(_ context.Context, itemID string) (int64, error) {
			if itemID == "item-1" {
				return 7, nil
			}
			return 0, nil
		},
	}
	count, err := repo.GetCommentCountByItemID(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 7 {
		t.Errorf("expected 7, got %d", count)
	}
}

func TestMockCommentRepository_ModerateCommentDefaultReturnsNil(t *testing.T) {
	repo := &testutil.MockCommentRepository{}
	if err := repo.ModerateComment(context.Background(), "c1", "approved"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockCommentRepository_ModerateCommentFnCalled(t *testing.T) {
	var capturedStatus string
	repo := &testutil.MockCommentRepository{
		ModerateCommentFn: func(_ context.Context, _, status string) error {
			capturedStatus = status
			return nil
		},
	}
	if err := repo.ModerateComment(context.Background(), "c1", "rejected"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStatus != "rejected" {
		t.Errorf("status: got %q, want %q", capturedStatus, "rejected")
	}
}