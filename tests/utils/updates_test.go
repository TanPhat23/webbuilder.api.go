package utils_test

import (
	"encoding/json"
	"testing"

	"my-go-app/pkg/utils"
)

// ─── BuildColumnUpdates ───────────────────────────────────────────────────────

func TestBuildColumnUpdates_AllowlistedScalarFields(t *testing.T) {
	updates := map[string]any{
		"name":        "My Project",
		"description": "A description",
	}
	allowed := map[string]string{
		"name":        "Name",
		"description": "Description",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols["Name"] != "My Project" {
		t.Errorf("Name: got %q, want %q", cols["Name"], "My Project")
	}
	if cols["Description"] != "A description" {
		t.Errorf("Description: got %q, want %q", cols["Description"], "A description")
	}
}

func TestBuildColumnUpdates_UnknownKeysAreIgnored(t *testing.T) {
	updates := map[string]any{
		"name":    "My Project",
		"unknown": "should be dropped",
	}
	allowed := map[string]string{
		"name": "Name",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 1 {
		t.Fatalf("expected 1 column, got %d: %v", len(cols), cols)
	}
	if _, present := cols["unknown"]; present {
		t.Error("unknown key should not appear in output")
	}
}

func TestBuildColumnUpdates_ObjectValueMarshalledToRawMessage(t *testing.T) {
	updates := map[string]any{
		"styles": map[string]any{"color": "red", "fontSize": 14},
	}
	allowed := map[string]string{
		"styles": "Styles",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := cols["Styles"].(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage for Styles, got %T", cols["Styles"])
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Styles RawMessage is not valid JSON: %v", err)
	}
	if decoded["color"] != "red" {
		t.Errorf("color: got %v, want %q", decoded["color"], "red")
	}
}

func TestBuildColumnUpdates_ArrayValueMarshalledToRawMessage(t *testing.T) {
	updates := map[string]any{
		"tags": []any{"a", "b", "c"},
	}
	allowed := map[string]string{
		"tags": "Tags",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := cols["Tags"].(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage for Tags, got %T", cols["Tags"])
	}

	var decoded []any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Tags RawMessage is not valid JSON: %v", err)
	}
	if len(decoded) != 3 {
		t.Errorf("expected 3 tags, got %d", len(decoded))
	}
}

func TestBuildColumnUpdates_EmptyUpdatesReturnsEmptyMap(t *testing.T) {
	cols, err := utils.BuildColumnUpdates(map[string]any{}, map[string]string{"name": "Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 0 {
		t.Errorf("expected empty map, got %v", cols)
	}
}

func TestBuildColumnUpdates_EmptyAllowlistReturnsEmptyMap(t *testing.T) {
	cols, err := utils.BuildColumnUpdates(map[string]any{"name": "x"}, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 0 {
		t.Errorf("expected empty map, got %v", cols)
	}
}

func TestBuildColumnUpdates_BoolAndNumericScalarsPassThrough(t *testing.T) {
	updates := map[string]any{
		"published": true,
		"order":     float64(5),
	}
	allowed := map[string]string{
		"published": "Published",
		"order":     "Order",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cols["Published"] != true {
		t.Errorf("Published: got %v, want true", cols["Published"])
	}
	if cols["Order"] != float64(5) {
		t.Errorf("Order: got %v, want 5", cols["Order"])
	}
}

func TestBuildColumnUpdates_NilValuePassesThrough(t *testing.T) {
	updates := map[string]any{
		"description": nil,
	}
	allowed := map[string]string{
		"description": "Description",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := cols["Description"]; !ok || v != nil {
		t.Errorf("Description: expected nil, got %v (present=%v)", v, ok)
	}
}

func TestBuildColumnUpdates_MixedAllowedAndForbiddenKeys(t *testing.T) {
	updates := map[string]any{
		"name":     "Allowed",
		"ownerId":  "should-be-stripped",
		"deletedAt": "should-be-stripped",
	}
	allowed := map[string]string{
		"name": "Name",
	}

	cols, err := utils.BuildColumnUpdates(updates, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 1 {
		t.Errorf("expected exactly 1 column, got %d: %v", len(cols), cols)
	}
	if cols["Name"] != "Allowed" {
		t.Errorf("Name: got %v, want %q", cols["Name"], "Allowed")
	}
}

// ─── RequireUpdates ───────────────────────────────────────────────────────────

func TestRequireUpdates_NonEmptyMapReturnsNil(t *testing.T) {
	err := utils.RequireUpdates(map[string]any{"name": "x"})
	if err != nil {
		t.Errorf("expected nil error for non-empty map, got: %v", err)
	}
}

func TestRequireUpdates_EmptyMapReturnsError(t *testing.T) {
	err := utils.RequireUpdates(map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty map, got nil")
	}
}

func TestRequireUpdates_NilMapReturnsError(t *testing.T) {
	err := utils.RequireUpdates(nil)
	if err == nil {
		t.Fatal("expected error for nil map, got nil")
	}
}

func TestRequireUpdates_SingleEntryIsValid(t *testing.T) {
	err := utils.RequireUpdates(map[string]any{"published": false})
	if err != nil {
		t.Errorf("single-entry map should be valid, got: %v", err)
	}
}