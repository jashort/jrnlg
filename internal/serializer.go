package internal

import (
	"fmt"
	"time"
)

// SerializeEntry converts a JournalEntry to markdown format
// Format: ## Monday 2006-01-02 3:04 PM MST
//
//	Body text
func SerializeEntry(entry *JournalEntry) string {
	header := FormatTimestamp(entry.Timestamp)
	return fmt.Sprintf("## %s\n\n%s\n", header, entry.Body)
}

// FormatTimestamp formats a timestamp for the entry header
// Preserves the original timezone as an abbreviation
// Format: Monday 2006-01-02 3:04 PM MST
// Example: Sunday 2026-02-08 8:31 AM PST
func FormatTimestamp(timestamp time.Time) string {
	return timestamp.Format("Monday 2006-01-02 3:04 PM MST")
}
