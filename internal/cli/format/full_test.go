package format

import (
	"strings"
	"testing"
	"time"

	"github.com/jashort/jrnlg/internal"
)

func TestFullFormatter_Empty(t *testing.T) {
	formatter := &FullFormatter{}
	entries := []*internal.JournalEntry{}

	result := formatter.Format(entries)

	expected := "Found 0 entries.\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFullFormatter_SingleEntry(t *testing.T) {
	formatter := &FullFormatter{}

	// Create test entry
	timestamp := time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC)
	entries := []*internal.JournalEntry{
		{
			Timestamp: timestamp,
			Tags:      []string{"work", "meeting"},
			Mentions:  []string{"alice"},
			Body:      "Had a great #meeting with @alice about #work today.",
		},
	}

	result := formatter.Format(entries)

	// Should contain header
	if !strings.Contains(result, "Found 1 entries:") {
		t.Error("Expected 'Found 1 entries:' in output")
	}

	// Should contain the formatted timestamp
	if !strings.Contains(result, "## Monday 2026-02-09") {
		t.Error("Expected formatted timestamp in output")
	}

	// Should contain the body
	if !strings.Contains(result, "Had a great #meeting with @alice about #work today.") {
		t.Error("Expected body text in output")
	}

	// Should NOT contain separator for single entry
	if strings.Contains(result, "---") {
		t.Error("Should not contain separator for single entry")
	}
}

func TestFullFormatter_MultipleEntries(t *testing.T) {
	formatter := &FullFormatter{}

	// Create test entries
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC),
			Tags:      []string{"feature"},
			Mentions:  []string{"bob"},
			Body:      "Working on new #feature with @bob.",
		},
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{"work"},
			Mentions:  []string{"alice"},
			Body:      "Discussed #work with @alice.",
		},
		{
			Timestamp: time.Date(2026, 2, 10, 9, 15, 0, 0, time.UTC),
			Tags:      []string{"review"},
			Mentions:  []string{},
			Body:      "Code #review completed.",
		},
	}

	result := formatter.Format(entries)

	// Should contain header with count
	if !strings.Contains(result, "Found 3 entries:") {
		t.Error("Expected 'Found 3 entries:' in output")
	}

	// Should contain all bodies
	if !strings.Contains(result, "Working on new #feature with @bob.") {
		t.Error("Expected first entry body in output")
	}
	if !strings.Contains(result, "Discussed #work with @alice.") {
		t.Error("Expected second entry body in output")
	}
	if !strings.Contains(result, "Code #review completed.") {
		t.Error("Expected third entry body in output")
	}

	// Should contain separators between entries (but not after last)
	separatorCount := strings.Count(result, "\n---\n")
	if separatorCount != 2 {
		t.Errorf("Expected 2 separators, found %d", separatorCount)
	}
}

func TestFullFormatter_MultilineBody(t *testing.T) {
	formatter := &FullFormatter{}

	body := "First line\n\nSecond paragraph\n\nThird paragraph"
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      body,
		},
	}

	result := formatter.Format(entries)

	// Should preserve multiline formatting
	if !strings.Contains(result, body) {
		t.Error("Expected multiline body to be preserved")
	}
}

func TestFullFormatter_TimezonePreservation(t *testing.T) {
	formatter := &FullFormatter{}

	// Create entry with PST timezone
	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 9, 14, 30, 0, 0, loc)

	entries := []*internal.JournalEntry{
		{
			Timestamp: timestamp,
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Test entry.",
		},
	}

	result := formatter.Format(entries)

	// Should contain timezone in formatted timestamp
	if !strings.Contains(result, "America/Los_Angeles") {
		t.Error("Expected timezone to be preserved in output")
	}
}

func TestFullFormatter_SpecialCharacters(t *testing.T) {
	formatter := &FullFormatter{}

	// Test entry with special characters that shouldn't break formatting
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Testing \"quotes\" and 'apostrophes' and <brackets> and &ampersands.",
		},
	}

	result := formatter.Format(entries)

	// Should contain the special characters as-is
	if !strings.Contains(result, "Testing \"quotes\" and 'apostrophes' and <brackets> and &ampersands.") {
		t.Error("Expected special characters to be preserved")
	}
}
