package format

import (
	"github.com/jashort/jrnlg/internal"
)

// Formatter is the interface for output formatters
type Formatter interface {
	Format(entries []*internal.JournalEntry) string
}

// GetFormatter returns the appropriate formatter based on format name
func GetFormatter(format string) Formatter {
	switch format {
	case "summary":
		return &SummaryFormatter{}
	case "json":
		return &JSONFormatter{}
	default:
		return &FullFormatter{}
	}
}
