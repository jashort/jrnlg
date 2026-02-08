package internal

import (
	"fmt"
	"time"
)

// SerializeEntry converts a JournalEntry to markdown format
// Format: ## Weekday YYYY-MM-DD H:MM AM/PM Timezone
//
//	Body text
func SerializeEntry(entry *JournalEntry) string {
	header := FormatTimestamp(entry.Timestamp)
	return fmt.Sprintf("## %s\n\n%s\n", header, entry.Body)
}

// FormatTimestamp formats a timestamp for the entry header
// Format: Weekday YYYY-MM-DD H:MM AM/PM Timezone
// Example: Sunday 2026-02-08 8:31 AM America/Los_Angeles
func FormatTimestamp(timestamp time.Time) string {
	// Extract components
	weekday := timestamp.Weekday().String()

	// Format date as YYYY-MM-DD
	date := timestamp.Format("2006-01-02")

	// Convert to 12-hour format
	hour := timestamp.Hour()
	var meridiem string
	if hour < 12 {
		meridiem = "AM"
		if hour == 0 {
			hour = 12 // midnight is 12 AM
		}
	} else {
		meridiem = "PM"
		if hour > 12 {
			hour = hour - 12
		}
		// noon is 12 PM (hour stays 12)
	}

	// Format time as H:MM (no leading zero for hour)
	minute := timestamp.Minute()
	timeStr := fmt.Sprintf("%d:%02d", hour, minute)

	// Get timezone location name
	location := timestamp.Location().String()

	return fmt.Sprintf("%s %s %s %s %s", weekday, date, timeStr, meridiem, location)
}
