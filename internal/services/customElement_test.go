package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func customElementSvc() (*CustomElementService, *fkCustomElementRepo) {
	r := &fkCustomElementRepo{}
	return NewCustomElementService(r), r
}

func customElementFound(r *fkCustomElementRepo) {
	r.OnGetByID = func(_ context.Context, id, userID string) (*models.CustomElement, error) {
		return &models.CustomElement{Id: id, UserId: userID}, nil
	}
}

func TestCustomElementGetListsAndByID(t *testing.T) {
	ctx := context.Background()
	s, r := customElementSvc()
	_, err := s.List(ctx, "", nil)
	wantErr(t, err, "userId is required")

	var gotPub *bool
	r.OnList = func(_ context.Context, u string, p *bool) ([]models.CustomElement, error) {
		gotPub = p
		return []models.CustomElement{{Id: "a"}}, nil
	}
	got, err := s.List(ctx, "u", boolp(true))
	wantNoErr(t, err)
	if len(got) != 1 || gotPub == nil || !*gotPub {
		t.Fatalf("got %v pub %v", got, gotPub)
	}
	r.OnList = func(context.Context, string, *bool) ([]models.CustomElement, error) { return nil, errBoom }
	if _, err = s.List(ctx, "u", nil); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	s2, r2 := customElementSvc()
	_, err = s2.GetByID(ctx, "", "u")
	wantErr(t, err, "custom element id is required")
	_, err = s2.GetByID(ctx, "i", "")
	wantErr(t, err, "userId is required")
	customElementFound(r2)
	item, err := s2.GetByID(ctx, "i", "u")
	wantNoErr(t, err)
	if item.Id != "i" || item.UserId != "u" {
		t.Fatalf("got %+v", item)
	}
}

func TestCustomElementCreateValidationAndDefaults(t *testing.T) {
	ctx := context.Background()
	s, r := customElementSvc()
	_, err := s.Create(ctx, nil)
	wantErr(t, err, "cannot be nil")
	_, err = s.Create(ctx, &models.CustomElement{Name: "  ", UserId: "u"})
	wantErr(t, err, "name is required")
	_, err = s.Create(ctx, &models.CustomElement{Name: "n"})
	wantErr(t, err, "userId is required")
	_, err = s.Create(ctx, &models.CustomElement{Name: "n", UserId: "u"})
	wantErr(t, err, "structure is required")
	wantCalls(t, r.Calls)

	var sent *models.CustomElement
	r.OnCreate = func(_ context.Context, e *models.CustomElement) (*models.CustomElement, error) {
		sent = e
		return e, nil
	}
	_, err = s.Create(ctx, &models.CustomElement{Name: "n", UserId: "u", Structure: []byte("{}")})
	wantNoErr(t, err)
	if string(sent.DefaultProps) != "{}" || sent.Version != "1.0.0" {
		t.Fatalf("defaults not applied: %+v", sent)
	}
	_, err = s.Create(ctx, &models.CustomElement{Name: "n", UserId: "u", Structure: []byte("{}"), DefaultProps: []byte(`{"a":1}`), Version: "2.0.0"})
	wantNoErr(t, err)
	if string(sent.DefaultProps) != `{"a":1}` || sent.Version != "2.0.0" {
		t.Fatalf("explicit values overwritten: %+v", sent)
	}
	r.OnCreate = func(context.Context, *models.CustomElement) (*models.CustomElement, error) { return nil, errBoom }
	if _, err = s.Create(ctx, &models.CustomElement{Name: "n", UserId: "u", Structure: []byte("{}")}); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestCustomElementUpdateValidationAndLookups(t *testing.T) {
	ctx := context.Background()
	ok := map[string]any{"name": "x"}
	s, r := customElementSvc()
	_, err := s.Update(ctx, "", "u", ok)
	wantErr(t, err, "id is required")
	_, err = s.Update(ctx, "i", "", ok)
	wantErr(t, err, "userId is required")
	_, err = s.Update(ctx, "i", "u", nil)
	wantErr(t, err, "no updates provided")
	wantCalls(t, r.Calls)

	r.OnGetByID = func(context.Context, string, string) (*models.CustomElement, error) { return nil, errBoom }
	if _, err := s.Update(ctx, "i", "u", ok); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	r.OnGetByID = func(context.Context, string, string) (*models.CustomElement, error) { return nil, nil }
	_, err = s.Update(ctx, "i", "u", ok)
	wantErr(t, err, "does not exist")
}

func TestCustomElementUpdateFieldsAndSuccess(t *testing.T) {
	ctx := context.Background()
	s, r := customElementSvc()
	customElementFound(r)
	_, err := s.Update(ctx, "i", "u", map[string]any{"name": " "})
	wantErr(t, err, "name cannot be empty")
	_, err = s.Update(ctx, "i", "u", map[string]any{"structure": []byte{}})
	wantErr(t, err, "structure cannot be empty")
	_, err = s.Update(ctx, "i", "u", map[string]any{"version": ""})
	wantErr(t, err, "version cannot be empty")
	wantCalls(t, r.Calls, "GetByID", "GetByID", "GetByID")

	var cols map[string]any
	r.OnUpdate = func(_ context.Context, id, u string, c map[string]any) (*models.CustomElement, error) {
		cols = c
		return &models.CustomElement{Id: id}, nil
	}
	got, err := s.Update(ctx, "i", "u", map[string]any{"name": "N", "isPublic": true, "bogus": 1, "structure": map[string]any{"k": "v"}})
	wantNoErr(t, err)
	if got.Id != "i" || cols["Name"] != "N" {
		t.Fatalf("cols = %v", cols)
	}
	if cols["IsPublic"] != true || len(cols) != 3 {
		t.Fatalf("cols = %v", cols)
	}
	if _, bad := cols["bogus"]; bad {
		t.Fatalf("bogus key retained: %v", cols)
	}
	if raw, isRaw := cols["Structure"].(json.RawMessage); !isRaw || string(raw) != `{"k":"v"}` {
		t.Fatalf("structure = %#v", cols["Structure"])
	}

	r.OnUpdate = func(context.Context, string, string, map[string]any) (*models.CustomElement, error) {
		return nil, errBoom
	}
	if _, err := s.Update(ctx, "i", "u", map[string]any{"name": "x"}); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestCustomElementDelete(t *testing.T) {
	ctx := context.Background()
	s, r := customElementSvc()
	wantErr(t, s.Delete(ctx, "", "u"), "id is required")
	wantErr(t, s.Delete(ctx, "i", ""), "userId is required")
	r.OnGetByID = func(context.Context, string, string) (*models.CustomElement, error) { return nil, errBoom }
	if err := s.Delete(ctx, "i", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	r.OnGetByID = func(context.Context, string, string) (*models.CustomElement, error) { return nil, nil }
	wantErr(t, s.Delete(ctx, "i", "u"), "does not exist")
	customElementFound(r)
	r.OnDelete = func(context.Context, string, string) error { return errBoom }
	if err := s.Delete(ctx, "i", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	r.OnDelete = func(context.Context, string, string) error { return nil }
	wantNoErr(t, s.Delete(ctx, "i", "u"))
}

func TestCustomElementGetPublic(t *testing.T) {
	ctx := context.Background()
	s, r := customElementSvc()
	_, err := s.ListPublic(ctx, nil, 0, 0)
	wantErr(t, err, "limit must be greater than zero")
	_, err = s.ListPublic(ctx, nil, 1, -1)
	wantErr(t, err, "offset cannot be negative")
	_, err = s.ListPublic(ctx, strp(" "), 1, 0)
	wantErr(t, err, "category cannot be empty")

	var gl, go_ int
	r.OnListPublic = func(_ context.Context, c *string, l, o int) ([]models.CustomElement, error) {
		gl, go_ = l, o
		return []models.CustomElement{{Id: "p"}}, nil
	}
	got, err := s.ListPublic(ctx, strp("ui"), 5, 10)
	wantNoErr(t, err)
	if len(got) != 1 || gl != 5 || go_ != 10 {
		t.Fatalf("got %v %d %d", got, gl, go_)
	}
}

func TestCustomElementDuplicate(t *testing.T) {
	ctx := context.Background()
	s, r := customElementSvc()
	_, err := s.Duplicate(ctx, "", "u", "n")
	wantErr(t, err, "id is required")
	_, err = s.Duplicate(ctx, "i", "", "n")
	wantErr(t, err, "userId is required")
	_, err = s.Duplicate(ctx, "i", "u", " ")
	wantErr(t, err, "newName is required")
	r.OnGetByID = func(context.Context, string, string) (*models.CustomElement, error) { return nil, errBoom }
	if _, err = s.Duplicate(ctx, "i", "u", "n"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	r.OnGetByID = func(context.Context, string, string) (*models.CustomElement, error) { return nil, nil }
	_, err = s.Duplicate(ctx, "i", "u", "n")
	wantErr(t, err, "does not exist")
	customElementFound(r)
	r.OnDuplicate = func(context.Context, string, string, string) (*models.CustomElement, error) { return nil, errBoom }
	if _, err = s.Duplicate(ctx, "i", "u", "n"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	r.OnDuplicate = func(_ context.Context, _, _, name string) (*models.CustomElement, error) {
		return &models.CustomElement{Name: name}, nil
	}
	got, err := s.Duplicate(ctx, "i", "u", "copy")
	wantNoErr(t, err)
	if got.Name != "copy" {
		t.Fatalf("got %+v", got)
	}
}

func TestCustomElementValidateVersion(t *testing.T) {
	s, r := customElementSvc()
	if r == nil {
		t.Fatal("missing repo")
	}
	for v, want := range map[string]string{"": "version is required", "1.0": "invalid version format", "a.b.c": "invalid version format", "1.2.3.4": "invalid version format"} {
		wantErr(t, s.ValidateVersion(v), want)
	}
	wantNoErr(t, s.ValidateVersion("1.20.3"))
}
