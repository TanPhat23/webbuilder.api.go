package workflowmigration

import (
	"encoding/json"
	"testing"
)

func TestTranslateHandlers(t *testing.T) {
	raw := json.RawMessage(`[{"id":"a","eventType":"onClick","actionType":"showNotification","config":{"type":"showNotification","message":"Saved","notificationType":"success"},"delay":50,"nextHandlers":[{"id":"b","eventType":"onClick","actionType":"resetForm","config":{"type":"resetForm","formElementId":"form"}}]}]`)
	steps, err := TranslateHandlers(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 3 || steps[0]["wait"] != float64(50) || steps[1]["toast"] != "Saved" || steps[2]["reset"] != "form" {
		t.Fatalf("unexpected steps: %#v", steps)
	}
}
func TestRejectUnsupportedBehavior(t *testing.T) {
	for _, raw := range []string{
		`[{"config":{"type":"unknown"}}]`,
		`[{"config":{"type":"resetForm"}}]`,
		`[{"config":{"type":"navigate","value":"/"},"conditions":[{"type":"customCode"}]}]`,
	} {
		if _, err := TranslateHandlers(json.RawMessage(raw)); err == nil {
			t.Errorf("silently accepted %s", raw)
		}
	}
}
func TestEventNames(t *testing.T) {
	if event, once, ok := TranslateEvent("onScroll"); !ok || event != "visible" || once == nil || *once {
		t.Fatal("scroll must repeat")
	}
	if event, _, ok := TranslateEvent("onClick"); !ok || event != "click" {
		t.Fatal("click mapping")
	}
	if _, _, ok := TranslateEvent("onUnknown"); ok {
		t.Fatal("unknown event accepted")
	}
}

func TestRejectUnmappedProperties(t *testing.T) {
	for _, raw := range []string{
		`[{"config":{"type":"navigate","value":"/","replaceHistory":true}}]`,
		`[{"config":{"type":"showNotification","message":"x","unexpected":"lost"}}]`,
		`[{"enabled":"wrong","config":{"type":"showNotification","message":"x"}}]`,
	} {
		if _, err := TranslateHandlers(json.RawMessage(raw)); err == nil {
			t.Errorf("silently accepted %s", raw)
		}
	}
}

func TestCustomCodeNeedsManualMigration(t *testing.T) {
	raw := json.RawMessage(`[{"config":{"type":"customCode","code":"return element.id"}}]`)
	if _, err := TranslateHandlers(raw); err == nil {
		t.Fatal("custom code has incompatible scopes and needs manual review")
	}
}

func TestRejectIncorrectLegacyFieldTypes(t *testing.T) {
	for _, raw := range []string{
		`[{"config":{"type":"showNotification","message":"x","duration":"bad"}}]`,
		`[{"config":{"type":"navigate","value":"/","openInNewTab":"yes"}}]`,
		`[{"config":{"type":"showNotification","message":"x"},"unknownFeature":true}]`,
		`[{"config":{"type":"scrollTo","target":"bogus","value":"a"}}]`,
	} {
		if _, err := TranslateHandlers(json.RawMessage(raw)); err == nil {
			t.Errorf("silently accepted %s", raw)
		}
	}
}

func TestRejectUnsupportedActionValues(t *testing.T) {
	for _, raw := range []string{
		`[{"config":{"type":"showNotification","message":"x","notificationType":"other"}}]`,
		`[{"config":{"type":"playAnimation","elementId":"x","animationType":"other"}}]`,
		`[{"config":{"type":"playAnimation","elementId":"x","animationType":"spin","duration":-1}}]`,
		`[{"config":{"type":"scrollTo","target":"elementId","value":5}}]`,
	} {
		if _, err := TranslateHandlers(json.RawMessage(raw)); err == nil {
			t.Errorf("silently accepted %s", raw)
		}
	}
}

func TestRejectLegacyValuesThatWouldBeSilentlyReinterpreted(t *testing.T) {
	for _, raw := range []string{
		`[{"config":{"type":"showNotification","message":"Hello {{vars.name}}"}}]`,
		`[{"config":{"type":"downloadFile","url":"/file","filename":42}}]`,
		`[{"config":{"type":"copyToClipboard","text":"text","successMessage":42}}]`,
		`[{"config":{"type":"showNotification","message":"x","notificationType":false}}]`,
	} {
		if _, err := TranslateHandlers(json.RawMessage(raw)); err == nil {
			t.Errorf("silently accepted %s", raw)
		}
	}
}
