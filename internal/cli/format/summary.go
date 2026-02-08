package format

import (
	"fmt"
	"strings"

	"github.com/jashort/jrnlg/internal"
)

// SummaryFormatter displays one line per entry with timestamp and preview
type SummaryFormatter struct{}

// Format returns entries in summary format (one line per entry)
// Format: YYYY-MM-DD H:MM PM MST | First 80 chars of body...
func (f *SummaryFormatter) Format(entries []*internal.JournalEntry) string {
	if len(entries) == 0 {
		return "Found 0 entries.\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d entries:\n\n", len(entries)))

	for _, entry := range entries {
		// Format timestamp with timezone abbreviation (MST format)
		// Format: YYYY-MM-DD H:MM PM TZ
		timestamp := entry.Timestamp.Format("2006-01-02 3:04 PM MST")

		// Get first non-empty line of body, trim whitespace
		bodyLines := strings.Split(entry.Body, "\n")
		firstLine := ""
		for _, line := range bodyLines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				firstLine = trimmed
				break
			}
		}

		// Truncate to 80 chars if longer
		preview := firstLine
		if len(preview) > 80 {
			preview = preview[:77] + "..."
		}

		// Write: timestamp | preview
		sb.WriteString(fmt.Sprintf("%s | %s\n", timestamp, preview))
	}

	return sb.String()
}
