package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("slug", validateSlug)
	validate.RegisterValidation("cron", validateCron)
	validate.RegisterValidation("password", validatePassword)
}

func Get() *validator.Validate {
	return validate
}

func Validate(s interface{}) error {
	return validate.Struct(s)
}

func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// Custom validators

func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}

func validateCron(fl validator.FieldLevel) bool {
	cron := fl.Field().String()
	parts := strings.Split(cron, " ")
	return len(parts) == 5 || len(parts) == 6
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// Error formatting
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   toSnakeCase(e.Field()),
				Message: formatMessage(e),
			})
		}
	}

	return errors
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "slug":
		return "Invalid slug format (use lowercase letters, numbers, and hyphens)"
	case "cron":
		return "Invalid cron expression"
	case "password":
		return "Password must be at least 8 characters with uppercase, lowercase, number, and special character"
	case "uuid":
		return "Invalid UUID format"
	default:
		return "Invalid value"
	}
}

func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
