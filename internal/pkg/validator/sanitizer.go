package validator

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// Sanitizer provides input sanitization utilities
type Sanitizer struct{}

// NewSanitizer creates a new sanitizer instance
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

var (
	// Patterns for sanitization
	multiSpaceRegex    = regexp.MustCompile(`\s+`)
	htmlTagRegex       = regexp.MustCompile(`<[^>]*>`)
	sqlInjectionRegex  = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|truncate|exec|execute|xp_|sp_|0x)`)
	scriptTagRegex     = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	eventHandlerRegex  = regexp.MustCompile(`(?i)\s+on\w+\s*=`)
	nullByteRegex      = regexp.MustCompile(`\x00`)
)

// SanitizeString performs basic string sanitization
func (s *Sanitizer) SanitizeString(input string) string {
	// Remove null bytes
	input = nullByteRegex.ReplaceAllString(input, "")
	// Trim whitespace
	input = strings.TrimSpace(input)
	// Normalize multiple spaces to single space
	input = multiSpaceRegex.ReplaceAllString(input, " ")
	return input
}

// SanitizeHTML escapes HTML entities
func (s *Sanitizer) SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// StripHTML removes all HTML tags
func (s *Sanitizer) StripHTML(input string) string {
	return htmlTagRegex.ReplaceAllString(input, "")
}

// SanitizeForDisplay sanitizes for safe display (HTML escape + strip dangerous tags)
func (s *Sanitizer) SanitizeForDisplay(input string) string {
	// Remove script tags first
	input = scriptTagRegex.ReplaceAllString(input, "")
	// Remove event handlers
	input = eventHandlerRegex.ReplaceAllString(input, " ")
	// Escape remaining HTML
	return html.EscapeString(input)
}

// SanitizeName sanitizes a name field (removes special characters, preserves spaces)
func (s *Sanitizer) SanitizeName(input string) string {
	input = s.SanitizeString(input)
	var result strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) || r == '-' || r == '\'' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SanitizeSlug sanitizes for URL slug (lowercase, alphanumeric, hyphens)
func (s *Sanitizer) SanitizeSlug(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	var result strings.Builder
	lastWasHyphen := false
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
			lastWasHyphen = false
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			if !lastWasHyphen && result.Len() > 0 {
				result.WriteRune('-')
				lastWasHyphen = true
			}
		}
	}
	slug := result.String()
	return strings.Trim(slug, "-")
}

// SanitizeEmail sanitizes email address
func (s *Sanitizer) SanitizeEmail(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	// Remove any characters that shouldn't be in an email
	var result strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '@' || r == '.' || r == '_' || r == '-' || r == '+' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SanitizeURL sanitizes URL input
func (s *Sanitizer) SanitizeURL(input string) string {
	input = strings.TrimSpace(input)
	// Remove javascript: and data: schemes
	lower := strings.ToLower(input)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "vbscript:") {
		return ""
	}
	return input
}

// ContainsSQLInjection checks if input contains potential SQL injection
func (s *Sanitizer) ContainsSQLInjection(input string) bool {
	return sqlInjectionRegex.MatchString(input)
}

// SanitizeJSON sanitizes JSON string values
func (s *Sanitizer) SanitizeJSON(input string) string {
	// Escape backslashes and quotes
	input = strings.ReplaceAll(input, "\\", "\\\\")
	input = strings.ReplaceAll(input, "\"", "\\\"")
	// Remove control characters except newline and tab
	var result strings.Builder
	for _, r := range input {
		if r == '\n' || r == '\t' || r >= 32 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SanitizeFilename sanitizes filename
func (s *Sanitizer) SanitizeFilename(input string) string {
	input = strings.TrimSpace(input)
	// Remove path separators and dangerous characters
	dangerous := []string{"/", "\\", "..", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, d := range dangerous {
		input = strings.ReplaceAll(input, d, "_")
	}
	// Limit length
	if len(input) > 255 {
		input = input[:255]
	}
	return input
}

// Package-level convenience functions

// Sanitize is a convenience function for basic string sanitization
func Sanitize(input string) string {
	return NewSanitizer().SanitizeString(input)
}

// SanitizeHTML is a convenience function for HTML escaping
func SanitizeHTML(input string) string {
	return NewSanitizer().SanitizeHTML(input)
}

// StripHTML is a convenience function for removing HTML tags
func StripHTML(input string) string {
	return NewSanitizer().StripHTML(input)
}

// SanitizeSlug is a convenience function for slug sanitization
func SanitizeSlug(input string) string {
	return NewSanitizer().SanitizeSlug(input)
}

// IsSQLInjection checks for SQL injection patterns
func IsSQLInjection(input string) bool {
	return NewSanitizer().ContainsSQLInjection(input)
}
