package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileSystemStorage implements journal entry storage using the filesystem
type FileSystemStorage struct {
	basePath  string
	config    *Config
	indexOnce sync.Once
	index     *Index
	indexErr  error
	mu        sync.RWMutex
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
	for i := 1; i < MaxCollisionAttempts; i++ {
		path := fmt.Sprintf("%s-%02d%s", basePath[:len(basePath)-len(MarkdownExt)], i, MarkdownExt)
		if entry, err := fs.parseFile(path); err == nil {
			return entry, nil
		}
		// If file doesn't exist, stop trying
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
	}

	return nil, fmt.Errorf("entry not found: %s", timestamp.Format(FileTimestampFormat))
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
		fs.config.Logger.Warn("skipping invalid file during list", "error", err)
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
	startYear, startMonth := DefaultStartYear, DefaultStartMonth
	endYear, endMonth := DefaultEndYear, DefaultEndMonth

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
// Returns empty string if too many collisions (>= MaxCollisionAttempts)
func (fs *FileSystemStorage) findAvailablePath(basePath string) string {
	// Check if base path is available
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	// Try appending suffixes
	baseWithoutExt := basePath[:len(basePath)-len(MarkdownExt)] // Remove extension
	for i := 1; i < MaxCollisionAttempts; i++ {
		path := fmt.Sprintf("%s-%02d%s", baseWithoutExt, i, MarkdownExt)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}

	return "" // Too many collisions
}

// ensureDirectories creates year and month directories if they don't exist
func (fs *FileSystemStorage) ensureDirectories(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, DirPermissions)
}

// writeAtomic writes content to a file atomically using temp file + rename
func (fs *FileSystemStorage) writeAtomic(filePath string, content []byte) error {
	// Create temp file in same directory
	dir := filepath.Dir(filePath)
	tmpFile := filepath.Join(dir, ".tmp-"+filepath.Base(filePath))

	// Write to temp file
	if err := os.WriteFile(tmpFile, content, FilePermissions); err != nil {
		return err
	}

	// Rename to final path (atomic on POSIX systems)
	if err := os.Rename(tmpFile, filePath); err != nil {
		err = errors.Join(err, os.Remove(tmpFile)) // Clean up temp file on error
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
	fs.indexOnce.Do(func() {
		files, err := fs.findFiles(filter)
		if err != nil {
			fs.indexErr = fmt.Errorf("failed to find files for indexing: %w", err)
			return
		}

		fs.index = NewIndex()
		fs.indexErr = fs.index.Build(files, fs.config.MaxParseWorkers, fs.parseFile)
	})

	return fs.index, fs.indexErr
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

// InvalidateIndex clears the search index, forcing a rebuild on next search
func (fs *FileSystemStorage) InvalidateIndex() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.index = nil
	fs.indexErr = nil
	fs.indexOnce = sync.Once{}
}

// GetIndex returns the existing index or builds a new one
// This is a public wrapper around getOrCreateIndex for use by CLI commands
func (fs *FileSystemStorage) GetIndex(filter EntryFilter) (*Index, error) {
	return fs.getOrCreateIndex(filter)
}

// GetTagStatistics returns tag usage counts across all entries
// Builds index if needed
func (fs *FileSystemStorage) GetTagStatistics() (map[string]int, error) {
	// Get or create index
	filter := EntryFilter{} // No filters, index all entries
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	return index.TagStatistics(), nil
}

// GetMentionStatistics returns mention usage counts across all entries
// Builds index if needed
func (fs *FileSystemStorage) GetMentionStatistics() (map[string]int, error) {
	// Get or create index
	filter := EntryFilter{} // No filters, index all entries
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	return index.MentionStatistics(), nil
}

// GetEntriesWithTag returns file paths for all entries with the specified tag
func (fs *FileSystemStorage) GetEntriesWithTag(tag string) ([]string, error) {
	// Get or create index
	filter := EntryFilter{} // No filters, index all entries
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	entries := index.GetEntriesForTag(tag)
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.FilePath
	}

	return paths, nil
}

// GetEntriesWithMention returns file paths for all entries with the specified mention
func (fs *FileSystemStorage) GetEntriesWithMention(mention string) ([]string, error) {
	// Get or create index
	filter := EntryFilter{} // No filters, index all entries
	index, err := fs.getOrCreateIndex(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	entries := index.GetEntriesForMention(mention)
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.FilePath
	}

	return paths, nil
}

// GetEntryPath returns the file path for an entry by timestamp
// Handles collision suffixes (-01, -02, etc.)
func (fs *FileSystemStorage) GetEntryPath(timestamp time.Time) (string, error) {
	// Build expected file path
	basePath := fs.buildFilePath(timestamp)

	// Try base path first
	if _, err := os.Stat(basePath); err == nil {
		return basePath, nil
	}

	// Try with collision suffixes (-01, -02, etc.)
	for i := 1; i < MaxCollisionAttempts; i++ {
		path := fmt.Sprintf("%s-%02d%s", basePath[:len(basePath)-len(MarkdownExt)], i, MarkdownExt)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
	}

	return "", fmt.Errorf("entry not found: %s", timestamp.Format(FileTimestampFormat))
}

// UpdateEntry updates an existing entry atomically
// The entry's timestamp must match the original (timestamp changes not allowed)
func (fs *FileSystemStorage) UpdateEntry(filePath string, newEntry *JournalEntry) error {
	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("entry not found: %s", filePath)
	}

	// Serialize new content
	markdown := SerializeEntry(newEntry)

	// Write atomically (overwrites old file)
	if err := fs.writeAtomic(filePath, []byte(markdown)); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	// Invalidate index
	fs.InvalidateIndex()

	return nil
}

// DeleteEntry removes a single entry by file path
func (fs *FileSystemStorage) DeleteEntry(filePath string) error {
	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("entry not found: %s", filePath)
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	// Invalidate index
	fs.InvalidateIndex()

	return nil
}

// DeleteEntries removes multiple entries matching the filter
// Returns list of deleted file paths and any errors encountered
func (fs *FileSystemStorage) DeleteEntries(filter EntryFilter) ([]string, error) {
	// Find all matching file paths
	files, err := fs.findFiles(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		return []string{}, nil
	}

	// Filter files by timestamp (same logic as ListEntries)
	var filesToDelete []string
	for _, filePath := range files {
		entry, err := fs.parseFile(filePath)
		if err != nil {
			// Skip invalid files
			continue
		}

		// Apply timestamp filter
		if !filter.Matches(entry.Timestamp) {
			continue
		}

		filesToDelete = append(filesToDelete, filePath)
	}

	// Delete each file and collect results
	var deleted []string
	var errs []error

	for _, filePath := range filesToDelete {
		if err := os.Remove(filePath); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete %s: %w", filePath, err))
			continue
		}

		deleted = append(deleted, filePath)
	}

	// Invalidate index if any deletions succeeded
	if len(deleted) > 0 {
		fs.InvalidateIndex()
	}

	// Return results
	if len(errs) > 0 {
		return deleted, errors.Join(errs...)
	}

	return deleted, nil
}

// replaceMetadataInEntries is a unified function for replacing tags or mentions
// Uses case-insensitive matching with proper word boundaries
// Returns list of updated file paths
func (fs *FileSystemStorage) replaceMetadataInEntries(oldValue, newValue, symbol string, filePaths []string, dryRun bool) ([]string, error) {
	if len(filePaths) == 0 {
		return []string{}, nil
	}

	// Build regex pattern for case-insensitive replacement
	// Pattern: {symbol}oldValue followed by non-word-char (but not hyphen/underscore)
	// This ensures we match #code but not #code-review
	pattern := fmt.Sprintf(`(?i)%s%s(?:[^a-zA-Z0-9_-]|$)`, regexp.QuoteMeta(symbol), regexp.QuoteMeta(oldValue))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	var updated []string
	var errs []error

	for _, filePath := range filePaths {
		// Read current entry
		entry, err := fs.parseFile(filePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read %s: %w", filePath, err))
			continue
		}

		// Replace metadata in body (case-insensitive)
		// Use ReplaceAllStringFunc to preserve the character after the metadata
		newBody := re.ReplaceAllStringFunc(entry.Body, func(match string) string {
			// The match includes symbol + value (case-insensitive) + possibly one trailing character
			// Find where the value ends (it starts with symbol and continues until non-metadata-char)
			if len(match) > 1 && match[0] == symbol[0] {
				// Skip the symbol and find the end of the value
				i := 1
				for i < len(match) && (match[i] >= 'a' && match[i] <= 'z' ||
					match[i] >= 'A' && match[i] <= 'Z' ||
					match[i] >= '0' && match[i] <= '9' ||
					match[i] == '_' || match[i] == '-') {
					i++
				}
				// Everything after position i is the trailing character(s)
				return symbol + newValue + match[i:]
			}
			return symbol + newValue
		})

		// Skip if no changes (shouldn't happen, but safety check)
		if newBody == entry.Body {
			continue
		}

		// Update entry body
		entry.Body = newBody

		// Re-parse to validate and auto-deduplicate metadata
		// This ensures metadata like "#code-review #code-review" become single instance
		serialized := SerializeEntry(entry)
		entry, err = ParseEntry(serialized)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse updated entry %s: %w", filePath, err))
			continue
		}

		// Write updated entry (unless dry run)
		if !dryRun {
			if err := fs.UpdateEntry(filePath, entry); err != nil {
				errs = append(errs, fmt.Errorf("failed to update %s: %w", filePath, err))
				continue
			}
		}

		updated = append(updated, filePath)
	}

	// Return results
	if len(errs) > 0 {
		return updated, errors.Join(errs...)
	}

	return updated, nil
}

// ReplaceTagInEntries replaces oldTag with newTag in all entries
// Uses case-insensitive matching with proper word boundaries
// Returns list of updated file paths
func (fs *FileSystemStorage) ReplaceTagInEntries(oldTag, newTag string, dryRun bool) ([]string, error) {
	// Get entries with old tag using index
	filePaths, err := fs.GetEntriesWithTag(oldTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries with tag: %w", err)
	}

	return fs.replaceMetadataInEntries(oldTag, newTag, "#", filePaths, dryRun)
}

// ReplaceMentionInEntries replaces oldMention with newMention in all entries
// Uses case-insensitive matching with proper word boundaries
// Returns list of updated file paths
func (fs *FileSystemStorage) ReplaceMentionInEntries(oldMention, newMention string, dryRun bool) ([]string, error) {
	// Get entries with old mention using index
	filePaths, err := fs.GetEntriesWithMention(oldMention)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries with mention: %w", err)
	}

	return fs.replaceMetadataInEntries(oldMention, newMention, "@", filePaths, dryRun)
}
