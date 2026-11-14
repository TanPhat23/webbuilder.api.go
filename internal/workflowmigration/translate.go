// Package workflowmigration translates legacy handler data without modifying it.
package workflowmigration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Step map[string]any

func TranslateEvent(name string) (string, *bool, bool) {
	if name == "onScroll" {
		repeat := false
		return "visible", &repeat, true
	}
	events := map[string]string{"onClick": "click", "onDoubleClick": "dblclick", "onMouseDown": "mousedown", "onMouseUp": "mouseup", "onMouseEnter": "mouseenter", "onMouseLeave": "mouseleave", "onFocus": "focus", "onBlur": "blur", "onChange": "change", "onInput": "input", "onSubmit": "submit", "onKeyDown": "keydown", "onKeyUp": "keyup", "onLoad": "load", "onError": "error"}
	value, ok := events[name]
	return value, nil, ok
}

// TranslateHandlers rejects behavior it cannot represent rather than silently
// dropping it. An error applies to the whole workflow, which must stay intact.
func TranslateHandlers(raw json.RawMessage) ([]Step, error) {
	var handlers []map[string]any
	if err := json.Unmarshal(raw, &handlers); err != nil {
		return nil, fmt.Errorf("invalid handlers: %w", err)
	}
	if handlers == nil {
		return nil, fmt.Errorf("handlers must be a list")
	}
	return translateList(handlers, 0)
}

func translateList(handlers []map[string]any, depth int) ([]Step, error) {
	if depth > 32 {
		return nil, fmt.Errorf("handler nesting exceeds 32")
	}
	result := []Step{}
	for i, handler := range handlers {
		handlerFields := map[string]bool{"id": true, "eventType": true, "actionType": true, "enabled": true, "config": true, "delay": true, "preventDefault": true, "stopPropagation": true, "conditions": true, "nextHandlers": true, "trueHandlers": true, "falseHandlers": true}
		for key := range handler {
			if !handlerFields[key] {
				return nil, fmt.Errorf("handler %d: unsupported field %s", i, key)
			}
		}
		if raw, exists := handler["enabled"]; exists && raw != nil {
			if _, ok := raw.(bool); !ok {
				return nil, fmt.Errorf("handler %d: invalid enabled", i)
			}
		}
		if enabled, ok := handler["enabled"].(bool); ok && !enabled {
			continue
		}
		for _, field := range []string{"conditions", "trueHandlers", "falseHandlers", "preventDefault", "stopPropagation"} {
			if v, ok := handler[field]; ok && v != nil && v != false {
				if a, ok := v.([]any); !ok || len(a) > 0 {
					return nil, fmt.Errorf("handler %d: unsupported %s", i, field)
				}
			}
		}
		if delay, ok := handler["delay"]; ok && delay != nil {
			n, valid := delay.(float64)
			if !valid || n < 0 || n > 60000 {
				return nil, fmt.Errorf("handler %d: invalid delay", i)
			}
			if n > 0 {
				result = append(result, Step{"wait": n})
			}
		}
		config, ok := handler["config"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("handler %d: missing config", i)
		}
		translated, err := translateAction(config)
		if err != nil {
			return nil, fmt.Errorf("handler %d: %w", i, err)
		}
		result = append(result, translated...)
		if next, exists := handler["nextHandlers"]; exists && next != nil {
			items, ok := next.([]any)
			if !ok {
				return nil, fmt.Errorf("handler %d: invalid nextHandlers", i)
			}
			children := make([]map[string]any, 0, len(items))
			for _, item := range items {
				child, ok := item.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("handler %d: invalid child", i)
				}
				children = append(children, child)
			}
			nested, err := translateList(children, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, nested...)
		}
	}
	return result, nil
}

func translateAction(c map[string]any) ([]Step, error) {
	kind, _ := c["type"].(string)
	allowed := map[string][]string{
		"navigate":    {"target", "value", "openInNewTab", "replaceHistory"},
		"showElement": {"elementId", "animationDuration"}, "hideElement": {"elementId", "animationDuration"}, "toggleElement": {"elementId", "animationDuration"},
		"showNotification": {"message", "notificationType", "duration"}, "customCode": {"code"},
		"copyToClipboard": {"text", "successMessage"}, "downloadFile": {"url", "filename"},
		"submitForm": {"formElementId"}, "resetForm": {"formElementId"}, "modal": {"action", "modalId"},
		"scrollTo":    {"target", "value", "behavior", "offsetY"},
		"toggleClass": {"elementId", "className", "animationDuration"}, "addClass": {"elementId", "className", "animationDuration"}, "removeClass": {"elementId", "className", "animationDuration"},
		"playAnimation": {"elementId", "animationType", "duration", "delay"},
	}
	allowedFields, known := allowed[kind]
	if !known {
		return nil, fmt.Errorf("unsupported action %q", kind)
	}
	permitted := map[string]bool{"type": true}
	for _, key := range allowedFields {
		permitted[key] = true
	}
	for key := range c {
		if !permitted[key] {
			return nil, fmt.Errorf("%s: unsupported field %s", kind, key)
		}
	}
	// Legacy action strings are literal. The new runtime interpolates templates,
	// so translating template-like values would change their meaning.
	for _, key := range []string{"target", "value", "elementId", "message", "notificationType", "code", "text", "successMessage", "url", "filename", "formElementId", "action", "modalId", "behavior", "className", "animationType"} {
		value, exists := c[key]
		if !exists || value == nil {
			continue
		}
		if key == "value" && kind == "scrollTo" && c["target"] == "position" {
			if _, ok := value.(float64); ok {
				continue
			}
		}
		v, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s: invalid %s", kind, key)
		}
		if strings.Contains(v, "{{") {
			return nil, fmt.Errorf("%s: %s contains a literal template marker requiring manual migration", kind, key)
		}
	}
	for _, key := range []string{"openInNewTab", "replaceHistory"} {
		if value, exists := c[key]; exists && value != nil {
			if _, ok := value.(bool); !ok {
				return nil, fmt.Errorf("%s: invalid %s", kind, key)
			}
		}
	}
	for _, key := range []string{"duration", "delay", "animationDuration", "offsetY"} {
		if value, exists := c[key]; exists && value != nil {
			if n, ok := value.(float64); !ok || (key != "offsetY" && n < 0) || (key == "duration" && n == 0) {
				return nil, fmt.Errorf("%s: invalid %s", kind, key)
			}
		}
	}
	str := func(key string) string { v, _ := c[key].(string); return v }
	required := func(key string) (string, error) {
		v := str(key)
		if strings.TrimSpace(v) == "" {
			return "", fmt.Errorf("%s: missing %s", kind, key)
		}
		return v, nil
	}
	var s Step
	switch kind {
	case "navigate":
		v, e := required("value")
		if e != nil {
			return nil, e
		}
		if replace, _ := c["replaceHistory"].(bool); replace {
			return nil, fmt.Errorf("navigate: replaceHistory needs manual review")
		}
		s = Step{"navigate": v}
		if x, ok := c["openInNewTab"].(bool); ok {
			s["newTab"] = x
		}
		if x, ok := c["replaceHistory"].(bool); ok {
			s["replace"] = x
		}
	case "showElement", "hideElement", "toggleElement":
		v, e := required("elementId")
		if e != nil {
			return nil, e
		}
		verb := map[string]string{"showElement": "show", "hideElement": "hide", "toggleElement": "toggle"}[kind]
		s = Step{verb: v}
		if d, ok := c["animationDuration"].(float64); ok {
			s["duration"] = d
		}
	case "showNotification":
		v, e := required("message")
		if e != nil {
			return nil, e
		}
		if kind := str("notificationType"); kind != "" && kind != "success" && kind != "error" && kind != "info" && kind != "warning" {
			return nil, fmt.Errorf("invalid notification type")
		}
		s = Step{"toast": v}
		if v := str("notificationType"); v != "" {
			s["kind"] = v
		}
		if d, ok := c["duration"].(float64); ok {
			s["duration"] = d
		}
	case "customCode":
		return nil, fmt.Errorf("customCode uses a different runtime scope and needs manual migration")
	case "copyToClipboard":
		v, e := required("text")
		if e != nil {
			return nil, e
		}
		if str("successMessage") != "" {
			return nil, fmt.Errorf("copy success message unsupported")
		}
		s = Step{"copy": v}
	case "downloadFile":
		v, e := required("url")
		if e != nil {
			return nil, e
		}
		s = Step{"download": v}
		if v := str("filename"); v != "" {
			s["filename"] = v
		}
	case "submitForm", "resetForm":
		v, e := required("formElementId")
		if e != nil {
			return nil, e
		}
		verb := "submit"
		if kind == "resetForm" {
			verb = "reset"
		}
		s = Step{verb: v}
	case "modal":
		v, e := required("action")
		if e != nil {
			return nil, e
		}
		if v != "open" && v != "close" {
			return nil, fmt.Errorf("invalid modal action")
		}
		s = Step{"modal": v}
		if v := str("modalId"); v != "" {
			s["target"] = v
		}
	case "scrollTo":
		if target := str("target"); target != "elementId" && target != "position" {
			return nil, fmt.Errorf("invalid scroll target")
		}
		if behavior := str("behavior"); behavior != "" && behavior != "smooth" && behavior != "auto" {
			return nil, fmt.Errorf("invalid scroll behavior")
		}
		v, ok := c["value"]
		if str("target") == "elementId" {
			id, valid := v.(string)
			if !valid || strings.TrimSpace(id) == "" {
				return nil, fmt.Errorf("invalid scroll element")
			}
		}
		if !ok {
			return nil, fmt.Errorf("missing scroll value")
		}
		if str("target") == "position" {
			switch n := v.(type) {
			case float64:
				v = n
			case string:
				return nil, fmt.Errorf("string position %q needs manual conversion", n)
			default:
				return nil, fmt.Errorf("invalid position")
			}
		}
		s = Step{"scroll": v}
		if v := str("behavior"); v != "" {
			s["smooth"] = v == "smooth"
		}
		if v, ok := c["offsetY"].(float64); ok {
			s["offset"] = v
		}
	case "toggleClass", "addClass", "removeClass":
		id, e := required("elementId")
		if e != nil {
			return nil, e
		}
		name, e := required("className")
		if e != nil {
			return nil, e
		}
		if !regexp.MustCompile(`^[A-Za-z0-9_:/\[\].%-]+$`).MatchString(name) {
			return nil, fmt.Errorf("invalid class name")
		}
		if _, ok := c["animationDuration"]; ok {
			return nil, fmt.Errorf("class animation duration unsupported")
		}
		s = Step{"class": map[string]string{"toggleClass": "toggle", "addClass": "add", "removeClass": "remove"}[kind], "target": id, "name": name}
	case "playAnimation":
		id, e := required("elementId")
		if e != nil {
			return nil, e
		}
		preset, e := required("animationType")
		if e != nil {
			return nil, e
		}
		switch preset {
		case "fadeIn", "slideIn", "bounce", "pulse", "shake", "spin":
		default:
			return nil, fmt.Errorf("invalid animation preset")
		}
		s = Step{"animate": id, "preset": preset}
		if d, ok := c["duration"].(float64); ok {
			s["duration"] = d
		}
		if d, ok := c["delay"].(float64); ok && d > 0 {
			return []Step{{"wait": d}, s}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported action %q", kind)
	}
	return []Step{s}, nil
}
