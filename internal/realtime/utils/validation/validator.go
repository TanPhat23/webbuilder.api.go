package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

const (
	MaxElementTreeDepth = 32
	MaxElementTreeNodes = 2000
	MaxMapDepth         = 16
)

var (
	validate *validator.Validate
	safeIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
)

func init() {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

// IsSafeID reports whether id is a safe identifier (1..128 chars of [A-Za-z0-9_-])
// with no Redis key delimiters, spaces, or path separators.
func IsSafeID(id string) bool {
	return safeIDRe.MatchString(id)
}

func Presence(p *domainEntity.PayloadPresence) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if p.UserID == "" {
		return fmt.Errorf("userId is required")
	}
	return nil
}

// ElementCreate validates the payload for adding a new element to the canvas
func ElementCreate(p *domainEntity.PayloadElementCreate) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	return ElementForCreate(&p.Element)
}

// ElementUpdate validates granular updates to properties or styles
func ElementUpdate(p *domainEntity.PayloadElementUpdate) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	if mapDepthExceeded(p.Styles, MaxMapDepth) {
		return fmt.Errorf("styles exceed maximum nesting depth of %d", MaxMapDepth)
	}
	if mapDepthExceeded(p.Settings, MaxMapDepth) {
		return fmt.Errorf("settings exceed maximum nesting depth of %d", MaxMapDepth)
	}
	return nil
}

// ElementMove validates reordering or reparenting an element
func ElementMove(p *domainEntity.PayloadElementMove) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	if p.Order < 0 {
		return fmt.Errorf("order must be non-negative")
	}
	if p.ParentID != nil && *p.ParentID != "" {
		if err := validate.Var(*p.ParentID, "uuid"); err != nil {
			return fmt.Errorf("invalid parentId: %w", err)
		}
	}
	return nil
}

// ElementDelete validates removal of an element
func ElementDelete(p *domainEntity.PayloadElementDelete) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	return nil
}

// ElementContent validates real-time text content updates
func ElementContent(p *domainEntity.PayloadElementContent) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	return nil
}

// Comment validates a new or updated comment
func Comment(p *domainEntity.PayloadComment) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ElementID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	if p.Comment.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

// ElementForCreate validates an Element struct for creation
func ElementForCreate(e *domainEntity.Element) error {
	if e == nil {
		return fmt.Errorf("element is nil")
	}
	// We check core fields for creation
	if e.Type == "" {
		return fmt.Errorf("element type is required")
	}
	if e.PageID == "" {
		return fmt.Errorf("pageId is required")
	}
	if !IsSafeID(e.PageID) {
		return fmt.Errorf("invalid pageId")
	}
	if err := ElementTree(e); err != nil {
		return err
	}
	if mapDepthExceeded(e.Styles, MaxMapDepth) {
		return fmt.Errorf("styles exceed maximum nesting depth of %d", MaxMapDepth)
	}
	if mapDepthExceeded(e.Settings, MaxMapDepth) {
		return fmt.Errorf("settings exceed maximum nesting depth of %d", MaxMapDepth)
	}
	var savedParent *string
	if e.ParentID != nil {
		// Treat an empty string or the special 'root' identifier as 'no parent'
		// for validation purposes so that clients can reference 'root' without
		// failing the UUID validation rule.
		v := strings.TrimSpace(*e.ParentID)
		if v == "" || v == "root" {
			savedParent = e.ParentID
			e.ParentID = nil
		}
	}
	if err := validate.Struct(e); err != nil {
		// restore before returning
		if savedParent != nil {
			e.ParentID = savedParent
		}
		return fmt.Errorf("invalid element for add: %w", err)
	}
	if savedParent != nil {
		e.ParentID = savedParent
	}
	return nil
}

// ElementTree enforces depth (max MaxElementTreeDepth) and total node count (max MaxElementTreeNodes) bounds on an element tree.
func ElementTree(e *domainEntity.Element) error {
	if e == nil {
		return nil
	}
	nodes := 0
	return validateElementNode(e, 1, &nodes)
}

func validateElementNode(e *domainEntity.Element, depth int, nodes *int) error {
	if depth > MaxElementTreeDepth {
		return fmt.Errorf("element tree exceeds maximum depth of %d", MaxElementTreeDepth)
	}
	*nodes++
	if *nodes > MaxElementTreeNodes {
		return fmt.Errorf("element tree exceeds maximum node count of %d", MaxElementTreeNodes)
	}
	for i := range e.Elements {
		if err := validateElementNode(&e.Elements[i], depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func mapDepthExceeded(m map[string]any, maxDepth int) bool {
	if len(m) == 0 {
		return false
	}
	return !withinMaxDepth(m, 1, maxDepth)
}

func withinMaxDepth(m map[string]any, depth, maxDepth int) bool {
	if depth > maxDepth {
		return false
	}
	for _, v := range m {
		if sub, ok := v.(map[string]any); ok {
			if !withinMaxDepth(sub, depth+1, maxDepth) {
				return false
			}
		}
	}
	return true
}

func ElementForUpdate(e *domainEntity.Element) error {
	if e == nil {
		return fmt.Errorf("element is nil")
	}
	if err := validate.Var(e.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	// Partial validation for updates
	if err := validate.StructPartial(e, "Type", "Order"); err != nil {
		return fmt.Errorf("invalid element for update: %w", err)
	}
	return nil
}
