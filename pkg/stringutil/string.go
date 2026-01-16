// Package stringutil provides string utility functions.
package stringutil

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if a string is not empty
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// Truncate truncates a string to the specified length
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}

// TruncateWithEllipsis truncates a string and adds ellipsis
func TruncateWithEllipsis(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// ToSlug converts a string to a URL-friendly slug
func ToSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	s = reg.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	words := splitIntoWords(s)
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = strings.Title(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	words := splitIntoWords(s)
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}
	return strings.Join(words, "")
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	words := splitIntoWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}

// ToKebabCase converts a string to kebab-case
func ToKebabCase(s string) string {
	words := splitIntoWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "-")
}

// splitIntoWords splits a string into words
func splitIntoWords(s string) []string {
	var words []string
	var word strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) && word.Len() > 0 {
			words = append(words, word.String())
			word.Reset()
		}

		if r == '_' || r == '-' || r == ' ' {
			if word.Len() > 0 {
				words = append(words, word.String())
				word.Reset()
			}
			continue
		}

		word.WriteRune(r)

		if i == len(s)-1 && word.Len() > 0 {
			words = append(words, word.String())
		}
	}

	return words
}

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// RandomHex generates a random hex string
func RandomHex(length int) string {
	bytes := make([]byte, (length+1)/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// Contains checks if a string contains a substring (case-insensitive)
func Contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ContainsAny checks if a string contains any of the substrings
func ContainsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(lower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// Reverse reverses a string
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Coalesce returns the first non-empty string
func Coalesce(strs ...string) string {
	for _, s := range strs {
		if IsNotEmpty(s) {
			return s
		}
	}
	return ""
}

// DefaultIfEmpty returns the default value if the string is empty
func DefaultIfEmpty(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// MaskString masks part of a string
func MaskString(s string, visibleStart, visibleEnd int, maskChar rune) string {
	if len(s) <= visibleStart+visibleEnd {
		return s
	}

	masked := make([]rune, len(s))
	runes := []rune(s)

	for i := range runes {
		if i < visibleStart || i >= len(runes)-visibleEnd {
			masked[i] = runes[i]
		} else {
			masked[i] = maskChar
		}
	}

	return string(masked)
}

// MaskEmail masks an email address
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local + "@" + domain
	}

	return local[0:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:] + "@" + domain
}

// SplitAndTrim splits a string and trims whitespace from each part
func SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
