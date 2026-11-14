package logic

import "testing"

func TestMergeWithLiteralNullDefaults(t *testing.T) {
	base := Empty()
	patch := map[string]any{"variables": map[string]any{"object": map[string]any{"type": "object", "default": map[string]any{"nested": map[string]any{}}}, "empty": map[string]any{"type": "object"}}}
	merged, err := MergeWithNulls(base, patch, [][]string{{"variables", "object", "default", "nested", "value"}, {"variables", "empty", "default"}})
	if err != nil {
		t.Fatal(err)
	}
	variables := merged["variables"].(map[string]any)
	if value, ok := variables["empty"].(map[string]any)["default"]; !ok || value != nil {
		t.Fatal("missing null default")
	}
	nested := variables["object"].(map[string]any)["default"].(map[string]any)["nested"].(map[string]any)
	if value, ok := nested["value"]; !ok || value != nil {
		t.Fatal("missing nested null")
	}
	if _, err := MergeWithNulls(base, patch, [][]string{{"flows", "f"}}); err == nil {
		t.Fatal("accepted a null outside a variable default")
	}
	if _, err := MergeWithNulls(base, patch, [][]string{{"variables", "missing", "default"}}); err == nil {
		t.Fatal("accepted a missing parent")
	}
}

func TestMergeAndValidateScopes(t *testing.T) {
	project := Document{"v": float64(1), "variables": map[string]any{"count": map[string]any{"type": "number", "default": float64(0)}}, "flows": map[string]any{}}
	page := Document{"v": float64(1), "variables": map[string]any{}, "flows": map[string]any{}}
	merged, err := Merge(page, map[string]any{"flows": map[string]any{"greet": map[string]any{"do": []any{map[string]any{"toast": "Hi"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateScopes(project, map[string]Document{"page": merged}); err != nil {
		t.Fatal(err)
	}
	if len(page["flows"].(map[string]any)) != 0 {
		t.Fatal("mutated original")
	}
	duplicate, err := Merge(merged, map[string]any{"variables": map[string]any{"count": map[string]any{"type": "number", "default": float64(1)}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateScopes(project, map[string]Document{"page": duplicate}); err == nil {
		t.Fatal("duplicate name accepted")
	}
	if _, err := Merge(page, map[string]any{"flows": map[string]any{"bad": map[string]any{"do": []any{map[string]any{"toast": 1}}}}}); err == nil {
		t.Fatal("invalid step accepted")
	}
	removed, err := Merge(merged, map[string]any{"flows": map[string]any{"greet": nil}})
	if err != nil {
		t.Fatal(err)
	}
	if len(removed["flows"].(map[string]any)) != 0 {
		t.Fatal("null did not remove flow")
	}
}

func TestRejectMalformedStepFields(t *testing.T) {
	for _, step := range []map[string]any{
		{"call": "api", "method": "FETCH", "url": "/"},
		{"call": "api", "method": "POST", "url": "/", "headers": map[string]any{"X": 42}},
		{"class": "add", "target": "x", "name": 42},
		{"animate": "x", "preset": "unknown"},
		{"toast": "Hi", "duration": "wrong"},
		{"set": "bad-name", "to": "1"},
		{"run": "bad-name"},
	} {
		doc := Empty()
		doc["flows"] = map[string]any{"f": map[string]any{"do": []any{step}}}
		if err := Validate(doc); err == nil {
			t.Errorf("accepted invalid step %#v", step)
		}
	}
}

func TestRejectInaccessibleVariableNames(t *testing.T) {
	doc := Empty()
	doc["variables"] = map[string]any{"constructor": map[string]any{"type": "string", "default": ""}}
	if err := Validate(doc); err == nil {
		t.Fatal("accepted inaccessible variable name")
	}
}
