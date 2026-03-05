package repositories_test

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"
)

// ─── Sentinel error identity ──────────────────────────────────────────────────

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrUserNotFound", repositories.ErrUserNotFound},
		{"ErrProjectNotFound", repositories.ErrProjectNotFound},
		{"ErrProjectUnauthorized", repositories.ErrProjectUnauthorized},
		{"ErrImageNotFound", repositories.ErrImageNotFound},
		{"ErrPageNotFound", repositories.ErrPageNotFound},
		{"ErrPageUnauthorized", repositories.ErrPageUnauthorized},
		{"ErrSnapshotNotFound", repositories.ErrSnapshotNotFound},
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i == j {
				continue
			}
			if errors.Is(a.err, b.err) {
				t.Errorf("%s and %s should be distinct sentinel errors", a.name, b.name)
			}
		}
	}
}

func TestSentinelErrors_ErrorsIsWorksForEachSentinel(t *testing.T) {
	cases := []struct {
		name     string
		sentinel error
	}{
		{"ErrUserNotFound", repositories.ErrUserNotFound},
		{"ErrProjectNotFound", repositories.ErrProjectNotFound},
		{"ErrProjectUnauthorized", repositories.ErrProjectUnauthorized},
		{"ErrImageNotFound", repositories.ErrImageNotFound},
		{"ErrPageNotFound", repositories.ErrPageNotFound},
		{"ErrPageUnauthorized", repositories.ErrPageUnauthorized},
		{"ErrSnapshotNotFound", repositories.ErrSnapshotNotFound},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// errors.Is on the sentinel itself must match.
			if !errors.Is(tc.sentinel, tc.sentinel) {
				t.Errorf("%s: errors.Is(sentinel, sentinel) returned false", tc.name)
			}

			// A wrapped version must also match.
			wrapped := errors.Join(errors.New("outer"), tc.sentinel)
			if !errors.Is(wrapped, tc.sentinel) {
				t.Errorf("%s: errors.Is(wrapped, sentinel) returned false", tc.name)
			}
		})
	}
}

func TestSentinelErrors_NonNilAndHaveMessage(t *testing.T) {
	sentinels := []error{
		repositories.ErrUserNotFound,
		repositories.ErrProjectNotFound,
		repositories.ErrProjectUnauthorized,
		repositories.ErrImageNotFound,
		repositories.ErrPageNotFound,
		repositories.ErrPageUnauthorized,
		repositories.ErrSnapshotNotFound,
	}

	for _, s := range sentinels {
		if s == nil {
			t.Errorf("sentinel must not be nil")
		}
		if s.Error() == "" {
			t.Errorf("sentinel %T must have a non-empty message", s)
		}
	}
}

// ─── MockUserRepository ───────────────────────────────────────────────────────

func TestMockUserRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockUserRepository{}
	ctx := context.Background()

	_, err := repo.GetUserByID(ctx, "uid-1")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("GetUserByID default: want ErrUserNotFound, got %v", err)
	}

	_, err = repo.GetUserByEmail(ctx, "test@example.com")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("GetUserByEmail default: want ErrUserNotFound, got %v", err)
	}

	_, err = repo.GetUserByUsername(ctx, "someuser")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("GetUserByUsername default: want ErrUserNotFound, got %v", err)
	}
}

func TestMockUserRepository_FnOverridesDefault(t *testing.T) {
	want := &models.User{Id: "uid-1", Email: "a@b.com"}
	repo := &testutil.MockUserRepository{
		GetUserByIDFn: func(_ context.Context, userID string) (*models.User, error) {
			if userID == "uid-1" {
				return want, nil
			}
			return nil, repositories.ErrUserNotFound
		},
	}

	got, err := repo.GetUserByID(context.Background(), "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Id: got %q, want %q", got.Id, want.Id)
	}

	_, err = repo.GetUserByID(context.Background(), "uid-unknown")
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("unknown ID: want ErrUserNotFound, got %v", err)
	}
}

func TestMockUserRepository_SearchUsersDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockUserRepository{}
	users, err := repo.SearchUsers(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty slice, got %d users", len(users))
	}
}

func TestMockUserRepository_SearchUsersFnReturnsResults(t *testing.T) {
	want := []models.User{{Id: "u1"}, {Id: "u2"}}
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, _ string) ([]models.User, error) {
			return want, nil
		},
	}

	got, err := repo.SearchUsers(context.Background(), "ali")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("expected %d users, got %d", len(want), len(got))
	}
}

// ─── MockProjectRepository ────────────────────────────────────────────────────

func TestMockProjectRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	ctx := context.Background()

	_, err := repo.GetProjectByID(ctx, "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetProjectByID default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.GetProjectWithAccess(ctx, "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetProjectWithAccess default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.GetPublicProjectByID(ctx, "p1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetPublicProjectByID default: want ErrProjectNotFound, got %v", err)
	}

	_, err = repo.GetProjectWithLock(ctx, "p1", "u1")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("GetProjectWithLock default: want ErrProjectNotFound, got %v", err)
	}
}

func TestMockProjectRepository_CreateProjectFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockProjectRepository{
		CreateProjectFn: func(_ context.Context, p *models.Project) error {
			called = true
			p.ID = "generated-id"
			return nil
		},
	}

	proj := &models.Project{Name: "Test", OwnerId: "u1"}
	if err := repo.CreateProject(context.Background(), proj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreateProjectFn was not called")
	}
	if proj.ID != "generated-id" {
		t.Errorf("ID: got %q, want %q", proj.ID, "generated-id")
	}
}

func TestMockProjectRepository_UpdateProjectFnReturnsProject(t *testing.T) {
	want := &models.Project{ID: "p1", Name: "Updated"}
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, projectID, _ string, _ map[string]any) (*models.Project, error) {
			if projectID == "p1" {
				return want, nil
			}
			return nil, repositories.ErrProjectNotFound
		},
	}

	got, err := repo.UpdateProject(context.Background(), "p1", "u1", map[string]any{"Name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}

	_, err = repo.UpdateProject(context.Background(), "p-missing", "u1", map[string]any{})
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Errorf("missing project: want ErrProjectNotFound, got %v", err)
	}
}

func TestMockProjectRepository_ExistsForUserReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	exists, err := repo.ExistsForUser(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false by default, got true")
	}
}

func TestMockProjectRepository_GetProjectsByUserIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockProjectRepository{}
	projects, err := repo.GetProjectsByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected empty slice, got %d", len(projects))
	}
}

// ─── MockImageRepository ──────────────────────────────────────────────────────

func TestMockImageRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	ctx := context.Background()

	_, err := repo.GetImageByID(ctx, "img-1", "u1")
	if !errors.Is(err, repositories.ErrImageNotFound) {
		t.Errorf("GetImageByID default: want ErrImageNotFound, got %v", err)
	}
}

func TestMockImageRepository_CreateImageDefaultReturnsSameImage(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	input := models.Image{ImageId: "img-1", UserId: "u1", ImageLink: "https://example.com/img.png"}

	got, err := repo.CreateImage(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ImageId != input.ImageId {
		t.Errorf("ImageId: got %q, want %q", got.ImageId, input.ImageId)
	}
}

func TestMockImageRepository_GetImagesByUserIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	images, err := repo.GetImagesByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected empty slice, got %d", len(images))
	}
}

func TestMockImageRepository_GetImagesByUserIDFnFilters(t *testing.T) {
	all := []models.Image{
		{ImageId: "img-1", UserId: "u1"},
		{ImageId: "img-2", UserId: "u2"},
		{ImageId: "img-3", UserId: "u1"},
	}
	repo := &testutil.MockImageRepository{
		GetImagesByUserIDFn: func(_ context.Context, userID string) ([]models.Image, error) {
			var out []models.Image
			for _, img := range all {
				if img.UserId == userID {
					out = append(out, img)
				}
			}
			return out, nil
		},
	}

	images, err := repo.GetImagesByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 2 {
		t.Errorf("expected 2 images for u1, got %d", len(images))
	}
}

func TestMockImageRepository_SoftDeleteImageFnCalled(t *testing.T) {
	called := false
	repo := &testutil.MockImageRepository{
		SoftDeleteImageFn: func(_ context.Context, imageID, userID string) error {
			called = true
			if imageID == "" || userID == "" {
				return errors.New("imageID and userID are required")
			}
			return nil
		},
	}

	if err := repo.SoftDeleteImage(context.Background(), "img-1", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("SoftDeleteImageFn was not called")
	}
}

func TestMockImageRepository_GetAllImagesDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockImageRepository{}
	images, err := repo.GetAllImages(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected empty slice, got %d", len(images))
	}
}

// ─── MockPageRepository ───────────────────────────────────────────────────────

func TestMockPageRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	ctx := context.Background()

	_, err := repo.GetPageByID(ctx, "pg-1", "p1")
	if !errors.Is(err, repositories.ErrPageNotFound) {
		t.Errorf("GetPageByID default: want ErrPageNotFound, got %v", err)
	}
}

func TestMockPageRepository_CreatePageFnIsInvoked(t *testing.T) {
	called := false
	repo := &testutil.MockPageRepository{
		CreatePageFn: func(_ context.Context, page *models.Page) error {
			called = true
			page.Id = "pg-created"
			return nil
		},
	}

	page := &models.Page{Name: "Home", ProjectId: "p1", Type: "page"}
	if err := repo.CreatePage(context.Background(), page); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CreatePageFn was not called")
	}
	if page.Id != "pg-created" {
		t.Errorf("Id: got %q, want %q", page.Id, "pg-created")
	}
}

func TestMockPageRepository_GetPagesByProjectIDDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockPageRepository{}
	pages, err := repo.GetPagesByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 0 {
		t.Errorf("expected empty slice, got %d", len(pages))
	}
}

// ─── MockSnapshotRepository ───────────────────────────────────────────────────

func TestMockSnapshotRepository_DefaultsReturnSentinel(t *testing.T) {
	repo := &testutil.MockSnapshotRepository{}
	ctx := context.Background()

	_, err := repo.GetSnapshotByID(ctx, "snap-1")
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

// ─── MockCollaboratorRepository ───────────────────────────────────────────────

func TestMockCollaboratorRepository_IsCollaboratorReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockCollaboratorRepository{}
	ok, err := repo.IsCollaborator(context.Background(), "p1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false by default")
	}
}

func TestMockCollaboratorRepository_IsCollaboratorFnOverride(t *testing.T) {
	collabs := map[string]bool{"p1:u1": true, "p1:u2": true}
	repo := &testutil.MockCollaboratorRepository{
		IsCollaboratorFn: func(_ context.Context, projectID, userID string) (bool, error) {
			return collabs[projectID+":"+userID], nil
		},
	}

	for _, tc := range []struct {
		project, user string
		want          bool
	}{
		{"p1", "u1", true},
		{"p1", "u2", true},
		{"p1", "u3", false},
		{"p2", "u1", false},
	} {
		got, err := repo.IsCollaborator(context.Background(), tc.project, tc.user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tc.want {
			t.Errorf("IsCollaborator(%q, %q): got %v, want %v", tc.project, tc.user, got, tc.want)
		}
	}
}

func TestMockCollaboratorRepository_UpdateCollaboratorRoleFnCalled(t *testing.T) {
	var capturedRole models.CollaboratorRole
	repo := &testutil.MockCollaboratorRepository{
		UpdateCollaboratorRoleFn: func(_ context.Context, id string, role models.CollaboratorRole) error {
			capturedRole = role
			return nil
		},
	}

	if err := repo.UpdateCollaboratorRole(context.Background(), "collab-1", models.RoleViewer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRole != models.RoleViewer {
		t.Errorf("role: got %q, want %q", capturedRole, models.RoleViewer)
	}
}

// ─── MockEventWorkflowRepository ─────────────────────────────────────────────

func TestMockEventWorkflowRepository_CountDefaultsToZero(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	n, err := repo.CountEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestMockEventWorkflowRepository_CheckNameExistsReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{}
	exists, err := repo.CheckIfWorkflowNameExists(context.Background(), "p1", "My Workflow", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false by default")
	}
}

func TestMockEventWorkflowRepository_CreateFnAssignsID(t *testing.T) {
	repo := &testutil.MockEventWorkflowRepository{
		CreateEventWorkflowFn: func(_ context.Context, wf *models.EventWorkflow) (*models.EventWorkflow, error) {
			wf.Id = "wf-generated"
			return wf, nil
		},
	}

	wf := &models.EventWorkflow{Name: "On Click", ProjectId: "p1"}
	got, err := repo.CreateEventWorkflow(context.Background(), wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "wf-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "wf-generated")
	}
}

func TestMockEventWorkflowRepository_GetEnabledFiltersByEnabled(t *testing.T) {
	all := []models.EventWorkflow{
		{Id: "wf-1", ProjectId: "p1", Enabled: true},
		{Id: "wf-2", ProjectId: "p1", Enabled: false},
		{Id: "wf-3", ProjectId: "p1", Enabled: true},
	}
	repo := &testutil.MockEventWorkflowRepository{
		GetEnabledEventWorkflowsByProjectIDFn: func(_ context.Context, projectID string) ([]models.EventWorkflow, error) {
			var out []models.EventWorkflow
			for _, wf := range all {
				if wf.ProjectId == projectID && wf.Enabled {
					out = append(out, wf)
				}
			}
			return out, nil
		},
	}

	got, err := repo.GetEnabledEventWorkflowsByProjectID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 enabled workflows, got %d", len(got))
	}
	for _, wf := range got {
		if !wf.Enabled {
			t.Errorf("workflow %q should be enabled", wf.Id)
		}
	}
}

// ─── MockElementEventWorkflowRepository ──────────────────────────────────────

func TestMockElementEventWorkflowRepository_GetAllDefaultsToEmptySlice(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	eews, err := repo.GetAllElementEventWorkflows(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eews) != 0 {
		t.Errorf("expected empty slice, got %d", len(eews))
	}
}

func TestMockElementEventWorkflowRepository_CheckLinkedReturnsFalseByDefault(t *testing.T) {
	repo := &testutil.MockElementEventWorkflowRepository{}
	linked, err := repo.CheckIfWorkflowLinkedToElement(context.Background(), "el-1", "wf-1", "onClick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if linked {
		t.Error("expected false by default")
	}
}

func TestMockElementEventWorkflowRepository_DeleteByElementIDFnCalled(t *testing.T) {
	var capturedElementID string
	repo := &testutil.MockElementEventWorkflowRepository{
		DeleteElementEventWorkflowsByElementIDFn: func(_ context.Context, elementID string) error {
			capturedElementID = elementID
			return nil
		},
	}

	if err := repo.DeleteElementEventWorkflowsByElementID(context.Background(), "el-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedElementID != "el-42" {
		t.Errorf("elementID: got %q, want %q", capturedElementID, "el-42")
	}
}

func TestMockElementEventWorkflowRepository_GetByPageIDFnFilters(t *testing.T) {
	all := []models.ElementEventWorkflow{
		{Id: "eew-1", ElementId: "el-a"},
		{Id: "eew-2", ElementId: "el-b"},
	}
	repo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByPageIDFn: func(_ context.Context, _ string) ([]models.ElementEventWorkflow, error) {
			return all, nil
		},
	}

	got, err := repo.GetElementEventWorkflowsByPageID(context.Background(), "pg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(all) {
		t.Errorf("expected %d, got %d", len(all), len(got))
	}
}

// ─── Interface satisfaction (compile-time) ────────────────────────────────────
//
// These assignments do nothing at runtime but cause a compile error if any mock
// stops satisfying its interface — giving instant feedback on missed methods.

// ─── additional sentinel errors added during repo refactor ───────────────────

func TestSentinelErrors_NewSentinelsExist(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"ErrCommentNotFound", repositories.ErrCommentNotFound},
		{"ErrCommentUnauthorized", repositories.ErrCommentUnauthorized},
		{"ErrInvitationNotFound", repositories.ErrInvitationNotFound},
		{"ErrInvitationExpired", repositories.ErrInvitationExpired},
		{"ErrInvitationInvalid", repositories.ErrInvitationInvalid},
		{"ErrContentTypeNotFound", repositories.ErrContentTypeNotFound},
		{"ErrContentFieldNotFound", repositories.ErrContentFieldNotFound},
		{"ErrContentItemNotFound", repositories.ErrContentItemNotFound},
		{"ErrCustomElementNotFound", repositories.ErrCustomElementNotFound},
		{"ErrCustomElementTypeNotFound", repositories.ErrCustomElementTypeNotFound},
		{"ErrElementNotFound", repositories.ErrElementNotFound},
		{"ErrSettingNotFound", repositories.ErrSettingNotFound},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Errorf("%s must not be nil", tc.name)
			}
			if tc.err.Error() == "" {
				t.Errorf("%s must have a non-empty message", tc.name)
			}
			if !errors.Is(tc.err, tc.err) {
				t.Errorf("%s: errors.Is(sentinel, sentinel) returned false", tc.name)
			}
		})
	}
}

func TestSentinelErrors_CommentSentinelsAreDistinct(t *testing.T) {
	if errors.Is(repositories.ErrCommentNotFound, repositories.ErrCommentUnauthorized) {
		t.Error("ErrCommentNotFound and ErrCommentUnauthorized must be distinct")
	}
}

func TestSentinelErrors_InvitationSentinelsAreDistinct(t *testing.T) {
	pairs := []struct{ a, b error }{
		{repositories.ErrInvitationNotFound, repositories.ErrInvitationExpired},
		{repositories.ErrInvitationNotFound, repositories.ErrInvitationInvalid},
		{repositories.ErrInvitationExpired, repositories.ErrInvitationInvalid},
	}
	for _, p := range pairs {
		if errors.Is(p.a, p.b) {
			t.Errorf("%v and %v should be distinct", p.a, p.b)
		}
	}
}

// ─── MockCommentRepository ────────────────────────────────────────────────────

type MockCommentRepository struct {
	CreateCommentFn           func(ctx context.Context, comment models.Comment) (*models.Comment, error)
	GetCommentByIDFn          func(ctx context.Context, id string) (*models.Comment, error)
	GetCommentsFn             func(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error)
	UpdateCommentFn           func(ctx context.Context, id, userID string, updates map[string]any) (*models.Comment, error)
	DeleteCommentFn           func(ctx context.Context, id, userID string) error
	CreateReactionFn          func(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error)
	DeleteReactionFn          func(ctx context.Context, commentID, userID, reactionType string) error
	GetReactionsByCommentIDFn func(ctx context.Context, commentID string) ([]models.CommentReaction, error)
	GetReactionSummaryFn      func(ctx context.Context, commentID string) ([]models.ReactionSummary, error)
	GetCommentCountByItemIDFn func(ctx context.Context, itemID string) (int64, error)
	ModerateCommentFn         func(ctx context.Context, id, status string) error
}

func (m *MockCommentRepository) CreateComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	if m.CreateCommentFn != nil {
		return m.CreateCommentFn(ctx, comment)
	}
	return &comment, nil
}
func (m *MockCommentRepository) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	if m.GetCommentByIDFn != nil {
		return m.GetCommentByIDFn(ctx, id)
	}
	return nil, repositories.ErrCommentNotFound
}
func (m *MockCommentRepository) GetComments(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error) {
	if m.GetCommentsFn != nil {
		return m.GetCommentsFn(ctx, filter)
	}
	return []models.Comment{}, 0, nil
}
func (m *MockCommentRepository) UpdateComment(ctx context.Context, id, userID string, updates map[string]any) (*models.Comment, error) {
	if m.UpdateCommentFn != nil {
		return m.UpdateCommentFn(ctx, id, userID, updates)
	}
	return nil, repositories.ErrCommentUnauthorized
}
func (m *MockCommentRepository) DeleteComment(ctx context.Context, id, userID string) error {
	if m.DeleteCommentFn != nil {
		return m.DeleteCommentFn(ctx, id, userID)
	}
	return nil
}
func (m *MockCommentRepository) CreateReaction(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error) {
	if m.CreateReactionFn != nil {
		return m.CreateReactionFn(ctx, reaction)
	}
	return &reaction, nil
}
func (m *MockCommentRepository) DeleteReaction(ctx context.Context, commentID, userID, reactionType string) error {
	if m.DeleteReactionFn != nil {
		return m.DeleteReactionFn(ctx, commentID, userID, reactionType)
	}
	return nil
}
func (m *MockCommentRepository) GetReactionsByCommentID(ctx context.Context, commentID string) ([]models.CommentReaction, error) {
	if m.GetReactionsByCommentIDFn != nil {
		return m.GetReactionsByCommentIDFn(ctx, commentID)
	}
	return []models.CommentReaction{}, nil
}
func (m *MockCommentRepository) GetReactionSummary(ctx context.Context, commentID string) ([]models.ReactionSummary, error) {
	if m.GetReactionSummaryFn != nil {
		return m.GetReactionSummaryFn(ctx, commentID)
	}
	return []models.ReactionSummary{}, nil
}
func (m *MockCommentRepository) GetCommentCountByItemID(ctx context.Context, itemID string) (int64, error) {
	if m.GetCommentCountByItemIDFn != nil {
		return m.GetCommentCountByItemIDFn(ctx, itemID)
	}
	return 0, nil
}
func (m *MockCommentRepository) ModerateComment(ctx context.Context, id, status string) error {
	if m.ModerateCommentFn != nil {
		return m.ModerateCommentFn(ctx, id, status)
	}
	return nil
}

var _ repositories.CommentRepositoryInterface = (*MockCommentRepository)(nil)

func TestMockCommentRepository_DefaultGetByIDReturnsSentinel(t *testing.T) {
	repo := &MockCommentRepository{}
	_, err := repo.GetCommentByID(context.Background(), "c1")
	if !errors.Is(err, repositories.ErrCommentNotFound) {
		t.Errorf("want ErrCommentNotFound, got %v", err)
	}
}

func TestMockCommentRepository_DefaultUpdateReturnsSentinel(t *testing.T) {
	repo := &MockCommentRepository{}
	_, err := repo.UpdateComment(context.Background(), "c1", "u1", map[string]any{"Content": "hi"})
	if !errors.Is(err, repositories.ErrCommentUnauthorized) {
		t.Errorf("want ErrCommentUnauthorized, got %v", err)
	}
}

func TestMockCommentRepository_CreateFnIsInvoked(t *testing.T) {
	called := false
	repo := &MockCommentRepository{
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
	repo := &MockCommentRepository{}
	comments, total, err := repo.GetComments(context.Background(), models.CommentFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 0 || total != 0 {
		t.Errorf("expected empty result, got %d comments, total=%d", len(comments), total)
	}
}

func TestMockCommentRepository_GetCommentsFnFilters(t *testing.T) {
	all := []models.Comment{
		{Id: "c1", ItemId: "item-1"},
		{Id: "c2", ItemId: "item-2"},
		{Id: "c3", ItemId: "item-1"},
	}
	repo := &MockCommentRepository{
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

func TestMockCommentRepository_ReactionSummaryFnReturnsData(t *testing.T) {
	want := []models.ReactionSummary{{Type: "like", Count: 5}, {Type: "heart", Count: 2}}
	repo := &MockCommentRepository{
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
}

// ─── MockInvitationRepository ─────────────────────────────────────────────────

type MockInvitationRepository struct {
	CreateInvitationFn               func(ctx context.Context, inv *models.Invitation) (*models.Invitation, error)
	GetInvitationsByProjectFn        func(ctx context.Context, projectID string) ([]models.Invitation, error)
	GetInvitationByIDFn              func(ctx context.Context, id string) (*models.Invitation, error)
	GetInvitationByTokenFn           func(ctx context.Context, token string) (*models.Invitation, error)
	AcceptInvitationFn               func(ctx context.Context, token, userID string) error
	DeleteInvitationFn               func(ctx context.Context, id string) error
	UpdateInvitationStatusFn         func(ctx context.Context, id string, status models.InvitationStatus) error
	CancelInvitationFn               func(ctx context.Context, id string) error
	GetPendingInvitationsByProjectFn func(ctx context.Context, projectID string) ([]models.Invitation, error)
}

func (m *MockInvitationRepository) CreateInvitation(ctx context.Context, inv *models.Invitation) (*models.Invitation, error) {
	if m.CreateInvitationFn != nil {
		return m.CreateInvitationFn(ctx, inv)
	}
	return inv, nil
}
func (m *MockInvitationRepository) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if m.GetInvitationsByProjectFn != nil {
		return m.GetInvitationsByProjectFn(ctx, projectID)
	}
	return []models.Invitation{}, nil
}
func (m *MockInvitationRepository) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	if m.GetInvitationByIDFn != nil {
		return m.GetInvitationByIDFn(ctx, id)
	}
	return nil, repositories.ErrInvitationNotFound
}
func (m *MockInvitationRepository) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	if m.GetInvitationByTokenFn != nil {
		return m.GetInvitationByTokenFn(ctx, token)
	}
	return nil, repositories.ErrInvitationNotFound
}
func (m *MockInvitationRepository) AcceptInvitation(ctx context.Context, token, userID string) error {
	if m.AcceptInvitationFn != nil {
		return m.AcceptInvitationFn(ctx, token, userID)
	}
	return nil
}
func (m *MockInvitationRepository) DeleteInvitation(ctx context.Context, id string) error {
	if m.DeleteInvitationFn != nil {
		return m.DeleteInvitationFn(ctx, id)
	}
	return nil
}
func (m *MockInvitationRepository) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	if m.UpdateInvitationStatusFn != nil {
		return m.UpdateInvitationStatusFn(ctx, id, status)
	}
	return nil
}
func (m *MockInvitationRepository) CancelInvitation(ctx context.Context, id string) error {
	if m.CancelInvitationFn != nil {
		return m.CancelInvitationFn(ctx, id)
	}
	return nil
}
func (m *MockInvitationRepository) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if m.GetPendingInvitationsByProjectFn != nil {
		return m.GetPendingInvitationsByProjectFn(ctx, projectID)
	}
	return []models.Invitation{}, nil
}

var _ repositories.InvitationRepositoryInterface = (*MockInvitationRepository)(nil)

func TestMockInvitationRepository_DefaultGetByIDReturnsSentinel(t *testing.T) {
	repo := &MockInvitationRepository{}
	_, err := repo.GetInvitationByID(context.Background(), "inv-1")
	if !errors.Is(err, repositories.ErrInvitationNotFound) {
		t.Errorf("want ErrInvitationNotFound, got %v", err)
	}
}

func TestMockInvitationRepository_DefaultGetByTokenReturnsSentinel(t *testing.T) {
	repo := &MockInvitationRepository{}
	_, err := repo.GetInvitationByToken(context.Background(), "tok-abc")
	if !errors.Is(err, repositories.ErrInvitationNotFound) {
		t.Errorf("want ErrInvitationNotFound, got %v", err)
	}
}

func TestMockInvitationRepository_CreateFnAssignsID(t *testing.T) {
	repo := &MockInvitationRepository{
		CreateInvitationFn: func(_ context.Context, inv *models.Invitation) (*models.Invitation, error) {
			inv.Id = "inv-generated"
			return inv, nil
		},
	}
	inv := &models.Invitation{Email: "a@b.com", ProjectId: "p1"}
	got, err := repo.CreateInvitation(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != "inv-generated" {
		t.Errorf("Id: got %q, want %q", got.Id, "inv-generated")
	}
}

func TestMockInvitationRepository_AcceptInvitationFnCalled(t *testing.T) {
	called := false
	repo := &MockInvitationRepository{
		AcceptInvitationFn: func(_ context.Context, token, userID string) error {
			called = true
			return nil
		},
	}
	if err := repo.AcceptInvitation(context.Background(), "tok-x", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("AcceptInvitationFn was not called")
	}
}

func TestMockInvitationRepository_GetPendingDefaultsToEmpty(t *testing.T) {
	repo := &MockInvitationRepository{}
	invs, err := repo.GetPendingInvitationsByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(invs) != 0 {
		t.Errorf("expected empty slice, got %d", len(invs))
	}
}

// ─── Compile-time interface satisfaction checks ───────────────────────────────

var (
	_ repositories.UserRepositoryInterface                 = (*testutil.MockUserRepository)(nil)
	_ repositories.ProjectRepositoryInterface              = (*testutil.MockProjectRepository)(nil)
	_ repositories.ImageRepositoryInterface                = (*testutil.MockImageRepository)(nil)
	_ repositories.PageRepositoryInterface                 = (*testutil.MockPageRepository)(nil)
	_ repositories.SnapshotRepositoryInterface             = (*testutil.MockSnapshotRepository)(nil)
	_ repositories.CollaboratorRepositoryInterface         = (*testutil.MockCollaboratorRepository)(nil)
	_ repositories.EventWorkflowRepositoryInterface        = (*testutil.MockEventWorkflowRepository)(nil)
	_ repositories.ElementEventWorkflowRepositoryInterface = (*testutil.MockElementEventWorkflowRepository)(nil)
	_ repositories.CommentRepositoryInterface              = (*MockCommentRepository)(nil)
	_ repositories.InvitationRepositoryInterface           = (*MockInvitationRepository)(nil)
)
