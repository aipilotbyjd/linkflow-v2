package validation

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidUUID     = errors.New("invalid UUID format")
	ErrInvalidSlug     = errors.New("invalid slug format (use lowercase letters, numbers, and hyphens)")
	ErrFieldRequired   = errors.New("field is required")
	ErrFieldTooLong    = errors.New("field exceeds maximum length")
	ErrFieldTooShort   = errors.New("field is below minimum length")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrFieldRequired
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrFieldRequired
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidUUID
	}
	return nil
}

// ValidateSlug validates a URL-friendly slug
func ValidateSlug(slug string) error {
	if strings.TrimSpace(slug) == "" {
		return ErrFieldRequired
	}
	if !slugRegex.MatchString(slug) {
		return ErrInvalidSlug
	}
	return nil
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(value string, min, max int) error {
	length := len(strings.TrimSpace(value))
	if min > 0 && length < min {
		return ErrFieldTooShort
	}
	if max > 0 && length > max {
		return ErrFieldTooLong
	}
	return nil
}

// ValidateRequired checks if a string field is not empty
func ValidateRequired(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrFieldRequired
	}
	return nil
}

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field string, err error) error {
	return ValidationError{
		Field:   field,
		Message: err.Error(),
	}
}
