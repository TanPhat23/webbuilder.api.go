// Package logic validates and merges page and project interaction documents.
package logic

import (
	"fmt"
	"math"
	"regexp"

	elementutils "my-go-app/internal/realtime/utils/element"
)

type Document map[string]any

// Revisioned is a stored document with its revision: 0 until first written,
// then incremented by every committed update of that document. Clients use it
// to drop a document older than one they already have.
type Revisioned struct {
	Doc      Document
	Revision int64
}

var nameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)

func Empty() Document {
	return Document{"v": float64(1), "variables": map[string]any{}, "flows": map[string]any{}}
}

// Merge applies the same recursive null-deleting patch semantics as element settings.
func Merge(base Document, patch map[string]any) (Document, error) {
	return MergeWithNulls(base, patch, nil)
}

// MergeWithNulls preserves merge-patch deletion semantics while allowing
// explicit literal null values inside variable defaults.
func MergeWithNulls(base Document, patch map[string]any, nullPaths [][]string) (Document, error) {
	if base == nil {
		base = Empty()
	}
	merged := Document(elementutils.MergeMap(map[string]any(base), patch))
	if len(nullPaths) > 1024 {
		return nil, fmt.Errorf("too many literal null paths")
	}
	for _, path := range nullPaths {
		if len(path) < 3 || len(path) > 64 || path[0] != "variables" || path[2] != "default" {
			return nil, fmt.Errorf("invalid literal null path")
		}
		current := map[string]any(merged)
		for _, key := range path[:len(path)-1] {
			child, ok := current[key].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("literal null path has no object parent")
			}
			copy := make(map[string]any, len(child))
			for k, v := range child {
				copy[k] = v
			}
			current[key] = copy
			current = copy
		}
		current[path[len(path)-1]] = nil
	}
	if err := Validate(merged); err != nil {
		return nil, err
	}
	return merged, nil
}

func obj(v any) (map[string]any, bool) { m, ok := v.(map[string]any); return m, ok && m != nil }
func only(m map[string]any, fields ...string) bool {
	allowed := make(map[string]bool, len(fields))
	for _, f := range fields {
		allowed[f] = true
	}
	for k := range m {
		if !allowed[k] {
			return false
		}
	}
	return true
}
func validDefault(kind string, value any) bool {
	switch kind {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		n, ok := value.(float64)
		return ok && !math.IsNaN(n) && !math.IsInf(n, 0)
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		if value == nil {
			return true
		}
		_, ok := obj(value)
		return ok
	case "list":
		_, ok := value.([]any)
		return ok
	}
	return false
}
func Validate(doc Document) error {
	if doc == nil || !only(doc, "v", "variables", "flows") || doc["v"] != float64(1) {
		return fmt.Errorf("invalid logic document version or fields")
	}
	variables, ok := obj(doc["variables"])
	if !ok {
		return fmt.Errorf("variables must be an object")
	}
	flows, ok := obj(doc["flows"])
	if !ok {
		return fmt.Errorf("flows must be an object")
	}
	for name, raw := range variables {
		if !nameRE.MatchString(name) || name == "__proto__" || name == "constructor" || name == "prototype" {
			return fmt.Errorf("invalid variable name %q", name)
		}
		def, ok := obj(raw)
		if !ok || !only(def, "type", "default", "lifetime") {
			return fmt.Errorf("invalid variable %q", name)
		}
		kind, ok := def["type"].(string)
		if !ok || !validDefault(kind, def["default"]) {
			return fmt.Errorf("invalid default for variable %q", name)
		}
		if _, exists := def["default"]; !exists {
			return fmt.Errorf("missing default for variable %q", name)
		}
		if lifetime, exists := def["lifetime"]; exists && lifetime != "page" && lifetime != "session" && lifetime != "local" {
			return fmt.Errorf("invalid lifetime for variable %q", name)
		}
	}
	for name, raw := range flows {
		if !nameRE.MatchString(name) {
			return fmt.Errorf("invalid flow name %q", name)
		}
		flow, ok := obj(raw)
		if !ok || !only(flow, "do") {
			return fmt.Errorf("invalid flow %q", name)
		}
		if err := validateSteps(flow["do"], 0); err != nil {
			return fmt.Errorf("flow %s: %w", name, err)
		}
	}
	return nil
}

// ValidateScopes enforces one namespace per definition kind across the project.
func ValidateScopes(project Document, pages map[string]Document) error {
	docs := []Document{project}
	for _, page := range pages {
		docs = append(docs, page)
	}
	seen := map[string]map[string]bool{"variables": {}, "flows": {}}
	for _, doc := range docs {
		if doc == nil {
			doc = Empty()
		}
		if err := Validate(doc); err != nil {
			return err
		}
		for _, kind := range []string{"variables", "flows"} {
			entries, _ := obj(doc[kind])
			for name := range entries {
				if seen[kind][name] {
					return fmt.Errorf("duplicate %s name %q", kind, name)
				}
				seen[kind][name] = true
			}
		}
	}
	return nil
}

var fields = map[string][]string{
	"navigate": {"newTab", "replace"}, "show": {"duration"}, "hide": {"duration"}, "toggle": {"duration"},
	"scroll": {"smooth", "offset"}, "modal": {"target"}, "toast": {"kind", "duration"}, "copy": {},
	"download": {"filename"}, "animate": {"preset", "duration"}, "class": {"target", "name"}, "set": {"to"},
	"call": {"method", "url", "headers", "body", "timeout"}, "submit": {}, "reset": {}, "wait": {},
	"if": {"then", "else"}, "run": {}, "code": {},
}

// A step may be stored before it is filled in, as the editor does: a field
// that is missing or "" is incomplete, not invalid (the client reports it as
// "incomplete"). Only a present value of the wrong type or shape is rejected,
// as are unknown fields, a missing or duplicated verb, and bad nesting.
// Templates are not parsed here; an unparseable one is incomplete on the
// client and fails only its own step when run.
func validateSteps(raw any, depth int) error {
	steps, ok := raw.([]any)
	if !ok || depth > 32 {
		return fmt.Errorf("invalid or deeply nested steps")
	}
	for i, rawStep := range steps {
		step, ok := obj(rawStep)
		if !ok {
			return fmt.Errorf("step %d must be an object", i)
		}
		verb := ""
		for k := range step {
			if _, exists := fields[k]; exists {
				if verb != "" {
					return fmt.Errorf("step %d has multiple verbs", i)
				}
				verb = k
			}
		}
		if verb == "" {
			return fmt.Errorf("step %d has no verb", i)
		}
		allowed := append([]string{"id", verb}, fields[verb]...)
		if !only(step, allowed...) {
			return fmt.Errorf("step %d has unknown fields", i)
		}
		filled := make(map[string]any, len(step))
		for k, v := range step {
			if v != "" {
				filled[k] = v
			}
		}
		if id, exists := filled["id"]; exists {
			s, ok := id.(string)
			if !ok || !nameRE.MatchString(s) {
				return fmt.Errorf("step %d has invalid id", i)
			}
		}
		if value, exists := filled[verb]; exists {
			if verb == "wait" {
				n, ok := value.(float64)
				if !ok || n < 0 || n > 60000 {
					return fmt.Errorf("invalid wait")
				}
			} else if verb == "scroll" {
				switch value.(type) {
				case string, float64:
				default:
					return fmt.Errorf("invalid scroll")
				}
			} else if _, ok := value.(string); !ok {
				return fmt.Errorf("invalid %s", verb)
			}
		}
		if err := validateStepFields(verb, filled); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
		if verb == "if" {
			// A missing branch is incomplete; a present one must be a step list.
			for _, branch := range []string{"then", "else"} {
				if v, ok := step[branch]; ok {
					if err := validateSteps(v, depth+1); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// The client's class-name rule: one class, or a template checked when it runs.
var classNameRE = regexp.MustCompile(`^(?:[A-Za-z0-9_:\-/\[\].%]+|[\s\S]*\{\{[\s\S]*)$`)

func validateStepFields(verb string, step map[string]any) error {
	str := func(key string, required bool) bool {
		v, exists := step[key]
		if !exists {
			return !required
		}
		s, ok := v.(string)
		return ok && (!required || s != "")
	}
	enum := func(key string, required bool, values ...string) bool {
		v, exists := step[key]
		if !exists {
			return !required
		}
		for _, a := range values {
			if v == a {
				return true
			}
		}
		return false
	}
	boolean := func(key string) bool {
		v, exists := step[key]
		if !exists {
			return true
		}
		_, ok := v.(bool)
		return ok
	}
	number := func(key string, min float64, positive bool, max float64) bool {
		v, exists := step[key]
		if !exists {
			return true
		}
		n, ok := v.(float64)
		return ok && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= min && (!positive || n > min) && n <= max
	}
	if !boolean("newTab") || !boolean("replace") || !boolean("smooth") || !number("duration", 0, verb == "animate" || verb == "toast", math.MaxFloat64) || !number("offset", -math.MaxFloat64, false, math.MaxFloat64) || !number("timeout", 0, true, 60000) {
		return fmt.Errorf("invalid option type or value")
	}
	switch verb {
	case "modal":
		if !enum("modal", false, "open", "close") || !str("target", false) {
			return fmt.Errorf("invalid modal")
		}
	case "toast":
		if !enum("kind", false, "success", "error", "info", "warning") {
			return fmt.Errorf("invalid toast kind")
		}
	case "download":
		if !str("filename", false) {
			return fmt.Errorf("invalid filename")
		}
	case "animate":
		if !enum("preset", false, "fadeIn", "slideIn", "bounce", "pulse", "shake", "spin") {
			return fmt.Errorf("invalid animation")
		}
	case "class":
		if !enum("class", false, "add", "remove", "toggle") || !str("target", false) || !str("name", false) {
			return fmt.Errorf("invalid class")
		}
		if name, ok := step["name"].(string); ok && !classNameRE.MatchString(name) {
			return fmt.Errorf("invalid class name")
		}
	case "set":
		if name, ok := step["set"].(string); (ok && !nameRE.MatchString(name)) || !str("to", false) {
			return fmt.Errorf("invalid set")
		}
	case "run":
		if name, ok := step["run"].(string); ok && !nameRE.MatchString(name) {
			return fmt.Errorf("invalid flow name")
		}
	case "call":
		if !enum("call", false, "api") || !enum("method", false, "GET", "POST", "PUT", "DELETE", "PATCH") || !str("url", false) || !str("body", false) {
			return fmt.Errorf("invalid call")
		}
		if h, exists := step["headers"]; exists {
			headers, ok := obj(h)
			if !ok {
				return fmt.Errorf("invalid headers")
			}
			for _, v := range headers {
				if _, ok := v.(string); !ok {
					return fmt.Errorf("invalid header value")
				}
			}
		}
	}
	return nil
}
