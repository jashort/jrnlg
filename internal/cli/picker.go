package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jashort/jrnlg/internal"
)

// PickEntry presents an interactive picker for selecting an entry from a list
// Returns the selected entry or error if canceled or invalid selection
func PickEntry(entries []*internal.JournalEntry) (*internal.JournalEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries to pick from")
	}

	// If only one entry, return it automatically
	if len(entries) == 1 {
		return entries[0], nil
	}

	// Display entries with numbers
	fmt.Printf("\nFound %d entries:\n\n", len(entries))

	// Display each entry with simple formatting
	for i, entry := range entries {
		fmt.Printf("%d. %s\n   %s\n\n",
			i+1,
			internal.FormatTimestamp(entry.Timestamp),
			TruncateBody(entry.Body, 70))
	}

	// Prompt for selection
	fmt.Printf("Select entry (1-%d, or 0 to cancel): ", len(entries))

	// Read input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Parse selection
	input = strings.TrimSpace(input)
	selection, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("invalid selection: must be a number")
	}

	// Validate selection
	if selection == 0 {
		return nil, fmt.Errorf("selection canceled")
	}

	if selection < 1 || selection > len(entries) {
		return nil, fmt.Errorf("invalid selection: must be 1-%d or 0 to cancel", len(entries))
	}

	// Return selected entry (convert 1-based to 0-based index)
	return entries[selection-1], nil
}
