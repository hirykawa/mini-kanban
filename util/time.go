// Package util provides shared utility functions for mini-kanban.
package util

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// DateTimeFormats are the supported date/time formats for parsing.
var DateTimeFormats = []string{
	"2006-01-02T15:04",
	"2006-01-02 15:04",
	"2006-01-02",
}

var relativeTimePattern = regexp.MustCompile(`^(?:(\d+)h)?(?:(\d+)m)?$`)

// ParseDateTime parses a date/datetime string in local timezone.
// Supports formats: "2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"
// Also supports relative time expressions: "2h", "30m", "1h30m"
func ParseDateTime(s string) (time.Time, error) {
	if t, ok := parseRelativeTime(s); ok {
		return t, nil
	}

	loc := time.Local
	for _, f := range DateTimeFormats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse %q (use YYYY-MM-DD, YYYY-MM-DD HH:MM, or relative like 2h, 30m)", s)
}

func parseRelativeTime(s string) (time.Time, bool) {
	matches := relativeTimePattern.FindStringSubmatch(s)
	if matches == nil || (matches[1] == "" && matches[2] == "") {
		return time.Time{}, false
	}

	var hours, minutes int
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}

	if hours == 0 && minutes == 0 {
		return time.Time{}, false
	}

	return time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute), true
}

// FormatUnix formats a Unix timestamp as RFC3339.
func FormatUnix(unix int64) string {
	return time.Unix(unix, 0).Local().Format(time.RFC3339)
}

// FormatUnixShort formats a Unix timestamp as YYYY-MM-DD.
func FormatUnixShort(unix int64) string {
	return time.Unix(unix, 0).Local().Format("2006-01-02")
}
