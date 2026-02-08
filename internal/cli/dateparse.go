package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

// ParseDate parses a date string in ISO 8601 format or natural language
// Supported ISO formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS
// Supported natural language: "yesterday", "last week", "3 days ago", "today", "tomorrow", etc.
func ParseDate(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try ISO 8601 formats first (faster and more reliable)
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return t, nil
		}
	}

	// Try natural language parsing
	result, err := parseNaturalLanguage(input)
	if err == nil {
		return result, nil
	}

	return time.Time{}, fmt.Errorf("could not parse date '%s' (use format: YYYY-MM-DD or natural language like 'yesterday')", input)
}

// parseNaturalLanguage parses natural language date strings
func parseNaturalLanguage(input string) (time.Time, error) {
	// Create parser with English and common rules
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	// Parse relative to current time
	now := time.Now()
	result, err := w.Parse(input, now)
	if err != nil {
		return time.Time{}, err
	}

	if result == nil {
		return time.Time{}, fmt.Errorf("could not parse natural language date: %s", input)
	}

	// Return the parsed time
	return result.Time, nil
}
