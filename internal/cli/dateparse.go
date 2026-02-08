package cli

import (
	"fmt"
	"time"
)

// ParseDate parses a date string in ISO 8601 format
// Supported formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS
// Phase 2 will add natural language parsing
func ParseDate(input string) (time.Time, error) {
	// Try common ISO 8601 formats
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

	return time.Time{}, fmt.Errorf("could not parse date '%s' (use format: YYYY-MM-DD)", input)
}
