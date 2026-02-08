package format

import (
	"fmt"
	"strings"

	"github.com/jashort/jrnlg/internal"
)

// FullFormatter displays complete entries with headers
type FullFormatter struct{}

// Format returns entries in full format with separators
func (f *FullFormatter) Format(entries []*internal.JournalEntry) string {
	if len(entries) == 0 {
		return "Found 0 entries.\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d entries:\n\n", len(entries)))

	for i, entry := range entries {
		// Write the entry header (timestamp with timezone)
		sb.WriteString("## ")
		sb.WriteString(internal.FormatTimestamp(entry.Timestamp))
		sb.WriteString("\n\n")

		// Write the body
		sb.WriteString(entry.Body)

		// Add separator between entries (but not after the last one)
		if i < len(entries)-1 {
			sb.WriteString("\n\n---\n\n")
		}
	}

	sb.WriteString("\n")
	return sb.String()
}
