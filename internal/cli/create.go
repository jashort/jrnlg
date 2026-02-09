package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal"
)

// CreateEntry opens an editor for the user to create a new journal entry
func (a *App) CreateEntry() error {
	// 1. Generate pre-populated template with current timestamp
	timestamp := time.Now()
	header := internal.FormatTimestamp(timestamp)
	template := fmt.Sprintf("## %s\n\n", header)

	// 2. Open editor with configured arguments
	content, err := OpenEditor(template, a.config.EditorArgs)
	if err != nil {
		return fmt.Errorf("cannot open editor: %w", err)
	}

	// 3. Handle empty entry (silent discard)
	if isEmptyEntry(content) {
		return nil // Exit silently
	}

	// 4. Parse entry
	entry, err := internal.ParseEntry(content)
	if err != nil {
		return fmt.Errorf("invalid entry format: %w", err)
	}

	// 5. Check for timestamp collision and warn user
	if a.hasCollision(entry.Timestamp) {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Another entry exists with timestamp %s\n",
			entry.Timestamp.Format("2006-01-02 3:04 PM"))
		_, _ = fmt.Fprintf(os.Stderr, "Saved with collision suffix\n")
	}

	// 6. Save entry
	if err := a.storage.SaveEntry(entry); err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	// 7. Confirmation
	fmt.Printf("Entry saved successfully.\n")
	fmt.Printf("Timestamp: %s\n", entry.Timestamp.Format("2006-01-02 3:04 PM"))

	return nil
}

// isEmptyEntry checks if the entry has no body text
func isEmptyEntry(content string) bool {
	entry, err := internal.ParseEntry(content)
	if err != nil {
		// If it doesn't parse, it's not empty (it's invalid)
		return false
	}
	return strings.TrimSpace(entry.Body) == ""
}

// hasCollision checks if an entry with this timestamp already exists
func (a *App) hasCollision(timestamp time.Time) bool {
	_, err := a.storage.GetEntry(timestamp)
	return err == nil // If no error, entry exists
}
