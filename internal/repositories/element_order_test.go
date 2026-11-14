package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// A replace has already deleted the project rows before flattening. The driver
// returns the empty result that its existing MAX(Order) query would observe.
type emptyConnector struct{}
type emptyConn struct{}
type emptyRows struct{}

func (emptyConnector) Connect(context.Context) (driver.Conn, error) { return emptyConn{}, nil }
func (emptyConnector) Driver() driver.Driver                        { return emptyDriver{} }

type emptyDriver struct{}

func (emptyDriver) Open(string) (driver.Conn, error)  { return emptyConn{}, nil }
func (emptyConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (emptyConn) Close() error                        { return nil }
func (emptyConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (emptyConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return emptyRows{}, nil
}
func (emptyRows) Columns() []string         { return []string{"ParentId", "max_order"} }
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func TestReplaceOrderingNormalizesEachSiblingGroup(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sql.OpenDB(emptyConnector{})}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	inputs := []map[string]any{
		{"id": "a", "type": "Frame", "pageId": "p1", "orderKey": "a0", "elements": []any{
			map[string]any{"id": "c", "type": "Text", "pageId": "p1", "orderKey": "a1"},
			map[string]any{"id": "d", "type": "Text", "pageId": "p1", "orderKey": "a0"},
		}},
		{"id": "x", "type": "Text", "pageId": "p2"},
		{"id": "b", "type": "Text", "pageId": "p1", "orderKey": "a1"},
		{"id": "y", "type": "Text", "pageId": "p2"},
	}
	var roots []models.EditorElement
	for _, input := range inputs {
		element, err := utils.ConvertToEditorElement(input)
		if err != nil {
			t.Fatal(err)
		}
		roots = append(roots, element)
	}
	got, err := (&ElementRepository{}).flattenElementsForInsert(context.Background(), db, roots, "project")
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]models.Element)
	for _, element := range got {
		byID[element.Id] = element
	}
	for _, pair := range [][2]string{{"a", "b"}, {"x", "y"}, {"c", "d"}} {
		first, second := byID[pair[0]], byID[pair[1]]
		if first.Order != 1 || second.Order != 2 || first.OrderKey == nil || second.OrderKey == nil || strings.Compare(*first.OrderKey, *second.OrderKey) >= 0 {
			t.Errorf("siblings %s, %s not ordered by array position: %+v %+v", pair[0], pair[1], first, second)
		}
	}
	if byID["a"].OrderKey == nil || *byID["a"].OrderKey != "a0" || byID["b"].OrderKey == nil || *byID["b"].OrderKey != "a1" {
		t.Error("valid existing keys should be preserved")
	}
}

func TestReplaceOrderingRegeneratesDuplicateAndInvalidKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []string
	}{
		{name: "duplicate", keys: []string{"a0", "a0"}},
		{name: "invalid", keys: []string{"a0", "not-a-key"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pageID := "page"
			roots := []models.EditorElement{
				&models.Element{Id: "first", Type: "Text", PageId: &pageID, OrderKey: &tc.keys[0]},
				&models.Element{Id: "second", Type: "Text", PageId: &pageID, OrderKey: &tc.keys[1]},
			}
			got, err := (&ElementRepository{}).flattenElementsForInsert(context.Background(), nil, roots, "project")
			if err != nil {
				t.Fatal(err)
			}
			if got[0].OrderKey == nil || got[1].OrderKey == nil || *got[0].OrderKey >= *got[1].OrderKey {
				t.Fatalf("expected ordered distinct keys, got %+v", got)
			}
		})
	}
}

func TestFlattenGeneratesMissingIDsAndPreservesSuppliedIDs(t *testing.T) {
	legacy := uuid.NewString()
	current := uuid.Must(uuid.NewV7()).String()
	roots := []models.EditorElement{
		&models.Element{Id: legacy, Type: "Text"},
		&models.Element{Id: current, Type: "Text"},
		&models.Element{Type: "Text"},
	}
	got, err := (&ElementRepository{}).flattenElementsForInsert(context.Background(), nil, roots, "project")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Id != legacy || got[1].Id != current {
		t.Fatalf("supplied IDs changed: %+v", got)
	}
	id, err := uuid.Parse(got[2].Id)
	if err != nil || id.Version() != 7 {
		t.Fatalf("expected generated v7 ID, got %q: %v", got[2].Id, err)
	}
}

func TestFlattenGeneratesNestedIDsAndParentReferences(t *testing.T) {
	root, err := utils.ConvertToEditorElement(map[string]any{"type": "Frame", "pageId": "page", "elements": []any{map[string]any{"type": "Text", "pageId": "page"}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := (&ElementRepository{}).flattenElementsForInsert(context.Background(), nil, []models.EditorElement{root}, "project")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected parent and child, got %d", len(got))
	}
	for _, e := range got {
		id, err := uuid.Parse(e.Id)
		if err != nil || id.Version() != 7 {
			t.Fatalf("expected v7 ID, got %q: %v", e.Id, err)
		}
	}
	if got[1].ParentId == nil || *got[1].ParentId != got[0].Id {
		t.Fatalf("wrong parent reference: %+v", got[1])
	}
}
