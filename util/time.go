// Package util provides shared utility functions for mini-kanban.
package util

import (
	"fmt"
	"time"
)

// DateTimeFormats are the supported date/time formats for parsing.
var DateTimeFormats = []string{
	"2006-01-02T15:04",
	"2006-01-02 15:04",
	"2006-01-02",
}

// ParseDateTime parses a date/datetime string in local timezone.
// Supports formats: "2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"
func ParseDateTime(s string) (time.Time, error) {
	loc := time.Local
	for _, f := range DateTimeFormats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse %q (use YYYY-MM-DD or YYYY-MM-DD HH:MM)", s)
}

// FormatUnix formats a Unix timestamp as RFC3339.
func FormatUnix(unix int64) string {
	return time.Unix(unix, 0).Local().Format(time.RFC3339)
}

// FormatUnixShort formats a Unix timestamp as YYYY-MM-DD.
func FormatUnixShort(unix int64) string {
	return time.Unix(unix, 0).Local().Format("2006-01-02")
}
