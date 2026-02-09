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

	// Find entry to edit using selector
	entrySelector := NewEntrySelector(a.storage)
	entry, filePath, err := entrySelector.SelectEntry(selector)
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
