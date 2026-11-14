package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLogicDocumentsAreIncludedInJSON(t *testing.T) {
	logic := json.RawMessage(`{"v":1,"variables":{},"flows":{}}`)
	for name, value := range map[string]any{
		"page":    Page{Logic: logic},
		"project": Project{Logic: logic},
	} {
		b, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.Contains(string(b), `"logic":{"v":1`) {
			t.Errorf("%s omits logic: %s", name, b)
		}
	}
}
