package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal"
)

// DeleteArgs contains parsed delete arguments
type DeleteArgs struct {
	Timestamp string     // Specific entry timestamp (YYYY-MM-DD-HH-MM-SS)
	FromDate  *time.Time // Filter: start date
	ToDate    *time.Time // Filter: end date
	Force     bool       // Skip confirmation
}

// DeleteEntries handles the delete command
func (a *App) DeleteEntries(args []string) error {
	// Parse arguments
	deleteArgs, err := parseDeleteArgs(args)
	if err != nil {
		return err
	}

	// Find entries to delete
	var entries []*internal.JournalEntry
	selector := NewEntrySelector(a.storage)

	if deleteArgs.Timestamp != "" {
		// Delete specific entry by timestamp
		entry, _, err := selector.SelectEntry(deleteArgs.Timestamp)
		if err != nil {
			return fmt.Errorf("entry not found: %s", deleteArgs.Timestamp)
		}

		entries = []*internal.JournalEntry{entry}
	} else {
		// Delete by filter
		filter := internal.EntryFilter{
			StartDate: deleteArgs.FromDate,
			EndDate:   deleteArgs.ToDate,
		}

		entries, err = selector.SelectEntries(filter)
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}
	}

	// Check if any entries found
	if len(entries) == 0 {
		return fmt.Errorf("no entries found to delete")
	}

	// Preview and confirm deletion
	confirmed, err := confirmDeletion(entries, deleteArgs.Force)
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

// parseDeleteArgs parses command-line arguments for delete command
func parseDeleteArgs(args []string) (DeleteArgs, error) {
	result := DeleteArgs{}

	// First argument (if not a flag) is the timestamp
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		result.Timestamp = args[0]
		args = args[1:] // Remove timestamp from args
	}

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-from", "--from":
			i++
			if i >= len(args) {
				return result, fmt.Errorf("%s requires a date argument", arg)
			}
			date, err := ParseDate(args[i])
			if err != nil {
				return result, fmt.Errorf("invalid date for %s: %w", arg, err)
			}
			// Truncate to start of day
			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			result.FromDate = &date

		case "-to", "--to":
			i++
			if i >= len(args) {
				return result, fmt.Errorf("%s requires a date argument", arg)
			}
			date, err := ParseDate(args[i])
			if err != nil {
				return result, fmt.Errorf("invalid date for %s: %w", arg, err)
			}
			// Truncate to end of day
			date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
			result.ToDate = &date

		case "--force", "-f":
			result.Force = true

		default:
			return result, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return result, nil
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
