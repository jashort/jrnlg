package cli

import (
	"fmt"
	"time"

	"github.com/jashort/jrnlg/internal"
)

// EntrySelector handles finding entries based on various selectors
type EntrySelector struct {
	storage *internal.FileSystemStorage
}

// NewEntrySelector creates a new entry selector
func NewEntrySelector(storage *internal.FileSystemStorage) *EntrySelector {
	return &EntrySelector{storage: storage}
}

// SelectEntry finds a single entry based on selector string.
// Selector can be:
// - "" (empty): most recent entry
// - "YYYY-MM-DD-HH-MM-SS": specific timestamp
// - Natural language date (yesterday, last week, etc.): filter and pick
// Returns: entry, filePath, error
func (s *EntrySelector) SelectEntry(selector string) (*internal.JournalEntry, string, error) {
	// Case 1: No selector → most recent entry
	if selector == "" {
		return s.selectMostRecent()
	}

	// Case 2: Timestamp format (YYYY-MM-DD-HH-MM-SS)
	if IsTimestampFormat(selector) {
		return s.selectByTimestamp(selector)
	}

	// Case 3: Natural language date → filter and pick
	return s.selectByDate(selector)
}

// SelectEntries finds multiple entries based on filter.
// Used by delete command for batch operations.
func (s *EntrySelector) SelectEntries(filter internal.EntryFilter) ([]*internal.JournalEntry, error) {
	entries, err := s.storage.ListEntries(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	return entries, nil
}

// selectMostRecent returns the most recent entry
func (s *EntrySelector) selectMostRecent() (*internal.JournalEntry, string, error) {
	// List all entries (entries are sorted oldest first)
	filter := internal.EntryFilter{}

	entries, err := s.storage.ListEntries(filter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list entries: %w", err)
	}

	if len(entries) == 0 {
		return nil, "", fmt.Errorf("no entries found")
	}

	// Get most recent (last entry in the sorted list)
	entry := entries[len(entries)-1]

	filePath, err := s.storage.GetEntryPath(entry.Timestamp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get entry path: %w", err)
	}

	return entry, filePath, nil
}

// selectByTimestamp finds an entry by its exact timestamp
func (s *EntrySelector) selectByTimestamp(timestampStr string) (*internal.JournalEntry, string, error) {
	timestamp, err := ParseTimestamp(timestampStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid timestamp format: %w", err)
	}

	entry, err := s.storage.GetEntry(timestamp)
	if err != nil {
		return nil, "", fmt.Errorf("entry not found: %s", timestampStr)
	}

	filePath, err := s.storage.GetEntryPath(timestamp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get entry path: %w", err)
	}

	return entry, filePath, nil
}

// selectByDate filters entries by date and picks one (interactive if multiple)
func (s *EntrySelector) selectByDate(dateStr string) (*internal.JournalEntry, string, error) {
	// Parse date
	date, err := ParseDate(dateStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid date: %w", err)
	}

	// Create filter for the day
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	filter := internal.EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	// List matching entries
	entries, err := s.storage.ListEntries(filter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list entries: %w", err)
	}

	if len(entries) == 0 {
		return nil, "", fmt.Errorf("no entries found for %s", dateStr)
	}

	// If only one entry, use it
	if len(entries) == 1 {
		entry := entries[0]
		filePath, err := s.storage.GetEntryPath(entry.Timestamp)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get entry path: %w", err)
		}
		return entry, filePath, nil
	}

	// Multiple entries → interactive picker
	fmt.Printf("Found %d entries for %s:\n", len(entries), dateStr)
	entry, err := PickEntry(entries)
	if err != nil {
		return nil, "", err
	}

	filePath, err := s.storage.GetEntryPath(entry.Timestamp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get entry path: %w", err)
	}

	return entry, filePath, nil
}
