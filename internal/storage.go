package internal

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal/patterns"
)

const (
	MaxTagLength     = 80
	MaxMentionLength = 80
)

type JournalEntry struct {
	Timestamp time.Time
	Tags      []string
	Mentions  []string
	Body      string
}

// EntryFilter specifies criteria for filtering journal entries
type EntryFilter struct {
	StartDate *time.Time // Inclusive start date (nil = no start limit)
	EndDate   *time.Time // Inclusive end date (nil = no end limit)
	Limit     int        // Maximum number of results (0 = no limit)
	Offset    int        // Number of results to skip (0 = no offset)
}

// Matches returns true if the given timestamp matches this filter's date range
func (f EntryFilter) Matches(timestamp time.Time) bool {
	if f.StartDate != nil && timestamp.Before(*f.StartDate) {
		return false
	}
	if f.EndDate != nil && timestamp.After(*f.EndDate) {
		return false
	}
	return true
}

func ParseEntry(input string) (*JournalEntry, error) {
	// Extract header and body
	header, bodyLines, err := extractHeader(input)
	if err != nil {
		return nil, err
	}

	// Parse timestamp from header
	timestamp, err := parseTimestamp(header)
	if err != nil {
		return nil, err
	}

	// Extract and validate body
	body := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	if body == "" {
		return nil, fmt.Errorf("empty body: entry must contain body text")
	}

	// Extract tags from body
	tags, err := extractTags(body)
	if err != nil {
		return nil, err
	}

	// Extract mentions from body
	mentions, err := extractMentions(body)
	if err != nil {
		return nil, err
	}

	return &JournalEntry{
		Timestamp: timestamp,
		Tags:      tags,
		Mentions:  mentions,
		Body:      body,
	}, nil
}

// extractHeader finds the header line and returns it along with remaining body lines
func extractHeader(input string) (string, []string, error) {
	lines := strings.Split(input, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "##") {
			// Found header
			header := strings.TrimSpace(strings.TrimPrefix(trimmed, "##"))
			bodyLines := lines[i+1:]
			return header, bodyLines, nil
		}
	}

	return "", nil, fmt.Errorf("missing header: expected line starting with '##'")
}

// parseTimestamp parses the header into a time.Time
// Expected format: "Monday 2006-01-02 3:04 PM MST"
func parseTimestamp(header string) (time.Time, error) {
	// Try parsing with the standard format
	// The format string "Monday 2006-01-02 3:04 PM MST" handles:
	// - Weekday name (Monday, Tuesday, etc.)
	// - Date in YYYY-MM-DD format
	// - Time in 12-hour format with no leading zero (3:04 not 03:04)
	// - AM/PM
	// - Timezone abbreviation (PST, EST, MST, UTC, etc.)

	const layout = "Monday 2006-01-02 3:04 PM MST"
	timestamp, err := time.Parse(layout, header)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: expected 'Monday 2006-01-02 3:04 PM MST', got: %s (error: %w)", header, err)
	}

	return timestamp, nil
}

// extractTags finds all hashtags in text
func extractTags(text string) ([]string, error) {
	matches := patterns.Tag.FindAllStringSubmatch(text, -1)

	tagMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			fullTag := match[1]

			// Normalize to lowercase
			tag := strings.ToLower(fullTag)

			// Validate length
			if len(tag) > MaxTagLength {
				return nil, fmt.Errorf("tag exceeds maximum length of %d characters: %s", MaxTagLength, tag)
			}

			tagMap[tag] = true
		}
	}

	// Convert to sorted slice
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return tags, nil
}

// extractMentions finds all @mentions in text (excluding emails)
func extractMentions(text string) ([]string, error) {
	matches := patterns.Mention.FindAllStringSubmatch(text, -1)

	mentionMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			mention := strings.ToLower(match[1])

			// Validate length
			if len(mention) > MaxMentionLength {
				return nil, fmt.Errorf("mention exceeds maximum length of %d characters: %s", MaxMentionLength, mention)
			}

			mentionMap[mention] = true
		}
	}

	// Convert to sorted slice
	mentions := make([]string, 0, len(mentionMap))
	for mention := range mentionMap {
		mentions = append(mentions, mention)
	}
	sort.Strings(mentions)

	return mentions, nil
}
