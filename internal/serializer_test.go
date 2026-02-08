package internal

import (
	"strings"
	"testing"
	"time"
)

func TestSerializeEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry *JournalEntry
		want  string
	}{
		{
			name: "basic entry with tags and mentions",
			entry: &JournalEntry{
				Timestamp: mustParseTime("2026-02-08T08:31:00", "America/Los_Angeles"),
				Tags:      []string{"work", "meeting"},
				Mentions:  []string{"alice", "bob"},
				Body:      "Had a meeting with @Alice and @Bob about #work #meeting.",
			},
			want: "## Sunday 2026-02-08 8:31 AM America/Los_Angeles\n\nHad a meeting with @Alice and @Bob about #work #meeting.\n",
		},
		{
			name: "entry with 12 AM (midnight)",
			entry: &JournalEntry{
				Timestamp: mustParseTime("2026-02-08T00:15:00", "America/Los_Angeles"),
				Tags:      []string{},
				Mentions:  []string{},
				Body:      "Midnight thoughts.",
			},
			want: "## Sunday 2026-02-08 12:15 AM America/Los_Angeles\n\nMidnight thoughts.\n",
		},
		{
			name: "entry with 12 PM (noon)",
			entry: &JournalEntry{
				Timestamp: mustParseTime("2026-02-08T12:00:00", "America/Los_Angeles"),
				Tags:      []string{},
				Mentions:  []string{},
				Body:      "Lunch time!",
			},
			want: "## Sunday 2026-02-08 12:00 PM America/Los_Angeles\n\nLunch time!\n",
		},
		{
			name: "entry with multiline body",
			entry: &JournalEntry{
				Timestamp: mustParseTime("2026-02-08T14:30:00", "America/Los_Angeles"),
				Tags:      []string{"personal"},
				Mentions:  []string{},
				Body:      "Line one.\nLine two.\nLine three.",
			},
			want: "## Sunday 2026-02-08 2:30 PM America/Los_Angeles\n\nLine one.\nLine two.\nLine three.\n",
		},
		{
			name: "entry with UTC timezone",
			entry: &JournalEntry{
				Timestamp: mustParseTime("2026-02-08T16:31:00", "UTC"),
				Tags:      []string{},
				Mentions:  []string{},
				Body:      "UTC entry.",
			},
			want: "## Sunday 2026-02-08 4:31 PM UTC\n\nUTC entry.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SerializeEntry(tt.entry)
			if got != tt.want {
				t.Errorf("SerializeEntry() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
	}{
		{
			name:     "basic entry",
			markdown: "## Sunday 2026-02-08 8:31 AM America/Los_Angeles\n\nWorked on some #things with @Alice.",
		},
		{
			name:     "entry with 12 AM",
			markdown: "## Sunday 2026-02-08 12:15 AM America/Los_Angeles\n\nMidnight thoughts.",
		},
		{
			name:     "entry with 12 PM",
			markdown: "## Sunday 2026-02-08 12:00 PM America/Los_Angeles\n\nLunch time!",
		},
		{
			name:     "entry with multiline body",
			markdown: "## Sunday 2026-02-08 2:30 PM America/Los_Angeles\n\nLine one.\nLine two.\nLine three.",
		},
		{
			name:     "entry with UTC timezone",
			markdown: "## Sunday 2026-02-08 4:31 PM UTC\n\nUTC entry.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse markdown to entry
			entry, err := ParseEntry(tt.markdown)
			if err != nil {
				t.Fatalf("ParseEntry() error = %v", err)
			}

			// Serialize entry back to markdown
			got := SerializeEntry(entry)

			// Compare (should be identical)
			if !strings.HasSuffix(got, "\n") {
				got = got + "\n"
			}
			if !strings.HasSuffix(tt.markdown, "\n") {
				tt.markdown = tt.markdown + "\n"
			}

			if got != tt.markdown {
				t.Errorf("Round-trip failed:\noriginal:\n%q\ngot:\n%q", tt.markdown, got)
			}
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
		want      string
	}{
		{
			name:      "morning time",
			timestamp: mustParseTime("2026-02-08T08:31:00", "America/Los_Angeles"),
			want:      "Sunday 2026-02-08 8:31 AM America/Los_Angeles",
		},
		{
			name:      "afternoon time",
			timestamp: mustParseTime("2026-02-08T14:22:00", "America/Los_Angeles"),
			want:      "Sunday 2026-02-08 2:22 PM America/Los_Angeles",
		},
		{
			name:      "midnight (12 AM)",
			timestamp: mustParseTime("2026-02-08T00:00:00", "America/Los_Angeles"),
			want:      "Sunday 2026-02-08 12:00 AM America/Los_Angeles",
		},
		{
			name:      "noon (12 PM)",
			timestamp: mustParseTime("2026-02-08T12:00:00", "America/Los_Angeles"),
			want:      "Sunday 2026-02-08 12:00 PM America/Los_Angeles",
		},
		{
			name:      "single digit minute",
			timestamp: mustParseTime("2026-02-08T09:05:00", "America/Los_Angeles"),
			want:      "Sunday 2026-02-08 9:05 AM America/Los_Angeles",
		},
		{
			name:      "UTC timezone",
			timestamp: mustParseTime("2026-02-08T16:31:00", "UTC"),
			want:      "Sunday 2026-02-08 4:31 PM UTC",
		},
		{
			name:      "different timezone",
			timestamp: mustParseTime("2026-02-08T14:31:00", "America/New_York"),
			want:      "Sunday 2026-02-08 2:31 PM America/New_York",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimestamp(tt.timestamp)
			if got != tt.want {
				t.Errorf("formatTimestamp() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper function to parse time in a specific location
func mustParseTime(timeStr, locationStr string) time.Time {
	location, err := time.LoadLocation(locationStr)
	if err != nil {
		panic(err)
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05", timeStr, location)
	if err != nil {
		panic(err)
	}
	return t
}
