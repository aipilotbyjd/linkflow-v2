package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// GetString extracts string from config map with default
func GetString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

// GetInt extracts int from config map with default
func GetInt(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	return defaultVal
}

// GetFloat extracts float64 from config map with default
func GetFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return defaultVal
}

// GetBool extracts bool from config map with default
func GetBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultVal
}

// GetMap extracts map from config with default empty map
func GetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}

// GetArray extracts array from config with default empty array
func GetArray(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key].([]interface{}); ok {
		return v
	}
	return []interface{}{}
}

// GetStringArray extracts string array from config
func GetStringArray(m map[string]interface{}, key string) []string {
	arr := GetArray(m, key)
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// ToFloat converts interface to float64
func ToFloat(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

// ToBool converts interface to bool
func ToBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && val != "false" && val != "0"
	case int, int64, float64:
		return ToFloat(v) != 0
	}
	return v != nil
}

// IsEmpty checks if value is empty
func IsEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	}
	return false
}

// GetNestedValue retrieves nested value from map using dot notation
func GetNestedValue(data interface{}, path string) interface{} {
	path = strings.TrimPrefix(path, "$json.")
	path = strings.TrimPrefix(path, "$input.")
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}

// ResolveValue resolves template expressions like {{$json.field}}
func ResolveValue(value interface{}, input map[string]interface{}) interface{} {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return value
	}
	if strings.HasPrefix(str, "{{") && strings.HasSuffix(str, "}}") {
		path := strings.TrimSpace(str[2 : len(str)-2])
		return GetNestedValue(input, path)
	}
	return value
}

// FormatString formats a string with named parameters from a map
func FormatString(template string, params map[string]interface{}) string {
	result := template
	for k, v := range params {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}
	return result
}

// MergeMap merges multiple maps into one
func MergeMap(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// CopyMap creates a shallow copy of a map
func CopyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
