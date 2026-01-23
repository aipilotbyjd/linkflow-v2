package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validator wraps go-playground/validator with custom error messages
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Use JSON tag names for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	// Register custom validator for json.RawMessage
	v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if val, ok := field.Interface().(json.RawMessage); ok {
			if len(val) == 0 {
				return nil // nil fails required validation
			}
			return string(val)
		}
		return nil
	}, json.RawMessage{})

	// Register custom validators
	RegisterCustomValidators(v)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct and returns user-friendly errors
func (v *Validator) Validate(s interface{}) []ValidationError {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Message: v.getMessage(err),
		})
	}
	return errors
}

// getMessage returns a user-friendly message for a validation error
func (v *Validator) getMessage(err validator.FieldError) string {
	field := formatFieldName(err.Field())

	// Fall back to tag-specific message
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "Please enter a valid email address"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", field, err.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, err.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, err.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "containsany":
		return fmt.Sprintf("%s must contain at least one special character", field)
	// Custom validators
	case "slug":
		return fmt.Sprintf("%s must be lowercase alphanumeric with hyphens (e.g., my-workspace)", field)
	case "cron":
		return fmt.Sprintf("%s must be a valid cron expression (e.g., 0 * * * *)", field)
	case "webhook_path":
		return fmt.Sprintf("%s must be a valid URL path (alphanumeric, slashes, hyphens, underscores)", field)
	case "variable_key":
		return fmt.Sprintf("%s must start with a letter or underscore and contain only alphanumeric characters and underscores", field)
	case "hex_color":
		return fmt.Sprintf("%s must be a valid hex color (e.g., #FF5733)", field)
	case "timezone":
		return fmt.Sprintf("%s must be a valid timezone (e.g., America/New_York, UTC)", field)
	case "credential_type":
		return fmt.Sprintf("%s must be a valid credential type (oauth2, api_key, basic, bearer, etc.)", field)
	case "sharing_scope":
		return fmt.Sprintf("%s must be one of: private, workspace, specific", field)
	case "share_permission":
		return fmt.Sprintf("%s must be one of: view, use, edit", field)
	case "workspace_role":
		return fmt.Sprintf("%s must be one of: owner, admin, editor, member, viewer", field)
	case "http_method":
		return fmt.Sprintf("%s must be a valid HTTP method (GET, POST, PUT, PATCH, DELETE)", field)
	case "ai_provider":
		return fmt.Sprintf("%s must be a valid AI provider (openai, anthropic, google, etc.)", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// formatFieldName converts camelCase to human readable
func formatFieldName(field string) string {
	// Common field name mappings
	mappings := map[string]string{
		"first_name":      "First name",
		"last_name":       "Last name",
		"firstName":       "First name",
		"lastName":        "Last name",
		"email":           "Email",
		"password":        "Password",
		"name":            "Name",
		"description":     "Description",
		"workflow_id":     "Workflow ID",
		"workflowId":      "Workflow ID",
		"workspace_id":    "Workspace ID",
		"workspaceId":     "Workspace ID",
		"cron_expression": "Cron expression",
		"cronExpression":  "Cron expression",
		"resource_id":     "Resource ID",
		"resourceId":      "Resource ID",
		"resource_type":   "Resource type",
		"resourceType":    "Resource type",
		"shared_with":     "Shared with email",
		"sharedWithEmail": "Shared with email",
		"url":             "URL",
		"method":          "Method",
		"path":            "Path",
		"type":            "Type",
		"timezone":        "Timezone",
		"permission":      "Permission",
		"code":            "Code",
		"node_id":         "Node ID",
		"nodeId":          "Node ID",
		"node_type":       "Node type",
		"nodeType":        "Node type",
		"role":            "Role",
		"planId":          "Plan ID",
		"plan_id":         "Plan ID",
		"streamName":      "Stream name",
		"score":           "Score",
		"data":            "Data",
		"refresh_token":   "Refresh token",
		"refreshToken":    "Refresh token",
		"token":           "Token",
		"new_password":    "New password",
		"newPassword":     "New password",
		"ids":             "IDs",
		"execution_ids":   "Execution IDs",
		"executionIds":    "Execution IDs",
	}

	if mapped, ok := mappings[field]; ok {
		return mapped
	}

	// Default: capitalize first letter
	if len(field) > 0 {
		return strings.ToUpper(field[:1]) + field[1:]
	}
	return field
}

// Global validator instance
var defaultValidator = New()

// Validate validates using the default validator
func Validate(s interface{}) []ValidationError {
	return defaultValidator.Validate(s)
}

// ValidateAndRespond validates and writes error response if invalid, returns true if valid
func ValidateAndRespond(w interface{ Header() map[string][]string }, s interface{}, writeError func([]ValidationError)) bool {
	errors := defaultValidator.Validate(s)
	if len(errors) > 0 {
		writeError(errors)
		return false
	}
	return true
}
