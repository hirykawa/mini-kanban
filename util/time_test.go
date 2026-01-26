package util

import (
	"testing"
	"time"
)

func TestParseDateTimeFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"date only", "2026-01-18", false},
		{"datetime with T", "2026-01-18T15:30", false},
		{"datetime with space", "2026-01-18 15:30", false},
		{"invalid format", "01-18-2026", true},
		{"invalid string", "not a date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseDateTimeRelative(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantHours   int
		wantMinutes int
		wantErr     bool
	}{
		{"hours only", "2h", 2, 0, false},
		{"minutes only", "30m", 0, 30, false},
		{"hours and minutes", "1h30m", 1, 30, false},
		{"large hours", "24h", 24, 0, false},
		{"large minutes", "120m", 0, 120, false},
		{"zero hours", "0h", 0, 0, true},
		{"zero minutes", "0m", 0, 0, true},
		{"zero both", "0h0m", 0, 0, true},
		{"empty string", "", 0, 0, true},
		{"just h", "h", 0, 0, true},
		{"just m", "m", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			result, err := ParseDateTime(tt.input)
			after := time.Now()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedDuration := time.Duration(tt.wantHours)*time.Hour + time.Duration(tt.wantMinutes)*time.Minute
				minExpected := before.Add(expectedDuration)
				maxExpected := after.Add(expectedDuration)

				if result.Before(minExpected) || result.After(maxExpected) {
					t.Errorf("ParseDateTime(%q) = %v, expected between %v and %v",
						tt.input, result, minExpected, maxExpected)
				}
			}
		})
	}
}

func TestParseDateTimeRelativeDoesNotBreakAbsolute(t *testing.T) {
	// Ensure absolute dates still work correctly
	loc := time.Local
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2026-01-18", time.Date(2026, 1, 18, 0, 0, 0, 0, loc)},
		{"2026-01-18T14:30", time.Date(2026, 1, 18, 14, 30, 0, 0, loc)},
		{"2026-01-18 14:30", time.Date(2026, 1, 18, 14, 30, 0, 0, loc)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDateTime(tt.input)
			if err != nil {
				t.Errorf("ParseDateTime(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !result.Equal(tt.expected) {
				t.Errorf("ParseDateTime(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
