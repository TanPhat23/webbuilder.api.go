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
	validate  *validator.Validate
	safeIDRe  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
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

// IsValidSafeID reports whether id is a safe identifier (1..128 chars of [A-Za-z0-9_-])
// with no Redis key delimiters, spaces, or path separators.
func IsValidSafeID(id string) bool {
	return safeIDRe.MatchString(id)
}

func ValidatePresence(p *domainEntity.PayloadPresence) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if p.UserID == "" {
		return fmt.Errorf("userId is required")
	}
	return nil
}

// ValidateElementCreate validates the payload for adding a new element to the canvas
func ValidateElementCreate(p *domainEntity.PayloadElementCreate) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	return ValidateElementForAdd(&p.Element)
}

// ValidateElementUpdate validates granular updates to properties or styles
func ValidateElementUpdate(p *domainEntity.PayloadElementUpdate) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	if mapExceedsDepth(p.Styles, MaxMapDepth) {
		return fmt.Errorf("styles exceed maximum nesting depth of %d", MaxMapDepth)
	}
	if mapExceedsDepth(p.Settings, MaxMapDepth) {
		return fmt.Errorf("settings exceed maximum nesting depth of %d", MaxMapDepth)
	}
	return nil
}

// ValidateElementMove validates reordering or reparenting an element
func ValidateElementMove(p *domainEntity.PayloadElementMove) error {
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

// ValidateElementDelete validates removal of an element
func ValidateElementDelete(p *domainEntity.PayloadElementDelete) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	return nil
}

// ValidateElementContent validates real-time text content updates
func ValidateElementContent(p *domainEntity.PayloadElementContent) error {
	if p == nil {
		return fmt.Errorf("payload is nil")
	}
	if err := validate.Var(p.ID, "required,uuid"); err != nil {
		return fmt.Errorf("invalid element id: %w", err)
	}
	return nil
}

// ValidateComment validates a new or updated comment
func ValidateComment(p *domainEntity.PayloadComment) error {
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

// ValidateElementForAdd validates an Element struct for creation
func ValidateElementForAdd(e *domainEntity.Element) error {
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
	if !IsValidSafeID(e.PageID) {
		return fmt.Errorf("invalid pageId")
	}
	nodes := 0
	if err := validateElementTreeBounds(e, 1, &nodes); err != nil {
		return err
	}
	if mapExceedsDepth(e.Styles, MaxMapDepth) {
		return fmt.Errorf("styles exceed maximum nesting depth of %d", MaxMapDepth)
	}
	if mapExceedsDepth(e.Settings, MaxMapDepth) {
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

func validateElementTreeBounds(e *domainEntity.Element, depth int, nodes *int) error {
	if depth > MaxElementTreeDepth {
		return fmt.Errorf("element tree exceeds maximum depth of %d", MaxElementTreeDepth)
	}
	*nodes++
	if *nodes > MaxElementTreeNodes {
		return fmt.Errorf("element tree exceeds maximum node count of %d", MaxElementTreeNodes)
	}
	for i := range e.Elements {
		if err := validateElementTreeBounds(&e.Elements[i], depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func mapExceedsDepth(m map[string]any, maxDepth int) bool {
	if len(m) == 0 {
		return false
	}
	return checkMapDepth(m, 1, maxDepth)
}

func checkMapDepth(m map[string]any, depth, maxDepth int) bool {
	if depth > maxDepth {
		return true
	}
	for _, v := range m {
		if sub, ok := v.(map[string]any); ok {
			if checkMapDepth(sub, depth+1, maxDepth) {
				return true
			}
		}
	}
	return false
}

func ValidateElementForUpdate(e *domainEntity.Element) error {
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
