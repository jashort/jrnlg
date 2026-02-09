package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal"
)

// EditEntry handles the edit command
// Selector can be:
// - Empty string: edit most recent entry
// - Timestamp (YYYY-MM-DD-HH-MM-SS): edit specific entry
// - Natural language date (yesterday, last week, etc.): filter and pick
func (a *App) EditEntry(args []string) error {
	// Parse selector (first argument)
	var selector string
	if len(args) > 0 {
		selector = args[0]
	}

	// Find entry to edit
	entry, filePath, err := a.findEntryToEdit(selector)
	if err != nil {
		return err
	}

	// Store original timestamp for validation
	originalTimestamp := entry.Timestamp

	// Serialize entry for editing
	content := internal.SerializeEntry(entry)

	// Open editor
	editedContent, err := OpenEditor(content, a.config.EditorArgs)
	if err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Check if entry was modified
	if strings.TrimSpace(editedContent) == strings.TrimSpace(content) {
		fmt.Println("No changes made.")
		return nil
	}

	// Parse edited content
	editedEntry, err := internal.ParseEntry(editedContent)
	if err != nil {
		return fmt.Errorf("invalid entry format after edit: %w", err)
	}

	// Validate timestamp wasn't changed
	if err := validateTimestampUnchanged(originalTimestamp, editedEntry.Timestamp); err != nil {
		return err
	}

	// Check if body is empty
	if strings.TrimSpace(editedEntry.Body) == "" {
		return fmt.Errorf("entry body cannot be empty")
	}

	// Update entry
	if err := a.storage.UpdateEntry(filePath, editedEntry); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	// Success
	fmt.Printf("Entry updated successfully.\n")
	fmt.Printf("Timestamp: %s\n", editedEntry.Timestamp.Format("2006-01-02 3:04 PM"))

	return nil
}

// findEntryToEdit locates the entry to edit based on the selector
// Returns the entry, its file path, and any error
func (a *App) findEntryToEdit(selector string) (*internal.JournalEntry, string, error) {
	// Case 1: No selector → most recent entry
	if selector == "" {
		return a.findMostRecentEntry()
	}

	// Case 2: Timestamp format (YYYY-MM-DD-HH-MM-SS)
	if isTimestampFormat(selector) {
		timestamp, err := parseTimestamp(selector)
		if err != nil {
			return nil, "", fmt.Errorf("invalid timestamp format: %w", err)
		}

		entry, err := a.storage.GetEntry(timestamp)
		if err != nil {
			return nil, "", fmt.Errorf("entry not found: %s", selector)
		}

		filePath, err := a.storage.GetEntryPath(timestamp)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get entry path: %w", err)
		}

		return entry, filePath, nil
	}

	// Case 3: Natural language date → filter and pick
	return a.findEntryByDate(selector)
}

// findMostRecentEntry returns the most recent entry
func (a *App) findMostRecentEntry() (*internal.JournalEntry, string, error) {
	// List all entries, limit 1, reverse order (newest first)
	filter := internal.EntryFilter{
		Limit: 1,
	}

	entries, err := a.storage.ListEntries(filter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list entries: %w", err)
	}

	if len(entries) == 0 {
		return nil, "", fmt.Errorf("no entries found")
	}

	// Get most recent (entries are sorted oldest first, so get last)
	entry := entries[len(entries)-1]

	filePath, err := a.storage.GetEntryPath(entry.Timestamp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get entry path: %w", err)
	}

	return entry, filePath, nil
}

// findEntryByDate filters entries by date and picks one (interactive if multiple)
func (a *App) findEntryByDate(dateStr string) (*internal.JournalEntry, string, error) {
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
	entries, err := a.storage.ListEntries(filter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list entries: %w", err)
	}

	if len(entries) == 0 {
		return nil, "", fmt.Errorf("no entries found for %s", dateStr)
	}

	// If only one entry, use it
	if len(entries) == 1 {
		entry := entries[0]
		filePath, err := a.storage.GetEntryPath(entry.Timestamp)
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

	filePath, err := a.storage.GetEntryPath(entry.Timestamp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get entry path: %w", err)
	}

	return entry, filePath, nil
}

// isTimestampFormat checks if a string matches the timestamp format YYYY-MM-DD-HH-MM-SS
func isTimestampFormat(s string) bool {
	// Simple check: 19 chars, dashes in right places
	if len(s) != 19 {
		return false
	}
	if s[4] != '-' || s[7] != '-' || s[10] != '-' || s[13] != '-' || s[16] != '-' {
		return false
	}
	return true
}

// parseTimestamp parses a timestamp string in format YYYY-MM-DD-HH-MM-SS
func parseTimestamp(s string) (time.Time, error) {
	// Parse as UTC (same as file naming)
	return time.Parse("2006-01-02-15-04-05", s)
}

// validateTimestampUnchanged ensures the timestamp hasn't been modified
func validateTimestampUnchanged(original, edited time.Time) error {
	// Compare timestamps (allow small differences due to parsing)
	// Format both to the same string format and compare
	originalStr := original.Format("Monday 2006-01-02 3:04 PM MST")
	editedStr := edited.Format("Monday 2006-01-02 3:04 PM MST")

	if originalStr != editedStr {
		return fmt.Errorf("timestamp cannot be changed during edit\nOriginal: %s\nEdited:   %s", originalStr, editedStr)
	}

	return nil
}
