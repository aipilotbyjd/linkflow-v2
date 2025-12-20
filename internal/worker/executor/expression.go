package executor

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type ExpressionEvaluator struct {
	program *vm.Program
}

var expressionRegex = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

type ExpressionContext struct {
	Input       interface{}            `expr:"$input"`
	JSON        interface{}            `expr:"$json"`
	Node        map[string]interface{} `expr:"$node"`
	Vars        map[string]interface{} `expr:"$vars"`
	Env         map[string]string      `expr:"$env"`
	Now         time.Time              `expr:"$now"`
	Today       string                 `expr:"$today"`
	Timestamp   int64                  `expr:"$timestamp"`
	ExecutionID string                 `expr:"$executionId"`
	WorkflowID  string                 `expr:"$workflowId"`
}

func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{}
}

func (e *ExpressionEvaluator) Evaluate(template string, ctx *ExpressionContext) (interface{}, error) {
	if !strings.Contains(template, "{{") {
		return template, nil
	}

	env := buildExpressionEnv(ctx)

	result := expressionRegex.ReplaceAllStringFunc(template, func(match string) string {
		expr := expressionRegex.FindStringSubmatch(match)[1]
		val, err := e.evaluateExpression(expr, env)
		if err != nil {
			return match
		}
		return fmt.Sprintf("%v", val)
	})

	if result == template {
		return e.evaluateExpression(strings.Trim(template, "{}"), env)
	}

	return result, nil
}

func (e *ExpressionEvaluator) evaluateExpression(expression string, env map[string]interface{}) (interface{}, error) {
	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("runtime error: %w", err)
	}

	return result, nil
}

func buildExpressionEnv(ctx *ExpressionContext) map[string]interface{} {
	env := map[string]interface{}{
		"$input":       ctx.Input,
		"$json":        ctx.JSON,
		"$node":        ctx.Node,
		"$vars":        ctx.Vars,
		"$env":         ctx.Env,
		"$now":         ctx.Now,
		"$today":       ctx.Today,
		"$timestamp":   ctx.Timestamp,
		"$executionId": ctx.ExecutionID,
		"$workflowId":  ctx.WorkflowID,

		// String functions
		"uppercase":  strings.ToUpper,
		"lowercase":  strings.ToLower,
		"trim":       strings.TrimSpace,
		"trimStart":  func(s, cutset string) string { return strings.TrimLeft(s, cutset) },
		"trimEnd":    func(s, cutset string) string { return strings.TrimRight(s, cutset) },
		"split":      strings.Split,
		"join":       strings.Join,
		"replace":    strings.ReplaceAll,
		"replaceOne": func(s, old, new string) string { return strings.Replace(s, old, new, 1) },
		"contains":   strings.Contains,
		"startsWith": strings.HasPrefix,
		"endsWith":   strings.HasSuffix,
		"substring": func(s string, start, end int) string {
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"length": func(v interface{}) int {
			switch val := v.(type) {
			case string:
				return len(val)
			case []interface{}:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},
		"padStart": func(s string, length int, pad string) string {
			for len(s) < length {
				s = pad + s
			}
			return s
		},
		"padEnd": func(s string, length int, pad string) string {
			for len(s) < length {
				s = s + pad
			}
			return s
		},
		"repeat": strings.Repeat,
		"reverse": func(s string) string {
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return string(runes)
		},
		"slug": func(s string) string {
			s = strings.ToLower(s)
			s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
			return strings.Trim(s, "-")
		},
		"camelCase": toCamelCase,
		"snakeCase": toSnakeCase,
		"kebabCase": toKebabCase,

		// Array functions
		"first": func(arr []interface{}) interface{} {
			if len(arr) > 0 {
				return arr[0]
			}
			return nil
		},
		"last": func(arr []interface{}) interface{} {
			if len(arr) > 0 {
				return arr[len(arr)-1]
			}
			return nil
		},
		"nth": func(arr []interface{}, n int) interface{} {
			if n >= 0 && n < len(arr) {
				return arr[n]
			}
			return nil
		},
		"slice": func(arr []interface{}, start, end int) []interface{} {
			if start < 0 {
				start = 0
			}
			if end > len(arr) {
				end = len(arr)
			}
			return arr[start:end]
		},
		"concat": func(arrs ...[]interface{}) []interface{} {
			var result []interface{}
			for _, arr := range arrs {
				result = append(result, arr...)
			}
			return result
		},
		"flatten": flattenArray,
		"unique": func(arr []interface{}) []interface{} {
			seen := make(map[string]bool)
			var result []interface{}
			for _, v := range arr {
				key := fmt.Sprintf("%v", v)
				if !seen[key] {
					seen[key] = true
					result = append(result, v)
				}
			}
			return result
		},
		"compact": func(arr []interface{}) []interface{} {
			var result []interface{}
			for _, v := range arr {
				if v != nil && v != "" && v != false {
					result = append(result, v)
				}
			}
			return result
		},
		"includes": func(arr []interface{}, val interface{}) bool {
			for _, v := range arr {
				if v == val {
					return true
				}
			}
			return false
		},
		"indexOf": func(arr []interface{}, val interface{}) int {
			for i, v := range arr {
				if v == val {
					return i
				}
			}
			return -1
		},
		"pluck": func(arr []interface{}, key string) []interface{} {
			var result []interface{}
			for _, item := range arr {
				if obj, ok := item.(map[string]interface{}); ok {
					result = append(result, obj[key])
				}
			}
			return result
		},

		// Math functions
		"round":   math.Round,
		"floor":   math.Floor,
		"ceil":    math.Ceil,
		"abs":     math.Abs,
		"min":     math.Min,
		"max":     math.Max,
		"pow":     math.Pow,
		"sqrt":    math.Sqrt,
		"log":     math.Log,
		"log10":   math.Log10,
		"exp":     math.Exp,
		"sin":     math.Sin,
		"cos":     math.Cos,
		"tan":     math.Tan,
		"random":  func() float64 { return float64(time.Now().UnixNano()%1000) / 1000 },
		"randInt": func(min, max int) int { return min + int(time.Now().UnixNano()%int64(max-min+1)) },
		"sum": func(arr []interface{}) float64 {
			var sum float64
			for _, v := range arr {
				sum += toFloat(v)
			}
			return sum
		},
		"avg": func(arr []interface{}) float64 {
			if len(arr) == 0 {
				return 0
			}
			var sum float64
			for _, v := range arr {
				sum += toFloat(v)
			}
			return sum / float64(len(arr))
		},
		"minArr": func(arr []interface{}) float64 {
			if len(arr) == 0 {
				return 0
			}
			min := toFloat(arr[0])
			for _, v := range arr[1:] {
				if f := toFloat(v); f < min {
					min = f
				}
			}
			return min
		},
		"maxArr": func(arr []interface{}) float64 {
			if len(arr) == 0 {
				return 0
			}
			max := toFloat(arr[0])
			for _, v := range arr[1:] {
				if f := toFloat(v); f > max {
					max = f
				}
			}
			return max
		},

		// Date functions
		"now":       func() time.Time { return time.Now() },
		"today":     func() string { return time.Now().Format("2006-01-02") },
		"timestamp": func() int64 { return time.Now().Unix() },
		"formatDate": func(t time.Time, format string) string {
			return t.Format(convertDateFormat(format))
		},
		"parseDate": func(s, format string) time.Time {
			t, _ := time.Parse(convertDateFormat(format), s)
			return t
		},
		"addDays": func(t time.Time, days int) time.Time {
			return t.AddDate(0, 0, days)
		},
		"addMonths": func(t time.Time, months int) time.Time {
			return t.AddDate(0, months, 0)
		},
		"addYears": func(t time.Time, years int) time.Time {
			return t.AddDate(years, 0, 0)
		},
		"addHours": func(t time.Time, hours int) time.Time {
			return t.Add(time.Duration(hours) * time.Hour)
		},
		"addMinutes": func(t time.Time, mins int) time.Time {
			return t.Add(time.Duration(mins) * time.Minute)
		},
		"diffDays": func(t1, t2 time.Time) int {
			return int(t1.Sub(t2).Hours() / 24)
		},
		"startOfDay": func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		},
		"endOfDay": func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		},
		"startOfMonth": func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		},
		"endOfMonth": func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, t.Location())
		},

		// Object functions
		"keys": func(obj map[string]interface{}) []string {
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			return keys
		},
		"values": func(obj map[string]interface{}) []interface{} {
			values := make([]interface{}, 0, len(obj))
			for _, v := range obj {
				values = append(values, v)
			}
			return values
		},
		"entries": func(obj map[string]interface{}) [][]interface{} {
			entries := make([][]interface{}, 0, len(obj))
			for k, v := range obj {
				entries = append(entries, []interface{}{k, v})
			}
			return entries
		},
		"merge": func(objs ...map[string]interface{}) map[string]interface{} {
			result := make(map[string]interface{})
			for _, obj := range objs {
				for k, v := range obj {
					result[k] = v
				}
			}
			return result
		},
		"pick": func(obj map[string]interface{}, keys ...string) map[string]interface{} {
			result := make(map[string]interface{})
			for _, k := range keys {
				if v, ok := obj[k]; ok {
					result[k] = v
				}
			}
			return result
		},
		"omit": func(obj map[string]interface{}, keys ...string) map[string]interface{} {
			result := make(map[string]interface{})
			keySet := make(map[string]bool)
			for _, k := range keys {
				keySet[k] = true
			}
			for k, v := range obj {
				if !keySet[k] {
					result[k] = v
				}
			}
			return result
		},
		"get": func(obj interface{}, path string) interface{} {
			return getNestedValue(obj, path)
		},
		"set": func(obj map[string]interface{}, path string, value interface{}) map[string]interface{} {
			setNestedValue(obj, path, value)
			return obj
		},
		"has": func(obj map[string]interface{}, key string) bool {
			_, ok := obj[key]
			return ok
		},

		// Type functions
		"isEmpty": func(v interface{}) bool {
			if v == nil {
				return true
			}
			switch val := v.(type) {
			case string:
				return val == ""
			case []interface{}:
				return len(val) == 0
			case map[string]interface{}:
				return len(val) == 0
			}
			return false
		},
		"isNumber": func(v interface{}) bool {
			switch v.(type) {
			case int, int64, float64:
				return true
			}
			return false
		},
		"isString": func(v interface{}) bool {
			_, ok := v.(string)
			return ok
		},
		"isArray": func(v interface{}) bool {
			_, ok := v.([]interface{})
			return ok
		},
		"isObject": func(v interface{}) bool {
			_, ok := v.(map[string]interface{})
			return ok
		},
		"isBoolean": func(v interface{}) bool {
			_, ok := v.(bool)
			return ok
		},
		"isNull": func(v interface{}) bool {
			return v == nil
		},
		"typeof": func(v interface{}) string {
			switch v.(type) {
			case nil:
				return "null"
			case bool:
				return "boolean"
			case int, int64, float64:
				return "number"
			case string:
				return "string"
			case []interface{}:
				return "array"
			case map[string]interface{}:
				return "object"
			default:
				return "unknown"
			}
		},

		// Conversion functions
		"toString": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},
		"toNumber": func(v interface{}) float64 {
			return toFloat(v)
		},
		"toInt": func(v interface{}) int {
			return int(toFloat(v))
		},
		"toBoolean": func(v interface{}) bool {
			switch val := v.(type) {
			case bool:
				return val
			case string:
				return val != "" && val != "false" && val != "0"
			case int, int64, float64:
				return toFloat(v) != 0
			}
			return v != nil
		},
		"toJSON": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"fromJSON": func(s string) interface{} {
			var v interface{}
			json.Unmarshal([]byte(s), &v)
			return v
		},

		// Utility functions
		"ifEmpty": func(v, defaultVal interface{}) interface{} {
			if v == nil || v == "" {
				return defaultVal
			}
			return v
		},
		"coalesce": func(vals ...interface{}) interface{} {
			for _, v := range vals {
				if v != nil && v != "" {
					return v
				}
			}
			return nil
		},
		"ternary": func(cond bool, trueVal, falseVal interface{}) interface{} {
			if cond {
				return trueVal
			}
			return falseVal
		},
		"uuid": generateUUID,
		"base64Encode": func(s string) string {
			return base64Encode(s)
		},
		"base64Decode": func(s string) string {
			return base64Decode(s)
		},
		"urlEncode": func(s string) string {
			return urlEncode(s)
		},
		"urlDecode": func(s string) string {
			return urlDecode(s)
		},
		"hash": func(s, algo string) string {
			return hashString(s, algo)
		},
	}

	return env
}

// Helper functions
func toFloat(v interface{}) float64 {
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

func flattenArray(arr []interface{}) []interface{} {
	var result []interface{}
	for _, v := range arr {
		if nested, ok := v.([]interface{}); ok {
			result = append(result, flattenArray(nested)...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func getNestedValue(obj interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := obj
	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}

func setNestedValue(obj map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := obj
	for i, part := range parts[:len(parts)-1] {
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			current[part] = make(map[string]interface{})
			current = current[part].(map[string]interface{})
		}
		_ = i
	}
	current[parts[len(parts)-1]] = value
}

func toCamelCase(s string) string {
	words := regexp.MustCompile(`[^a-zA-Z0-9]+`).Split(s, -1)
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
	}
	if len(words[0]) > 0 {
		words[0] = strings.ToLower(words[0])
	}
	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

func convertDateFormat(format string) string {
	replacements := map[string]string{
		"YYYY": "2006", "YY": "06",
		"MM": "01", "M": "1",
		"DD": "02", "D": "2",
		"HH": "15", "H": "15",
		"hh": "03", "h": "3",
		"mm": "04", "m": "4",
		"ss": "05", "s": "5",
		"SSS": "000",
		"A": "PM", "a": "pm",
	}
	for k, v := range replacements {
		format = strings.ReplaceAll(format, k, v)
	}
	return format
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func base64Decode(s string) string {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	return string(decoded)
}

func urlEncode(s string) string {
	return url.QueryEscape(s)
}

func urlDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}

func hashString(s, algo string) string {
	switch strings.ToLower(algo) {
	case "md5":
		h := md5.Sum([]byte(s))
		return hex.EncodeToString(h[:])
	case "sha1":
		h := sha1.Sum([]byte(s))
		return hex.EncodeToString(h[:])
	case "sha256":
		h := sha256.Sum256([]byte(s))
		return hex.EncodeToString(h[:])
	case "sha512":
		h := sha512.Sum512([]byte(s))
		return hex.EncodeToString(h[:])
	default:
		h := sha256.Sum256([]byte(s))
		return hex.EncodeToString(h[:])
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
