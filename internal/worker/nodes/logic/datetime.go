package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// DateTimeNode performs date/time operations
type DateTimeNode struct{}

func (n *DateTimeNode) Type() string {
	return "logic.datetime"
}

func (n *DateTimeNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	operation := core.GetString(config, "operation", "now")

	switch operation {
	case "now":
		return n.now(config)
	case "format":
		return n.format(config, input)
	case "parse":
		return n.parse(config, input)
	case "add":
		return n.add(config, input)
	case "subtract":
		return n.subtract(config, input)
	case "diff":
		return n.diff(config, input)
	case "startOf":
		return n.startOf(config, input)
	case "endOf":
		return n.endOf(config, input)
	case "compare":
		return n.compare(config, input)
	case "extract":
		return n.extract(config, input)
	default:
		return n.now(config)
	}
}

func (n *DateTimeNode) now(config map[string]interface{}) (map[string]interface{}, error) {
	now := time.Now()
	timezone := core.GetString(config, "timezone", "UTC")

	if loc, err := time.LoadLocation(timezone); err == nil {
		now = now.In(loc)
	}

	format := core.GetString(config, "format", time.RFC3339)
	goFormat := convertToGoFormat(format)

	return map[string]interface{}{
		"datetime":  now.Format(goFormat),
		"timestamp": now.Unix(),
		"iso":       now.Format(time.RFC3339),
		"date":      now.Format("2006-01-02"),
		"time":      now.Format("15:04:05"),
		"year":      now.Year(),
		"month":     int(now.Month()),
		"day":       now.Day(),
		"hour":      now.Hour(),
		"minute":    now.Minute(),
		"second":    now.Second(),
		"weekday":   now.Weekday().String(),
		"timezone":  timezone,
	}, nil
}

func (n *DateTimeNode) format(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	t, err := parseDateTime(dateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	format := core.GetString(config, "format", time.RFC3339)
	timezone := core.GetString(config, "timezone", "")

	if timezone != "" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			t = t.In(loc)
		}
	}

	goFormat := convertToGoFormat(format)

	return map[string]interface{}{
		"formatted": t.Format(goFormat),
		"timestamp": t.Unix(),
		"iso":       t.Format(time.RFC3339),
	}, nil
}

func (n *DateTimeNode) parse(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	inputFormat := core.GetString(config, "inputFormat", "")

	var t time.Time
	var err error

	if inputFormat != "" {
		goFormat := convertToGoFormat(inputFormat)
		t, err = time.Parse(goFormat, fmt.Sprintf("%v", dateInput))
	} else {
		t, err = parseDateTime(dateInput)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	return map[string]interface{}{
		"datetime":  t.Format(time.RFC3339),
		"timestamp": t.Unix(),
		"year":      t.Year(),
		"month":     int(t.Month()),
		"day":       t.Day(),
		"hour":      t.Hour(),
		"minute":    t.Minute(),
		"second":    t.Second(),
		"weekday":   t.Weekday().String(),
		"valid":     true,
	}, nil
}

func (n *DateTimeNode) add(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	t, err := parseDateTime(dateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	amount := core.GetInt(config, "amount", 0)
	unit := core.GetString(config, "unit", "days")

	t = addDuration(t, amount, unit)

	return map[string]interface{}{
		"datetime":  t.Format(time.RFC3339),
		"timestamp": t.Unix(),
		"iso":       t.Format(time.RFC3339),
		"date":      t.Format("2006-01-02"),
	}, nil
}

func (n *DateTimeNode) subtract(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	t, err := parseDateTime(dateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	amount := core.GetInt(config, "amount", 0)
	unit := core.GetString(config, "unit", "days")

	t = addDuration(t, -amount, unit)

	return map[string]interface{}{
		"datetime":  t.Format(time.RFC3339),
		"timestamp": t.Unix(),
		"iso":       t.Format(time.RFC3339),
		"date":      t.Format("2006-01-02"),
	}, nil
}

func (n *DateTimeNode) diff(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	date1Input := getDateInput(config, input)
	date2Input := config["date2"]
	if date2Input == nil {
		date2Input = input["date2"]
	}

	t1, err := parseDateTime(date1Input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date1: %w", err)
	}

	t2, err := parseDateTime(date2Input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date2: %w", err)
	}

	diff := t2.Sub(t1)

	return map[string]interface{}{
		"milliseconds": diff.Milliseconds(),
		"seconds":      int64(diff.Seconds()),
		"minutes":      int64(diff.Minutes()),
		"hours":        int64(diff.Hours()),
		"days":         int64(diff.Hours() / 24),
		"weeks":        int64(diff.Hours() / (24 * 7)),
	}, nil
}

func (n *DateTimeNode) startOf(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	t, err := parseDateTime(dateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	unit := core.GetString(config, "unit", "day")

	switch unit {
	case "year":
		t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	case "month":
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case "week":
		weekday := int(t.Weekday())
		t = t.AddDate(0, 0, -weekday)
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "day":
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "hour":
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case "minute":
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	}

	return map[string]interface{}{
		"datetime":  t.Format(time.RFC3339),
		"timestamp": t.Unix(),
	}, nil
}

func (n *DateTimeNode) endOf(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	t, err := parseDateTime(dateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	unit := core.GetString(config, "unit", "day")

	switch unit {
	case "year":
		t = time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
	case "month":
		t = time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, t.Location())
	case "week":
		weekday := int(t.Weekday())
		t = t.AddDate(0, 0, 6-weekday)
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
	case "day":
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
	case "hour":
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 999999999, t.Location())
	case "minute":
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 999999999, t.Location())
	}

	return map[string]interface{}{
		"datetime":  t.Format(time.RFC3339),
		"timestamp": t.Unix(),
	}, nil
}

func (n *DateTimeNode) compare(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	date1Input := getDateInput(config, input)
	date2Input := config["date2"]
	if date2Input == nil {
		date2Input = input["date2"]
	}

	t1, err := parseDateTime(date1Input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date1: %w", err)
	}

	t2, err := parseDateTime(date2Input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date2: %w", err)
	}

	return map[string]interface{}{
		"isBefore":      t1.Before(t2),
		"isAfter":       t1.After(t2),
		"isEqual":       t1.Equal(t2),
		"isSameDay":     t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay(),
		"isSameMonth":   t1.Year() == t2.Year() && t1.Month() == t2.Month(),
		"isSameYear":    t1.Year() == t2.Year(),
	}, nil
}

func (n *DateTimeNode) extract(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	dateInput := getDateInput(config, input)
	t, err := parseDateTime(dateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	_, week := t.ISOWeek()

	return map[string]interface{}{
		"year":        t.Year(),
		"month":       int(t.Month()),
		"monthName":   t.Month().String(),
		"day":         t.Day(),
		"hour":        t.Hour(),
		"minute":      t.Minute(),
		"second":      t.Second(),
		"millisecond": t.Nanosecond() / 1000000,
		"weekday":     int(t.Weekday()),
		"weekdayName": t.Weekday().String(),
		"week":        week,
		"dayOfYear":   t.YearDay(),
		"quarter":     (int(t.Month())-1)/3 + 1,
		"isLeapYear":  t.Year()%4 == 0 && (t.Year()%100 != 0 || t.Year()%400 == 0),
		"daysInMonth": time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day(),
	}, nil
}

// Helper functions

func getDateInput(config map[string]interface{}, input map[string]interface{}) interface{} {
	if date := config["date"]; date != nil {
		return date
	}
	if date := input["date"]; date != nil {
		return date
	}
	if date := input["datetime"]; date != nil {
		return date
	}
	if date := input["timestamp"]; date != nil {
		return date
	}
	return time.Now().Format(time.RFC3339)
}

func parseDateTime(input interface{}) (time.Time, error) {
	switch v := input.(type) {
	case time.Time:
		return v, nil
	case int64:
		return time.Unix(v, 0), nil
	case int:
		return time.Unix(int64(v), 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case string:
		// Try common formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"01/02/2006",
			"02/01/2006",
			"Jan 2, 2006",
			"January 2, 2006",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}

		// Try as Unix timestamp
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(ts, 0), nil
		}

		return time.Time{}, fmt.Errorf("unable to parse date: %s", v)
	default:
		return time.Time{}, fmt.Errorf("unsupported date type: %T", input)
	}
}

func addDuration(t time.Time, amount int, unit string) time.Time {
	switch unit {
	case "years", "year":
		return t.AddDate(amount, 0, 0)
	case "months", "month":
		return t.AddDate(0, amount, 0)
	case "weeks", "week":
		return t.AddDate(0, 0, amount*7)
	case "days", "day":
		return t.AddDate(0, 0, amount)
	case "hours", "hour":
		return t.Add(time.Duration(amount) * time.Hour)
	case "minutes", "minute":
		return t.Add(time.Duration(amount) * time.Minute)
	case "seconds", "second":
		return t.Add(time.Duration(amount) * time.Second)
	default:
		return t
	}
}

func convertToGoFormat(format string) string {
	// Convert common format tokens to Go format
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
		"Z":    "Z07:00",
		"A":    "PM",
		"a":    "pm",
	}

	result := format
	for token, goToken := range replacements {
		result = replaceAll(result, token, goToken)
	}
	return result
}

func replaceAll(s, old, new string) string {
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
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

// Note: DateTimeNode is registered in logic/init.go
