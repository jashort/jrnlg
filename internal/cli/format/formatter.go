package format

import (
	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
)

// Formatter formats journal entries for display
type Formatter interface {
	Format(entries []*internal.JournalEntry, c *color.Colorizer) string
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
