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

var handlerAllowedFields = map[string]bool{
	"id": true, "eventType": true, "actionType": true, "enabled": true, "config": true,
	"delay": true, "preventDefault": true, "stopPropagation": true, "conditions": true,
	"nextHandlers": true, "trueHandlers": true, "falseHandlers": true,
}

var unsupportedNonEmptyHandlerFields = []string{
	"conditions", "trueHandlers", "falseHandlers", "preventDefault", "stopPropagation",
}

func translateList(handlers []map[string]any, depth int) ([]Step, error) {
	if depth > 32 {
		return nil, fmt.Errorf("handler nesting exceeds 32")
	}
	result := []Step{}
	for i, handler := range handlers {
		steps, err := translateSingleHandler(i, handler, depth)
		if err != nil {
			return nil, err
		}
		result = append(result, steps...)
	}
	return result, nil
}

func translateSingleHandler(index int, handler map[string]any, depth int) ([]Step, error) {
	if err := validateHandlerFields(index, handler); err != nil {
		return nil, err
	}
	if enabled, ok := handler["enabled"].(bool); ok && !enabled {
		return nil, nil
	}
	steps, err := extractHandlerDelayStep(index, handler)
	if err != nil {
		return nil, err
	}
	config, ok := handler["config"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("handler %d: missing config", index)
	}
	translated, err := translateAction(config)
	if err != nil {
		return nil, fmt.Errorf("handler %d: %w", index, err)
	}
	steps = append(steps, translated...)
	nested, err := translateNextHandlers(index, handler["nextHandlers"], depth)
	if err != nil {
		return nil, err
	}
	return append(steps, nested...), nil
}

func validateHandlerFields(index int, handler map[string]any) error {
	for key := range handler {
		if !handlerAllowedFields[key] {
			return fmt.Errorf("handler %d: unsupported field %s", index, key)
		}
	}
	if raw, exists := handler["enabled"]; exists && raw != nil {
		if _, ok := raw.(bool); !ok {
			return fmt.Errorf("handler %d: invalid enabled", index)
		}
	}
	for _, field := range unsupportedNonEmptyHandlerFields {
		if v, ok := handler[field]; ok && v != nil && v != false {
			if a, ok := v.([]any); !ok || len(a) > 0 {
				return fmt.Errorf("handler %d: unsupported %s", index, field)
			}
		}
	}
	return nil
}

func extractHandlerDelayStep(index int, handler map[string]any) ([]Step, error) {
	delay, ok := handler["delay"]
	if !ok || delay == nil {
		return nil, nil
	}
	n, valid := delay.(float64)
	if !valid || n < 0 || n > 60000 {
		return nil, fmt.Errorf("handler %d: invalid delay", index)
	}
	if n > 0 {
		return []Step{{"wait": n}}, nil
	}
	return nil, nil
}

func translateNextHandlers(index int, rawNext any, depth int) ([]Step, error) {
	if rawNext == nil {
		return nil, nil
	}
	items, ok := rawNext.([]any)
	if !ok {
		return nil, fmt.Errorf("handler %d: invalid nextHandlers", index)
	}
	children := make([]map[string]any, 0, len(items))
	for _, item := range items {
		child, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("handler %d: invalid child", index)
		}
		children = append(children, child)
	}
	return translateList(children, depth+1)
}

var actionAllowedFields = map[string][]string{
	"navigate":         {"target", "value", "openInNewTab", "replaceHistory"},
	"showElement":      {"elementId", "animationDuration"},
	"hideElement":      {"elementId", "animationDuration"},
	"toggleElement":    {"elementId", "animationDuration"},
	"showNotification": {"message", "notificationType", "duration"},
	"customCode":       {"code"},
	"copyToClipboard":  {"text", "successMessage"},
	"downloadFile":     {"url", "filename"},
	"submitForm":       {"formElementId"},
	"resetForm":        {"formElementId"},
	"modal":            {"action", "modalId"},
	"scrollTo":         {"target", "value", "behavior", "offsetY"},
	"toggleClass":      {"elementId", "className", "animationDuration"},
	"addClass":         {"elementId", "className", "animationDuration"},
	"removeClass":      {"elementId", "className", "animationDuration"},
	"playAnimation":    {"elementId", "animationType", "duration", "delay"},
}

var (
	actionStringKeys = []string{
		"target", "value", "elementId", "message", "notificationType", "code", "text",
		"successMessage", "url", "filename", "formElementId", "action", "modalId",
		"behavior", "className", "animationType",
	}
	actionBoolKeys        = []string{"openInNewTab", "replaceHistory"}
	actionNumberKeys      = []string{"duration", "delay", "animationDuration", "offsetY"}
	classNamePattern      = regexp.MustCompile(`^[A-Za-z0-9_:/\[\].%-]+$`)
	visibilityVerbByKind  = map[string]string{"showElement": "show", "hideElement": "hide", "toggleElement": "toggle"}
	classOpVerbByKind     = map[string]string{"toggleClass": "toggle", "addClass": "add", "removeClass": "remove"}
	validToastTypes       = map[string]bool{"success": true, "error": true, "info": true, "warning": true}
	validAnimationPresets = map[string]bool{"fadeIn": true, "slideIn": true, "bounce": true, "pulse": true, "shake": true, "spin": true}
)

func cfgStr(c map[string]any, key string) string {
	v, _ := c[key].(string)
	return v
}

func requireCfgStr(c map[string]any, kind, key string) (string, error) {
	v := cfgStr(c, key)
	if strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("%s: missing %s", kind, key)
	}
	return v, nil
}

func validateActionConfig(kind string, c map[string]any) error {
	allowedFields, known := actionAllowedFields[kind]
	if !known {
		return fmt.Errorf("unsupported action %q", kind)
	}
	if err := validateActionFieldNames(kind, c, allowedFields); err != nil {
		return err
	}
	if err := validateActionStringValues(kind, c); err != nil {
		return err
	}
	if err := validateActionBoolValues(kind, c); err != nil {
		return err
	}
	return validateActionNumberValues(kind, c)
}

func validateActionFieldNames(kind string, c map[string]any, allowedFields []string) error {
	permitted := make(map[string]bool, len(allowedFields)+1)
	permitted["type"] = true
	for _, key := range allowedFields {
		permitted[key] = true
	}
	for key := range c {
		if !permitted[key] {
			return fmt.Errorf("%s: unsupported field %s", kind, key)
		}
	}
	return nil
}

// Legacy action strings are literal. The new runtime interpolates templates,
// so translating template-like values would change their meaning.
func validateActionStringValues(kind string, c map[string]any) error {
	for _, key := range actionStringKeys {
		value, exists := c[key]
		if !exists || value == nil {
			continue
		}
		if err := validateSingleActionString(kind, key, value, c["target"]); err != nil {
			return err
		}
	}
	return nil
}

func validateSingleActionString(kind, key string, value, target any) error {
	if key == "value" && kind == "scrollTo" && target == "position" {
		if _, ok := value.(float64); ok {
			return nil
		}
	}
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: invalid %s", kind, key)
	}
	if strings.Contains(v, "{{") {
		return fmt.Errorf("%s: %s contains a literal template marker requiring manual migration", kind, key)
	}
	return nil
}

func validateActionBoolValues(kind string, c map[string]any) error {
	for _, key := range actionBoolKeys {
		if value, exists := c[key]; exists && value != nil {
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("%s: invalid %s", kind, key)
			}
		}
	}
	return nil
}

func validateActionNumberValues(kind string, c map[string]any) error {
	for _, key := range actionNumberKeys {
		if value, exists := c[key]; exists && value != nil {
			if !isValidActionNumber(key, value) {
				return fmt.Errorf("%s: invalid %s", kind, key)
			}
		}
	}
	return nil
}

func isValidActionNumber(key string, value any) bool {
	n, ok := value.(float64)
	if !ok {
		return false
	}
	if key != "offsetY" && n < 0 {
		return false
	}
	if key == "duration" && n == 0 {
		return false
	}
	return true
}

func translateAction(c map[string]any) ([]Step, error) {
	kind := cfgStr(c, "type")
	if err := validateActionConfig(kind, c); err != nil {
		return nil, err
	}

	switch kind {
	case "navigate":
		return translateNavigateAction(c)
	case "showElement", "hideElement", "toggleElement":
		return translateVisibilityAction(kind, c)
	case "showNotification":
		return translateNotificationAction(c)
	case "customCode":
		return nil, fmt.Errorf("customCode uses a different runtime scope and needs manual migration")
	case "copyToClipboard":
		return translateCopyAction(c)
	case "downloadFile":
		return translateDownloadAction(c)
	case "submitForm", "resetForm":
		return translateFormAction(kind, c)
	case "modal":
		return translateModalAction(c)
	case "scrollTo":
		return translateScrollAction(c)
	case "toggleClass", "addClass", "removeClass":
		return translateClassAction(kind, c)
	case "playAnimation":
		return translatePlayAnimationAction(c)
	default:
		return nil, fmt.Errorf("unsupported action %q", kind)
	}
}

func translateNavigateAction(c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, "navigate", "value")
	if err != nil {
		return nil, err
	}
	if replace, _ := c["replaceHistory"].(bool); replace {
		return nil, fmt.Errorf("navigate: replaceHistory needs manual review")
	}
	s := Step{"navigate": v}
	if x, ok := c["openInNewTab"].(bool); ok {
		s["newTab"] = x
	}
	if x, ok := c["replaceHistory"].(bool); ok {
		s["replace"] = x
	}
	return []Step{s}, nil
}

func translateVisibilityAction(kind string, c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, kind, "elementId")
	if err != nil {
		return nil, err
	}
	s := Step{visibilityVerbByKind[kind]: v}
	if d, ok := c["animationDuration"].(float64); ok {
		s["duration"] = d
	}
	return []Step{s}, nil
}

func translateNotificationAction(c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, "showNotification", "message")
	if err != nil {
		return nil, err
	}
	toastKind := cfgStr(c, "notificationType")
	if toastKind != "" && !validToastTypes[toastKind] {
		return nil, fmt.Errorf("invalid notification type")
	}
	s := Step{"toast": v}
	if toastKind != "" {
		s["kind"] = toastKind
	}
	if d, ok := c["duration"].(float64); ok {
		s["duration"] = d
	}
	return []Step{s}, nil
}

func translateCopyAction(c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, "copyToClipboard", "text")
	if err != nil {
		return nil, err
	}
	if cfgStr(c, "successMessage") != "" {
		return nil, fmt.Errorf("copy success message unsupported")
	}
	return []Step{{"copy": v}}, nil
}

func translateDownloadAction(c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, "downloadFile", "url")
	if err != nil {
		return nil, err
	}
	s := Step{"download": v}
	if filename := cfgStr(c, "filename"); filename != "" {
		s["filename"] = filename
	}
	return []Step{s}, nil
}

func translateFormAction(kind string, c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, kind, "formElementId")
	if err != nil {
		return nil, err
	}
	verb := "submit"
	if kind == "resetForm" {
		verb = "reset"
	}
	return []Step{{verb: v}}, nil
}

func translateModalAction(c map[string]any) ([]Step, error) {
	v, err := requireCfgStr(c, "modal", "action")
	if err != nil {
		return nil, err
	}
	if v != "open" && v != "close" {
		return nil, fmt.Errorf("invalid modal action")
	}
	s := Step{"modal": v}
	if target := cfgStr(c, "modalId"); target != "" {
		s["target"] = target
	}
	return []Step{s}, nil
}

func translateScrollAction(c map[string]any) ([]Step, error) {
	target := cfgStr(c, "target")
	if target != "elementId" && target != "position" {
		return nil, fmt.Errorf("invalid scroll target")
	}
	behavior := cfgStr(c, "behavior")
	if behavior != "" && behavior != "smooth" && behavior != "auto" {
		return nil, fmt.Errorf("invalid scroll behavior")
	}
	v, err := validateScrollValue(target, c)
	if err != nil {
		return nil, err
	}
	s := Step{"scroll": v}
	if behavior != "" {
		s["smooth"] = behavior == "smooth"
	}
	if offset, ok := c["offsetY"].(float64); ok {
		s["offset"] = offset
	}
	return []Step{s}, nil
}

func validateScrollValue(target string, c map[string]any) (any, error) {
	v, ok := c["value"]
	if target == "elementId" {
		id, valid := v.(string)
		if !valid || strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("invalid scroll element")
		}
	}
	if !ok {
		return nil, fmt.Errorf("missing scroll value")
	}
	if target == "position" {
		return parseScrollPosition(v)
	}
	return v, nil
}

func parseScrollPosition(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case string:
		return 0, fmt.Errorf("string position %q needs manual conversion", n)
	default:
		return 0, fmt.Errorf("invalid position")
	}
}

func translateClassAction(kind string, c map[string]any) ([]Step, error) {
	id, err := requireCfgStr(c, kind, "elementId")
	if err != nil {
		return nil, err
	}
	name, err := requireCfgStr(c, kind, "className")
	if err != nil {
		return nil, err
	}
	if !classNamePattern.MatchString(name) {
		return nil, fmt.Errorf("invalid class name")
	}
	if _, ok := c["animationDuration"]; ok {
		return nil, fmt.Errorf("class animation duration unsupported")
	}
	return []Step{{"class": classOpVerbByKind[kind], "target": id, "name": name}}, nil
}

func translatePlayAnimationAction(c map[string]any) ([]Step, error) {
	id, err := requireCfgStr(c, "playAnimation", "elementId")
	if err != nil {
		return nil, err
	}
	preset, err := requireCfgStr(c, "playAnimation", "animationType")
	if err != nil {
		return nil, err
	}
	if !validAnimationPresets[preset] {
		return nil, fmt.Errorf("invalid animation preset")
	}
	s := Step{"animate": id, "preset": preset}
	if d, ok := c["duration"].(float64); ok {
		s["duration"] = d
	}
	if d, ok := c["delay"].(float64); ok && d > 0 {
		return []Step{{"wait": d}, s}, nil
	}
	return []Step{s}, nil
}
