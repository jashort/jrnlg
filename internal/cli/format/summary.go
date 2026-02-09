package format

import (
	"fmt"
	"strings"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
)

// SummaryFormatter displays one line per entry with timestamp and preview
type SummaryFormatter struct{}

// Format returns entries in summary format (one line per entry)
// Format: YYYY-MM-DD H:MM PM MST | First 80 chars of body...
func (f *SummaryFormatter) Format(entries []*internal.JournalEntry, c *color.Colorizer) string {
	if len(entries) == 0 {
		return "Found 0 entries.\n"
	}

	var sb strings.Builder
	sb.WriteString(c.Dim(fmt.Sprintf("Found %d entries:\n\n", len(entries))))

	for _, entry := range entries {
		// Format timestamp with timezone abbreviation (MST format)
		// Format: YYYY-MM-DD H:MM PM TZ
		timestamp := c.Timestamp(entry.Timestamp.Format("2006-01-02 3:04 PM MST"))

		// Get first non-empty line of body, trim whitespace
		preview := getFirstLine(entry.Body)
		if len(preview) > 80 {
			preview = preview[:77] + "..."
		}

		// Dim separator
		separator := c.Dim(" | ")

		// Write: timestamp | preview
		sb.WriteString(fmt.Sprintf("%s%s%s\n", timestamp, separator, preview))
	}

	return sb.String()
}

// getFirstLine extracts the first non-empty line from body text
func getFirstLine(body string) string {
	bodyLines := strings.Split(body, "\n")
	for _, line := range bodyLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
