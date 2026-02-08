package format

import (
	"encoding/json"
	"fmt"

	"github.com/jashort/jrnlg/internal"
)

// JSONFormatter displays entries as JSON
type JSONFormatter struct{}

// jsonEntry is the JSON representation of a journal entry
type jsonEntry struct {
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
	Mentions  []string `json:"mentions"`
	Body      string   `json:"body"`
}

// Format returns entries in JSON format
func (f *JSONFormatter) Format(entries []*internal.JournalEntry) string {
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
