package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type visibleProjectCase struct {
	name       string
	userID     string
	access     func(context.Context, string, string) (*models.Project, error)
	public     func(context.Context, string) (*models.Project, error)
	wantErr    error
	wantAccess bool
	wantPublic bool
}

func runVisibleProjectCases(t *testing.T, tests []visibleProjectCase) {
	t.Helper()
	ctx := context.Background()
	for _, tc := range tests {
		repo := &fkProjectRepo{OnGetWithAccess: tc.access, OnGetPublicByID: tc.public}
		err := ensureProjectVisible(ctx, repo, "p1", tc.userID)
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s: err = %v, want %v", tc.name, err, tc.wantErr)
		}
		if tc.wantErr == nil && err != nil {
			t.Fatalf("%s: unexpected err = %v", tc.name, err)
		}
		if repo.Called("GetWithAccess") != tc.wantAccess {
			t.Fatalf("%s: access called = %v, want %v", tc.name, repo.Called("GetWithAccess"), tc.wantAccess)
		}
		if repo.Called("GetPublicByID") != tc.wantPublic {
			t.Fatalf("%s: public called = %v, want %v", tc.name, repo.Called("GetPublicByID"), tc.wantPublic)
		}
	}
}

func TestEnsureProjectVisibleMembersAndErrors(t *testing.T) {
	project := &models.Project{ID: "p1"}
	runVisibleProjectCases(t, []visibleProjectCase{
		{
			name:       "member passes without public lookup",
			userID:     "u1",
			access:     func(context.Context, string, string) (*models.Project, error) { return project, nil },
			wantAccess: true,
		},
		{
			name:   "not found falls back to published project",
			userID: "u1",
			access: func(context.Context, string, string) (*models.Project, error) {
				return nil, repositories.ErrProjectNotFound
			},
			public:     func(context.Context, string) (*models.Project, error) { return project, nil },
			wantAccess: true, wantPublic: true,
		},
		{
			name:   "unauthorized and unpublished hides project",
			userID: "u1",
			access: func(context.Context, string, string) (*models.Project, error) {
				return nil, repositories.ErrProjectUnauthorized
			},
			public:     func(context.Context, string) (*models.Project, error) { return nil, nil },
			wantErr:    repositories.ErrProjectNotFound,
			wantAccess: true, wantPublic: true,
		},
		{
			name:       "nil project without error falls back to public",
			userID:     "u1",
			access:     func(context.Context, string, string) (*models.Project, error) { return nil, nil },
			public:     func(context.Context, string) (*models.Project, error) { return project, nil },
			wantAccess: true, wantPublic: true,
		},
		{
			name:       "unexpected access error is returned and public is not consulted",
			userID:     "u1",
			access:     func(context.Context, string, string) (*models.Project, error) { return nil, errBoom },
			wantErr:    errBoom,
			wantAccess: true,
		},
	})
}

func TestEnsureProjectVisibleAnonymous(t *testing.T) {
	project := &models.Project{ID: "p1"}
	runVisibleProjectCases(t, []visibleProjectCase{
		{
			name:       "anonymous user skips access check for published project",
			public:     func(context.Context, string) (*models.Project, error) { return project, nil },
			wantPublic: true,
		},
		{
			name:       "anonymous user cannot see unpublished project",
			public:     func(context.Context, string) (*models.Project, error) { return nil, nil },
			wantErr:    repositories.ErrProjectNotFound,
			wantPublic: true,
		},
		{
			name:       "public lookup error is returned",
			public:     func(context.Context, string) (*models.Project, error) { return nil, errBoom },
			wantErr:    errBoom,
			wantPublic: true,
		},
	})
}

func assertProjectGetter(
	t *testing.T,
	call func(*ProjectService, string, string) (*models.Project, error),
	hook func(*fkProjectRepo, func(context.Context, string, string) (*models.Project, error)),
) {
	t.Helper()
	emptyRepo := &fkProjectRepo{}
	svc := NewProjectService(emptyRepo, nil, nil)
	_, err := call(svc, "", "u1")
	wantErr(t, err, "projectId is required")
	_, err = call(svc, "p1", "")
	wantErr(t, err, "userId is required")
	wantCalls(t, emptyRepo.Calls)

	repo := &fkProjectRepo{}
	hook(repo, func(_ context.Context, p, u string) (*models.Project, error) {
		return &models.Project{ID: p, OwnerId: u}, nil
	})
	got, err := call(NewProjectService(repo, nil, nil), "p1", "u1")
	wantNoErr(t, err)
	if got.ID != "p1" || got.OwnerId != "u1" {
		t.Fatalf("got %+v", got)
	}
	hook(repo, func(context.Context, string, string) (*models.Project, error) { return nil, errBoom })
	_, err = call(NewProjectService(repo, nil, nil), "p1", "u1")
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestProjectServiceGetters(t *testing.T) {
	ctx := context.Background()
	repo := &fkProjectRepo{OnGetPublicByID: func(_ context.Context, id string) (*models.Project, error) {
		return &models.Project{ID: id}, nil
	}}
	p, err := NewProjectService(repo, nil, nil).GetPublicByID(ctx, "p1")
	wantNoErr(t, err)
	if p.ID != "p1" {
		t.Fatalf("id = %q", p.ID)
	}

	assertProjectGetter(t,
		func(s *ProjectService, p, u string) (*models.Project, error) { return s.GetByID(ctx, p, u) },
		func(r *fkProjectRepo, f func(context.Context, string, string) (*models.Project, error)) {
			r.OnGetByID = f
		})
	assertProjectGetter(t,
		func(s *ProjectService, p, u string) (*models.Project, error) { return s.GetWithAccess(ctx, p, u) },
		func(r *fkProjectRepo, f func(context.Context, string, string) (*models.Project, error)) {
			r.OnGetWithAccess = f
		})
	assertProjectGetter(t,
		func(s *ProjectService, p, u string) (*models.Project, error) { return s.GetWithLock(ctx, p, u) },
		func(r *fkProjectRepo, f func(context.Context, string, string) (*models.Project, error)) {
			r.OnGetWithLock = f
		})

	nilRepo := &fkProjectRepo{
		OnGetByID:       func(context.Context, string, string) (*models.Project, error) { return nil, nil },
		OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) { return nil, nil },
	}
	svc := NewProjectService(nilRepo, nil, nil)
	_, err = svc.GetByID(ctx, "p1", "u1")
	wantErr(t, err, "project does not exist")
	_, err = svc.GetWithAccess(ctx, "p1", "u1")
	wantErr(t, err, "project does not exist")
}

func TestProjectServiceListings(t *testing.T) {
	ctx := context.Background()
	repo := &fkProjectRepo{
		OnListByUser: func(_ context.Context, u string) ([]models.Project, error) {
			return []models.Project{{ID: "mine", OwnerId: u}}, nil
		},
		OnListByCollaborator: func(_ context.Context, u string) ([]models.Project, error) {
			return []models.Project{{ID: "shared"}}, nil
		},
	}
	svc := NewProjectService(repo, nil, nil)

	got, err := svc.ListByUser(ctx, "u1")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].ID != "mine" {
		t.Fatalf("got %+v", got)
	}
	got, err = svc.ListByCollaborator(ctx, "u1")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].ID != "shared" {
		t.Fatalf("got %+v", got)
	}

	_, err = svc.ListByUser(ctx, "")
	wantErr(t, err, "userId is required")
	_, err = svc.ListByCollaborator(ctx, "")
	wantErr(t, err, "userId is required")
	wantCalls(t, repo.Calls, "ListByUser", "ListByCollaborator")
}

func TestProjectServiceGetProjectPages(t *testing.T) {
	ctx := context.Background()

	t.Run("requires projectId", func(t *testing.T) {
		_, err := NewProjectService(&fkProjectRepo{}, nil, nil).ListPages(ctx, "", "u1")
		wantErr(t, err, "projectId is required")
	})
	t.Run("invisible project is wrapped and pages are not read", func(t *testing.T) {
		repo := &fkProjectRepo{
			OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) {
				return nil, repositories.ErrProjectNotFound
			},
			OnGetPublicByID: func(context.Context, string) (*models.Project, error) { return nil, nil },
		}
		_, err := NewProjectService(repo, nil, nil).ListPages(ctx, "p1", "u1")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		wantErr(t, err, "failed to verify project")
		if repo.Called("ListPages") {
			t.Fatal("pages read for invisible project")
		}
	})
	t.Run("visible project returns pages", func(t *testing.T) {
		repo := &fkProjectRepo{
			OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) {
				return &models.Project{ID: "p1"}, nil
			},
			OnListPages: func(_ context.Context, p, u string) ([]models.Page, error) {
				return []models.Page{{Id: "pg", ProjectId: p}}, nil
			},
		}
		pages, err := NewProjectService(repo, nil, nil).ListPages(ctx, "p1", "u1")
		wantNoErr(t, err)
		if len(pages) != 1 || pages[0].ProjectId != "p1" {
			t.Fatalf("pages = %+v", pages)
		}
	})
}

func TestProjectServiceCreateErrorPaths(t *testing.T) {
	ctx := context.Background()
	repo := &fkProjectRepo{OnCreate: func(context.Context, *models.Project) error { return errBoom }}
	got, err := NewProjectService(repo, nil, nil).Create(ctx, &models.Project{Name: "n", OwnerId: "o"})
	if !errors.Is(err, errBoom) || got != nil {
		t.Fatalf("got %v, err %v", got, err)
	}

	failRepo := &fkProjectRepo{OnCreateWithPage: func(context.Context, *models.Project, *models.Page) error { return errBoom }}
	got, err = NewProjectService(failRepo, nil, nil).CreateWithPage(ctx, &models.Project{Name: "n", OwnerId: "o"})
	if !errors.Is(err, errBoom) || got != nil {
		t.Fatalf("got %v, err %v", got, err)
	}
}

func TestProjectServiceCreateWithPageValidationAndEmptyPage(t *testing.T) {
	ctx := context.Background()
	for name, p := range map[string]*models.Project{
		"nil":           nil,
		"missing name":  {OwnerId: "o"},
		"missing owner": {Name: "n"},
	} {
		repo := &fkProjectRepo{}
		_, err := NewProjectService(repo, nil, nil).CreateWithPage(ctx, p)
		wantErr(t, err, "project name and ownerId are required")
		if len(repo.Calls) != 0 {
			t.Fatalf("%s: repo called", name)
		}
	}

	var gotPage *models.Page
	repo := &fkProjectRepo{OnCreateWithPage: func(_ context.Context, _ *models.Project, pg *models.Page) error {
		gotPage = pg
		return nil
	}}
	_, err := NewProjectService(repo, nil, nil).CreateWithPage(ctx, &models.Project{Name: "n", OwnerId: "o"})
	wantNoErr(t, err)
	if gotPage.Type != "page" || gotPage.Name != "" || string(gotPage.Styles) != "{}" {
		t.Fatalf("page = %+v", gotPage)
	}
}

func TestProjectServiceUpdateProjectValidationAndAccessDenied(t *testing.T) {
	ctx := context.Background()
	svc := NewProjectService(&fkProjectRepo{}, nil, nil)
	_, err := svc.Update(ctx, "", "u", &models.Project{})
	wantErr(t, err, "projectId is required")
	_, err = svc.Update(ctx, "p", "", &models.Project{})
	wantErr(t, err, "userId is required")
	_, err = svc.Update(ctx, "p", "u", nil)
	wantErr(t, err, "project cannot be nil")

	deniedRepo := &fkProjectRepo{OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) {
		return nil, repositories.ErrProjectUnauthorized
	}}
	_, err = NewProjectService(deniedRepo, nil, nil).Update(ctx, "p1", "u1", &models.Project{Name: "x"})
	if !errors.Is(err, repositories.ErrProjectUnauthorized) {
		t.Fatalf("err = %v", err)
	}
	if deniedRepo.Called("Update") {
		t.Fatal("Update called without access")
	}
}

func TestProjectServiceUpdateProjectFieldInheritanceAndOverride(t *testing.T) {
	ctx := context.Background()
	desc, sub := strp("old description"), strp("old-sub")
	styles, header := rawp(`{"a":1}`), rawp(`{"h":1}`)
	existing := &models.Project{ID: "p1", Name: "old", Description: desc, Subdomain: sub, Styles: styles, Header: header}

	var got map[string]any
	captureRepo := &fkProjectRepo{
		OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) { return existing, nil },
		OnUpdate: func(_ context.Context, _, _ string, u map[string]any) (*models.Project, error) {
			got = u
			return &models.Project{ID: "p1"}, nil
		},
	}
	_, err := NewProjectService(captureRepo, nil, nil).Update(ctx, "p1", "u1", &models.Project{Published: true})
	wantNoErr(t, err)
	if got["Name"] != "old" || got["Description"] != desc || got["Subdomain"] != sub {
		t.Fatalf("updates = %+v", got)
	}
	if got["Styles"] != styles || got["Header"] != header || got["Published"] != true {
		t.Fatalf("updates = %+v", got)
	}

	newDesc := strp("new")
	_, err = NewProjectService(captureRepo, nil, nil).Update(ctx, "p1", "u1", &models.Project{Name: "fresh", Description: newDesc})
	wantNoErr(t, err)
	if got["Name"] != "fresh" || got["Description"] != newDesc {
		t.Fatalf("updates = %+v", got)
	}
}

func TestProjectServiceAccessGuardedMutations(t *testing.T) {
	ctx := context.Background()
	type op struct {
		name string
		repo string
		run  func(*ProjectService, string, string) error
	}
	ops := []op{
		{"Delete", "Delete", func(s *ProjectService, p, u string) error { return s.Delete(ctx, p, u) }},
		{"HardDelete", "HardDelete", func(s *ProjectService, p, u string) error { return s.HardDelete(ctx, p, u) }},
		{"Restore", "Restore", func(s *ProjectService, p, u string) error { return s.Restore(ctx, p, u) }},
	}
	for _, o := range ops {
		newRepo := func(accessErr, opErr error) *fkProjectRepo {
			return &fkProjectRepo{
				OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) {
					if accessErr != nil {
						return nil, accessErr
					}
					return &models.Project{ID: "p1"}, nil
				},
				OnDelete:     func(context.Context, string, string) error { return opErr },
				OnHardDelete: func(context.Context, string, string) error { return opErr },
				OnRestore:    func(context.Context, string, string) error { return opErr },
			}
		}
		t.Run(o.name+" validation", func(t *testing.T) {
			svc := NewProjectService(newRepo(nil, nil), nil, nil)
			wantErr(t, o.run(svc, "", "u"), "projectId is required")
			wantErr(t, o.run(svc, "p", ""), "userId is required")
		})
		t.Run(o.name+" denied", func(t *testing.T) {
			repo := newRepo(repositories.ErrProjectUnauthorized, nil)
			err := o.run(NewProjectService(repo, nil, nil), "p1", "u1")
			if !errors.Is(err, repositories.ErrProjectUnauthorized) {
				t.Fatalf("err = %v", err)
			}
			if repo.Called(o.repo) {
				t.Fatalf("%s reached repository without access", o.name)
			}
		})
		t.Run(o.name+" success", func(t *testing.T) {
			repo := newRepo(nil, nil)
			wantNoErr(t, o.run(NewProjectService(repo, nil, nil), "p1", "u1"))
			wantCalls(t, repo.Calls, "GetWithAccess", o.repo)
		})
		t.Run(o.name+" repo error", func(t *testing.T) {
			err := o.run(NewProjectService(newRepo(nil, errBoom), nil, nil), "p1", "u1")
			if !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestProjectServiceExistsForUserAndAccess(t *testing.T) {
	ctx := context.Background()
	repo := &fkProjectRepo{
		OnExistsForUser: func(_ context.Context, p, u string) (bool, error) { return p == "p1" && u == "u1", nil },
		OnGetAccess: func(_ context.Context, p, u string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: p, UserId: u, Role: models.RoleViewer}, nil
		},
	}
	svc := NewProjectService(repo, nil, nil)

	_, err := svc.ExistsForUser(ctx, "", "u1")
	wantErr(t, err, "projectId is required")
	_, err = svc.ExistsForUser(ctx, "p1", "")
	wantErr(t, err, "userId is required")
	ok, err := svc.ExistsForUser(ctx, "p1", "u1")
	wantNoErr(t, err)
	if !ok {
		t.Fatal("expected project to exist")
	}
	ok, err = svc.ExistsForUser(ctx, "p2", "u1")
	wantNoErr(t, err)
	if ok {
		t.Fatal("unexpected existence")
	}

	access, err := svc.GetAccess(ctx, "p1", "u1")
	wantNoErr(t, err)
	if access.Role != models.RoleViewer || access.ProjectId != "p1" {
		t.Fatalf("access = %+v", access)
	}
}

// ---- PageService ----

func newPageSvc(pages *fkPageRepo, projects *fkProjectRepo) *PageService {
	return NewPageService(pages, projects)
}

func publishedProject() *fkProjectRepo {
	return &fkProjectRepo{
		OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
		OnGetPublicByID: func(_ context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Published: true}, nil
		},
	}
}

func TestPageServiceGetPagesByProjectID(t *testing.T) {
	ctx := context.Background()
	all := []models.Page{{Id: "a", Type: "page"}, {Id: "b", Type: models.PageTypeComponent}, {Id: "c", Type: "page"}}
	pages := func() *fkPageRepo {
		return &fkPageRepo{OnListByProject: func(context.Context, string) ([]models.Page, error) { return all, nil }}
	}

	t.Run("requires projectId", func(t *testing.T) {
		_, err := newPageSvc(pages(), &fkProjectRepo{}).ListByProject(ctx, "", "u")
		wantErr(t, err, "projectId is required")
	})
	t.Run("anonymous callers only get routable pages", func(t *testing.T) {
		got, err := newPageSvc(pages(), publishedProject()).ListByProject(ctx, "p1", "")
		wantNoErr(t, err)
		if len(got) != 2 || got[0].Id != "a" || got[1].Id != "c" {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("signed-in members also get component pages", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return &models.Project{ID: "p1"}, nil
		}}
		got, err := newPageSvc(pages(), projects).ListByProject(ctx, "p1", "u1")
		wantNoErr(t, err)
		if len(got) != 3 {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("unpublished project is hidden", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetPublicByID: func(context.Context, string) (*models.Project, error) { return nil, nil }}
		pr := pages()
		_, err := newPageSvc(pr, projects).ListByProject(ctx, "p1", "")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		if len(pr.Calls) != 0 {
			t.Fatal("pages read for hidden project")
		}
	})
	t.Run("repository error is returned", func(t *testing.T) {
		pr := &fkPageRepo{OnListByProject: func(context.Context, string) ([]models.Page, error) { return nil, errBoom }}
		_, err := newPageSvc(pr, publishedProject()).ListByProject(ctx, "p1", "")
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestPageServiceGetPageByID(t *testing.T) {
	ctx := context.Background()
	var gotProject string
	pr := &fkPageRepo{OnGetByID: func(_ context.Context, id, project string) (*models.Page, error) {
		gotProject = project
		return &models.Page{Id: id}, nil
	}}
	svc := newPageSvc(pr, nil)
	_, err := svc.GetByID(ctx, "")
	wantErr(t, err, "pageId is required")
	p, err := svc.GetByID(ctx, "pg")
	wantNoErr(t, err)
	if p.Id != "pg" || gotProject != "" {
		t.Fatalf("page %+v project filter %q", p, gotProject)
	}
}

func TestPageServiceGetPublicPageByID(t *testing.T) {
	ctx := context.Background()
	pubProject := func(p *models.Project, err error) *fkProjectRepo {
		return &fkProjectRepo{OnGetPublicByID: func(context.Context, string) (*models.Project, error) { return p, err }}
	}
	pageRepo := func(p *models.Page, err error) *fkPageRepo {
		return &fkPageRepo{OnGetByID: func(context.Context, string, string) (*models.Page, error) { return p, err }}
	}
	tests := []struct {
		name     string
		projects *fkProjectRepo
		pages    *fkPageRepo
		project  string
		page     string
		want     error
		wantMsg  string
	}{
		{"missing ids", pubProject(nil, nil), pageRepo(nil, nil), "", "pg", nil, "projectId and pageId are required"},
		{"missing page id", pubProject(nil, nil), pageRepo(nil, nil), "p", "", nil, "projectId and pageId are required"},
		{"project lookup error", pubProject(nil, errBoom), pageRepo(nil, nil), "p", "pg", errBoom, ""},
		{"unpublished project", pubProject(nil, nil), pageRepo(nil, nil), "p", "pg", repositories.ErrProjectNotFound, ""},
		{"page lookup error", pubProject(&models.Project{}, nil), pageRepo(nil, errBoom), "p", "pg", errBoom, ""},
		{"missing page", pubProject(&models.Project{}, nil), pageRepo(nil, nil), "p", "pg", repositories.ErrPageNotFound, ""},
		{"component pages are not public", pubProject(&models.Project{}, nil), pageRepo(&models.Page{Id: "pg", Type: models.PageTypeComponent}, nil), "p", "pg", repositories.ErrPageNotFound, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newPageSvc(tc.pages, tc.projects).GetPublicByID(ctx, tc.project, tc.page)
			if got != nil {
				t.Fatalf("got page %+v", got)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.wantMsg != "" {
				wantErr(t, err, tc.wantMsg)
			}
		})
	}

	t.Run("routable page of a published project", func(t *testing.T) {
		var gotProject string
		pr := &fkPageRepo{OnGetByID: func(_ context.Context, id, project string) (*models.Page, error) {
			gotProject = project
			return &models.Page{Id: id, Type: "page"}, nil
		}}
		got, err := newPageSvc(pr, pubProject(&models.Project{}, nil)).GetPublicByID(ctx, "p", "pg")
		wantNoErr(t, err)
		if got.Id != "pg" || gotProject != "p" {
			t.Fatalf("page %+v scoped to project %q", got, gotProject)
		}
	})
}

func newValidTestPage() *models.Page {
	return &models.Page{ProjectId: "p", Name: "Home", Type: "page"}
}

func TestPageServiceCreatePageValidation(t *testing.T) {
	ctx := context.Background()
	validation := map[string]func(*models.Page) *models.Page{
		"page cannot be nil":    func(*models.Page) *models.Page { return nil },
		"projectId is required": func(p *models.Page) *models.Page { p.ProjectId = ""; return p },
		"page name is required": func(p *models.Page) *models.Page { p.Name = ""; return p },
		"page type is required": func(p *models.Page) *models.Page { p.Type = ""; return p },
	}
	for msg, mutate := range validation {
		pr := &fkPageRepo{}
		_, err := newPageSvc(pr, &fkProjectRepo{}).Create(ctx, mutate(newValidTestPage()))
		wantErr(t, err, msg)
		if len(pr.Calls) != 0 {
			t.Fatalf("%s: page repo called", msg)
		}
	}
}

func TestPageServiceCreatePageLookupAndInsert(t *testing.T) {
	ctx := context.Background()

	errProjects := &fkProjectRepo{OnGetPublicByID: func(context.Context, string) (*models.Project, error) { return nil, errBoom }}
	_, err := newPageSvc(&fkPageRepo{}, errProjects).Create(ctx, newValidTestPage())
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	missingProjects := &fkProjectRepo{OnGetPublicByID: func(context.Context, string) (*models.Project, error) { return nil, nil }}
	pr := &fkPageRepo{}
	_, err = newPageSvc(pr, missingProjects).Create(ctx, newValidTestPage())
	wantErr(t, err, "project does not exist")
	if len(pr.Calls) != 0 {
		t.Fatal("page created for missing project")
	}

	projects := &fkProjectRepo{OnGetPublicByID: func(context.Context, string) (*models.Project, error) { return &models.Project{}, nil }}
	var stored *models.Page
	pr = &fkPageRepo{OnCreate: func(_ context.Context, p *models.Page) error { stored = p; return nil }}
	in := newValidTestPage()
	got, err := newPageSvc(pr, projects).Create(ctx, in)
	wantNoErr(t, err)
	if got != in || stored != in {
		t.Fatal("page not stored/returned as given")
	}
	pr.OnCreate = func(context.Context, *models.Page) error { return errBoom }
	got, err = newPageSvc(pr, projects).Create(ctx, newValidTestPage())
	if !errors.Is(err, errBoom) || got != nil {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestPageServiceUpdates(t *testing.T) {
	ctx := context.Background()

	t.Run("Update forces the path ID onto the page", func(t *testing.T) {
		var stored *models.Page
		pr := &fkPageRepo{OnUpdate: func(_ context.Context, p *models.Page) error { stored = p; return nil }}
		svc := newPageSvc(pr, nil)
		got, err := svc.Update(ctx, "pg", &models.Page{Id: "spoofed", Name: "x"})
		wantNoErr(t, err)
		if got.Id != "pg" || stored.Id != "pg" {
			t.Fatalf("id = %q / %q", got.Id, stored.Id)
		}
		_, err = svc.Update(ctx, "", &models.Page{})
		wantErr(t, err, "pageId is required")
		_, err = svc.Update(ctx, "pg", nil)
		wantErr(t, err, "page cannot be nil")
		pr.OnUpdate = func(context.Context, *models.Page) error { return errBoom }
		got, err = svc.Update(ctx, "pg", &models.Page{})
		if !errors.Is(err, errBoom) || got != nil {
			t.Fatalf("got %v err %v", got, err)
		}
	})

	t.Run("UpdateFields", func(t *testing.T) {
		var gotUpdates map[string]any
		pr := &fkPageRepo{OnUpdateFields: func(_ context.Context, id string, u map[string]any) error { gotUpdates = u; return nil }}
		svc := newPageSvc(pr, nil)
		wantErr(t, svc.UpdateFields(ctx, "", map[string]any{"a": 1}), "pageId is required")
		wantErr(t, svc.UpdateFields(ctx, "pg", nil), "no updates provided")
		wantErr(t, svc.UpdateFields(ctx, "pg", map[string]any{}), "no updates provided")
		wantNoErr(t, svc.UpdateFields(ctx, "pg", map[string]any{"Name": "n"}))
		if gotUpdates["Name"] != "n" {
			t.Fatalf("updates = %+v", gotUpdates)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		pr := &fkPageRepo{OnDelete: func(_ context.Context, id string) error { return nil }}
		svc := newPageSvc(pr, nil)
		wantErr(t, svc.Delete(ctx, ""), "pageId is required")
		wantNoErr(t, svc.Delete(ctx, "pg"))
		wantCalls(t, pr.Calls, "Delete")
	})
}

func TestPageServiceDeleteByProject(t *testing.T) {
	ctx := context.Background()

	t.Run("validation", func(t *testing.T) {
		svc := newPageSvc(&fkPageRepo{}, &fkProjectRepo{})
		wantErr(t, svc.DeleteForUser(ctx, "", "p", "u"), "pageId is required")
		wantErr(t, svc.DeleteForUser(ctx, "pg", "", "u"), "projectId is required")
		wantErr(t, svc.DeleteForUser(ctx, "pg", "p", ""), "userId is required")
		wantErr(t, svc.DeleteInProject(ctx, "", "p"), "pageId is required")
		wantErr(t, svc.DeleteInProject(ctx, "pg", ""), "projectId is required")
	})
	t.Run("access denied blocks delete", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{Role: models.RoleViewer}, nil
		}}
		pr := &fkPageRepo{}
		err := newPageSvc(pr, projects).DeleteForUser(ctx, "pg", "p", "u")
		if err == nil {
			t.Fatal("expected forbidden error")
		}
		if len(pr.Calls) != 0 {
			t.Fatal("page deleted without access")
		}
	})
	t.Run("authorized delete forwards user", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{Role: models.RoleEditor}, nil
		}}
		var got [3]string
		pr := &fkPageRepo{OnDeleteForUser: func(_ context.Context, pg, p, u string) error { got = [3]string{pg, p, u}; return nil }}
		wantNoErr(t, newPageSvc(pr, projects).DeleteForUser(ctx, "pg", "p", "u"))
		if got != [3]string{"pg", "p", "u"} {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("without verification passes empty user", func(t *testing.T) {
		var got [3]string
		pr := &fkPageRepo{OnDeleteForUser: func(_ context.Context, pg, p, u string) error { got = [3]string{pg, p, u}; return nil }}
		wantNoErr(t, newPageSvc(pr, &fkProjectRepo{}).DeleteInProject(ctx, "pg", "p"))
		if got != [3]string{"pg", "p", ""} {
			t.Fatalf("got %v", got)
		}
	})
}

func TestPageServiceCreatePageAsUser(t *testing.T) {
	ctx := context.Background()
	access := func(role models.CollaboratorRole, err error) *fkProjectRepo {
		return &fkProjectRepo{OnGetAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
			if err != nil {
				return nil, err
			}
			return &models.ProjectAccess{Role: role}, nil
		}}
	}
	valid := func() *models.Page { return &models.Page{ProjectId: "p", Name: "n", Type: "page"} }

	t.Run("nil page", func(t *testing.T) {
		_, err := newPageSvc(&fkPageRepo{}, &fkProjectRepo{}).CreateForUser(ctx, nil, "u")
		wantErr(t, err, "page cannot be nil")
	})

	for _, tc := range []struct {
		name    string
		project *fkProjectRepo
		want    error
		created bool
	}{
		{"owner may create", access(models.RoleOwner, nil), nil, true},
		{"editor may create", access(models.RoleEditor, nil), nil, true},
		{"viewer is unauthorized", access(models.RoleViewer, nil), authz.ErrForbidden, false},
		{"stranger sees unauthorized not not-found", access("", repositories.ErrProjectNotFound), authz.ErrForbidden, false},
		{"other access errors pass through", access("", errBoom), errBoom, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			created := false
			pr := &fkPageRepo{OnCreate: func(context.Context, *models.Page) error { created = true; return nil }}
			got, err := newPageSvc(pr, tc.project).CreateForUser(ctx, valid(), "u")
			if tc.want == nil {
				wantNoErr(t, err)
				if got == nil {
					t.Fatal("no page returned")
				}
			} else if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if created != tc.created {
				t.Fatalf("created = %v, want %v", created, tc.created)
			}
		})
	}

	t.Run("authorized caller still needs name and type", func(t *testing.T) {
		svc := newPageSvc(&fkPageRepo{}, access(models.RoleEditor, nil))
		_, err := svc.CreateForUser(ctx, &models.Page{ProjectId: "p", Type: "page"}, "u")
		wantErr(t, err, "page name is required")
		_, err = svc.CreateForUser(ctx, &models.Page{ProjectId: "p", Name: "n"}, "u")
		wantErr(t, err, "page type is required")
	})
}
