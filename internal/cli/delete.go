package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal"
)

// executeDelete performs the actual deletion logic
func (a *App) executeDelete(selector string, fromDate, toDate *time.Time, force bool) error {
	// Find entries to delete
	var entries []*internal.JournalEntry
	entrySelector := NewEntrySelector(a.storage)

	if selector != "" {
		// Delete specific entry by timestamp
		entry, _, err := entrySelector.SelectEntry(selector)
		if err != nil {
			return fmt.Errorf("entry not found: %s", selector)
		}

		entries = []*internal.JournalEntry{entry}
	} else {
		// Delete by filter
		filter := internal.EntryFilter{
			StartDate: fromDate,
			EndDate:   toDate,
		}

		var err error
		entries, err = entrySelector.SelectEntries(filter)
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}
	}

	// Check if any entries found
	if len(entries) == 0 {
		return fmt.Errorf("no entries found to delete")
	}

	// Preview and confirm deletion
	confirmed, err := confirmDeletion(entries, force)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Deletion canceled.")
		return nil
	}

	// Delete entries
	var deletedPaths []string
	var deleteErrors []string

	for _, entry := range entries {
		filePath, err := a.storage.GetEntryPath(entry.Timestamp)
		if err != nil {
			deleteErrors = append(deleteErrors, fmt.Sprintf("Failed to get path for %s: %v",
				entry.Timestamp.Format("2006-01-02 15:04:05"), err))
			continue
		}

		if err := a.storage.DeleteEntry(filePath); err != nil {
			deleteErrors = append(deleteErrors, fmt.Sprintf("Failed to delete %s: %v", filePath, err))
			continue
		}

		deletedPaths = append(deletedPaths, filePath)
	}

	// Report results
	if len(deletedPaths) > 0 {
		fmt.Printf("Successfully deleted %d entr", len(deletedPaths))
		if len(deletedPaths) == 1 {
			fmt.Println("y.")
		} else {
			fmt.Println("ies.")
		}
	}

	if len(deleteErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors encountered:\n")
		for _, errMsg := range deleteErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", errMsg)
		}
		return fmt.Errorf("failed to delete %d of %d entries", len(deleteErrors), len(entries))
	}

	return nil
}

// confirmDeletion shows preview and asks for confirmation unless force is true
func confirmDeletion(entries []*internal.JournalEntry, force bool) (bool, error) {
	// Show preview
	if len(entries) == 1 {
		fmt.Printf("\nThe following entry will be deleted:\n\n")
	} else {
		fmt.Printf("\nThe following %d entries will be deleted:\n\n", len(entries))
	}

	for i, entry := range entries {
		fmt.Printf("%d. %s\n   %s\n\n",
			i+1,
			internal.FormatTimestamp(entry.Timestamp),
			TruncateBody(entry.Body, 70))
	}

	// Skip confirmation if force flag
	if force {
		return true, nil
	}

	// Ask for confirmation
	fmt.Printf("Delete these entries? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}
