package format

import (
	"strings"
	"testing"
	"time"

	"github.com/jashort/jrnlg/internal"
)

func TestSummaryFormatter_Empty(t *testing.T) {
	formatter := &SummaryFormatter{}
	var entries []*internal.JournalEntry

	result := formatter.Format(entries)

	expected := "Found 0 entries.\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSummaryFormatter_SingleEntry(t *testing.T) {
	formatter := &SummaryFormatter{}

	timestamp := time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC)
	entries := []*internal.JournalEntry{
		{
			Timestamp: timestamp,
			Tags:      []string{"work"},
			Mentions:  []string{"alice"},
			Body:      "Had a great #meeting with @alice about #work today.",
		},
	}

	result := formatter.Format(entries)

	// Should contain header
	if !strings.Contains(result, "Found 1 entries:") {
		t.Error("Expected 'Found 1 entries:' in output")
	}

	// Should contain timestamp in YYYY-MM-DD HH:MM TZ format
	if !strings.Contains(result, "2026-02-09 2:30 PM UTC") {
		t.Error("Expected timestamp in YYYY-MM-DD H:MM PM TZ format")
	}

	// Should contain pipe separator
	if !strings.Contains(result, "|") {
		t.Error("Expected pipe separator")
	}

	// Should contain the body text
	if !strings.Contains(result, "Had a great #meeting with @alice about #work today.") {
		t.Error("Expected body text in output")
	}
}

func TestSummaryFormatter_MultipleEntries(t *testing.T) {
	formatter := &SummaryFormatter{}

	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC),
			Body:      "First entry",
		},
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      "Second entry",
		},
		{
			Timestamp: time.Date(2026, 2, 10, 9, 15, 0, 0, time.UTC),
			Body:      "Third entry",
		},
	}

	result := formatter.Format(entries)

	// Should contain header with count
	if !strings.Contains(result, "Found 3 entries:") {
		t.Error("Expected 'Found 3 entries:' in output")
	}

	// Should be on separate lines (one per entry)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	// 1 header line + blank line + 3 entry lines = 5 lines
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}

	// Each entry line should contain timestamp and body
	if !strings.Contains(result, "2026-02-08 10:00 AM UTC") {
		t.Error("Expected first entry timestamp")
	}
	if !strings.Contains(result, "2026-02-09 2:30 PM UTC") {
		t.Error("Expected second entry timestamp")
	}
	if !strings.Contains(result, "2026-02-10 9:15 AM UTC") {
		t.Error("Expected third entry timestamp")
	}
}

func TestSummaryFormatter_Truncation(t *testing.T) {
	formatter := &SummaryFormatter{}

	// Create entry with body longer than 80 chars
	longBody := strings.Repeat("a", 100)
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      longBody,
		},
	}

	result := formatter.Format(entries)

	// Should truncate to 77 chars + "..."
	if strings.Contains(result, strings.Repeat("a", 80)) {
		t.Error("Expected body to be truncated")
	}

	if !strings.Contains(result, "...") {
		t.Error("Expected ellipsis after truncation")
	}

	// Extract the summary line
	lines := strings.Split(result, "\n")
	var summaryLine string
	for _, line := range lines {
		if strings.Contains(line, " | ") {
			summaryLine = line
			break
		}
	}

	// The part after the pipe should be at most 80 chars
	parts := strings.Split(summaryLine, " | ")
	if len(parts) < 2 {
		t.Fatal("Expected pipe separator in output")
	}

	preview := parts[1]
	if len(preview) > 80 {
		t.Errorf("Expected preview to be max 80 chars, got %d", len(preview))
	}
}

func TestSummaryFormatter_ExactlyEightyChars(t *testing.T) {
	formatter := &SummaryFormatter{}

	// Create entry with body exactly 80 chars (should not be truncated)
	body := strings.Repeat("a", 80)
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      body,
		},
	}

	result := formatter.Format(entries)

	// Should NOT add ellipsis for exactly 80 chars
	if strings.Contains(result, "...") {
		t.Error("Should not truncate body of exactly 80 chars")
	}

	// Should contain full body
	if !strings.Contains(result, body) {
		t.Error("Expected full body text")
	}
}

func TestSummaryFormatter_UnderEightyChars(t *testing.T) {
	formatter := &SummaryFormatter{}

	// Create entry with body under 80 chars
	body := "This is a short entry."
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      body,
		},
	}

	result := formatter.Format(entries)

	// Should NOT truncate
	if !strings.Contains(result, body) {
		t.Error("Expected full body text")
	}

	// Should NOT add ellipsis
	if strings.Contains(result, "...") {
		t.Error("Should not add ellipsis for short body")
	}
}

func TestSummaryFormatter_MultilineBody(t *testing.T) {
	formatter := &SummaryFormatter{}

	// Entry with multiple lines - should only show first line
	body := "First line of text\nSecond line\nThird line"
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      body,
		},
	}

	result := formatter.Format(entries)

	// Should contain first line
	if !strings.Contains(result, "First line of text") {
		t.Error("Expected first line of body")
	}

	// Should NOT contain second and third lines
	if strings.Contains(result, "Second line") {
		t.Error("Should not show second line in summary")
	}
	if strings.Contains(result, "Third line") {
		t.Error("Should not show third line in summary")
	}
}

func TestSummaryFormatter_WhitespaceHandling(t *testing.T) {
	formatter := &SummaryFormatter{}

	// Entry with leading/trailing whitespace
	body := "   Text with spaces   "
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      body,
		},
	}

	result := formatter.Format(entries)

	// Should trim whitespace
	if !strings.Contains(result, "Text with spaces") {
		t.Error("Expected trimmed text")
	}

	// Should not have leading spaces in preview
	parts := strings.Split(result, " | ")
	if len(parts) >= 2 {
		preview := parts[1]
		if strings.HasPrefix(preview, "   ") {
			t.Error("Should not have leading whitespace in preview")
		}
		if strings.HasSuffix(strings.TrimSpace(preview), "   ") {
			t.Error("Should not have trailing whitespace in preview")
		}
	}
}

func TestSummaryFormatter_EmptyLines(t *testing.T) {
	formatter := &SummaryFormatter{}

	// Entry with empty lines at start
	body := "\n\nActual content here"
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Body:      body,
		},
	}

	result := formatter.Format(entries)

	// Should skip empty lines and show first non-empty line
	if !strings.Contains(result, "Actual content here") {
		t.Error("Expected first non-empty line")
	}
}
