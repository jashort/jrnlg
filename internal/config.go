package internal

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config holds configuration for journal storage
type Config struct {
	StoragePath     string // Path to store journal entries
	IndexCacheSize  int    // Maximum number of entries to cache in memory
	ParallelParse   bool   // Enable parallel parsing of entries
	MaxParseWorkers int    // Maximum number of parallel parsing workers
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

	return config, nil
}
