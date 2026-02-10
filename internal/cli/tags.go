package cli

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
)

// MetadataType represents the type of metadata (tag or mention)
type MetadataType string

const (
	MetadataTypeTag     MetadataType = "tag"
	MetadataTypeMention MetadataType = "mention"
)

// Symbol returns the symbol prefix for this metadata type (# or @)
func (m MetadataType) Symbol() string {
	if m == MetadataTypeTag {
		return "#"
	}
	return "@"
}

// Name returns the string name of the metadata type
func (m MetadataType) Name() string {
	return string(m)
}

// MaxLength returns the maximum length for this metadata type
func (m MetadataType) MaxLength() int {
	if m == MetadataTypeTag {
		return internal.MaxTagLength
	}
	return internal.MaxMentionLength
}

// listMentions displays all mentions with their counts
func (a *App) listMentions(orphanedOnly bool) error {
	return a.listMetadata(MetadataTypeMention, orphanedOnly)
}

// listTags displays all tags with their counts
func (a *App) listTags(orphanedOnly bool) error {
	return a.listMetadata(MetadataTypeTag, orphanedOnly)
}

// listMetadata is the unified function for listing tags or mentions
func (a *App) listMetadata(metadataType MetadataType, orphanedOnly bool) error {
	// Get statistics based on type
	var stats map[string]int
	var err error
	if metadataType == MetadataTypeTag {
		stats, err = a.storage.GetTagStatistics()
	} else {
		stats, err = a.storage.GetMentionStatistics()
	}

	if err != nil {
		return fmt.Errorf("failed to get %s statistics: %w", metadataType.Name(), err)
	}

	if len(stats) == 0 {
		fmt.Printf("No %ss found.\n", metadataType.Name())
		return nil
	}

	// Filter orphaned if requested
	if orphanedOnly {
		filtered := make(map[string]int)
		for name, count := range stats {
			if count == 1 {
				filtered[name] = count
			}
		}
		stats = filtered

		if len(stats) == 0 {
			fmt.Printf("No orphaned %ss found.\n", metadataType.Name())
			return nil
		}
	}

	// Sort alphabetically
	sorted := sortStatisticsAlpha(stats)

	// Format output
	colorizer := color.New(color.Auto)
	for _, item := range sorted {
		// Apply appropriate colorization
		var displayName string
		if metadataType == MetadataTypeTag {
			displayName = colorizer.Tag(metadataType.Symbol() + item.name)
		} else {
			displayName = colorizer.Mention(metadataType.Symbol() + item.name)
		}

		fmt.Printf("%s (%d %s)\n",
			displayName,
			item.count,
			plural("entry", item.count),
		)
	}

	return nil
}

// renameTags handles the tag rename subcommand (Kong-compatible signature)
func (a *App) renameTags(oldName, newName string, dryRun, force bool) error {
	return a.renameMetadata(oldName, newName, MetadataTypeTag, dryRun, force)
}

// renameMentions handles the mention rename subcommand (Kong-compatible signature)
func (a *App) renameMentions(oldName, newName string, dryRun, force bool) error {
	return a.renameMetadata(oldName, newName, MetadataTypeMention, dryRun, force)
}

// renameMetadata is the unified function for renaming tags or mentions
func (a *App) renameMetadata(oldName, newName string, metadataType MetadataType, dryRun, force bool) error {
	// Validate formats
	if err := validateMetadataName(oldName, metadataType); err != nil {
		return fmt.Errorf("invalid old %s: %w", metadataType.Name(), err)
	}
	if err := validateMetadataName(newName, metadataType); err != nil {
		return fmt.Errorf("invalid new %s: %w", metadataType.Name(), err)
	}

	// Get entries based on type
	var filePaths []string
	var err error
	if metadataType == MetadataTypeTag {
		filePaths, err = a.storage.GetEntriesWithTag(oldName)
	} else {
		filePaths, err = a.storage.GetEntriesWithMention(oldName)
	}

	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		fmt.Printf("No entries found with %s%s\n", metadataType.Symbol(), oldName)
		return nil
	}

	// Check if new name already exists (WARN - merging will occur)
	var existingNew []string
	if metadataType == MetadataTypeTag {
		existingNew, _ = a.storage.GetEntriesWithTag(newName)
	} else {
		existingNew, _ = a.storage.GetEntriesWithMention(newName)
	}

	if len(existingNew) > 0 {
		fmt.Printf("⚠ Warning: %s%s already exists in %d %s (%ss will be merged)\n\n",
			metadataType.Symbol(),
			newName,
			len(existingNew),
			plural("entry", len(existingNew)),
			metadataType.Name(),
		)
	}

	// Show preview (first 5 entries)
	fmt.Printf("Found %d %s with %s%s:\n\n",
		len(filePaths),
		plural("entry", len(filePaths)),
		metadataType.Symbol(),
		oldName,
	)

	if !force && !dryRun {
		showPreview(filePaths, 5)
		if len(filePaths) > 5 {
			fmt.Printf("... and %d more\n\n", len(filePaths)-5)
		}
	}

	// Dry run
	if dryRun {
		fmt.Printf("Would rename %s%s to %s%s in %d %s\n",
			metadataType.Symbol(),
			oldName,
			metadataType.Symbol(),
			newName,
			len(filePaths),
			plural("entry", len(filePaths)),
		)
		return nil
	}

	// Confirmation
	if !force {
		fmt.Printf("Rename %s%s to %s%s in %d %s? (y/N): ",
			metadataType.Symbol(),
			oldName,
			metadataType.Symbol(),
			newName,
			len(filePaths),
			plural("entry", len(filePaths)),
		)

		if !promptYes() {
			fmt.Println("Canceled")
			return nil
		}
	}

	// Execute with progress message
	if len(filePaths) > 1 {
		fmt.Printf("Updating %d %s...\n",
			len(filePaths),
			plural("entry", len(filePaths)),
		)
	}

	// Call appropriate replace function
	var updated []string
	if metadataType == MetadataTypeTag {
		updated, err = a.storage.ReplaceTagInEntries(oldName, newName, false)
	} else {
		updated, err = a.storage.ReplaceMentionInEntries(oldName, newName, false)
	}

	if err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	// Success message
	fmt.Printf("✓ Updated %d %s\n",
		len(updated),
		plural("entry", len(updated)),
	)

	return nil
}

// Helper types and functions

type statItem struct {
	name  string
	count int
}

func sortStatisticsAlpha(stats map[string]int) []statItem {
	items := make([]statItem, 0, len(stats))
	for name, count := range stats {
		items = append(items, statItem{name, count})
	}

	// Sort alphabetically by name
	sort.Slice(items, func(i, j int) bool {
		return items[i].name < items[j].name
	})

	return items
}

func showPreview(filePaths []string, maxCount int) {
	count := min(maxCount, len(filePaths))
	for i := 0; i < count; i++ {
		// Parse the file directly
		content, err := os.ReadFile(filePaths[i])
		if err != nil {
			continue
		}

		entry, err := internal.ParseEntry(string(content))
		if err != nil {
			continue
		}

		timestamp := internal.FormatTimestamp(entry.Timestamp)
		body := TruncateBody(entry.Body, 60)
		fmt.Printf("%d. %s\n   %s\n\n", i+1, timestamp, body)
	}
}

func promptYes() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	r := strings.ToLower(strings.TrimSpace(response))
	return r == "y" || r == "yes"
}

func plural(word string, count int) string {
	if count == 1 {
		return word
	}
	// Handle special cases
	if word == "entry" {
		return "entries"
	}
	return word + "s"
}

// validateMetadataName is the unified validation function for tags and mentions
func validateMetadataName(name string, metadataType MetadataType) error {
	if name == "" {
		return fmt.Errorf("%s cannot be empty", metadataType.Name())
	}

	// Must start with letter
	if !isLetter(rune(name[0])) {
		return fmt.Errorf("%s must start with a letter", metadataType.Name())
	}

	// Can only contain alphanumeric, underscore, hyphen
	pattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("%s can only contain letters, numbers, underscores, and hyphens", metadataType.Name())
	}

	if len(name) > metadataType.MaxLength() {
		return fmt.Errorf("%s exceeds maximum length of %d characters", metadataType.Name(), metadataType.MaxLength())
	}

	return nil
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
