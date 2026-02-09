package color

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ANSI escape codes
const (
	reset = "\033[0m"
	bold  = "\033[1m"
	dim   = "\033[2m"

	// Foreground colors
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

// Mode determines when colors are used
type Mode int

const (
	Auto   Mode = iota // Auto-detect terminal (default)
	Always             // Always use colors
	Never              // Never use colors
)

// Colorizer applies ANSI colors to text
type Colorizer struct {
	enabled bool
}

// New creates a Colorizer with the specified mode
func New(mode Mode) *Colorizer {
	enabled := false

	switch mode {
	case Always:
		enabled = true
	case Never:
		enabled = false
	case Auto:
		// Enable if: stdout is a terminal AND NO_COLOR is not set
		enabled = term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("NO_COLOR") == ""
	}

	return &Colorizer{enabled: enabled}
}

// Enabled returns whether colors are active
func (c *Colorizer) Enabled() bool {
	return c.enabled
}

// Timestamp colors timestamps in cyan
func (c *Colorizer) Timestamp(s string) string {
	if !c.enabled {
		return s
	}
	return cyan + s + reset
}

// Tag colors hashtags in green
func (c *Colorizer) Tag(s string) string {
	if !c.enabled {
		return s
	}
	return green + s + reset
}

// Mention colors @mentions in yellow
func (c *Colorizer) Mention(s string) string {
	if !c.enabled {
		return s
	}
	return yellow + s + reset
}

// Dim renders text in dim gray
func (c *Colorizer) Dim(s string) string {
	if !c.enabled {
		return s
	}
	return gray + s + reset
}

// Bold makes text bold
func (c *Colorizer) Bold(s string) string {
	if !c.enabled {
		return s
	}
	return bold + s + reset
}

// Separator colors separator lines in dim gray
func (c *Colorizer) Separator(s string) string {
	return c.Dim(s)
}

// ParseMode converts a string to a Mode
func ParseMode(s string) (Mode, error) {
	switch s {
	case "auto":
		return Auto, nil
	case "always":
		return Always, nil
	case "never":
		return Never, nil
	default:
		return Auto, fmt.Errorf("invalid color mode %q: must be auto, always, or never", s)
	}
}

// Default colorizer for convenience functions
var defaultColorizer = New(Auto)

// Cyan returns text in cyan (for headers)
func Cyan(s string) string {
	if !defaultColorizer.enabled {
		return s
	}
	return cyan + s + reset
}

// Green returns text in green (for numbers/highlights)
func Green(s string) string {
	if !defaultColorizer.enabled {
		return s
	}
	return green + s + reset
}

// Yellow returns text in yellow
func Yellow(s string) string {
	if !defaultColorizer.enabled {
		return s
	}
	return yellow + s + reset
}

// Red returns text in red
func Red(s string) string {
	if !defaultColorizer.enabled {
		return s
	}
	return red + s + reset
}
