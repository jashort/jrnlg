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

// HandleTagsCommand routes the tags subcommands (legacy entry point)
func (a *App) HandleTagsCommand(args []string) error {
	if len(args) == 0 {
		// Default: list tags
		return a.listTags(false)
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		// Parse orphaned flag
		orphaned := false
		for _, arg := range args[1:] {
			if arg == "--orphaned" {
				orphaned = true
			}
		}
		return a.listTags(orphaned)
	case "rename":
		return a.renameTagsLegacy(args[1:])
	default:
		// If first arg doesn't look like a flag, treat it as unknown subcommand
		if !strings.HasPrefix(subcommand, "-") {
			return fmt.Errorf("unknown subcommand: %s\nUsage: jrnlg tags [list|rename] [options]", subcommand)
		}
		// Otherwise treat as flags for list command
		orphaned := false
		for _, arg := range args {
			if arg == "--orphaned" {
				orphaned = true
			}
		}
		return a.listTags(orphaned)
	}
}

// HandleMentionsCommand routes the mentions subcommands (legacy entry point)
func (a *App) HandleMentionsCommand(args []string) error {
	if len(args) == 0 {
		// Default: list mentions
		return a.listMentions(false)
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		// Parse orphaned flag
		orphaned := false
		for _, arg := range args[1:] {
			if arg == "--orphaned" {
				orphaned = true
			}
		}
		return a.listMentions(orphaned)
	case "rename":
		return a.renameMentionsLegacy(args[1:])
	default:
		// If first arg doesn't look like a flag, treat it as unknown subcommand
		if !strings.HasPrefix(subcommand, "-") {
			return fmt.Errorf("unknown subcommand: %s\nUsage: jrnlg mentions [list|rename] [options]", subcommand)
		}
		// Otherwise treat as flags for list command
		orphaned := false
		for _, arg := range args {
			if arg == "--orphaned" {
				orphaned = true
			}
		}
		return a.listMentions(orphaned)
	}
}

// listMentions displays all mentions with their counts
func (a *App) listMentions(orphanedOnly bool) error {
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

// listTags displays all tags with their counts
func (a *App) listTags(orphanedOnly bool) error {
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

// renameTags handles the tag rename subcommand (Kong-compatible signature)
func (a *App) renameTags(oldName, newName string, dryRun, force bool) error {
	// Validate tag formats
	if err := validateTagName(oldName); err != nil {
		return fmt.Errorf("invalid old tag: %w", err)
	}
	if err := validateTagName(newName); err != nil {
		return fmt.Errorf("invalid new tag: %w", err)
	}

	// Check if old tag exists
	filePaths, err := a.storage.GetEntriesWithTag(oldName)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		fmt.Printf("No entries found with #%s\n", oldName)
		return nil
	}

	// Check if new tag already exists (WARN - merging will occur)
	existingNew, _ := a.storage.GetEntriesWithTag(newName)
	if len(existingNew) > 0 {
		fmt.Printf("⚠ Warning: #%s already exists in %d %s (tags will be merged)\n\n",
			newName,
			len(existingNew),
			plural("entry", len(existingNew)),
		)
	}

	// Show preview (first 5 entries)
	fmt.Printf("Found %d %s with #%s:\n\n",
		len(filePaths),
		plural("entry", len(filePaths)),
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
		fmt.Printf("Would rename #%s to #%s in %d %s\n",
			oldName,
			newName,
			len(filePaths),
			plural("entry", len(filePaths)),
		)
		return nil
	}

	// Confirmation
	if !force {
		fmt.Printf("Rename #%s to #%s in %d %s? (y/N): ",
			oldName,
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

	updated, err := a.storage.ReplaceTagInEntries(
		oldName,
		newName,
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

// renameTagsLegacy handles the tag rename subcommand (legacy args-based signature)
func (a *App) renameTagsLegacy(args []string) error {
	// Parse args: OLD NEW [--dry-run] [--force]
	tagArgs, err := parseRenameArgs(args)
	if err != nil {
		return err
	}
	return a.renameTags(tagArgs.OldName, tagArgs.NewName, tagArgs.DryRun, tagArgs.Force)
}

// renameMentions handles the mention rename subcommand (Kong-compatible signature)
func (a *App) renameMentions(oldName, newName string, dryRun, force bool) error {
	// Validate mention formats
	if err := validateMentionName(oldName); err != nil {
		return fmt.Errorf("invalid old mention: %w", err)
	}
	if err := validateMentionName(newName); err != nil {
		return fmt.Errorf("invalid new mention: %w", err)
	}

	// Check if old mention exists
	filePaths, err := a.storage.GetEntriesWithMention(oldName)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		fmt.Printf("No entries found with @%s\n", oldName)
		return nil
	}

	// Check if new mention already exists (WARN - merging will occur)
	existingNew, _ := a.storage.GetEntriesWithMention(newName)
	if len(existingNew) > 0 {
		fmt.Printf("⚠ Warning: @%s already exists in %d %s (mentions will be merged)\n\n",
			newName,
			len(existingNew),
			plural("entry", len(existingNew)),
		)
	}

	// Show preview (first 5 entries)
	fmt.Printf("Found %d %s with @%s:\n\n",
		len(filePaths),
		plural("entry", len(filePaths)),
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
		fmt.Printf("Would rename @%s to @%s in %d %s\n",
			oldName,
			newName,
			len(filePaths),
			plural("entry", len(filePaths)),
		)
		return nil
	}

	// Confirmation
	if !force {
		fmt.Printf("Rename @%s to @%s in %d %s? (y/N): ",
			oldName,
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

	updated, err := a.storage.ReplaceMentionInEntries(
		oldName,
		newName,
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

// renameMentionsLegacy handles the mention rename subcommand (legacy args-based signature)
func (a *App) renameMentionsLegacy(args []string) error {
	// Parse args: OLD NEW [--dry-run] [--force]
	mentionArgs, err := parseRenameArgs(args)
	if err != nil {
		return err
	}
	return a.renameMentions(mentionArgs.OldName, mentionArgs.NewName, mentionArgs.DryRun, mentionArgs.Force)
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
