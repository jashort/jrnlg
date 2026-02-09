package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal/cli/color"
)

// SearchArgs contains parsed search arguments
type SearchArgs struct {
	Tags      []string
	Mentions  []string
	Keywords  []string
	FromDate  *time.Time
	ToDate    *time.Time
	Limit     int
	Offset    int
	Format    string // "full", "summary", "json"
	Reverse   bool
	ColorMode color.Mode // Color mode: auto, always, never
}

// parseSearchArgs parses command-line arguments into SearchArgs
func parseSearchArgs(args []string) (SearchArgs, error) {
	result := SearchArgs{
		Format:    "full",     // Default format
		Limit:     0,          // No limit by default
		ColorMode: color.Auto, // Default color mode
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle flags
		if strings.HasPrefix(arg, "-") {
			// Skip global flags (already handled in Run())
			if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
				continue
			}

			// Validate flag is known
			if !isKnownFlag(arg) {
				return result, unknownFlagError(arg)
			}

			// Parse the flag
			var err error
			i, err = parseFlag(arg, args, i, &result)
			if err != nil {
				return result, err
			}
		} else {
			// Parse search terms
			parseSearchTerm(arg, &result)
		}
	}

	return result, nil
}

// parseFlag handles parsing a single flag and its value (if applicable)
func parseFlag(flag string, args []string, index int, result *SearchArgs) (int, error) {
	switch flag {
	case "-from", "--from":
		index++
		if index >= len(args) {
			return index, fmt.Errorf("%s requires a date argument. Example: --from today", flag)
		}
		date, err := ParseDate(args[index])
		if err != nil {
			return index, fmt.Errorf("invalid date for %s: %w\n\nExamples: today, yesterday, \"3 days ago\", 2024-01-01", flag, err)
		}
		// Truncate to start of day (00:00:00) for date comparison
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		result.FromDate = &date

	case "-to", "--to":
		index++
		if index >= len(args) {
			return index, fmt.Errorf("%s requires a date argument. Example: --to today", flag)
		}
		date, err := ParseDate(args[index])
		if err != nil {
			return index, fmt.Errorf("invalid date for %s: %w\n\nExamples: today, yesterday, \"3 days ago\", 2024-01-01", flag, err)
		}
		// Truncate to end of day (23:59:59) for inclusive date comparison
		date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
		result.ToDate = &date

	case "-n", "--limit":
		index++
		if index >= len(args) {
			return index, fmt.Errorf("%s requires a number. Example: -n 10", flag)
		}
		limit, err := strconv.Atoi(args[index])
		if err != nil {
			return index, fmt.Errorf("invalid limit: %q is not a valid number", args[index])
		}
		if limit < 0 {
			return index, fmt.Errorf("invalid limit: must be a positive number (got %d)", limit)
		}
		result.Limit = limit

	case "--offset":
		index++
		if index >= len(args) {
			return index, fmt.Errorf("--offset requires a number. Example: --offset 5")
		}
		offset, err := strconv.Atoi(args[index])
		if err != nil {
			return index, fmt.Errorf("invalid offset: %q is not a valid number", args[index])
		}
		if offset < 0 {
			return index, fmt.Errorf("invalid offset: must be a positive number (got %d)", offset)
		}
		result.Offset = offset

	case "--summary":
		result.Format = "summary"

	case "--format":
		index++
		if index >= len(args) {
			return index, fmt.Errorf("--format requires an argument. Valid values: full, summary, json")
		}
		format := args[index]
		if format != "full" && format != "summary" && format != "json" {
			return index, fmt.Errorf("invalid format: %q (must be full, summary, or json)", format)
		}
		result.Format = format

	case "-r", "--reverse":
		result.Reverse = true

	case "--color":
		index++
		if index >= len(args) {
			return index, fmt.Errorf("--color requires an argument: auto, always, never")
		}
		mode, err := color.ParseMode(args[index])
		if err != nil {
			return index, err
		}
		result.ColorMode = mode

	default:
		return index, unknownFlagError(flag)
	}

	return index, nil
}

// parseSearchTerm extracts tags, mentions, or keywords from a search term
func parseSearchTerm(term string, result *SearchArgs) {
	if strings.HasPrefix(term, "#") {
		// Tag
		tag := strings.TrimPrefix(term, "#")
		if tag != "" {
			result.Tags = append(result.Tags, tag)
		}
	} else if strings.HasPrefix(term, "@") {
		// Mention
		mention := strings.TrimPrefix(term, "@")
		if mention != "" {
			result.Mentions = append(result.Mentions, mention)
		}
	} else {
		// Keyword
		result.Keywords = append(result.Keywords, term)
	}
}
