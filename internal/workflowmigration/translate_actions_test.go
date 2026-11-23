package workflowmigration

import (
	"encoding/json"
	"strings"
	"testing"
)

func translateActionsRun(t *testing.T, raw string) (string, error) {
	t.Helper()
	steps, err := TranslateHandlers(json.RawMessage(raw))
	if err != nil {
		return "", err
	}
	b, merr := json.Marshal(steps)
	if merr != nil {
		t.Fatal(merr)
	}
	return string(b), nil
}

func translateActionsHandler(config string) string {
	return `[{"config":` + config + `}]`
}

func TestTranslateActionsProduceExpectedSteps(t *testing.T) {
	cases := []struct{ name, config, want string }{
		{"navigate", `{"type":"navigate","value":"/x"}`, `[{"navigate":"/x"}]`},
		{"navigate new tab", `{"type":"navigate","value":"/x","openInNewTab":true}`, `[{"navigate":"/x","newTab":true}]`},
		{"navigate replace false", `{"type":"navigate","value":"/x","replaceHistory":false}`, `[{"navigate":"/x","replace":false}]`},
		{"show", `{"type":"showElement","elementId":"a","animationDuration":200}`, `[{"duration":200,"show":"a"}]`},
		{"hide", `{"type":"hideElement","elementId":"a"}`, `[{"hide":"a"}]`},
		{"toggle", `{"type":"toggleElement","elementId":"a"}`, `[{"toggle":"a"}]`},
		{"toast", `{"type":"showNotification","message":"m","notificationType":"warning","duration":3}`, `[{"duration":3,"kind":"warning","toast":"m"}]`},
		{"toast plain", `{"type":"showNotification","message":"m"}`, `[{"toast":"m"}]`},
		{"copy", `{"type":"copyToClipboard","text":"t"}`, `[{"copy":"t"}]`},
		{"download", `{"type":"downloadFile","url":"/f"}`, `[{"download":"/f"}]`},
		{"download named", `{"type":"downloadFile","url":"/f","filename":"a.txt"}`, `[{"download":"/f","filename":"a.txt"}]`},
		{"submit", `{"type":"submitForm","formElementId":"f"}`, `[{"submit":"f"}]`},
		{"reset", `{"type":"resetForm","formElementId":"f"}`, `[{"reset":"f"}]`},
		{"modal open", `{"type":"modal","action":"open","modalId":"m"}`, `[{"modal":"open","target":"m"}]`},
		{"modal close", `{"type":"modal","action":"close"}`, `[{"modal":"close"}]`},
		{"scroll element", `{"type":"scrollTo","target":"elementId","value":"el","behavior":"smooth","offsetY":-10}`, `[{"offset":-10,"scroll":"el","smooth":true}]`},
		{"scroll auto", `{"type":"scrollTo","target":"elementId","value":"el","behavior":"auto"}`, `[{"scroll":"el","smooth":false}]`},
		{"scroll position", `{"type":"scrollTo","target":"position","value":120}`, `[{"scroll":120}]`},
		{"class toggle", `{"type":"toggleClass","elementId":"a","className":"md:w-1/2"}`, `[{"class":"toggle","name":"md:w-1/2","target":"a"}]`},
		{"class add", `{"type":"addClass","elementId":"a","className":"x"}`, `[{"class":"add","name":"x","target":"a"}]`},
		{"class remove", `{"type":"removeClass","elementId":"a","className":"x"}`, `[{"class":"remove","name":"x","target":"a"}]`},
		{"animate", `{"type":"playAnimation","elementId":"a","animationType":"spin","duration":2}`, `[{"animate":"a","duration":2,"preset":"spin"}]`},
		{"animate delayed", `{"type":"playAnimation","elementId":"a","animationType":"fadeIn","delay":100}`, `[{"wait":100},{"animate":"a","preset":"fadeIn"}]`},
		{"animate zero delay", `{"type":"playAnimation","elementId":"a","animationType":"bounce","delay":0}`, `[{"animate":"a","preset":"bounce"}]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := translateActionsRun(t, translateActionsHandler(c.config))
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %s, want %s", got, c.want)
			}
		})
	}
}

func TestTranslateActionsRejections(t *testing.T) {
	cases := []struct{ name, config, want string }{
		{"no type", `{}`, "unsupported action"},
		{"unknown field", `{"type":"hideElement","elementId":"a","x":1}`, "unsupported field x"},
		{"non-string id", `{"type":"hideElement","elementId":5}`, "invalid elementId"},
		{"template marker", `{"type":"navigate","value":"/{{id}}"}`, "template marker"},
		{"bool wrong type", `{"type":"navigate","value":"/","replaceHistory":"x"}`, "invalid replaceHistory"},
		{"negative duration", `{"type":"showElement","elementId":"a","animationDuration":-1}`, "invalid animationDuration"},
		{"zero duration", `{"type":"playAnimation","elementId":"a","animationType":"spin","duration":0}`, "invalid duration"},
		{"offset not number", `{"type":"scrollTo","target":"elementId","value":"a","offsetY":"x"}`, "invalid offsetY"},
		{"navigate missing value", `{"type":"navigate"}`, "missing value"},
		{"navigate blank value", `{"type":"navigate","value":"  "}`, "missing value"},
		{"navigate replace", `{"type":"navigate","value":"/","replaceHistory":true}`, "manual review"},
		{"show missing id", `{"type":"showElement"}`, "missing elementId"},
		{"toast missing message", `{"type":"showNotification"}`, "missing message"},
		{"toast bad kind", `{"type":"showNotification","message":"m","notificationType":"fatal"}`, "invalid notification type"},
		{"custom code", `{"type":"customCode","code":"x"}`, "manual migration"},
		{"copy missing", `{"type":"copyToClipboard"}`, "missing text"},
		{"copy success message", `{"type":"copyToClipboard","text":"t","successMessage":"ok"}`, "success message unsupported"},
		{"download missing", `{"type":"downloadFile"}`, "missing url"},
		{"submit missing", `{"type":"submitForm"}`, "missing formElementId"},
		{"modal missing", `{"type":"modal"}`, "missing action"},
		{"modal bad action", `{"type":"modal","action":"toggle"}`, "invalid modal action"},
		{"scroll no target", `{"type":"scrollTo","value":"a"}`, "invalid scroll target"},
		{"scroll bad behavior", `{"type":"scrollTo","target":"elementId","value":"a","behavior":"fast"}`, "invalid scroll behavior"},
		{"scroll blank element", `{"type":"scrollTo","target":"elementId","value":" "}`, "invalid scroll element"},
		{"scroll element number", `{"type":"scrollTo","target":"elementId","value":5}`, "invalid value"},
		{"scroll position missing", `{"type":"scrollTo","target":"position"}`, "missing scroll value"},
		{"scroll position string", `{"type":"scrollTo","target":"position","value":"50%"}`, "needs manual conversion"},
		{"class missing id", `{"type":"addClass","className":"x"}`, "missing elementId"},
		{"class missing name", `{"type":"addClass","elementId":"a"}`, "missing className"},
		{"class bad name", `{"type":"addClass","elementId":"a","className":"a b"}`, "invalid class name"},
		{"class animation", `{"type":"addClass","elementId":"a","className":"x","animationDuration":5}`, "animation duration unsupported"},
		{"animate missing id", `{"type":"playAnimation","animationType":"spin"}`, "missing elementId"},
		{"animate missing type", `{"type":"playAnimation","elementId":"a"}`, "missing animationType"},
		{"animate bad preset", `{"type":"playAnimation","elementId":"a","animationType":"wobble"}`, "invalid animation preset"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := translateActionsRun(t, translateActionsHandler(c.config))
			if err == nil || !strings.Contains(err.Error(), c.want) || !strings.HasPrefix(err.Error(), "handler 0: ") {
				t.Fatalf("err = %v, want containing %q", err, c.want)
			}
		})
	}
}

func TestTranslateHandlersStructure(t *testing.T) {
	good := map[string]string{
		"disabled handler skipped":        `[{"enabled":false,"config":{"type":"bogus"}}]`,
		"empty list":                      `[]`,
		"null enabled, false flags, none": `[{"enabled":null,"preventDefault":false,"stopPropagation":null,"conditions":[],"trueHandlers":[],"falseHandlers":[],"nextHandlers":null,"delay":null,"config":{"type":"hideElement","elementId":"a"}}]`,
	}
	wantGood := map[string]string{
		"disabled handler skipped":        `[]`,
		"empty list":                      `[]`,
		"null enabled, false flags, none": `[{"hide":"a"}]`,
	}
	for name, raw := range good {
		got, err := translateActionsRun(t, raw)
		if err != nil || got != wantGood[name] {
			t.Errorf("%s: got %s, %v; want %s", name, got, err, wantGood[name])
		}
	}
	cfg := `"config":{"type":"hideElement","elementId":"a"}`
	bad := map[string]string{
		"not json":          `nope`,
		"object not list":   `{}`,
		"null":              `null`,
		"unknown field":     `[{` + cfg + `,"zzz":1}]`,
		"enabled string":    `[{"enabled":"yes",` + cfg + `}]`,
		"conditions":        `[{"conditions":[{}],` + cfg + `}]`,
		"conditions scalar": `[{"conditions":"x",` + cfg + `}]`,
		"preventDefault":    `[{"preventDefault":true,` + cfg + `}]`,
		"stopPropagation":   `[{"stopPropagation":true,` + cfg + `}]`,
		"trueHandlers":      `[{"trueHandlers":[1],` + cfg + `}]`,
		"falseHandlers":     `[{"falseHandlers":[1],` + cfg + `}]`,
		"delay string":      `[{"delay":"1",` + cfg + `}]`,
		"delay negative":    `[{"delay":-1,` + cfg + `}]`,
		"delay too long":    `[{"delay":60001,` + cfg + `}]`,
		"missing config":    `[{"id":"x"}]`,
		"bad action":        `[{"config":{"type":"bogus"}}]`,
		"next not list":     `[{` + cfg + `,"nextHandlers":"x"}]`,
		"child not object":  `[{` + cfg + `,"nextHandlers":[1]}]`,
		"child fails":       `[{` + cfg + `,"nextHandlers":[{"config":{"type":"bogus"}}]}]`,
	}
	for name, raw := range bad {
		if _, err := TranslateHandlers(json.RawMessage(raw)); err == nil {
			t.Errorf("%s: accepted %s", name, raw)
		}
	}
	if _, err := TranslateHandlers(json.RawMessage(`nope`)); err == nil || !strings.HasPrefix(err.Error(), "invalid handlers: ") {
		t.Errorf("err = %v", err)
	}
	if _, err := TranslateHandlers(json.RawMessage(`null`)); err == nil || err.Error() != "handlers must be a list" {
		t.Errorf("err = %v", err)
	}
}

func TestTranslateHandlersDelayAndNestingOrder(t *testing.T) {
	raw := `[{"delay":10,"config":{"type":"hideElement","elementId":"a"},"nextHandlers":[{"delay":0,"config":{"type":"showElement","elementId":"b"}}]}]`
	got, err := translateActionsRun(t, raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != `[{"wait":10},{"hide":"a"},{"show":"b"}]` {
		t.Fatalf("got %s", got)
	}
}

func TestTranslateHandlersNestingLimit(t *testing.T) {
	nest := func(levels int) string {
		s := `{"config":{"type":"hideElement","elementId":"a"}}`
		for i := 0; i < levels; i++ {
			s = `{"config":{"type":"hideElement","elementId":"a"},"nextHandlers":[` + s + `]}`
		}
		return "[" + s + "]"
	}
	if _, err := TranslateHandlers(json.RawMessage(nest(32))); err != nil {
		t.Fatalf("depth 32 must be accepted: %v", err)
	}
	if _, err := TranslateHandlers(json.RawMessage(nest(33))); err == nil || !strings.Contains(err.Error(), "nesting exceeds 32") {
		t.Fatalf("depth 33 must be rejected, got %v", err)
	}
}

func TestTranslateEventMapping(t *testing.T) {
	want := map[string]string{"onClick": "click", "onDoubleClick": "dblclick", "onMouseDown": "mousedown", "onMouseUp": "mouseup", "onMouseEnter": "mouseenter", "onMouseLeave": "mouseleave", "onFocus": "focus", "onBlur": "blur", "onChange": "change", "onInput": "input", "onSubmit": "submit", "onKeyDown": "keydown", "onKeyUp": "keyup", "onLoad": "load", "onError": "error"}
	for in, out := range want {
		got, once, ok := TranslateEvent(in)
		if !ok || got != out || once != nil {
			t.Errorf("TranslateEvent(%q) = %q, %v, %v", in, got, once, ok)
		}
	}
	if got, _, ok := TranslateEvent(""); ok || got != "" {
		t.Errorf("empty event accepted: %q", got)
	}
}
