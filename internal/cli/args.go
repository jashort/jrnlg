package cli

import (
	"time"

	"github.com/jashort/jrnlg/internal/cli/color"
)

// SearchArgs contains parsed search arguments
type SearchArgs struct {
	Tags      []string
	Mentions  []string
	Keywords  []string
	FromDate  *time.Time
	ToDate    *time.Time
	Limit     int
	Offset    int
	Format    string // "full", "summary", "json"
	Reverse   bool
	ColorMode color.Mode // Color mode: auto, always, never
}
