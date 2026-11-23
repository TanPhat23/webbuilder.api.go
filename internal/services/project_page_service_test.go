package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

func TestEnsureProjectVisible(t *testing.T) {
	ctx := context.Background()
	project := &models.Project{ID: "p1"}

	tests := []struct {
		name       string
		userID     string
		access     func(context.Context, string, string) (*models.Project, error)
		public     func(context.Context, string) (*models.Project, error)
		wantErr    error
		wantAccess bool
		wantPublic bool
	}{
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fkProjectRepo{OnGetProjectWithAccess: tc.access, OnGetPublicProjectByID: tc.public}
			err := ensureProjectVisible(ctx, repo, "p1", tc.userID)
			if !errors.Is(err, tc.wantErr) || (tc.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if repo.Called("GetProjectWithAccess") != tc.wantAccess {
				t.Fatalf("access called = %v, want %v", repo.Called("GetProjectWithAccess"), tc.wantAccess)
			}
			if repo.Called("GetPublicProjectByID") != tc.wantPublic {
				t.Fatalf("public called = %v, want %v", repo.Called("GetPublicProjectByID"), tc.wantPublic)
			}
		})
	}
}

func TestProjectServiceGetters(t *testing.T) {
	ctx := context.Background()

	t.Run("GetPublicProjectByID forwards", func(t *testing.T) {
		repo := &fkProjectRepo{OnGetPublicProjectByID: func(_ context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id}, nil
		}}
		p, err := NewProjectService(repo, nil, nil).GetPublicProjectByID(ctx, "p1")
		wantNoErr(t, err)
		if p.ID != "p1" {
			t.Fatalf("id = %q", p.ID)
		}
	})

	for _, tc := range []struct {
		name string
		call func(*ProjectService, string, string) (*models.Project, error)
		hook func(*fkProjectRepo, func(context.Context, string, string) (*models.Project, error))
	}{
		{"GetProjectByID", func(s *ProjectService, p, u string) (*models.Project, error) { return s.GetProjectByID(ctx, p, u) },
			func(r *fkProjectRepo, f func(context.Context, string, string) (*models.Project, error)) {
				r.OnGetProjectByID = f
			}},
		{"GetProjectWithAccess", func(s *ProjectService, p, u string) (*models.Project, error) {
			return s.GetProjectWithAccess(ctx, p, u)
		},
			func(r *fkProjectRepo, f func(context.Context, string, string) (*models.Project, error)) {
				r.OnGetProjectWithAccess = f
			}},
		{"GetProjectWithLock", func(s *ProjectService, p, u string) (*models.Project, error) { return s.GetProjectWithLock(ctx, p, u) },
			func(r *fkProjectRepo, f func(context.Context, string, string) (*models.Project, error)) {
				r.OnGetProjectWithLock = f
			}},
	} {
		t.Run(tc.name+" validation", func(t *testing.T) {
			repo := &fkProjectRepo{}
			svc := NewProjectService(repo, nil, nil)
			_, err := tc.call(svc, "", "u1")
			wantErr(t, err, "projectId is required")
			_, err = tc.call(svc, "p1", "")
			wantErr(t, err, "userId is required")
			wantCalls(t, repo.Calls)
		})
		t.Run(tc.name+" success and repo error", func(t *testing.T) {
			repo := &fkProjectRepo{}
			tc.hook(repo, func(_ context.Context, p, u string) (*models.Project, error) {
				return &models.Project{ID: p, OwnerId: u}, nil
			})
			got, err := tc.call(NewProjectService(repo, nil, nil), "p1", "u1")
			wantNoErr(t, err)
			if got.ID != "p1" || got.OwnerId != "u1" {
				t.Fatalf("got %+v", got)
			}
			tc.hook(repo, func(context.Context, string, string) (*models.Project, error) { return nil, errBoom })
			_, err = tc.call(NewProjectService(repo, nil, nil), "p1", "u1")
			if !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}

	t.Run("GetProjectByID and GetProjectWithAccess reject nil project", func(t *testing.T) {
		repo := &fkProjectRepo{
			OnGetProjectByID:       func(context.Context, string, string) (*models.Project, error) { return nil, nil },
			OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) { return nil, nil },
		}
		svc := NewProjectService(repo, nil, nil)
		_, err := svc.GetProjectByID(ctx, "p1", "u1")
		wantErr(t, err, "project does not exist")
		_, err = svc.GetProjectWithAccess(ctx, "p1", "u1")
		wantErr(t, err, "project does not exist")
	})
}

func TestProjectServiceListings(t *testing.T) {
	ctx := context.Background()
	repo := &fkProjectRepo{
		OnGetProjectsByUserID: func(_ context.Context, u string) ([]models.Project, error) {
			return []models.Project{{ID: "mine", OwnerId: u}}, nil
		},
		OnGetCollaboratorProjects: func(_ context.Context, u string) ([]models.Project, error) {
			return []models.Project{{ID: "shared"}}, nil
		},
	}
	svc := NewProjectService(repo, nil, nil)

	got, err := svc.GetProjectsByUserID(ctx, "u1")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].ID != "mine" {
		t.Fatalf("got %+v", got)
	}
	got, err = svc.GetCollaboratorProjects(ctx, "u1")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].ID != "shared" {
		t.Fatalf("got %+v", got)
	}

	_, err = svc.GetProjectsByUserID(ctx, "")
	wantErr(t, err, "userId is required")
	_, err = svc.GetCollaboratorProjects(ctx, "")
	wantErr(t, err, "userId is required")
	wantCalls(t, repo.Calls, "GetProjectsByUserID", "GetCollaboratorProjects")
}

func TestProjectServiceGetProjectPages(t *testing.T) {
	ctx := context.Background()

	t.Run("requires projectId", func(t *testing.T) {
		_, err := NewProjectService(&fkProjectRepo{}, nil, nil).GetProjectPages(ctx, "", "u1")
		wantErr(t, err, "projectId is required")
	})
	t.Run("invisible project is wrapped and pages are not read", func(t *testing.T) {
		repo := &fkProjectRepo{
			OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
				return nil, repositories.ErrProjectNotFound
			},
			OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) { return nil, nil },
		}
		_, err := NewProjectService(repo, nil, nil).GetProjectPages(ctx, "p1", "u1")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		wantErr(t, err, "failed to verify project")
		if repo.Called("GetProjectPages") {
			t.Fatal("pages read for invisible project")
		}
	})
	t.Run("visible project returns pages", func(t *testing.T) {
		repo := &fkProjectRepo{
			OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
				return &models.Project{ID: "p1"}, nil
			},
			OnGetProjectPages: func(_ context.Context, p, u string) ([]models.Page, error) {
				return []models.Page{{Id: "pg", ProjectId: p}}, nil
			},
		}
		pages, err := NewProjectService(repo, nil, nil).GetProjectPages(ctx, "p1", "u1")
		wantNoErr(t, err)
		if len(pages) != 1 || pages[0].ProjectId != "p1" {
			t.Fatalf("pages = %+v", pages)
		}
	})
}

func TestProjectServiceCreateErrorPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateProject repository failure", func(t *testing.T) {
		repo := &fkProjectRepo{OnCreateProject: func(context.Context, *models.Project) error { return errBoom }}
		got, err := NewProjectService(repo, nil, nil).CreateProject(ctx, &models.Project{Name: "n", OwnerId: "o"})
		if !errors.Is(err, errBoom) || got != nil {
			t.Fatalf("got %v, err %v", got, err)
		}
	})

	t.Run("CreateProjectWithPage validation", func(t *testing.T) {
		for name, p := range map[string]*models.Project{
			"nil":           nil,
			"missing name":  {OwnerId: "o"},
			"missing owner": {Name: "n"},
		} {
			repo := &fkProjectRepo{}
			_, err := NewProjectService(repo, nil, nil).CreateProjectWithPage(ctx, p)
			wantErr(t, err, "project name and ownerId are required")
			if len(repo.Calls) != 0 {
				t.Fatalf("%s: repo called", name)
			}
		}
	})

	t.Run("CreateProjectWithPage creates an empty component-free first page", func(t *testing.T) {
		var gotPage *models.Page
		repo := &fkProjectRepo{OnCreateProjectWithPage: func(_ context.Context, _ *models.Project, pg *models.Page) error {
			gotPage = pg
			return nil
		}}
		_, err := NewProjectService(repo, nil, nil).CreateProjectWithPage(ctx, &models.Project{Name: "n", OwnerId: "o"})
		wantNoErr(t, err)
		if gotPage.Type != "page" || gotPage.Name != "" || string(gotPage.Styles) != "{}" {
			t.Fatalf("page = %+v", gotPage)
		}
	})

	t.Run("CreateProjectWithPage repository failure", func(t *testing.T) {
		repo := &fkProjectRepo{OnCreateProjectWithPage: func(context.Context, *models.Project, *models.Page) error { return errBoom }}
		got, err := NewProjectService(repo, nil, nil).CreateProjectWithPage(ctx, &models.Project{Name: "n", OwnerId: "o"})
		if !errors.Is(err, errBoom) || got != nil {
			t.Fatalf("got %v, err %v", got, err)
		}
	})
}

func TestProjectServiceUpdateProject(t *testing.T) {
	ctx := context.Background()
	desc, sub := strp("old description"), strp("old-sub")
	styles, header := rawp(`{"a":1}`), rawp(`{"h":1}`)
	existing := &models.Project{ID: "p1", Name: "old", Description: desc, Subdomain: sub, Styles: styles, Header: header}

	t.Run("validation", func(t *testing.T) {
		svc := NewProjectService(&fkProjectRepo{}, nil, nil)
		_, err := svc.UpdateProject(ctx, "", "u", &models.Project{})
		wantErr(t, err, "projectId is required")
		_, err = svc.UpdateProject(ctx, "p", "", &models.Project{})
		wantErr(t, err, "userId is required")
		_, err = svc.UpdateProject(ctx, "p", "u", nil)
		wantErr(t, err, "project cannot be nil")
	})

	t.Run("blank fields inherit from existing project", func(t *testing.T) {
		var got map[string]any
		repo := &fkProjectRepo{
			OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) { return existing, nil },
			OnUpdateProject: func(_ context.Context, _, _ string, u map[string]any) (*models.Project, error) {
				got = u
				return &models.Project{ID: "p1"}, nil
			},
		}
		_, err := NewProjectService(repo, nil, nil).UpdateProject(ctx, "p1", "u1", &models.Project{Published: true})
		wantNoErr(t, err)
		if got["Name"] != "old" || got["Description"] != desc || got["Subdomain"] != sub || got["Styles"] != styles || got["Header"] != header {
			t.Fatalf("updates = %+v", got)
		}
		if got["Published"] != true {
			t.Fatalf("Published not forwarded: %+v", got)
		}
	})

	t.Run("supplied fields override existing", func(t *testing.T) {
		var got map[string]any
		repo := &fkProjectRepo{
			OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) { return existing, nil },
			OnUpdateProject: func(_ context.Context, _, _ string, u map[string]any) (*models.Project, error) {
				got = u
				return nil, nil
			},
		}
		newDesc := strp("new")
		_, err := NewProjectService(repo, nil, nil).UpdateProject(ctx, "p1", "u1", &models.Project{Name: "fresh", Description: newDesc})
		wantNoErr(t, err)
		if got["Name"] != "fresh" || got["Description"] != newDesc {
			t.Fatalf("updates = %+v", got)
		}
	})

	t.Run("no access means no write", func(t *testing.T) {
		repo := &fkProjectRepo{OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return nil, repositories.ErrProjectUnauthorized
		}}
		_, err := NewProjectService(repo, nil, nil).UpdateProject(ctx, "p1", "u1", &models.Project{Name: "x"})
		if !errors.Is(err, repositories.ErrProjectUnauthorized) {
			t.Fatalf("err = %v", err)
		}
		if repo.Called("UpdateProject") {
			t.Fatal("UpdateProject called without access")
		}
	})
}

func TestProjectServiceAccessGuardedMutations(t *testing.T) {
	ctx := context.Background()
	type op struct {
		name string
		repo string
		run  func(*ProjectService, string, string) error
	}
	ops := []op{
		{"DeleteProject", "DeleteProject", func(s *ProjectService, p, u string) error { return s.DeleteProject(ctx, p, u) }},
		{"HardDeleteProject", "HardDeleteProject", func(s *ProjectService, p, u string) error { return s.HardDeleteProject(ctx, p, u) }},
		{"RestoreProject", "RestoreProject", func(s *ProjectService, p, u string) error { return s.RestoreProject(ctx, p, u) }},
	}
	for _, o := range ops {
		newRepo := func(accessErr, opErr error) *fkProjectRepo {
			return &fkProjectRepo{
				OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
					if accessErr != nil {
						return nil, accessErr
					}
					return &models.Project{ID: "p1"}, nil
				},
				OnDeleteProject:     func(context.Context, string, string) error { return opErr },
				OnHardDeleteProject: func(context.Context, string, string) error { return opErr },
				OnRestoreProject:    func(context.Context, string, string) error { return opErr },
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
			wantCalls(t, repo.Calls, "GetProjectWithAccess", o.repo)
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
		OnGetProjectAccess: func(_ context.Context, p, u string) (*models.ProjectAccess, error) {
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
	ok, _ = svc.ExistsForUser(ctx, "p2", "u1")
	if ok {
		t.Fatal("unexpected existence")
	}

	access, err := svc.GetProjectAccess(ctx, "p1", "u1")
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
		OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
		OnGetPublicProjectByID: func(_ context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Published: true}, nil
		},
	}
}

func TestPageServiceGetPagesByProjectID(t *testing.T) {
	ctx := context.Background()
	all := []models.Page{{Id: "a", Type: "page"}, {Id: "b", Type: models.PageTypeComponent}, {Id: "c", Type: "page"}}
	pages := func() *fkPageRepo {
		return &fkPageRepo{OnGetPagesByProjectID: func(context.Context, string) ([]models.Page, error) { return all, nil }}
	}

	t.Run("requires projectId", func(t *testing.T) {
		_, err := newPageSvc(pages(), &fkProjectRepo{}).GetPagesByProjectID(ctx, "", "u")
		wantErr(t, err, "projectId is required")
	})
	t.Run("anonymous callers only get routable pages", func(t *testing.T) {
		got, err := newPageSvc(pages(), publishedProject()).GetPagesByProjectID(ctx, "p1", "")
		wantNoErr(t, err)
		if len(got) != 2 || got[0].Id != "a" || got[1].Id != "c" {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("signed-in members also get component pages", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return &models.Project{ID: "p1"}, nil
		}}
		got, err := newPageSvc(pages(), projects).GetPagesByProjectID(ctx, "p1", "u1")
		wantNoErr(t, err)
		if len(got) != 3 {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("unpublished project is hidden", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) { return nil, nil }}
		pr := pages()
		_, err := newPageSvc(pr, projects).GetPagesByProjectID(ctx, "p1", "")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		if len(pr.Calls) != 0 {
			t.Fatal("pages read for hidden project")
		}
	})
	t.Run("repository error is returned", func(t *testing.T) {
		pr := &fkPageRepo{OnGetPagesByProjectID: func(context.Context, string) ([]models.Page, error) { return nil, errBoom }}
		_, err := newPageSvc(pr, publishedProject()).GetPagesByProjectID(ctx, "p1", "")
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestPageServiceGetPageByID(t *testing.T) {
	ctx := context.Background()
	var gotProject string
	pr := &fkPageRepo{OnGetPageByID: func(_ context.Context, id, project string) (*models.Page, error) {
		gotProject = project
		return &models.Page{Id: id}, nil
	}}
	svc := newPageSvc(pr, nil)
	_, err := svc.GetPageByID(ctx, "")
	wantErr(t, err, "pageId is required")
	p, err := svc.GetPageByID(ctx, "pg")
	wantNoErr(t, err)
	if p.Id != "pg" || gotProject != "" {
		t.Fatalf("page %+v project filter %q", p, gotProject)
	}
}

func TestPageServiceGetPublicPageByID(t *testing.T) {
	ctx := context.Background()
	pubProject := func(p *models.Project, err error) *fkProjectRepo {
		return &fkProjectRepo{OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) { return p, err }}
	}
	pageRepo := func(p *models.Page, err error) *fkPageRepo {
		return &fkPageRepo{OnGetPageByID: func(context.Context, string, string) (*models.Page, error) { return p, err }}
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
			got, err := newPageSvc(tc.pages, tc.projects).GetPublicPageByID(ctx, tc.project, tc.page)
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
		pr := &fkPageRepo{OnGetPageByID: func(_ context.Context, id, project string) (*models.Page, error) {
			gotProject = project
			return &models.Page{Id: id, Type: "page"}, nil
		}}
		got, err := newPageSvc(pr, pubProject(&models.Project{}, nil)).GetPublicPageByID(ctx, "p", "pg")
		wantNoErr(t, err)
		if got.Id != "pg" || gotProject != "p" {
			t.Fatalf("page %+v scoped to project %q", got, gotProject)
		}
	})
}

func TestPageServiceCreatePage(t *testing.T) {
	ctx := context.Background()
	valid := func() *models.Page { return &models.Page{ProjectId: "p", Name: "Home", Type: "page"} }

	validation := map[string]func(*models.Page) *models.Page{
		"page cannot be nil":    func(*models.Page) *models.Page { return nil },
		"projectId is required": func(p *models.Page) *models.Page { p.ProjectId = ""; return p },
		"page name is required": func(p *models.Page) *models.Page { p.Name = ""; return p },
		"page type is required": func(p *models.Page) *models.Page { p.Type = ""; return p },
	}
	for msg, mutate := range validation {
		t.Run(msg, func(t *testing.T) {
			pr := &fkPageRepo{}
			_, err := newPageSvc(pr, &fkProjectRepo{}).CreatePage(ctx, mutate(valid()))
			wantErr(t, err, msg)
			if len(pr.Calls) != 0 {
				t.Fatal("page repo called")
			}
		})
	}

	t.Run("project lookup error", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) { return nil, errBoom }}
		_, err := newPageSvc(&fkPageRepo{}, projects).CreatePage(ctx, valid())
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("project does not exist", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) { return nil, nil }}
		pr := &fkPageRepo{}
		_, err := newPageSvc(pr, projects).CreatePage(ctx, valid())
		wantErr(t, err, "project does not exist")
		if len(pr.Calls) != 0 {
			t.Fatal("page created for missing project")
		}
	})
	t.Run("success and insert failure", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) { return &models.Project{}, nil }}
		var stored *models.Page
		pr := &fkPageRepo{OnCreatePage: func(_ context.Context, p *models.Page) error { stored = p; return nil }}
		in := valid()
		got, err := newPageSvc(pr, projects).CreatePage(ctx, in)
		wantNoErr(t, err)
		if got != in || stored != in {
			t.Fatal("page not stored/returned as given")
		}
		pr.OnCreatePage = func(context.Context, *models.Page) error { return errBoom }
		got, err = newPageSvc(pr, projects).CreatePage(ctx, valid())
		if !errors.Is(err, errBoom) || got != nil {
			t.Fatalf("got %v err %v", got, err)
		}
	})
}

func TestPageServiceUpdates(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdatePage forces the path ID onto the page", func(t *testing.T) {
		var stored *models.Page
		pr := &fkPageRepo{OnUpdatePage: func(_ context.Context, p *models.Page) error { stored = p; return nil }}
		svc := newPageSvc(pr, nil)
		got, err := svc.UpdatePage(ctx, "pg", &models.Page{Id: "spoofed", Name: "x"})
		wantNoErr(t, err)
		if got.Id != "pg" || stored.Id != "pg" {
			t.Fatalf("id = %q / %q", got.Id, stored.Id)
		}
		_, err = svc.UpdatePage(ctx, "", &models.Page{})
		wantErr(t, err, "pageId is required")
		_, err = svc.UpdatePage(ctx, "pg", nil)
		wantErr(t, err, "page cannot be nil")
		pr.OnUpdatePage = func(context.Context, *models.Page) error { return errBoom }
		got, err = svc.UpdatePage(ctx, "pg", &models.Page{})
		if !errors.Is(err, errBoom) || got != nil {
			t.Fatalf("got %v err %v", got, err)
		}
	})

	t.Run("UpdatePageFields", func(t *testing.T) {
		var gotUpdates map[string]any
		pr := &fkPageRepo{OnUpdatePageFields: func(_ context.Context, id string, u map[string]any) error { gotUpdates = u; return nil }}
		svc := newPageSvc(pr, nil)
		wantErr(t, svc.UpdatePageFields(ctx, "", map[string]any{"a": 1}), "pageId is required")
		wantErr(t, svc.UpdatePageFields(ctx, "pg", nil), "no updates provided")
		wantErr(t, svc.UpdatePageFields(ctx, "pg", map[string]any{}), "no updates provided")
		wantNoErr(t, svc.UpdatePageFields(ctx, "pg", map[string]any{"Name": "n"}))
		if gotUpdates["Name"] != "n" {
			t.Fatalf("updates = %+v", gotUpdates)
		}
	})

	t.Run("DeletePage", func(t *testing.T) {
		pr := &fkPageRepo{OnDeletePage: func(_ context.Context, id string) error { return nil }}
		svc := newPageSvc(pr, nil)
		wantErr(t, svc.DeletePage(ctx, ""), "pageId is required")
		wantNoErr(t, svc.DeletePage(ctx, "pg"))
		wantCalls(t, pr.Calls, "DeletePage")
	})
}

func TestPageServiceDeleteByProject(t *testing.T) {
	ctx := context.Background()

	t.Run("validation", func(t *testing.T) {
		svc := newPageSvc(&fkPageRepo{}, &fkProjectRepo{})
		wantErr(t, svc.DeletePageByProjectID(ctx, "", "p", "u"), "pageId is required")
		wantErr(t, svc.DeletePageByProjectID(ctx, "pg", "", "u"), "projectId is required")
		wantErr(t, svc.DeletePageByProjectID(ctx, "pg", "p", ""), "userId is required")
		wantErr(t, svc.DeletePageByProjectIDWithoutVerification(ctx, "", "p"), "pageId is required")
		wantErr(t, svc.DeletePageByProjectIDWithoutVerification(ctx, "pg", ""), "projectId is required")
	})
	t.Run("access denied blocks delete", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return nil, repositories.ErrProjectUnauthorized
		}}
		pr := &fkPageRepo{}
		err := newPageSvc(pr, projects).DeletePageByProjectID(ctx, "pg", "p", "u")
		if !errors.Is(err, repositories.ErrProjectUnauthorized) {
			t.Fatalf("err = %v", err)
		}
		if len(pr.Calls) != 0 {
			t.Fatal("page deleted without access")
		}
	})
	t.Run("authorized delete forwards user", func(t *testing.T) {
		projects := &fkProjectRepo{OnGetProjectWithAccess: func(context.Context, string, string) (*models.Project, error) {
			return &models.Project{}, nil
		}}
		var got [3]string
		pr := &fkPageRepo{OnDeletePageByProjectID: func(_ context.Context, pg, p, u string) error { got = [3]string{pg, p, u}; return nil }}
		wantNoErr(t, newPageSvc(pr, projects).DeletePageByProjectID(ctx, "pg", "p", "u"))
		if got != [3]string{"pg", "p", "u"} {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("without verification passes empty user", func(t *testing.T) {
		var got [3]string
		pr := &fkPageRepo{OnDeletePageByProjectID: func(_ context.Context, pg, p, u string) error { got = [3]string{pg, p, u}; return nil }}
		wantNoErr(t, newPageSvc(pr, &fkProjectRepo{}).DeletePageByProjectIDWithoutVerification(ctx, "pg", "p"))
		if got != [3]string{"pg", "p", ""} {
			t.Fatalf("got %v", got)
		}
	})
}

func TestPageServiceCreatePageAsUser(t *testing.T) {
	ctx := context.Background()
	access := func(role models.CollaboratorRole, err error) *fkProjectRepo {
		return &fkProjectRepo{OnGetProjectAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
			if err != nil {
				return nil, err
			}
			return &models.ProjectAccess{Role: role}, nil
		}}
	}
	valid := func() *models.Page { return &models.Page{ProjectId: "p", Name: "n", Type: "page"} }

	t.Run("nil page", func(t *testing.T) {
		_, err := newPageSvc(&fkPageRepo{}, &fkProjectRepo{}).CreatePageAsUser(ctx, nil, "u")
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
		{"viewer is unauthorized", access(models.RoleViewer, nil), repositories.ErrProjectUnauthorized, false},
		{"stranger sees unauthorized not not-found", access("", repositories.ErrProjectNotFound), repositories.ErrProjectUnauthorized, false},
		{"other access errors pass through", access("", errBoom), errBoom, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			created := false
			pr := &fkPageRepo{OnCreatePage: func(context.Context, *models.Page) error { created = true; return nil }}
			got, err := newPageSvc(pr, tc.project).CreatePageAsUser(ctx, valid(), "u")
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
		_, err := svc.CreatePageAsUser(ctx, &models.Page{ProjectId: "p", Type: "page"}, "u")
		wantErr(t, err, "page name is required")
		_, err = svc.CreatePageAsUser(ctx, &models.Page{ProjectId: "p", Name: "n"}, "u")
		wantErr(t, err, "page type is required")
	})
}
