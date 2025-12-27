package validator

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Common regex patterns
var (
	slugRegex        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	webhookPathRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	expressionRegex  = regexp.MustCompile(`\{\{\s*\$[a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*|\[[^\]]+\])*\s*\}\}`)
	jsonPathRegex    = regexp.MustCompile(`^\$\.?[a-zA-Z0-9_\[\].*@?()]+$`)
	nodeIDRegex      = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func init() {
	validate = validator.New()

	// Register custom validators (errors indicate programmer error, so panic)
	validators := map[string]validator.Func{
		"slug":         validateSlug,
		"cron":         validateCron,
		"password":     validatePassword,
		"timezone":     validateTimezone,
		"webhook_path": validateWebhookPath,
		"expression":   validateExpression,
		"json_path":    validateJSONPath,
		"node_id":      validateNodeID,
		"safe_string":  validateSafeString,
		"url_or_expr":  validateURLOrExpression,
	}

	for name, fn := range validators {
		if err := validate.RegisterValidation(name, fn); err != nil {
			panic("failed to register " + name + " validator: " + err.Error())
		}
	}
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
	return slugRegex.MatchString(slug)
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

// validateTimezone checks if the string is a valid IANA timezone
func validateTimezone(fl validator.FieldLevel) bool {
	tz := fl.Field().String()
	if tz == "" {
		return false
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}

// validateWebhookPath checks if the string is a safe webhook path
func validateWebhookPath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if len(path) < 1 || len(path) > 100 {
		return false
	}
	return webhookPathRegex.MatchString(path)
}

// validateExpression validates n8n-style expressions {{ $json.field }}
func validateExpression(fl validator.FieldLevel) bool {
	expr := fl.Field().String()
	if expr == "" {
		return true // Empty is valid (no expression)
	}
	// Check if it contains valid expression syntax
	return expressionRegex.MatchString(expr)
}

// validateJSONPath validates JSONPath expressions
func validateJSONPath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return true
	}
	return jsonPathRegex.MatchString(path)
}

// validateNodeID validates workflow node IDs
func validateNodeID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	if len(id) < 1 || len(id) > 100 {
		return false
	}
	return nodeIDRegex.MatchString(id)
}

// validateSafeString checks for potentially dangerous characters (XSS prevention)
func validateSafeString(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	// Check for script tags and common XSS patterns
	dangerous := []string{
		"<script", "</script", "javascript:", "onerror=", "onload=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=",
	}
	lower := strings.ToLower(s)
	for _, d := range dangerous {
		if strings.Contains(lower, d) {
			return false
		}
	}
	return true
}

// validateURLOrExpression validates URL or allows expressions
func validateURLOrExpression(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	// Allow expressions
	if strings.Contains(value, "{{") && strings.Contains(value, "}}") {
		return true
	}
	// Otherwise must be valid URL
	_, err := url.ParseRequestURI(value)
	return err == nil
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
	case "timezone":
		return "Invalid timezone (use IANA format like 'America/New_York')"
	case "webhook_path":
		return "Invalid webhook path (use alphanumeric characters, underscores, and hyphens)"
	case "expression":
		return "Invalid expression syntax (use {{ $json.field }} format)"
	case "json_path":
		return "Invalid JSONPath expression"
	case "node_id":
		return "Invalid node ID (use alphanumeric characters, underscores, and hyphens)"
	case "safe_string":
		return "Value contains potentially unsafe content"
	case "url_or_expr":
		return "Invalid URL or expression"
	case "oneof":
		return "Value must be one of: " + e.Param()
	case "url":
		return "Invalid URL format"
	case "len":
		return "Value must be exactly " + e.Param() + " characters"
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
