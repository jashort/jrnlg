package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileSystemStorage implements journal entry storage using the filesystem
type FileSystemStorage struct {
	basePath string
	config   *Config
	index    *Index
	mu       sync.RWMutex
}

// NewFileSystemStorage creates a new filesystem-based storage
func NewFileSystemStorage(basePath string, config *Config) *FileSystemStorage {
	if config == nil {
		config = DefaultConfig()
	}
	return &FileSystemStorage{
		basePath: basePath,
		config:   config,
	}
}

// SaveEntry writes a journal entry to disk
// The entry is stored in: <basePath>/<year>/<month>/YYYY-MM-DD-HH-MM-SS.md
// Timestamp is converted to UTC for consistent file naming and sorting
func (fs *FileSystemStorage) SaveEntry(entry *JournalEntry) error {
	// Build file path (uses UTC for consistent naming)
	filePath := fs.buildFilePath(entry.Timestamp)

	// Handle collision - if file exists, append suffix
	filePath = fs.findAvailablePath(filePath)
	if filePath == "" {
		return fmt.Errorf("too many entries with same timestamp")
	}

	// Ensure directories exist
	if err := fs.ensureDirectories(filePath); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Serialize entry to markdown
	markdown := SerializeEntry(entry)

	// Write atomically (temp file + rename)
	if err := fs.writeAtomic(filePath, []byte(markdown)); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	return nil
}

// GetEntry retrieves a journal entry by timestamp
// Searches for files matching the timestamp (including collision suffixes)
func (fs *FileSystemStorage) GetEntry(timestamp time.Time) (*JournalEntry, error) {
	// Build expected file path
	basePath := fs.buildFilePath(timestamp)

	// Try base path first
	if entry, err := fs.parseFile(basePath); err == nil {
		return entry, nil
	}

	// Try with collision suffixes (-01, -02, etc.)
	for i := 1; i < 100; i++ {
		path := fmt.Sprintf("%s-%02d.md", basePath[:len(basePath)-3], i)
		if entry, err := fs.parseFile(path); err == nil {
			return entry, nil
		}
		// If file doesn't exist, stop trying
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
	}

	return nil, fmt.Errorf("entry not found: %s", timestamp.Format("2006-01-02 15:04:05"))
}

// ListEntries retrieves all entries matching the given filter
// Entries are sorted by timestamp (oldest first)
// Supports date range filtering, limit, and offset
func (fs *FileSystemStorage) ListEntries(filter EntryFilter) ([]*JournalEntry, error) {
	// Find all matching files
	files, err := fs.findFiles(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	// Parse files in parallel
	entries := make([]*JournalEntry, 0, len(files))
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	// Use worker pool for parsing
	maxWorkers := fs.config.MaxParseWorkers
	if !fs.config.ParallelParse {
		maxWorkers = 1
	}

	jobs := make(chan string, len(files))

	// Start workers
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				entry, err := fs.parseFile(filePath)
				if err != nil {
					// Log warning but continue (skip invalid files)
					errChan <- fmt.Errorf("skipping invalid file %s: %w", filePath, err)
					continue
				}

				// Filter by timestamp (additional check)
				if !filter.Matches(entry.Timestamp) {
					continue
				}

				mu.Lock()
				entries = append(entries, entry)
				mu.Unlock()
			}
		}()
	}

	// Send jobs
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(errChan)

	// Log any errors (non-fatal)
	for err := range errChan {
		// TODO: Add proper logging when we have a logger
		_ = err
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	// Apply offset and limit
	start := filter.Offset
	if start > len(entries) {
		return []*JournalEntry{}, nil
	}

	end := len(entries)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}

	return entries[start:end], nil
}

// findFiles locates all markdown files in the date range specified by the filter
// Optimizes by only scanning relevant year/month directories
func (fs *FileSystemStorage) findFiles(filter EntryFilter) ([]string, error) {
	var files []string

	// Determine year and month ranges to scan
	startYear, startMonth := 1900, 1
	endYear, endMonth := 2100, 12

	if filter.StartDate != nil {
		utc := filter.StartDate.UTC()
		startYear, startMonth = utc.Year(), int(utc.Month())
	}

	if filter.EndDate != nil {
		utc := filter.EndDate.UTC()
		endYear, endMonth = utc.Year(), int(utc.Month())
	}

	// Scan each year/month directory
	for year := startYear; year <= endYear; year++ {
		yearPath := filepath.Join(fs.basePath, fmt.Sprintf("%04d", year))

		// Check if year directory exists
		if _, err := os.Stat(yearPath); os.IsNotExist(err) {
			continue
		}

		// Determine month range for this year
		firstMonth, lastMonth := 1, 12
		if year == startYear {
			firstMonth = startMonth
		}
		if year == endYear {
			lastMonth = endMonth
		}

		for month := firstMonth; month <= lastMonth; month++ {
			monthPath := filepath.Join(yearPath, fmt.Sprintf("%02d", month))

			// Check if month directory exists
			if _, err := os.Stat(monthPath); os.IsNotExist(err) {
				continue
			}

			// Read all files in this month directory
			entries, err := os.ReadDir(monthPath)
			if err != nil {
				continue // Skip directories we can't read
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				// Only include .md files
				if !isMarkdownFile(entry.Name()) {
					continue
				}

				filePath := filepath.Join(monthPath, entry.Name())
				files = append(files, filePath)
			}
		}
	}

	return files, nil
}

// buildFilePath constructs the file path for an entry
// Format: <basePath>/<year>/<month>/YYYY-MM-DD-HH-MM-SS.md
// Uses UTC time for consistent naming across timezones
func (fs *FileSystemStorage) buildFilePath(timestamp time.Time) string {
	utc := timestamp.UTC()
	year := fmt.Sprintf("%04d", utc.Year())
	month := fmt.Sprintf("%02d", int(utc.Month()))
	filename := fmt.Sprintf("%04d-%02d-%02d-%02d-%02d-%02d.md",
		utc.Year(), utc.Month(), utc.Day(),
		utc.Hour(), utc.Minute(), utc.Second())

	return filepath.Join(fs.basePath, year, month, filename)
}

// findAvailablePath checks if file exists and appends suffix if needed
// Returns empty string if too many collisions (>= 100)
func (fs *FileSystemStorage) findAvailablePath(basePath string) string {
	// Check if base path is available
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	// Try appending suffixes
	baseWithoutExt := basePath[:len(basePath)-3] // Remove .md
	for i := 1; i < 100; i++ {
		path := fmt.Sprintf("%s-%02d.md", baseWithoutExt, i)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}

	return "" // Too many collisions
}

// ensureDirectories creates year and month directories if they don't exist
func (fs *FileSystemStorage) ensureDirectories(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}

// writeAtomic writes content to a file atomically using temp file + rename
func (fs *FileSystemStorage) writeAtomic(filePath string, content []byte) error {
	// Create temp file in same directory
	dir := filepath.Dir(filePath)
	tmpFile := filepath.Join(dir, ".tmp-"+filepath.Base(filePath))

	// Write to temp file
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		return err
	}

	// Rename to final path (atomic on POSIX systems)
	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return err
	}

	return nil
}

// parseFile reads and parses a single entry file
func (fs *FileSystemStorage) parseFile(filePath string) (*JournalEntry, error) {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse entry
	entry, err := ParseEntry(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return entry, nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isMarkdownFile checks if a file has .md extension
func isMarkdownFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".md")
}

// SearchByTags searches for entries that have ALL the specified tags (AND logic)
// Builds index on first search if not already built
func (fs *FileSystemStorage) SearchByTags(tags []string, filter EntryFilter) ([]*JournalEntry, error) {
	// Get or create index
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, err
	}

	// Search index
	indexedEntries := index.SearchByTags(tags)

	// Convert to full entries and apply filter
	return fs.indexedEntriesToFull(indexedEntries, filter)
}

// SearchByMentions searches for entries that have ALL the specified mentions (AND logic)
// Builds index on first search if not already built
func (fs *FileSystemStorage) SearchByMentions(mentions []string, filter EntryFilter) ([]*JournalEntry, error) {
	// Get or create index
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, err
	}

	// Search index
	indexedEntries := index.SearchByMentions(mentions)

	// Convert to full entries and apply filter
	return fs.indexedEntriesToFull(indexedEntries, filter)
}

// SearchByKeyword searches for entries whose body contains the keyword (case-insensitive)
// Builds index on first search if not already built
func (fs *FileSystemStorage) SearchByKeyword(keyword string, filter EntryFilter) ([]*JournalEntry, error) {
	// Get or create index
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, err
	}

	// Search index
	indexedEntries := index.SearchByKeyword(keyword)

	// Convert to full entries and apply filter
	return fs.indexedEntriesToFull(indexedEntries, filter)
}

// getOrCreateIndex returns the existing index or builds a new one
// Only builds index for the files in the specified date range
func (fs *FileSystemStorage) getOrCreateIndex(filter EntryFilter) (*Index, error) {
	fs.mu.RLock()
	if fs.index != nil {
		fs.mu.RUnlock()
		return fs.index, nil
	}
	fs.mu.RUnlock()

	// Need to build index
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Double-check after acquiring write lock
	if fs.index != nil {
		return fs.index, nil
	}

	// Find all files in date range
	files, err := fs.findFiles(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for indexing: %w", err)
	}

	// Build index
	index := NewIndex()
	err = index.Build(files, fs.config.MaxParseWorkers, fs.parseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	fs.index = index
	return index, nil
}

// indexedEntriesToFull converts IndexedEntry results to full JournalEntry objects
// Applies date filter, sorting, limit, and offset
func (fs *FileSystemStorage) indexedEntriesToFull(indexed []*IndexedEntry, filter EntryFilter) ([]*JournalEntry, error) {
	var entries []*JournalEntry

	for _, ie := range indexed {
		// Apply timestamp filter
		if !filter.Matches(ie.Timestamp) {
			continue
		}

		// Read full entry from file
		entry, err := fs.parseFile(ie.FilePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		entries = append(entries, entry)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	// Apply offset and limit
	start := filter.Offset
	if start > len(entries) {
		return []*JournalEntry{}, nil
	}

	end := len(entries)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}

	return entries[start:end], nil
}
