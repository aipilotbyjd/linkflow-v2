package validation

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
)

// Custom validation patterns
var (
	slugRegex        = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	webhookPathRegex = regexp.MustCompile(`^[a-zA-Z0-9/_-]+$`)
	variableKeyRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	hexColorRegex    = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
)

// RegisterCustomValidators registers all custom validators on the given validator instance
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("slug", validateSlug)
	v.RegisterValidation("cron", validateCron)
	v.RegisterValidation("webhook_path", validateWebhookPath)
	v.RegisterValidation("variable_key", validateVariableKey)
	v.RegisterValidation("hex_color", validateHexColor)
	v.RegisterValidation("timezone", validateTimezone)
	v.RegisterValidation("credential_type", validateCredentialType)
	v.RegisterValidation("sharing_scope", validateSharingScope)
	v.RegisterValidation("share_permission", validateSharePermission)
	v.RegisterValidation("workspace_role", validateWorkspaceRole)
	v.RegisterValidation("http_method", validateHTTPMethod)
	v.RegisterValidation("ai_provider", validateAIProvider)
}

// validateSlug validates a URL-safe slug (lowercase alphanumeric with hyphens)
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if slug == "" {
		return true // Let required tag handle empty
	}
	return slugRegex.MatchString(slug)
}

// validateCron validates a cron expression
func validateCron(fl validator.FieldLevel) bool {
	expr := fl.Field().String()
	if expr == "" {
		return true // Let required tag handle empty
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	return err == nil
}

// validateWebhookPath validates a webhook path (URL-safe, no path traversal)
func validateWebhookPath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return true // Let required tag handle empty
	}
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return false
	}
	// Check for valid characters
	return webhookPathRegex.MatchString(path)
}

// validateVariableKey validates a variable key (alphanumeric, underscores, starts with letter/underscore)
func validateVariableKey(fl validator.FieldLevel) bool {
	key := fl.Field().String()
	if key == "" {
		return true // Let required tag handle empty
	}
	return variableKeyRegex.MatchString(key)
}

// validateHexColor validates a hex color code
func validateHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	if color == "" {
		return true // Optional
	}
	return hexColorRegex.MatchString(color)
}

// validateTimezone validates a timezone string
func validateTimezone(fl validator.FieldLevel) bool {
	tz := fl.Field().String()
	if tz == "" {
		return true // Optional, will default to UTC
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}

// validateCredentialType validates credential types
func validateCredentialType(fl validator.FieldLevel) bool {
	credType := fl.Field().String()
	if credType == "" {
		return true // Let required tag handle empty
	}
	validTypes := map[string]bool{
		"oauth2":          true,
		"api_key":         true,
		"basic":           true,
		"bearer":          true,
		"custom":          true,
		"aws":             true,
		"gcp":             true,
		"azure":           true,
		"database":        true,
		"ssh":             true,
		"smtp":            true,
		"webhook":         true,
		"certificate":     true,
		"jwt":             true,
		"service_account": true,
	}
	return validTypes[credType]
}

// validateSharingScope validates credential sharing scopes
func validateSharingScope(fl validator.FieldLevel) bool {
	scope := fl.Field().String()
	if scope == "" {
		return true // Will default
	}
	validScopes := map[string]bool{
		"private":   true,
		"workspace": true,
		"specific":  true,
	}
	return validScopes[scope]
}

// validateSharePermission validates share permissions
func validateSharePermission(fl validator.FieldLevel) bool {
	perm := fl.Field().String()
	if perm == "" {
		return true // Let required tag handle empty
	}
	validPerms := map[string]bool{
		"view": true,
		"use":  true,
		"edit": true,
	}
	return validPerms[perm]
}

// validateWorkspaceRole validates workspace member roles
func validateWorkspaceRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	if role == "" {
		return true // Let required tag handle empty
	}
	validRoles := map[string]bool{
		"owner":  true,
		"admin":  true,
		"editor": true,
		"member": true,
		"viewer": true,
	}
	return validRoles[role]
}

// validateHTTPMethod validates HTTP methods
func validateHTTPMethod(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	if method == "" {
		return true // Let required tag handle empty
	}
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	return validMethods[strings.ToUpper(method)]
}

// validateAIProvider validates AI provider identifiers
func validateAIProvider(fl validator.FieldLevel) bool {
	provider := fl.Field().String()
	if provider == "" {
		return true // Let required tag handle empty
	}
	validProviders := map[string]bool{
		"openai":       true,
		"anthropic":    true,
		"google":       true,
		"azure_openai": true,
		"groq":         true,
		"mistral":      true,
		"cohere":       true,
		"together":     true,
	}
	return validProviders[provider]
}
