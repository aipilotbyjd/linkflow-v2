package transform

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// DateTimeNode performs date/time operations
type DateTimeNode struct{}

func NewDateTimeNode() *DateTimeNode {
	return &DateTimeNode{}
}

func (n *DateTimeNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	operation, _ := params["operation"].(string)
	timezone, _ := params["timezone"].(string)
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	switch operation {
	case "now":
		now := time.Now().In(loc)
		return types.JSON{
			"iso":        now.Format(time.RFC3339),
			"unix":       now.Unix(),
			"unix_milli": now.UnixMilli(),
			"year":       now.Year(),
			"month":      int(now.Month()),
			"day":        now.Day(),
			"hour":       now.Hour(),
			"minute":     now.Minute(),
			"second":     now.Second(),
			"weekday":    now.Weekday().String(),
			"success":    true,
		}, nil

	case "parse":
		dateStr, _ := params["date_string"].(string)
		format, _ := params["format"].(string)
		if format == "" {
			format = time.RFC3339
		}
		format = convertFormat(format)

		t, err := time.ParseInLocation(format, dateStr, loc)
		if err != nil {
			return types.JSON{"error": err.Error(), "success": false}, nil
		}
		return types.JSON{
			"iso":     t.Format(time.RFC3339),
			"unix":    t.Unix(),
			"success": true,
		}, nil

	case "format":
		unix := int64(toFloat(params["unix"]))
		format, _ := params["format"].(string)
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		format = convertFormat(format)

		t := time.Unix(unix, 0).In(loc)
		return types.JSON{
			"formatted": t.Format(format),
			"success":   true,
		}, nil

	case "add":
		unix := int64(toFloat(params["unix"]))
		amount := int(toFloat(params["amount"]))
		unit, _ := params["unit"].(string)

		t := time.Unix(unix, 0).In(loc)
		switch unit {
		case "seconds":
			t = t.Add(time.Duration(amount) * time.Second)
		case "minutes":
			t = t.Add(time.Duration(amount) * time.Minute)
		case "hours":
			t = t.Add(time.Duration(amount) * time.Hour)
		case "days":
			t = t.AddDate(0, 0, amount)
		case "months":
			t = t.AddDate(0, amount, 0)
		case "years":
			t = t.AddDate(amount, 0, 0)
		}
		return types.JSON{
			"iso":     t.Format(time.RFC3339),
			"unix":    t.Unix(),
			"success": true,
		}, nil

	case "diff":
		unix1 := int64(toFloat(params["unix1"]))
		unix2 := int64(toFloat(params["unix2"]))
		unit, _ := params["unit"].(string)

		t1 := time.Unix(unix1, 0)
		t2 := time.Unix(unix2, 0)
		diff := t2.Sub(t1)

		var result float64
		switch unit {
		case "seconds":
			result = diff.Seconds()
		case "minutes":
			result = diff.Minutes()
		case "hours":
			result = diff.Hours()
		case "days":
			result = diff.Hours() / 24
		default:
			result = diff.Seconds()
		}
		return types.JSON{
			"difference": result,
			"unit":       unit,
			"success":    true,
		}, nil

	case "start_of":
		unix := int64(toFloat(params["unix"]))
		unit, _ := params["unit"].(string)

		t := time.Unix(unix, 0).In(loc)
		switch unit {
		case "day":
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		case "week":
			weekday := int(t.Weekday())
			t = time.Date(t.Year(), t.Month(), t.Day()-weekday, 0, 0, 0, 0, loc)
		case "month":
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		case "year":
			t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
		}
		return types.JSON{
			"iso":     t.Format(time.RFC3339),
			"unix":    t.Unix(),
			"success": true,
		}, nil

	default:
		return types.JSON{"error": "unknown operation", "success": false}, nil
	}
}

func (n *DateTimeNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.datetime",
		Name:        "Date/Time",
		Description: "Parse, format, and manipulate dates and times",
		Category:    "transform",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "options", Description: "Date operation", Required: true, Options: []wtypes.ParamOption{
				{Value: "now", Name: "Current Time"},
				{Value: "parse", Name: "Parse String"},
				{Value: "format", Name: "Format Date"},
				{Value: "add", Name: "Add Duration"},
				{Value: "diff", Name: "Difference"},
				{Value: "start_of", Name: "Start Of Period"},
			}},
			{Name: "timezone", Type: "string", Description: "Timezone", Default: "UTC"},
			{Name: "date_string", Type: "string", Description: "Date string to parse"},
			{Name: "format", Type: "string", Description: "Date format"},
			{Name: "unix", Type: "number", Description: "Unix timestamp"},
			{Name: "unix1", Type: "number", Description: "First timestamp"},
			{Name: "unix2", Type: "number", Description: "Second timestamp"},
			{Name: "amount", Type: "number", Description: "Amount to add"},
			{Name: "unit", Type: "options", Description: "Time unit", Options: []wtypes.ParamOption{
				{Value: "seconds", Name: "Seconds"},
				{Value: "minutes", Name: "Minutes"},
				{Value: "hours", Name: "Hours"},
				{Value: "days", Name: "Days"},
				{Value: "months", Name: "Months"},
				{Value: "years", Name: "Years"},
			}},
		},
	}
}

// Convert common format tokens to Go format
func convertFormat(format string) string {
	replacements := map[string]string{
		"YYYY": "2006",
		"YY":   "06",
		"MM":   "01",
		"DD":   "02",
		"HH":   "15",
		"hh":   "03",
		"mm":   "04",
		"ss":   "05",
		"SSS":  "000",
	}
	for k, v := range replacements {
		format = replaceAll(format, k, v)
	}
	return format
}

func replaceAll(s, old, new string) string {
	for {
		i := indexOf(s, old)
		if i < 0 {
			break
		}
		s = s[:i] + new + s[i+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
