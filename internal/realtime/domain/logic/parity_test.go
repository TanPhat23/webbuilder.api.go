package logic

import (
	"encoding/json"
	"os"
	"testing"
)

// The same fixtures are checked by the client validator
// (webbuilderv2 tests/interactions-logic-parity.test.ts): both sides must
// agree on what may be stored.
func TestParityFixtures(t *testing.T) {
	raw, err := os.ReadFile("testdata/parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Name   string         `json:"name"`
			Accept bool           `json:"accept"`
			Doc    map[string]any `json:"doc"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Cases) == 0 {
		t.Fatal("no fixtures")
	}
	for _, c := range fixtures.Cases {
		err := Validate(Document(c.Doc))
		if (err == nil) != c.Accept {
			t.Errorf("%s: accept=%v, got err=%v", c.Name, c.Accept, err)
		}
	}
}
