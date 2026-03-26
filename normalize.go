package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func normalizeDate(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	if _, err := strconv.ParseInt(input, 10, 64); err == nil {
		return input
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return fmt.Sprintf("%d", t.UnixMilli())
		}
	}

	return input
}
