package cli

import (
	"strings"
	"time"
)

// TruncateBody truncates the body text to a maximum length, adding "..." if truncated.
// It only shows the first line of the body.
func TruncateBody(body string, maxLen int) string {
	// Get first line only
	lines := strings.Split(body, "\n")
	firstLine := strings.TrimSpace(lines[0])

	if len(firstLine) <= maxLen {
		return firstLine
	}

	return firstLine[:maxLen] + "..."
}

// IsTimestampFormat checks if a string matches the timestamp format YYYY-MM-DD-HH-MM-SS
func IsTimestampFormat(s string) bool {
	// Simple check: 19 chars, dashes in right places
	if len(s) != 19 {
		return false
	}
	if s[4] != '-' || s[7] != '-' || s[10] != '-' || s[13] != '-' || s[16] != '-' {
		return false
	}
	return true
}

// ParseTimestamp parses a timestamp string in format YYYY-MM-DD-HH-MM-SS
// The timestamp is parsed as UTC to match the file naming convention
func ParseTimestamp(s string) (time.Time, error) {
	// Parse as UTC (same as file naming)
	return time.Parse("2006-01-02-15-04-05", s)
}
