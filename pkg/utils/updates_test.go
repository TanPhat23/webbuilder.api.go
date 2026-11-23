package utils

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestBuildColumnUpdates(t *testing.T) {
	allowed := map[string]string{"name": "Name", "styles": "Styles", "tags": "Tags", "order": "Order"}
	updates := map[string]any{
		"name":      "Home",
		"styles":    map[string]any{"color": "red"},
		"tags":      []any{"a", "b"},
		"order":     float64(3),
		"forbidden": "must be dropped",
	}
	cols, err := BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 4 {
		t.Fatalf("want 4 columns, got %v", cols)
	}
	if _, bad := cols["forbidden"]; bad {
		t.Error("key outside the allowlist leaked")
	}
	if cols["Name"] != "Home" || cols["Order"] != float64(3) {
		t.Errorf("scalars should pass through unchanged: %v", cols)
	}
	styles, ok := cols["Styles"].(json.RawMessage)
	if !ok || string(styles) != `{"color":"red"}` {
		t.Errorf("objects must become json.RawMessage, got %#v", cols["Styles"])
	}
	tags, ok := cols["Tags"].(json.RawMessage)
	if !ok || string(tags) != `["a","b"]` {
		t.Errorf("arrays must become json.RawMessage, got %#v", cols["Tags"])
	}
}

func TestBuildColumnUpdatesEmptyAndMissing(t *testing.T) {
	cols, err := BuildColumnUpdates(map[string]any{"other": 1}, map[string]string{"name": "Name"})
	if err != nil || cols == nil || len(cols) != 0 {
		t.Fatalf("want empty non-nil map, got %v (%v)", cols, err)
	}
	cols, err = BuildColumnUpdates(nil, map[string]string{"name": "Name"})
	if err != nil || len(cols) != 0 {
		t.Fatalf("nil updates: %v (%v)", cols, err)
	}
	cols, err = BuildColumnUpdates(map[string]any{"name": nil}, map[string]string{"name": "Name"})
	if err != nil || !reflect.DeepEqual(cols, map[string]any{"Name": nil}) {
		t.Fatalf("explicit null must be kept so the column can be cleared: %v (%v)", cols, err)
	}
}

func TestBuildColumnUpdatesRejectsUnmarshalableValue(t *testing.T) {
	// A NaN inside a nested object cannot be encoded as JSON.
	updates := map[string]any{"styles": map[string]any{"x": nan()}}
	_, err := BuildColumnUpdates(updates, map[string]string{"styles": "Styles"})
	var fe *fiber.Error
	if !errors.As(err, &fe) || fe.Code != fiber.StatusBadRequest {
		t.Fatalf("want 400 fiber error, got %v", err)
	}
	if fe.Message == "" || fe.Message[:len("Invalid value for field styles")] != "Invalid value for field styles" {
		t.Errorf("message should name the field: %q", fe.Message)
	}
}

func nan() float64 {
	zero := 0.0
	return zero / zero
}
