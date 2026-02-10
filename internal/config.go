package internal

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config holds configuration for journal storage
type Config struct {
	StoragePath     string       // Path to store journal entries
	IndexCacheSize  int          // Maximum number of entries to cache in memory
	ParallelParse   bool         // Enable parallel parsing of entries
	MaxParseWorkers int          // Maximum number of parallel parsing workers
	EditorArgs      []string     // Additional arguments to pass to the editor
	Logger          *slog.Logger // Structured logger
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Create logger with text handler to stderr
	// Default level is INFO, can be configured via JRNLG_LOG_LEVEL env var
	logLevel := slog.LevelInfo
	if level := os.Getenv("JRNLG_LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN", "WARNING":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(handler)

	return &Config{
		StoragePath:     filepath.Join(homeDir, ".jrnlg", "entries"),
		IndexCacheSize:  DefaultIndexCacheSize,
		ParallelParse:   true,
		MaxParseWorkers: runtime.NumCPU(),
		Logger:          logger,
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
