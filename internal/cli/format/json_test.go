package format

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jashort/jrnlg/internal"
)

func TestJSONFormatter_Empty(t *testing.T) {
	formatter := &JSONFormatter{}
	var entries []*internal.JournalEntry

	result := formatter.Format(entries)

	// Should be valid JSON
	var parsed []interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	// Should be empty array
	if len(parsed) != 0 {
		t.Errorf("Expected empty array, got %d elements", len(parsed))
	}
}

func TestJSONFormatter_SingleEntry(t *testing.T) {
	formatter := &JSONFormatter{}

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

	// Should be valid JSON
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	// Should have one element
	if len(parsed) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(parsed))
	}

	entry := parsed[0]

	// Check timestamp field
	if _, ok := entry["timestamp"]; !ok {
		t.Error("Expected 'timestamp' field")
	}

	// Check tags field
	tags, ok := entry["tags"].([]interface{})
	if !ok {
		t.Fatal("Expected 'tags' to be array")
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	// Check mentions field
	mentions, ok := entry["mentions"].([]interface{})
	if !ok {
		t.Fatal("Expected 'mentions' to be array")
	}
	if len(mentions) != 1 {
		t.Errorf("Expected 1 mention, got %d", len(mentions))
	}

	// Check body field
	body, ok := entry["body"].(string)
	if !ok {
		t.Fatal("Expected 'body' to be string")
	}
	if body != "Had a great #meeting with @alice about #work today." {
		t.Errorf("Unexpected body: %s", body)
	}
}

func TestJSONFormatter_MultipleEntries(t *testing.T) {
	formatter := &JSONFormatter{}

	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC),
			Tags:      []string{"feature"},
			Mentions:  []string{"bob"},
			Body:      "Working on #feature with @bob.",
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

	// Should be valid JSON
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	// Should have three elements
	if len(parsed) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(parsed))
	}

	// Check first entry
	if parsed[0]["body"] != "Working on #feature with @bob." {
		t.Error("First entry body mismatch")
	}

	// Check second entry
	if parsed[1]["body"] != "Discussed #work with @alice." {
		t.Error("Second entry body mismatch")
	}

	// Check third entry
	if parsed[2]["body"] != "Code #review completed." {
		t.Error("Third entry body mismatch")
	}
}

func TestJSONFormatter_RFC3339Timestamp(t *testing.T) {
	formatter := &JSONFormatter{}

	// Create entry with specific timestamp
	timestamp := time.Date(2026, 2, 9, 14, 30, 45, 0, time.UTC)
	entries := []*internal.JournalEntry{
		{
			Timestamp: timestamp,
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Test entry.",
		},
	}

	result := formatter.Format(entries)

	// Should contain RFC3339 formatted timestamp
	if !strings.Contains(result, "2026-02-09T14:30:45Z") {
		t.Error("Expected RFC3339 timestamp format")
	}

	// Should be parseable as RFC3339
	var parsed []map[string]interface{}
	_ = json.Unmarshal([]byte(result), &parsed)

	timestampStr, ok := parsed[0]["timestamp"].(string)
	if !ok {
		t.Fatal("Expected timestamp to be string")
	}

	parsedTime, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		t.Errorf("Expected RFC3339 parseable timestamp, got error: %v", err)
	}

	if !parsedTime.Equal(timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", timestamp, parsedTime)
	}
}

func TestJSONFormatter_TimezonePreservation(t *testing.T) {
	formatter := &JSONFormatter{}

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

	var parsed []map[string]interface{}
	_ = json.Unmarshal([]byte(result), &parsed)

	timestampStr := parsed[0]["timestamp"].(string)

	// Should include timezone offset
	if !strings.Contains(timestampStr, "-08:00") && !strings.Contains(timestampStr, "-07:00") {
		t.Error("Expected timezone offset in timestamp")
	}
}

func TestJSONFormatter_EmptyArrays(t *testing.T) {
	formatter := &JSONFormatter{}

	// Entry with no tags or mentions
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Simple entry.",
		},
	}

	result := formatter.Format(entries)

	var parsed []map[string]interface{}
	_ = json.Unmarshal([]byte(result), &parsed)

	entry := parsed[0]

	// Tags should be empty array, not null
	tags, ok := entry["tags"].([]interface{})
	if !ok {
		t.Error("Expected tags to be array")
	}
	if len(tags) != 0 {
		t.Errorf("Expected empty tags array, got %d elements", len(tags))
	}

	// Mentions should be empty array, not null
	mentions, ok := entry["mentions"].([]interface{})
	if !ok {
		t.Error("Expected mentions to be array")
	}
	if len(mentions) != 0 {
		t.Errorf("Expected empty mentions array, got %d elements", len(mentions))
	}
}

func TestJSONFormatter_SpecialCharacters(t *testing.T) {
	formatter := &JSONFormatter{}

	// Entry with special characters that need JSON escaping
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Testing \"quotes\" and \n newlines and \t tabs and \\ backslashes.",
		},
	}

	result := formatter.Format(entries)

	// Should be valid JSON (properly escaped)
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON with escaped characters, got error: %v", err)
	}

	// Body should be properly unescaped when parsed
	body := parsed[0]["body"].(string)
	if !strings.Contains(body, "\"quotes\"") {
		t.Error("Expected quotes in body")
	}
	if !strings.Contains(body, "\n") {
		t.Error("Expected newline in body")
	}
	if !strings.Contains(body, "\t") {
		t.Error("Expected tab in body")
	}
	if !strings.Contains(body, "\\") {
		t.Error("Expected backslash in body")
	}
}

func TestJSONFormatter_Indentation(t *testing.T) {
	formatter := &JSONFormatter{}

	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{"work"},
			Mentions:  []string{"alice"},
			Body:      "Test entry.",
		},
	}

	result := formatter.Format(entries)

	// Should be indented (contain newlines and spaces for readability)
	if !strings.Contains(result, "\n") {
		t.Error("Expected indented JSON (with newlines)")
	}

	// Should have proper indentation (2 spaces)
	if !strings.Contains(result, "  \"timestamp\"") {
		t.Error("Expected 2-space indentation")
	}
}

func TestJSONFormatter_UTF8Characters(t *testing.T) {
	formatter := &JSONFormatter{}

	// Entry with UTF-8 characters
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Testing emoji 😀 and unicode characters: ñ, ü, 日本語",
		},
	}

	result := formatter.Format(entries)

	// Should be valid JSON
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON with UTF-8, got error: %v", err)
	}

	// UTF-8 should be preserved
	body := parsed[0]["body"].(string)
	if !strings.Contains(body, "😀") {
		t.Error("Expected emoji to be preserved")
	}
	if !strings.Contains(body, "日本語") {
		t.Error("Expected Japanese characters to be preserved")
	}
}
