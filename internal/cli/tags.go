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

// HandleTagsCommand routes the tags subcommands
func (a *App) HandleTagsCommand(args []string) error {
	if len(args) == 0 {
		// Default: list tags
		return a.listTags([]string{})
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		return a.listTags(args[1:])
	case "rename":
		return a.renameTags(args[1:])
	default:
		// If first arg doesn't look like a flag, treat it as unknown subcommand
		if !strings.HasPrefix(subcommand, "-") {
			return fmt.Errorf("unknown subcommand: %s\nUsage: jrnlg tags [list|rename] [options]", subcommand)
		}
		// Otherwise treat as flags for list command
		return a.listTags(args)
	}
}

// HandleMentionsCommand routes the mentions subcommands
func (a *App) HandleMentionsCommand(args []string) error {
	if len(args) == 0 {
		// Default: list mentions
		return a.listMentions([]string{})
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		return a.listMentions(args[1:])
	case "rename":
		return a.renameMentions(args[1:])
	default:
		// If first arg doesn't look like a flag, treat it as unknown subcommand
		if !strings.HasPrefix(subcommand, "-") {
			return fmt.Errorf("unknown subcommand: %s\nUsage: jrnlg mentions [list|rename] [options]", subcommand)
		}
		// Otherwise treat as flags for list command
		return a.listMentions(args)
	}
}

// listTags displays all tags with their counts
func (a *App) listTags(args []string) error {
	// Parse flags
	orphanedOnly := false
	for _, arg := range args {
		if arg == "--orphaned" {
			orphanedOnly = true
		}
	}

	// Get statistics
	stats, err := a.storage.GetTagStatistics()
	if err != nil {
		return fmt.Errorf("failed to get tag statistics: %w", err)
	}

	if len(stats) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	// Filter orphaned if requested
	if orphanedOnly {
		filtered := make(map[string]int)
		for tag, count := range stats {
			if count == 1 {
				filtered[tag] = count
			}
		}
		stats = filtered

		if len(stats) == 0 {
			fmt.Println("No orphaned tags found.")
			return nil
		}
	}

	// Sort alphabetically
	sorted := sortStatisticsAlpha(stats)

	// Format output
	colorizer := color.New(color.Auto)
	for _, item := range sorted {
		fmt.Printf("%s (%d %s)\n",
			colorizer.Tag("#"+item.name),
			item.count,
			plural("entry", item.count),
		)
	}

	return nil
}

// listMentions displays all mentions with their counts
func (a *App) listMentions(args []string) error {
	// Parse flags
	orphanedOnly := false
	for _, arg := range args {
		if arg == "--orphaned" {
			orphanedOnly = true
		}
	}

	// Get statistics
	stats, err := a.storage.GetMentionStatistics()
	if err != nil {
		return fmt.Errorf("failed to get mention statistics: %w", err)
	}

	if len(stats) == 0 {
		fmt.Println("No mentions found.")
		return nil
	}

	// Filter orphaned if requested
	if orphanedOnly {
		filtered := make(map[string]int)
		for mention, count := range stats {
			if count == 1 {
				filtered[mention] = count
			}
		}
		stats = filtered

		if len(stats) == 0 {
			fmt.Println("No orphaned mentions found.")
			return nil
		}
	}

	// Sort alphabetically
	sorted := sortStatisticsAlpha(stats)

	// Format output
	colorizer := color.New(color.Auto)
	for _, item := range sorted {
		fmt.Printf("%s (%d %s)\n",
			colorizer.Mention("@"+item.name),
			item.count,
			plural("entry", item.count),
		)
	}

	return nil
}

// renameTags handles the tag rename subcommand
func (a *App) renameTags(args []string) error {
	// Parse args: OLD NEW [--dry-run] [--force]
	tagArgs, err := parseRenameArgs(args)
	if err != nil {
		return err
	}

	// Validate tag formats
	if err := validateTagName(tagArgs.OldName); err != nil {
		return fmt.Errorf("invalid old tag: %w", err)
	}
	if err := validateTagName(tagArgs.NewName); err != nil {
		return fmt.Errorf("invalid new tag: %w", err)
	}

	// Check if old tag exists
	filePaths, err := a.storage.GetEntriesWithTag(tagArgs.OldName)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		fmt.Printf("No entries found with #%s\n", tagArgs.OldName)
		return nil
	}

	// Check if new tag already exists (WARN - merging will occur)
	existingNew, _ := a.storage.GetEntriesWithTag(tagArgs.NewName)
	if len(existingNew) > 0 {
		fmt.Printf("⚠ Warning: #%s already exists in %d %s (tags will be merged)\n\n",
			tagArgs.NewName,
			len(existingNew),
			plural("entry", len(existingNew)),
		)
	}

	// Show preview (first 5 entries)
	fmt.Printf("Found %d %s with #%s:\n\n",
		len(filePaths),
		plural("entry", len(filePaths)),
		tagArgs.OldName,
	)

	if !tagArgs.Force && !tagArgs.DryRun {
		showPreview(filePaths, 5)
		if len(filePaths) > 5 {
			fmt.Printf("... and %d more\n\n", len(filePaths)-5)
		}
	}

	// Dry run
	if tagArgs.DryRun {
		fmt.Printf("Would rename #%s to #%s in %d %s\n",
			tagArgs.OldName,
			tagArgs.NewName,
			len(filePaths),
			plural("entry", len(filePaths)),
		)
		return nil
	}

	// Confirmation
	if !tagArgs.Force {
		fmt.Printf("Rename #%s to #%s in %d %s? (y/N): ",
			tagArgs.OldName,
			tagArgs.NewName,
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

	updated, err := a.storage.ReplaceTagInEntries(
		tagArgs.OldName,
		tagArgs.NewName,
		false,
	)
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

// renameMentions handles the mention rename subcommand
func (a *App) renameMentions(args []string) error {
	// Parse args: OLD NEW [--dry-run] [--force]
	mentionArgs, err := parseRenameArgs(args)
	if err != nil {
		return err
	}

	// Validate mention formats
	if err := validateMentionName(mentionArgs.OldName); err != nil {
		return fmt.Errorf("invalid old mention: %w", err)
	}
	if err := validateMentionName(mentionArgs.NewName); err != nil {
		return fmt.Errorf("invalid new mention: %w", err)
	}

	// Check if old mention exists
	filePaths, err := a.storage.GetEntriesWithMention(mentionArgs.OldName)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		fmt.Printf("No entries found with @%s\n", mentionArgs.OldName)
		return nil
	}

	// Check if new mention already exists (WARN - merging will occur)
	existingNew, _ := a.storage.GetEntriesWithMention(mentionArgs.NewName)
	if len(existingNew) > 0 {
		fmt.Printf("⚠ Warning: @%s already exists in %d %s (mentions will be merged)\n\n",
			mentionArgs.NewName,
			len(existingNew),
			plural("entry", len(existingNew)),
		)
	}

	// Show preview (first 5 entries)
	fmt.Printf("Found %d %s with @%s:\n\n",
		len(filePaths),
		plural("entry", len(filePaths)),
		mentionArgs.OldName,
	)

	if !mentionArgs.Force && !mentionArgs.DryRun {
		showPreview(filePaths, 5)
		if len(filePaths) > 5 {
			fmt.Printf("... and %d more\n\n", len(filePaths)-5)
		}
	}

	// Dry run
	if mentionArgs.DryRun {
		fmt.Printf("Would rename @%s to @%s in %d %s\n",
			mentionArgs.OldName,
			mentionArgs.NewName,
			len(filePaths),
			plural("entry", len(filePaths)),
		)
		return nil
	}

	// Confirmation
	if !mentionArgs.Force {
		fmt.Printf("Rename @%s to @%s in %d %s? (y/N): ",
			mentionArgs.OldName,
			mentionArgs.NewName,
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

	updated, err := a.storage.ReplaceMentionInEntries(
		mentionArgs.OldName,
		mentionArgs.NewName,
		false,
	)
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

type renameArgs struct {
	OldName string
	NewName string
	DryRun  bool
	Force   bool
}

type statItem struct {
	name  string
	count int
}

func parseRenameArgs(args []string) (*renameArgs, error) {
	result := &renameArgs{}

	// Extract non-flag arguments
	var positional []string
	for _, arg := range args {
		if arg == "--dry-run" {
			result.DryRun = true
		} else if arg == "--force" {
			result.Force = true
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		} else {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	if len(positional) != 2 {
		return nil, fmt.Errorf("usage: jrnlg tags rename OLD NEW [--dry-run] [--force]")
	}

	result.OldName = positional[0]
	result.NewName = positional[1]

	return result, nil
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

func validateTagName(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	// Must start with letter
	if !isLetter(rune(tag[0])) {
		return fmt.Errorf("tag must start with a letter")
	}

	// Can only contain alphanumeric, underscore, hyphen
	pattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !pattern.MatchString(tag) {
		return fmt.Errorf("tag can only contain letters, numbers, underscores, and hyphens")
	}

	if len(tag) > internal.MaxTagLength {
		return fmt.Errorf("tag exceeds maximum length of %d characters", internal.MaxTagLength)
	}

	return nil
}

func validateMentionName(mention string) error {
	if mention == "" {
		return fmt.Errorf("mention cannot be empty")
	}

	// Must start with letter
	if !isLetter(rune(mention[0])) {
		return fmt.Errorf("mention must start with a letter")
	}

	// Can only contain alphanumeric, underscore, hyphen
	pattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !pattern.MatchString(mention) {
		return fmt.Errorf("mention can only contain letters, numbers, underscores, and hyphens")
	}

	if len(mention) > internal.MaxMentionLength {
		return fmt.Errorf("mention exceeds maximum length of %d characters", internal.MaxMentionLength)
	}

	return nil
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
