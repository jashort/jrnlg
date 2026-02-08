package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SearchArgs contains parsed search arguments
type SearchArgs struct {
	Tags     []string
	Mentions []string
	Keywords []string
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
	Format   string // "full", "summary", "json"
	Reverse  bool
}

// parseSearchArgs parses command-line arguments into SearchArgs
func parseSearchArgs(args []string) (SearchArgs, error) {
	result := SearchArgs{
		Format: "full", // Default format
		Limit:  0,      // No limit by default
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle flags
		if strings.HasPrefix(arg, "-") {
			switch arg {
			case "-from":
				i++
				if i >= len(args) {
					return result, fmt.Errorf("-from requires a date argument")
				}
				date, err := ParseDate(args[i])
				if err != nil {
					return result, fmt.Errorf("invalid date for -from: %w", err)
				}
				// Truncate to start of day (00:00:00) for date comparison
				date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
				result.FromDate = &date

			case "-to":
				i++
				if i >= len(args) {
					return result, fmt.Errorf("-to requires a date argument")
				}
				date, err := ParseDate(args[i])
				if err != nil {
					return result, fmt.Errorf("invalid date for -to: %w", err)
				}
				// Truncate to end of day (23:59:59) for inclusive date comparison
				date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
				result.ToDate = &date

			case "-n", "--limit":
				i++
				if i >= len(args) {
					return result, fmt.Errorf("%s requires a number", arg)
				}
				limit, err := strconv.Atoi(args[i])
				if err != nil || limit < 0 {
					return result, fmt.Errorf("invalid limit: must be a positive number")
				}
				result.Limit = limit

			case "--offset":
				i++
				if i >= len(args) {
					return result, fmt.Errorf("--offset requires a number")
				}
				offset, err := strconv.Atoi(args[i])
				if err != nil || offset < 0 {
					return result, fmt.Errorf("invalid offset: must be a positive number")
				}
				result.Offset = offset

			case "--summary":
				result.Format = "summary"

			case "--format":
				i++
				if i >= len(args) {
					return result, fmt.Errorf("--format requires an argument")
				}
				format := args[i]
				if format != "full" && format != "summary" && format != "json" {
					return result, fmt.Errorf("invalid format: %s (must be full, summary, or json)", format)
				}
				result.Format = format

			case "-r", "--reverse":
				result.Reverse = true

			default:
				return result, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// Parse search terms
			if strings.HasPrefix(arg, "#") {
				// Tag
				tag := strings.TrimPrefix(arg, "#")
				if tag != "" {
					result.Tags = append(result.Tags, tag)
				}
			} else if strings.HasPrefix(arg, "@") {
				// Mention
				mention := strings.TrimPrefix(arg, "@")
				if mention != "" {
					result.Mentions = append(result.Mentions, mention)
				}
			} else {
				// Keyword
				result.Keywords = append(result.Keywords, arg)
			}
		}
	}

	return result, nil
}
