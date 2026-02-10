package internal

import "os"

// File system permissions
const (
	// DirPermissions are the permissions for created directories
	DirPermissions os.FileMode = 0755
	// FilePermissions are the permissions for created files
	FilePermissions os.FileMode = 0644
)

// Collision handling limits
const (
	// MaxCollisionAttempts is the maximum number of timestamp collision suffixes to try
	// before giving up. Entries with the same timestamp get suffixes: -01, -02, etc.
	MaxCollisionAttempts = 100
)

// Date range constants
const (
	// DefaultStartYear is the earliest year to scan when no start date is specified
	DefaultStartYear = 1900
	// DefaultEndYear is the latest year to scan when no end date is specified
	DefaultEndYear = 2100
	// DefaultStartMonth is the first month (January)
	DefaultStartMonth = 1
	// DefaultEndMonth is the last month (December)
	DefaultEndMonth = 12
)

// Index configuration
const (
	// DefaultIndexCacheSize is the default maximum number of entries to cache in memory
	DefaultIndexCacheSize = 10000
)

// Date and time formats
const (
	// TimestampLayout is the format for entry headers
	// Format: Monday 2006-01-02 3:04 PM MST
	TimestampLayout = "Monday 2006-01-02 3:04 PM MST"

	// FileTimestampFormat is the format for entry filenames (UTC)
	// Format: 2006-01-02-15-04-05
	FileTimestampFormat = "2006-01-02 15:04:05"
)

// File extensions
const (
	// MarkdownExt is the file extension for journal entries
	MarkdownExt = ".md"
)

// Statistics configuration
const (
	// TopItemsLimit is the default number of top tags/mentions to show in statistics
	TopItemsLimit = 5
	// PercentageMultiplier is used to convert ratios to percentages
	PercentageMultiplier = 100.0
)

// Time of day boundaries (hours in 24-hour format)
const (
	// MorningStart is the hour when morning begins (5 AM)
	MorningStart = 5
	// AfternoonStart is the hour when afternoon begins (12 PM / noon)
	AfternoonStart = 12
	// EveningStart is the hour when evening begins (5 PM)
	EveningStart = 17
	// NightStart is the hour when night begins (9 PM)
	NightStart = 21
	// HoursInDay is the total number of hours in a day
	HoursInDay = 24
)

// Worker pool configuration
const (
	// MinWorkers is the minimum number of workers for parallel operations
	MinWorkers = 1
)
