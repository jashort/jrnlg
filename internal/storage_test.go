package internal

import (
	"strings"
	"testing"
	"time"
)

func Test_ParseEntry_Valid(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantTags        []string
		wantMentions    []string
		wantBodyContain string
		wantYear        int
		wantMonth       time.Month
		wantDay         int
		wantHour        int
		wantMinute      int
	}{
		{
			name: "original example with tags, mentions, and email",
			input: `
## Sunday 2026-02-08 8:31 AM PST

Worked on some #things with @Alice. Need to email bob@example.com.
`,
			wantTags:        []string{"things"},
			wantMentions:    []string{"alice"},
			wantBodyContain: "Worked on some #things",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         8,
			wantHour:        8,
			wantMinute:      31,
		},
		{
			name: "multiple tags and mentions",
			input: `
## Monday 2026-02-09 2:15 PM EST

Working with @Bob and @Alice on #golang #testing #project_v2.
`,
			wantTags:        []string{"golang", "project_v2", "testing"},
			wantMentions:    []string{"alice", "bob"},
			wantBodyContain: "Working with",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         9,
			wantHour:        14, // 2 PM in 24-hour
			wantMinute:      15,
		},
		{
			name: "case-insensitive deduplication",
			input: `
## Tuesday 2026-02-10 10:00 AM CST

Met @Alice and @alice and @ALICE. Tags: #Work #work #WORK.
`,
			wantTags:        []string{"work"},
			wantMentions:    []string{"alice"},
			wantBodyContain: "Met @Alice",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         10,
			wantHour:        10,
			wantMinute:      0,
		},
		{
			name: "tags and mentions with underscores",
			input: `
## Wednesday 2026-02-11 5:45 PM UTC

Working on #my_project with @john_doe and @jane_smith.
`,
			wantTags:        []string{"my_project"},
			wantMentions:    []string{"jane_smith", "john_doe"},
			wantBodyContain: "Working on",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         11,
			wantHour:        17, // 5 PM in 24-hour
			wantMinute:      45,
		},
		{
			name: "hyphenated tags are single tags",
			input: `
## Thursday 2026-02-12 9:00 AM MST

Topics: #machine-learning and #data-science.
`,
			wantTags:        []string{"data-science", "machine-learning"},
			wantMentions:    []string{},
			wantBodyContain: "Topics:",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         12,
			wantHour:        9,
			wantMinute:      0,
		},
		{
			name: "no tags or mentions",
			input: `
## Friday 2026-02-13 11:30 AM MST

Just a regular journal entry with no special markers.
`,
			wantTags:        []string{},
			wantMentions:    []string{},
			wantBodyContain: "regular journal",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         13,
			wantHour:        11,
			wantMinute:      30,
		},
		{
			name: "emails should NOT be mentions",
			input: `
## Saturday 2026-02-14 3:00 PM PST

Contact alice@example.com and bob@company.org for details.
`,
			wantTags:        []string{},
			wantMentions:    []string{},
			wantBodyContain: "Contact alice@example.com",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         14,
			wantHour:        15, // 3 PM
			wantMinute:      0,
		},
		{
			name: "12 PM (noon)",
			input: `
## Sunday 2026-02-15 12:00 PM PST

Noon entry.
`,
			wantTags:        []string{},
			wantMentions:    []string{},
			wantBodyContain: "Noon entry",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         15,
			wantHour:        12, // noon stays 12
			wantMinute:      0,
		},
		{
			name: "12 AM (midnight)",
			input: `
## Monday 2026-02-16 12:00 AM PST

Midnight entry.
`,
			wantTags:        []string{},
			wantMentions:    []string{},
			wantBodyContain: "Midnight entry",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         16,
			wantHour:        0, // midnight is 0
			wantMinute:      0,
		},
		{
			name: "multiline body",
			input: `
## Tuesday 2026-02-17 8:00 AM PST

First line.
Second line with #tag1.
Third line with @Person.
`,
			wantTags:        []string{"tag1"},
			wantMentions:    []string{"person"},
			wantBodyContain: "First line.\nSecond line",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         17,
			wantHour:        8,
			wantMinute:      0,
		},
		{
			name: "single character tags and mentions",
			input: `
## Wednesday 2026-02-18 1:00 PM PST

Tag #a and mention @b are valid.
`,
			wantTags:        []string{"a"},
			wantMentions:    []string{"b"},
			wantBodyContain: "Tag #a",
			wantYear:        2026,
			wantMonth:       2,
			wantDay:         18,
			wantHour:        13, // 1 PM
			wantMinute:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseEntry(tt.input)
			if err != nil {
				t.Fatalf("ParseEntry() error = %v, want nil", err)
			}

			// Check timestamp components
			if entry.Timestamp.Year() != tt.wantYear {
				t.Errorf("Year = %d, want %d", entry.Timestamp.Year(), tt.wantYear)
			}
			if entry.Timestamp.Month() != tt.wantMonth {
				t.Errorf("Month = %d, want %d", entry.Timestamp.Month(), tt.wantMonth)
			}
			if entry.Timestamp.Day() != tt.wantDay {
				t.Errorf("Day = %d, want %d", entry.Timestamp.Day(), tt.wantDay)
			}
			if entry.Timestamp.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d", entry.Timestamp.Hour(), tt.wantHour)
			}
			if entry.Timestamp.Minute() != tt.wantMinute {
				t.Errorf("Minute = %d, want %d", entry.Timestamp.Minute(), tt.wantMinute)
			}

			// Check tags
			if len(entry.Tags) != len(tt.wantTags) {
				t.Errorf("Tags length = %d, want %d. Got: %v, Want: %v",
					len(entry.Tags), len(tt.wantTags), entry.Tags, tt.wantTags)
			} else {
				for i, tag := range entry.Tags {
					if tag != tt.wantTags[i] {
						t.Errorf("Tags[%d] = %s, want %s", i, tag, tt.wantTags[i])
					}
				}
			}

			// Check mentions
			if len(entry.Mentions) != len(tt.wantMentions) {
				t.Errorf("Mentions length = %d, want %d. Got: %v, Want: %v",
					len(entry.Mentions), len(tt.wantMentions), entry.Mentions, tt.wantMentions)
			} else {
				for i, mention := range entry.Mentions {
					if mention != tt.wantMentions[i] {
						t.Errorf("Mentions[%d] = %s, want %s", i, mention, tt.wantMentions[i])
					}
				}
			}

			// Check body content
			if !strings.Contains(entry.Body, tt.wantBodyContain) {
				t.Errorf("Body does not contain %q. Got: %s", tt.wantBodyContain, entry.Body)
			}
		})
	}
}

func Test_ParseEntry_Errors(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErrMsg string
	}{
		{
			name: "missing header",
			input: `
This is just body text with no header.
`,
			wantErrMsg: "missing header",
		},
		{
			name: "empty body",
			input: `
## Friday 2026-02-13 8:00 AM PST

`,
			wantErrMsg: "empty body",
		},
		{
			name: "invalid date format",
			input: `
## Friday 02/13/2026 8:00 AM PST

Body text.
`,
			wantErrMsg: "invalid",
		},
		{
			name: "tag exceeds 80 characters",
			input: `
## Sunday 2026-02-08 8:00 AM PST

This has a #thisisaverylongtagnamethatexceedstheeightycharacterlimitandthereforeisnotvalidatall tag.
`,
			wantErrMsg: "tag exceeds maximum length",
		},
		{
			name: "mention exceeds 80 characters",
			input: `
## Sunday 2026-02-08 8:00 AM PST

Met with @thisisaverylongmentionnamethatexceedstheeightycharacterlimitandthereforeisnotvalid today.
`,
			wantErrMsg: "mention exceeds maximum length",
		},
		{
			name: "missing AM/PM (wrong number of components)",
			input: `
## Sunday 2026-02-08 8:00 PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
		{
			name: "invalid meridiem",
			input: `
## Sunday 2026-02-08 8:00 XM PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
		{
			name: "invalid hour (13)",
			input: `
## Sunday 2026-02-08 13:00 PM PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
		{
			name: "invalid minute (60)",
			input: `
## Sunday 2026-02-08 8:60 AM PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
		{
			name: "invalid minute (negative)",
			input: `
## Sunday 2026-02-08 8:-5 AM PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
		{
			name: "missing time colon",
			input: `
## Sunday 2026-02-08 800 AM PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
		{
			name: "invalid date (wrong separator)",
			input: `
## Sunday 2026/02/08 8:00 AM PST

Body text.
`,
			wantErrMsg: "invalid timestamp format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseEntry(tt.input)
			if err == nil {
				t.Fatalf("ParseEntry() error = nil, want error containing %q. Got entry: %+v", tt.wantErrMsg, entry)
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("ParseEntry() error = %v, want error containing %q", err, tt.wantErrMsg)
			}
		})
	}
}
