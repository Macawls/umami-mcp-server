package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// normalizeDate converts ISO 8601 date strings to Unix millisecond timestamps.
// If the input is already a numeric timestamp, it is returned as-is.
// Supported formats: "2026-03-23", "2026-03-23T14:30:00Z", "2026-03-23T14:30:00+01:00"
func normalizeDate(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	// Already a number (timestamp) — return as-is
	if _, err := strconv.ParseInt(input, 10, 64); err == nil {
		return input
	}

	// Try parsing as ISO 8601 formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // full ISO with timezone
		"2006-01-02T15:04:05Z",      // UTC
		"2006-01-02T15:04:05",       // no timezone (assume UTC)
		"2006-01-02",                // date only (assume start of day UTC)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return fmt.Sprintf("%d", t.UnixMilli())
		}
	}

	// Nothing matched — return original, let Umami API handle the error
	return input
}
