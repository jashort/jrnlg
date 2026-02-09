package format

import (
	"fmt"
	"strings"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
	"github.com/jashort/jrnlg/internal/patterns"
)

// FullFormatter displays complete entries with headers
type FullFormatter struct{}

// Format returns entries in full format with separators
func (f *FullFormatter) Format(entries []*internal.JournalEntry, c *color.Colorizer) string {
	if len(entries) == 0 {
		return "Found 0 entries.\n"
	}

	var sb strings.Builder
	sb.WriteString(c.Dim(fmt.Sprintf("Found %d entries:\n\n", len(entries))))

	for i, entry := range entries {
		// Write the entry header (timestamp with timezone)
		sb.WriteString(c.Bold("## "))
		sb.WriteString(c.Timestamp(internal.FormatTimestamp(entry.Timestamp)))
		sb.WriteString("\n\n")

		// Write the body with colorized tags and mentions
		body := colorizeBody(entry.Body, c)
		sb.WriteString(body)

		// Add separator between entries (but not after the last one)
		if i < len(entries)-1 {
			sb.WriteString("\n\n")
			sb.WriteString(c.Separator("---"))
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// colorizeBody highlights #tags and @mentions in body text
func colorizeBody(body string, c *color.Colorizer) string {
	if !c.Enabled() {
		return body
	}

	// Colorize all #tags in green
	body = patterns.Tag.ReplaceAllStringFunc(body, func(match string) string {
		return c.Tag(match)
	})

	// Colorize all @mentions in yellow
	body = patterns.Mention.ReplaceAllStringFunc(body, func(match string) string {
		return c.Mention(match)
	})

	return body
}
