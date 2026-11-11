package models

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestElementRetainsOptionalOrderKey(t *testing.T) {
	var element Element
	if err := json.Unmarshal([]byte(`{"id":"one","type":"Text","order":3,"orderKey":"a1V"}`), &element); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(element)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if result["orderKey"] != "a1V" || result["order"] != float64(3) {
		t.Fatalf("order values not preserved: %s", encoded)
	}

	var legacy Element
	if err := json.Unmarshal([]byte(`{"id":"legacy","type":"Text","order":2}`), &legacy); err != nil {
		t.Fatal(err)
	}
	encoded, err = json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"orderKey"`) {
		t.Fatalf("missing key should be omitted: %s", encoded)
	}
}

func TestElementOrderKeyDatabaseMapping(t *testing.T) {
	parsed, err := schema.Parse(&Element{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	field := parsed.LookUpField("OrderKey")
	if field == nil || field.DBName != "OrderKey" || field.NotNull {
		t.Fatalf("expected nullable OrderKey column, got %#v", field)
	}
}
