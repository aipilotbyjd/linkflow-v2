package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validator provides validation functionality
type Validator struct {
	errors []ValidationError
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// New creates a new validator
func New() *Validator {
	return &Validator{errors: make([]ValidationError, 0)}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() []ValidationError {
	return v.errors
}

// Error returns the first error message
func (v *Validator) Error() string {
	if len(v.errors) == 0 {
		return ""
	}
	return v.errors[0].Error()
}

// AddError adds a validation error
func (v *Validator) AddError(field, message, code string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// Required validates that a string is not empty
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required", "required")
	}
	return v
}

// RequiredIf validates that a string is not empty if condition is true
func (v *Validator) RequiredIf(field, value string, condition bool) *Validator {
	if condition && strings.TrimSpace(value) == "" {
		v.AddError(field, "is required", "required")
	}
	return v
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, fmt.Sprintf("must be at least %d characters", min), "min_length")
	}
	return v
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("must be at most %d characters", max), "max_length")
	}
	return v
}

// Email validates an email address
func (v *Validator) Email(field, value string) *Validator {
	if value == "" {
		return v
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		v.AddError(field, "must be a valid email address", "email")
	}
	return v
}

// URL validates a URL
func (v *Validator) URL(field, value string) *Validator {
	if value == "" {
		return v
	}
	_, err := url.ParseRequestURI(value)
	if err != nil {
		v.AddError(field, "must be a valid URL", "url")
	}
	return v
}

// UUID validates a UUID
func (v *Validator) UUID(field, value string) *Validator {
	if value == "" {
		return v
	}
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidPattern.MatchString(value) {
		v.AddError(field, "must be a valid UUID", "uuid")
	}
	return v
}

// Pattern validates against a regex pattern
func (v *Validator) Pattern(field, value, pattern, message string) *Validator {
	if value == "" {
		return v
	}
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.AddError(field, message, "pattern")
	}
	return v
}

// In validates that value is in allowed list
func (v *Validator) In(field, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.AddError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")), "in")
	return v
}

// Min validates minimum numeric value
func (v *Validator) Min(field string, value, min int) *Validator {
	if value < min {
		v.AddError(field, fmt.Sprintf("must be at least %d", min), "min")
	}
	return v
}

// Max validates maximum numeric value
func (v *Validator) Max(field string, value, max int) *Validator {
	if value > max {
		v.AddError(field, fmt.Sprintf("must be at most %d", max), "max")
	}
	return v
}

// Between validates numeric value is between min and max
func (v *Validator) Between(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("must be between %d and %d", min, max), "between")
	}
	return v
}

// Password validates password strength
func (v *Validator) Password(field, value string) *Validator {
	if value == "" {
		return v
	}

	if len(value) < 8 {
		v.AddError(field, "must be at least 8 characters", "password_length")
		return v
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(value)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(value)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(value)

	if !hasUpper || !hasLower || !hasNumber {
		v.AddError(field, "must contain uppercase, lowercase, and number", "password_strength")
	}
	return v
}

// Slug validates a URL-safe slug
func (v *Validator) Slug(field, value string) *Validator {
	if value == "" {
		return v
	}
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	if !slugPattern.MatchString(value) {
		v.AddError(field, "must be a valid slug (lowercase letters, numbers, hyphens)", "slug")
	}
	return v
}

// CronExpression validates a cron expression
func (v *Validator) CronExpression(field, value string) *Validator {
	if value == "" {
		return v
	}
	parts := strings.Fields(value)
	if len(parts) != 5 && len(parts) != 6 {
		v.AddError(field, "must be a valid cron expression", "cron")
	}
	return v
}

// Validate is a helper that returns error if validation failed
func Validate(fn func(*Validator)) error {
	v := New()
	fn(v)
	if v.HasErrors() {
		return &ValidationErrors{Errors: v.Errors()}
	}
	return nil
}

// ValidationErrors wraps multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}
