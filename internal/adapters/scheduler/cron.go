package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// CronParser parses cron expressions
type CronParser struct {
	parser cron.Parser
}

// NewCronParser creates a new cron parser with standard options
func NewCronParser() *CronParser {
	return &CronParser{
		parser: cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		),
	}
}

// Parse parses a cron expression and returns a schedule
func (p *CronParser) Parse(expression string) (cron.Schedule, error) {
	return p.parser.Parse(expression)
}

// NextRun returns the next run time for the given expression from now
func (p *CronParser) NextRun(expression string) (time.Time, error) {
	return p.NextRunFrom(expression, time.Now())
}

// NextRunFrom returns the next run time for the given expression from the specified time
func (p *CronParser) NextRunFrom(expression string, from time.Time) (time.Time, error) {
	schedule, err := p.parser.Parse(expression)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	return schedule.Next(from), nil
}

// NextRunInTimezone returns the next run time in the specified timezone
func (p *CronParser) NextRunInTimezone(expression, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %w", err)
	}

	now := time.Now().In(loc)
	schedule, err := p.parser.Parse(expression)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}

	return schedule.Next(now), nil
}

// Validate checks if a cron expression is valid
func (p *CronParser) Validate(expression string) error {
	_, err := p.parser.Parse(expression)
	return err
}

// Describe returns a human-readable description of the cron expression
func (p *CronParser) Describe(expression string) string {
	parts := strings.Fields(expression)
	if len(parts) != 5 {
		return "Invalid cron expression"
	}

	// Handle special cases
	switch expression {
	case "* * * * *":
		return "Every minute"
	case "0 * * * *":
		return "Every hour"
	case "0 0 * * *":
		return "Every day at midnight"
	case "0 0 * * 0":
		return "Every Sunday at midnight"
	case "0 0 1 * *":
		return "First day of every month at midnight"
	}

	// Build description from parts
	minute, hour, dom, month, dow := parts[0], parts[1], parts[2], parts[3], parts[4]

	var desc []string

	// Minutes
	if minute == "*" {
		desc = append(desc, "every minute")
	} else if minute == "0" {
		// Skip
	} else {
		desc = append(desc, fmt.Sprintf("at minute %s", minute))
	}

	// Hours
	if hour == "*" {
		desc = append(desc, "every hour")
	} else if hour != "*" && hour != "0" {
		desc = append(desc, fmt.Sprintf("at hour %s", hour))
	}

	// Day of month
	if dom != "*" {
		desc = append(desc, fmt.Sprintf("on day %s", dom))
	}

	// Month
	if month != "*" {
		desc = append(desc, fmt.Sprintf("in month %s", month))
	}

	// Day of week
	if dow != "*" {
		days := map[string]string{
			"0": "Sunday", "1": "Monday", "2": "Tuesday",
			"3": "Wednesday", "4": "Thursday", "5": "Friday", "6": "Saturday",
		}
		if day, ok := days[dow]; ok {
			desc = append(desc, fmt.Sprintf("on %s", day))
		}
	}

	if len(desc) == 0 {
		return "Custom schedule"
	}

	return strings.Join(desc, ", ")
}

// Common cron expressions
var CommonSchedules = map[string]string{
	"every_minute":     "* * * * *",
	"every_5_minutes":  "*/5 * * * *",
	"every_15_minutes": "*/15 * * * *",
	"every_30_minutes": "*/30 * * * *",
	"every_hour":       "0 * * * *",
	"every_day":        "0 0 * * *",
	"every_week":       "0 0 * * 0",
	"every_month":      "0 0 1 * *",
	"weekdays":         "0 9 * * 1-5",
	"weekends":         "0 9 * * 0,6",
}
