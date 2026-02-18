package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FieldError holds the per-field validation failure detail.
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidationError is a custom error type that carries structured field-level
// validation failures. Fiber's custom ErrorHandler (in configs) detects this
// type and formats it as a 422 JSON response.
type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// NewValidationError converts go-playground/validator ValidationErrors into
// our own ValidationError type, producing human-readable per-field messages.
func NewValidationError(errs validator.ValidationErrors) *ValidationError {
	fields := make([]FieldError, 0, len(errs))
	for _, fe := range errs {
		fields = append(fields, FieldError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: humanizeValidationError(fe),
		})
	}
	return &ValidationError{Fields: fields}
}

// humanizeValidationError turns a single validator.FieldError into a readable
// sentence. Extend the switch as you add more validation tags to your structs.
func humanizeValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", fe.Field())
	case "numeric":
		return fmt.Sprintf("%s must be a numeric value", fe.Field())
	default:
		return fmt.Sprintf("%s failed validation on rule '%s'", fe.Field(), fe.Tag())
	}
}
