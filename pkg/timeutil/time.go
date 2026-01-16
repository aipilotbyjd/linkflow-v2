// Package timeutil provides time-related utility functions.
package timeutil

import (
	"fmt"
	"time"
)

// Now returns current time in UTC
func Now() time.Time {
	return time.Now().UTC()
}

// StartOfDay returns the start of the day for the given time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for the given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday) for the given time
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -weekday+1))
}

// EndOfWeek returns the end of the week (Sunday) for the given time
func EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return EndOfDay(t.AddDate(0, 0, 7-weekday))
}

// StartOfMonth returns the start of the month for the given time
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month for the given time
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// StartOfYear returns the start of the year for the given time
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year for the given time
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// DaysAgo returns the time N days ago
func DaysAgo(n int) time.Time {
	return Now().AddDate(0, 0, -n)
}

// DaysFromNow returns the time N days from now
func DaysFromNow(n int) time.Time {
	return Now().AddDate(0, 0, n)
}

// HoursAgo returns the time N hours ago
func HoursAgo(n int) time.Time {
	return Now().Add(-time.Duration(n) * time.Hour)
}

// MinutesAgo returns the time N minutes ago
func MinutesAgo(n int) time.Time {
	return Now().Add(-time.Duration(n) * time.Minute)
}

// DurationSince returns the duration since the given time
func DurationSince(t time.Time) time.Duration {
	return Now().Sub(t)
}

// DurationUntil returns the duration until the given time
func DurationUntil(t time.Time) time.Duration {
	return t.Sub(Now())
}

// IsToday checks if the given time is today
func IsToday(t time.Time) bool {
	now := Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsYesterday checks if the given time is yesterday
func IsYesterday(t time.Time) bool {
	yesterday := Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay()
}

// IsFuture checks if the given time is in the future
func IsFuture(t time.Time) bool {
	return t.After(Now())
}

// IsPast checks if the given time is in the past
func IsPast(t time.Time) bool {
	return t.Before(Now())
}

// FormatRFC3339 formats time as RFC3339
func FormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseRFC3339 parses RFC3339 formatted time
func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// FormatDate formats time as YYYY-MM-DD
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseDate parses YYYY-MM-DD formatted date
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// FormatDateTime formats time as YYYY-MM-DD HH:MM:SS
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseDateTime parses YYYY-MM-DD HH:MM:SS formatted datetime
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

// HumanDuration returns a human-readable duration
func HumanDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// Milliseconds returns the time as Unix milliseconds
func Milliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// FromMilliseconds creates a time from Unix milliseconds
func FromMilliseconds(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
