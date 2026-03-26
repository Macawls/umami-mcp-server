package main

import (
	"strconv"
	"testing"
	"time"
)

func TestNormalizeDate_Timestamps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1774220400000", "1774220400000"},
		{"1742688000000", "1742688000000"},
		{"0", "0"},
	}

	for _, tt := range tests {
		result := normalizeDate(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeDate(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeDate_ISODates(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month time.Month
		day   int
	}{
		{"2026-03-23", 2026, time.March, 23},
		{"2025-01-01", 2025, time.January, 1},
		{"2026-12-31", 2026, time.December, 31},
	}

	for _, tt := range tests {
		result := normalizeDate(tt.input)
		ms, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			t.Errorf("normalizeDate(%q) = %q, expected numeric timestamp", tt.input, result)
			continue
		}

		parsed := time.UnixMilli(ms).UTC()
		if parsed.Year() != tt.year || parsed.Month() != tt.month || parsed.Day() != tt.day {
			t.Errorf("normalizeDate(%q) = %s, want %d-%02d-%02d",
				tt.input, parsed.Format("2006-01-02"), tt.year, tt.month, tt.day)
		}
	}
}

func TestNormalizeDate_ISODateTime(t *testing.T) {
	tests := []struct {
		input string
		hour  int
		min   int
	}{
		{"2026-03-23T14:30:00Z", 14, 30},
		{"2026-03-23T00:00:00Z", 0, 0},
		{"2026-03-23T23:59:59Z", 23, 59},
	}

	for _, tt := range tests {
		result := normalizeDate(tt.input)
		ms, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			t.Errorf("normalizeDate(%q) = %q, expected numeric timestamp", tt.input, result)
			continue
		}

		parsed := time.UnixMilli(ms).UTC()
		if parsed.Hour() != tt.hour || parsed.Minute() != tt.min {
			t.Errorf("normalizeDate(%q) time = %02d:%02d, want %02d:%02d",
				tt.input, parsed.Hour(), parsed.Minute(), tt.hour, tt.min)
		}
	}
}

func TestNormalizeDate_ISOWithTimezone(t *testing.T) {
	result := normalizeDate("2026-03-23T14:30:00+02:00")
	ms, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Fatalf("normalizeDate returned non-numeric: %q", result)
	}

	parsed := time.UnixMilli(ms).UTC()
	// 14:30 +02:00 = 12:30 UTC
	if parsed.Hour() != 12 || parsed.Minute() != 30 {
		t.Errorf("expected 12:30 UTC, got %02d:%02d", parsed.Hour(), parsed.Minute())
	}
}

func TestNormalizeDate_EmptyAndInvalid(t *testing.T) {
	if result := normalizeDate(""); result != "" {
		t.Errorf("normalizeDate(\"\") = %q, want \"\"", result)
	}
	if result := normalizeDate("  "); result != "" {
		t.Errorf("normalizeDate(\"  \") = %q, want \"\"", result)
	}
	// Invalid input returned as-is
	if result := normalizeDate("not-a-date"); result != "not-a-date" {
		t.Errorf("normalizeDate(\"not-a-date\") = %q, want \"not-a-date\"", result)
	}
}

func TestNormalizeDate_LLMBugScenario(t *testing.T) {
	// This is the exact bug: LLM calculates 1742688000000 for "2026-03-23"
	// but that's actually 2025-03-23. With ISO dates, this can't happen.
	result := normalizeDate("2026-03-23")
	ms, _ := strconv.ParseInt(result, 10, 64)
	parsed := time.UnixMilli(ms).UTC()

	if parsed.Year() != 2026 {
		t.Errorf("normalizeDate(\"2026-03-23\") resolved to year %d, want 2026", parsed.Year())
	}
	if parsed.Month() != time.March || parsed.Day() != 23 {
		t.Errorf("normalizeDate(\"2026-03-23\") resolved to %s, want 2026-03-23", parsed.Format("2006-01-02"))
	}
}
