package internal

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	MaxTagLength     = 80
	MaxMentionLength = 80
)

var (
	// Matches: #letter followed by alphanumeric/underscore/hyphen
	// Hyphens will be split later to create multiple tags
	tagRegex = regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9_-]*)`)
	// Matches: @letter followed by alphanumeric/underscore, where @ is not preceded by alphanumeric
	// This excludes emails like bob@example.com
	mentionRegex = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])@([a-zA-Z][a-zA-Z0-9_]*)`)
)

type JournalEntry struct {
	Timestamp time.Time
	Tags      []string
	Mentions  []string
	Body      string
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
// Expected format: "Weekday YYYY-MM-DD H:MM AM/PM Location"
func parseTimestamp(header string) (time.Time, error) {
	parts := strings.Fields(header)

	if len(parts) != 5 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: expected '## Weekday YYYY-MM-DD H:MM AM/PM Location', got: %s", header)
	}

	weekdayStr := parts[0]
	dateStr := parts[1]
	timeStr := parts[2]
	meridiem := parts[3]
	locationStr := parts[4]

	// Validate meridiem
	if meridiem != "AM" && meridiem != "PM" {
		return time.Time{}, fmt.Errorf("invalid meridiem: expected AM or PM, got: %s", meridiem)
	}

	// Parse date (YYYY-MM-DD)
	dateParts := strings.Split(dateStr, "-")
	if len(dateParts) != 3 {
		return time.Time{}, fmt.Errorf("invalid date format: expected YYYY-MM-DD, got: %s", dateStr)
	}

	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year: %s", dateParts[0])
	}

	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month: %s", dateParts[1])
	}

	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day: %s", dateParts[2])
	}

	// Parse time (H:MM or HH:MM)
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: expected H:MM or HH:MM, got: %s", timeStr)
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour: %s", timeParts[0])
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minute: %s", timeParts[1])
	}

	// Validate hour and minute ranges
	if hour < 1 || hour > 12 {
		return time.Time{}, fmt.Errorf("invalid hour: must be between 1 and 12, got: %d", hour)
	}
	if minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("invalid minute: must be between 0 and 59, got: %d", minute)
	}

	// Convert to 24-hour format
	if meridiem == "AM" {
		if hour == 12 {
			hour = 0 // 12 AM is midnight (00:00)
		}
	} else { // PM
		if hour != 12 {
			hour += 12 // 1 PM = 13:00, 11 PM = 23:00, but 12 PM stays 12:00
		}
	}

	// Load timezone location
	location, err := time.LoadLocation(locationStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("unknown location: %s: %w", locationStr, err)
	}

	// Create time.Time
	timestamp := time.Date(year, time.Month(month), day, hour, minute, 0, 0, location)

	// Validate weekday matches
	if err := validateWeekday(weekdayStr, timestamp); err != nil {
		return time.Time{}, err
	}

	return timestamp, nil
}

// validateWeekday checks if the weekday string matches the actual date
func validateWeekday(weekdayStr string, date time.Time) error {
	expectedWeekday := date.Weekday().String()
	if weekdayStr != expectedWeekday {
		return fmt.Errorf("weekday mismatch: expected %s, got %s for date %s",
			expectedWeekday, weekdayStr, date.Format("2006-01-02"))
	}
	return nil
}

// extractTags finds all hashtags in text
func extractTags(text string) ([]string, error) {
	matches := tagRegex.FindAllStringSubmatch(text, -1)

	tagMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			fullTag := match[1]

			// Split on hyphens to create multiple tags
			// e.g., "machine-learning" becomes ["machine", "learning"]
			words := strings.Split(fullTag, "-")
			for _, word := range words {
				// Skip empty strings from consecutive hyphens
				if word == "" {
					continue
				}

				tag := strings.ToLower(word)

				// Validate length
				if len(tag) > MaxTagLength {
					return nil, fmt.Errorf("tag exceeds maximum length of %d characters: %s", MaxTagLength, tag)
				}

				// Validate tag starts with a letter (after splitting)
				if len(tag) > 0 && (tag[0] < 'a' || tag[0] > 'z') {
					continue // Skip tags that don't start with a letter after splitting
				}

				tagMap[tag] = true
			}
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
	matches := mentionRegex.FindAllStringSubmatch(text, -1)

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
