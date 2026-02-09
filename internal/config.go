package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config holds configuration for journal storage
type Config struct {
	StoragePath     string   // Path to store journal entries
	IndexCacheSize  int      // Maximum number of entries to cache in memory
	ParallelParse   bool     // Enable parallel parsing of entries
	MaxParseWorkers int      // Maximum number of parallel parsing workers
	EditorArgs      []string // Additional arguments to pass to the editor
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &Config{
		StoragePath:     filepath.Join(homeDir, ".jrnlg", "entries"),
		IndexCacheSize:  10000,
		ParallelParse:   true,
		MaxParseWorkers: runtime.NumCPU(),
	}
}

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Override with environment variable if set
	if storagePath := os.Getenv("JRNLG_STORAGE_PATH"); storagePath != "" {
		config.StoragePath = storagePath
	}

	// Parse editor arguments from environment variable
	if editorArgs := os.Getenv("JRNLG_EDITOR_ARGS"); editorArgs != "" {
		config.EditorArgs = parseEditorArgs(editorArgs)
	}

	return config, nil
}

// parseEditorArgs parses a string of editor arguments into a slice
// Supports quoted arguments to handle spaces, e.g., "+startinsert" "+call cursor(3,1)"
func parseEditorArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range input {
		switch {
		case (r == '"' || r == '\'') && !inQuote:
			// Start of quoted section
			inQuote = true
			quoteChar = r
		case r == quoteChar && inQuote:
			// End of quoted section
			inQuote = false
			quoteChar = 0
		case r == ' ' && !inQuote:
			// Space outside quotes - separator
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			// Regular character
			current.WriteRune(r)
		}
	}

	// Add final argument if any
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
