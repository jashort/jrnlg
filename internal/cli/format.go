package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
	"github.com/jashort/jrnlg/internal/patterns"
)

// FormatEntries formats entries based on the specified format type
func FormatEntries(entries []*internal.JournalEntry, format string, colorizer *color.Colorizer) string {
	switch format {
	case "summary":
		return formatSummary(entries, colorizer)
	case "json":
		return formatJSON(entries)
	default: // "full"
		return formatFull(entries, colorizer)
	}
}

// formatFull displays complete entries with headers
func formatFull(entries []*internal.JournalEntry, c *color.Colorizer) string {
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

// formatSummary displays one line per entry with timestamp and preview
// Format: YYYY-MM-DD H:MM PM MST | First 80 chars of body...
func formatSummary(entries []*internal.JournalEntry, c *color.Colorizer) string {
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

// jsonEntry is the JSON representation of a journal entry
type jsonEntry struct {
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
	Mentions  []string `json:"mentions"`
	Body      string   `json:"body"`
}

// formatJSON returns entries in JSON format
func formatJSON(entries []*internal.JournalEntry) string {
	// Convert to JSON-friendly format
	jsonEntries := make([]jsonEntry, len(entries))
	for i, entry := range entries {
		jsonEntries[i] = jsonEntry{
			Timestamp: entry.Timestamp.Format("2006-01-02T15:04:05Z07:00"), // RFC3339
			Tags:      entry.Tags,
			Mentions:  entry.Mentions,
			Body:      entry.Body,
		}
	}

	// Marshal with indentation for readability
	jsonBytes, err := json.MarshalIndent(jsonEntries, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v\n", err)
	}

	return string(jsonBytes) + "\n"
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
