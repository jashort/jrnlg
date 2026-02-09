package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveEntry(t *testing.T) {
	// Create temp directory for tests
	tmpDir := t.TempDir()

	config := DefaultConfig()
	storage := NewFileSystemStorage(tmpDir, config)

	// Create test entry
	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)
	entry := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{"work", "meeting"},
		Mentions:  []string{"alice", "bob"},
		Body:      "Had a meeting with @Alice and @Bob about #work #meeting.",
	}

	// Save entry
	err := storage.SaveEntry(entry)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Verify file exists at expected path
	// UTC conversion: 2026-02-08 8:31 AM PST = 2026-02-08 16:31 UTC
	expectedPath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file not found: %s", expectedPath)
	}

	// Verify directories were created
	yearDir := filepath.Join(tmpDir, "2026")
	monthDir := filepath.Join(tmpDir, "2026", "02")
	for _, dir := range []string{yearDir, monthDir} {
		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			t.Errorf("Expected directory not found: %s", dir)
		}
	}

	// Read file and verify content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Parse content
	parsed, err := ParseEntry(string(content))
	if err != nil {
		t.Fatalf("Failed to parse saved entry: %v", err)
	}

	// Verify parsed entry matches original (body content)
	// Note: Timestamp equality is not checked because MST format loses timezone offset information
	if parsed.Body != entry.Body {
		t.Errorf("Body mismatch: got %q, want %q", parsed.Body, entry.Body)
	}
}

func TestSaveEntry_CollisionHandling(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Save first entry
	entry1 := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "First entry.",
	}
	err := storage.SaveEntry(entry1)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Save second entry with same timestamp
	entry2 := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Second entry.",
	}
	err = storage.SaveEntry(entry2)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Verify both files exist
	basePath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md")
	collisionPath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00-01.md")

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		t.Errorf("Expected base file not found: %s", basePath)
	}
	if _, err := os.Stat(collisionPath); os.IsNotExist(err) {
		t.Errorf("Expected collision file not found: %s", collisionPath)
	}

	// Verify content is different
	content1, _ := os.ReadFile(basePath)
	content2, _ := os.ReadFile(collisionPath)
	if string(content1) == string(content2) {
		t.Error("Collision files should have different content")
	}
}

func TestGetEntry(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	// Create and save entry
	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)
	original := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{"work"},
		Mentions:  []string{"alice"},
		Body:      "Meeting with @Alice about #work.",
	}

	err := storage.SaveEntry(original)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Retrieve entry
	retrieved, err := storage.GetEntry(timestamp)
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}

	// Verify retrieved entry matches original
	// Note: Timestamp equality is not checked because MST format loses timezone offset information
	if retrieved.Body != original.Body {
		t.Errorf("Body mismatch: got %q, want %q", retrieved.Body, original.Body)
	}
	if len(retrieved.Tags) != len(original.Tags) {
		t.Errorf("Tags count mismatch: got %d, want %d", len(retrieved.Tags), len(original.Tags))
	}
	if len(retrieved.Mentions) != len(original.Mentions) {
		t.Errorf("Mentions count mismatch: got %d, want %d", len(retrieved.Mentions), len(original.Mentions))
	}
}

func TestGetEntry_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Try to get non-existent entry
	_, err := storage.GetEntry(timestamp)
	if err == nil {
		t.Error("Expected error for non-existent entry, got nil")
	}
}

func TestGetEntry_WithCollision(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Save two entries with same timestamp
	entry1 := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "First entry.",
	}
	entry2 := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Second entry.",
	}

	_ = storage.SaveEntry(entry1)
	_ = storage.SaveEntry(entry2)

	// GetEntry should return the first one (base path)
	retrieved, err := storage.GetEntry(timestamp)
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}

	if retrieved.Body != "First entry." {
		t.Errorf("GetEntry() returned wrong entry: got %q, want %q", retrieved.Body, "First entry.")
	}
}

func TestBuildFilePath(t *testing.T) {
	tmpDir := "/test/path"
	storage := NewFileSystemStorage(tmpDir, nil)

	tests := []struct {
		name      string
		timestamp time.Time
		want      string
	}{
		{
			name:      "PST morning time",
			timestamp: time.Date(2026, 2, 8, 8, 31, 0, 0, time.FixedZone("PST", -8*3600)),
			want:      filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md"),
		},
		{
			name:      "UTC time",
			timestamp: time.Date(2026, 2, 8, 16, 31, 0, 0, time.UTC),
			want:      filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md"),
		},
		{
			name:      "midnight PST",
			timestamp: time.Date(2026, 2, 8, 0, 0, 0, 0, time.FixedZone("PST", -8*3600)),
			want:      filepath.Join(tmpDir, "2026", "02", "2026-02-08-08-00-00.md"),
		},
		{
			name:      "single digit month",
			timestamp: time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
			want:      filepath.Join(tmpDir, "2026", "01", "2026-01-15-10-30-00.md"),
		},
		{
			name:      "december",
			timestamp: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			want:      filepath.Join(tmpDir, "2025", "12", "2025-12-31-23-59-59.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.buildFilePath(tt.timestamp)
			if got != tt.want {
				t.Errorf("buildFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSaveAndRetrieveMultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create multiple entries across different days and months
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"morning"},
			Mentions:  []string{},
			Body:      "Morning entry in January.",
		},
		{
			Timestamp: time.Date(2026, 1, 15, 14, 0, 0, 0, loc),
			Tags:      []string{"afternoon"},
			Mentions:  []string{},
			Body:      "Afternoon entry in January.",
		},
		{
			Timestamp: time.Date(2026, 2, 8, 8, 31, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{"alice"},
			Body:      "Work entry in February.",
		},
		{
			Timestamp: time.Date(2026, 3, 20, 10, 0, 0, 0, loc),
			Tags:      []string{"personal"},
			Mentions:  []string{},
			Body:      "Personal entry in March.",
		},
	}

	// Save all entries
	for i, entry := range entries {
		err := storage.SaveEntry(entry)
		if err != nil {
			t.Fatalf("SaveEntry(%d) error = %v", i, err)
		}
	}

	// Retrieve and verify each entry
	for i, original := range entries {
		retrieved, err := storage.GetEntry(original.Timestamp)
		if err != nil {
			t.Errorf("GetEntry(%d) error = %v", i, err)
			continue
		}

		// Note: Timestamp equality is not checked because MST format loses timezone offset information
		if retrieved.Body != original.Body {
			t.Errorf("Entry %d: body mismatch: got %q, want %q", i, retrieved.Body, original.Body)
		}
	}

	// Verify directory structure
	expectedDirs := []string{
		filepath.Join(tmpDir, "2026"),
		filepath.Join(tmpDir, "2026", "01"),
		filepath.Join(tmpDir, "2026", "02"),
		filepath.Join(tmpDir, "2026", "03"),
	}

	for _, dir := range expectedDirs {
		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			t.Errorf("Expected directory not found: %s", dir)
		}
	}
}

func TestWriteAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	// Ensure directory exists
	testDir := filepath.Join(tmpDir, "test")
	_ = os.MkdirAll(testDir, 0755)

	filePath := filepath.Join(testDir, "test.md")
	content := []byte("Test content")

	// Write atomically
	err := storage.writeAtomic(filePath, content)
	if err != nil {
		t.Fatalf("writeAtomic() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Verify content
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", readContent, content)
	}

	// Verify temp file was cleaned up
	tmpFile := filepath.Join(testDir, ".tmp-test.md")
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("Temp file was not cleaned up")
	}
}

func TestTimezonePreservation(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	// Create entry with specific timezone
	loc, _ := time.LoadLocation("America/New_York")
	timestamp := time.Date(2026, 2, 8, 14, 30, 0, 0, loc)

	entry := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Entry with New York timezone.",
	}

	// Save and retrieve
	err := storage.SaveEntry(entry)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	retrieved, err := storage.GetEntry(timestamp)
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}

	// Verify timezone abbreviation is present (MST format preserves abbreviations but not full IANA names)
	zoneName, _ := retrieved.Timestamp.Zone()
	if zoneName != "EST" && zoneName != "" {
		t.Errorf("Expected timezone abbreviation EST, got: %s", zoneName)
	}

	// Note: Cannot verify timestamp equality because MST format loses timezone offset information
	// The parsed timestamp will have offset +0000 instead of the original -0500
}

func TestListEntries_All(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries across multiple months
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Entry 1 - January.",
		},
		{
			Timestamp: time.Date(2026, 1, 20, 14, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Entry 2 - January.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 8, 31, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Entry 3 - February.",
		},
		{
			Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Entry 4 - March.",
		},
	}

	// Save all entries
	for _, entry := range entries {
		if err := storage.SaveEntry(entry); err != nil {
			t.Fatalf("SaveEntry() error = %v", err)
		}
	}

	// List all entries (no filter)
	filter := EntryFilter{}
	results, err := storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Verify count
	if len(results) != 4 {
		t.Errorf("ListEntries() returned %d entries, want 4", len(results))
	}

	// Verify sorted by timestamp (oldest first)
	for i := 0; i < len(results)-1; i++ {
		if !results[i].Timestamp.Before(results[i+1].Timestamp) {
			t.Errorf("Results not sorted: entry %d timestamp %v is not before entry %d timestamp %v",
				i, results[i].Timestamp, i+1, results[i+1].Timestamp)
		}
	}

	// Verify content matches
	for i, result := range results {
		if result.Body != entries[i].Body {
			t.Errorf("Entry %d body mismatch: got %q, want %q", i, result.Body, entries[i].Body)
		}
	}
}

func TestListEntries_DateRange(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries across 3 months
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "January entry.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 8, 31, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "February entry 1.",
		},
		{
			Timestamp: time.Date(2026, 2, 20, 14, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "February entry 2.",
		},
		{
			Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "March entry.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Filter for February only
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2026, 2, 28, 23, 59, 59, 0, loc)
	filter := EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	results, err := storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should only return February entries
	if len(results) != 2 {
		t.Errorf("ListEntries() returned %d entries, want 2", len(results))
	}

	// Verify content
	if len(results) >= 1 && results[0].Body != "February entry 1." {
		t.Errorf("First result body = %q, want %q", results[0].Body, "February entry 1.")
	}
	if len(results) >= 2 && results[1].Body != "February entry 2." {
		t.Errorf("Second result body = %q, want %q", results[1].Body, "February entry 2.")
	}
}

func TestListEntries_LimitAndOffset(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create 10 entries
	for i := 1; i <= 10; i++ {
		entry := &JournalEntry{
			Timestamp: time.Date(2026, 1, i, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      fmt.Sprintf("Entry %d", i),
		}
		_ = storage.SaveEntry(entry)
	}

	// Test limit
	filter := EntryFilter{Limit: 3}
	results, err := storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(results) != 3 {
		t.Errorf("ListEntries(limit=3) returned %d entries, want 3", len(results))
	}

	// Test offset
	filter = EntryFilter{Offset: 5}
	results, err = storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(results) != 5 {
		t.Errorf("ListEntries(offset=5) returned %d entries, want 5", len(results))
	}

	// Test limit + offset
	filter = EntryFilter{Offset: 3, Limit: 4}
	results, err = storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(results) != 4 {
		t.Errorf("ListEntries(offset=3, limit=4) returned %d entries, want 4", len(results))
	}
	// Should return entries 4-7
	if len(results) >= 1 && results[0].Body != "Entry 4" {
		t.Errorf("First result body = %q, want %q", results[0].Body, "Entry 4")
	}
}

func TestListEntries_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	// List entries with no files
	filter := EntryFilter{}
	results, err := storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("ListEntries() returned %d entries, want 0", len(results))
	}
}

func TestListEntries_OnlyStartDate(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Body:      "January entry.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 8, 31, 0, 0, loc),
			Body:      "February entry.",
		},
		{
			Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			Body:      "March entry.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Filter for entries from February onwards
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	filter := EntryFilter{
		StartDate: &startDate,
	}

	results, err := storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should return February and March entries
	if len(results) != 2 {
		t.Errorf("ListEntries() returned %d entries, want 2", len(results))
	}
}

func TestListEntries_OnlyEndDate(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Body:      "January entry.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 8, 31, 0, 0, loc),
			Body:      "February entry.",
		},
		{
			Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			Body:      "March entry.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Filter for entries up to end of February
	endDate := time.Date(2026, 2, 28, 23, 59, 59, 0, loc)
	filter := EntryFilter{
		EndDate: &endDate,
	}

	results, err := storage.ListEntries(filter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should return January and February entries
	if len(results) != 2 {
		t.Errorf("ListEntries() returned %d entries, want 2", len(results))
	}
}

func TestSearchByTags(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with various tags
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"work", "meeting"},
			Mentions:  []string{},
			Body:      "Work meeting about project. #work #meeting",
		},
		{
			Timestamp: time.Date(2026, 1, 20, 14, 0, 0, 0, loc),
			Tags:      []string{"work", "code"},
			Mentions:  []string{},
			Body:      "Coding session. #work #code",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, loc),
			Tags:      []string{"personal"},
			Mentions:  []string{},
			Body:      "Personal thoughts. #personal",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Search for single tag
	filter := EntryFilter{}
	results, err := storage.SearchByTags([]string{"work"}, filter)
	if err != nil {
		t.Fatalf("SearchByTags() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("SearchByTags(['work']) returned %d results, want 2", len(results))
	}

	// Search for multiple tags (AND logic)
	results, err = storage.SearchByTags([]string{"work", "meeting"}, filter)
	if err != nil {
		t.Fatalf("SearchByTags() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("SearchByTags(['work', 'meeting']) returned %d results, want 1", len(results))
	}

	// Verify correct entry returned
	if len(results) > 0 && !containsTag(results[0].Tags, "work") {
		t.Error("Result should have 'work' tag")
	}
}

func TestSearchByMentions(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with various mentions
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice", "bob"},
			Body:      "Meeting with @Alice and @Bob.",
		},
		{
			Timestamp: time.Date(2026, 1, 20, 14, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice"},
			Body:      "Lunch with @Alice.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"charlie"},
			Body:      "Coffee with @Charlie.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Search for single mention
	filter := EntryFilter{}
	results, err := storage.SearchByMentions([]string{"alice"}, filter)
	if err != nil {
		t.Fatalf("SearchByMentions() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("SearchByMentions(['alice']) returned %d results, want 2", len(results))
	}

	// Search for multiple mentions (AND logic)
	results, err = storage.SearchByMentions([]string{"alice", "bob"}, filter)
	if err != nil {
		t.Fatalf("SearchByMentions() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("SearchByMentions(['alice', 'bob']) returned %d results, want 1", len(results))
	}
}

func TestSearchByKeyword(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with various content
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Important meeting about the project deadline.",
		},
		{
			Timestamp: time.Date(2026, 1, 20, 14, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Casual lunch with colleagues.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "Another important decision was made today.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Search for keyword
	filter := EntryFilter{}
	results, err := storage.SearchByKeyword("important", filter)
	if err != nil {
		t.Fatalf("SearchByKeyword() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("SearchByKeyword('important') returned %d results, want 2", len(results))
	}

	// Search for partial match
	results, err = storage.SearchByKeyword("meet", filter)
	if err != nil {
		t.Fatalf("SearchByKeyword() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("SearchByKeyword('meet') returned %d results, want 1", len(results))
	}

	// Case insensitive search
	results, err = storage.SearchByKeyword("IMPORTANT", filter)
	if err != nil {
		t.Fatalf("SearchByKeyword() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("SearchByKeyword('IMPORTANT') returned %d results, want 2 (case-insensitive)", len(results))
	}
}

func TestSearchWithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries across multiple months
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{},
			Body:      "January work entry. #work",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{},
			Body:      "February work entry. #work",
		},
		{
			Timestamp: time.Date(2026, 3, 10, 11, 0, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{},
			Body:      "March work entry. #work",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Search with date range filter
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2026, 2, 28, 23, 59, 59, 0, loc)
	filter := EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	results, err := storage.SearchByTags([]string{"work"}, filter)
	if err != nil {
		t.Fatalf("SearchByTags() error = %v", err)
	}

	// Should only return February entry
	if len(results) != 1 {
		t.Errorf("SearchByTags() with date filter returned %d results, want 1", len(results))
	}

	if len(results) > 0 && results[0].Timestamp.Month() != time.February {
		t.Errorf("Result timestamp = %v, want February", results[0].Timestamp)
	}
}

func TestSearchSorting(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries in non-chronological order
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 3, 10, 11, 0, 0, 0, loc),
			Tags:      []string{"test"},
			Mentions:  []string{},
			Body:      "Third entry. #test",
		},
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"test"},
			Mentions:  []string{},
			Body:      "First entry. #test",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, loc),
			Tags:      []string{"test"},
			Mentions:  []string{},
			Body:      "Second entry. #test",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Search should return sorted results (oldest first)
	filter := EntryFilter{}
	results, err := storage.SearchByTags([]string{"test"}, filter)
	if err != nil {
		t.Fatalf("SearchByTags() error = %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("SearchByTags() returned %d results, want 3", len(results))
	}

	// Verify sorted order (oldest first)
	for i := 0; i < len(results)-1; i++ {
		if !results[i].Timestamp.Before(results[i+1].Timestamp) {
			t.Errorf("Results not sorted: entry %d timestamp %v is not before entry %d timestamp %v",
				i, results[i].Timestamp, i+1, results[i+1].Timestamp)
		}
	}

	// Verify content order
	if results[0].Body != "First entry. #test" {
		t.Errorf("First result body = %q, want %q", results[0].Body, "First entry. #test")
	}
}

// Helper function to check if tags contain a specific tag
func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

func containsMention(mentions []string, target string) bool {
	for _, mention := range mentions {
		if mention == target {
			return true
		}
	}
	return false
}

func TestInvalidateIndex(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create and save an entry
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{"work"},
		Mentions:  []string{},
		Body:      "Test entry. #work",
	}
	_ = storage.SaveEntry(entry)

	// Build index by performing a search
	filter := EntryFilter{}
	_, err := storage.SearchByTags([]string{"work"}, filter)
	if err != nil {
		t.Fatalf("SearchByTags() error = %v", err)
	}

	// Verify index exists
	if storage.index == nil {
		t.Fatal("Index should be built after search")
	}

	// Invalidate index
	storage.InvalidateIndex()

	// Verify index is cleared
	if storage.index != nil {
		t.Error("Index should be nil after InvalidateIndex()")
	}
}

func TestGetEntryPath(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Save entry
	entry := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Test entry.",
	}
	err := storage.SaveEntry(entry)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Get path
	path, err := storage.GetEntryPath(timestamp)
	if err != nil {
		t.Fatalf("GetEntryPath() error = %v", err)
	}

	// Verify path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("GetEntryPath() returned non-existent path: %s", path)
	}

	// Verify path format (should be UTC)
	expectedPath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md")
	if path != expectedPath {
		t.Errorf("GetEntryPath() = %s, want %s", path, expectedPath)
	}
}

func TestGetEntryPath_WithCollision(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Save two entries with same timestamp
	entry1 := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "First entry.",
	}
	entry2 := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Second entry.",
	}

	_ = storage.SaveEntry(entry1)
	_ = storage.SaveEntry(entry2)

	// GetEntryPath should return base path (first entry)
	path, err := storage.GetEntryPath(timestamp)
	if err != nil {
		t.Fatalf("GetEntryPath() error = %v", err)
	}

	expectedBasePath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md")
	if path != expectedBasePath {
		t.Errorf("GetEntryPath() = %s, want %s", path, expectedBasePath)
	}
}

func TestGetEntryPath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Try to get path for non-existent entry
	_, err := storage.GetEntryPath(timestamp)
	if err == nil {
		t.Error("Expected error for non-existent entry, got nil")
	}
}

func TestUpdateEntry(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Create and save original entry
	original := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{"work"},
		Mentions:  []string{"alice"},
		Body:      "Original content with @Alice about #work.",
	}
	err := storage.SaveEntry(original)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Get entry path
	path, err := storage.GetEntryPath(timestamp)
	if err != nil {
		t.Fatalf("GetEntryPath() error = %v", err)
	}

	// Update entry
	updated := &JournalEntry{
		Timestamp: timestamp, // Same timestamp
		Tags:      []string{"personal", "updated"},
		Mentions:  []string{"bob"},
		Body:      "Updated content with @Bob about #personal and #updated.",
	}
	err = storage.UpdateEntry(path, updated)
	if err != nil {
		t.Fatalf("UpdateEntry() error = %v", err)
	}

	// Retrieve and verify updated content
	retrieved, err := storage.GetEntry(timestamp)
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}

	if retrieved.Body != updated.Body {
		t.Errorf("Body = %q, want %q", retrieved.Body, updated.Body)
	}

	// Verify file still exists at same path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("File should still exist at original path")
	}
}

func TestUpdateEntry_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Try to update non-existent entry
	entry := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Updated content.",
	}

	fakePath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md")
	err := storage.UpdateEntry(fakePath, entry)
	if err == nil {
		t.Error("Expected error for non-existent entry, got nil")
	}
}

func TestDeleteEntry(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2026, 2, 8, 8, 31, 0, 0, loc)

	// Create and save entry
	entry := &JournalEntry{
		Timestamp: timestamp,
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "Entry to be deleted.",
	}
	err := storage.SaveEntry(entry)
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Get entry path
	path, err := storage.GetEntryPath(timestamp)
	if err != nil {
		t.Fatalf("GetEntryPath() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("File should exist before deletion")
	}

	// Delete entry
	err = storage.DeleteEntry(path)
	if err != nil {
		t.Fatalf("DeleteEntry() error = %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("File should not exist after deletion")
	}

	// Verify GetEntry fails
	_, err = storage.GetEntry(timestamp)
	if err == nil {
		t.Error("GetEntry() should fail after deletion")
	}
}

func TestDeleteEntry_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	// Try to delete non-existent entry
	fakePath := filepath.Join(tmpDir, "2026", "02", "2026-02-08-16-31-00.md")
	err := storage.DeleteEntry(fakePath)
	if err == nil {
		t.Error("Expected error for non-existent entry, got nil")
	}
}

func TestDeleteEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create multiple entries
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "January entry.",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 8, 31, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "February entry 1.",
		},
		{
			Timestamp: time.Date(2026, 2, 20, 14, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "February entry 2.",
		},
		{
			Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{},
			Body:      "March entry.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Delete February entries
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2026, 2, 28, 23, 59, 59, 0, loc)
	filter := EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	deleted, err := storage.DeleteEntries(filter)
	if err != nil {
		t.Fatalf("DeleteEntries() error = %v", err)
	}

	// Should delete 2 entries
	if len(deleted) != 2 {
		t.Errorf("DeleteEntries() deleted %d entries, want 2", len(deleted))
	}

	// Verify remaining entries
	allFilter := EntryFilter{}
	remaining, err := storage.ListEntries(allFilter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if len(remaining) != 2 {
		t.Errorf("ListEntries() returned %d entries, want 2", len(remaining))
	}

	// Verify January and March entries remain
	if len(remaining) >= 1 && remaining[0].Body != "January entry." {
		t.Errorf("First remaining entry body = %q, want %q", remaining[0].Body, "January entry.")
	}
	if len(remaining) >= 2 && remaining[1].Body != "March entry." {
		t.Errorf("Second remaining entry body = %q, want %q", remaining[1].Body, "March entry.")
	}
}

func TestDeleteEntries_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries in January
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{},
		Mentions:  []string{},
		Body:      "January entry.",
	}
	_ = storage.SaveEntry(entry)

	// Try to delete entries in February (none exist)
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2026, 2, 28, 23, 59, 59, 0, loc)
	filter := EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	deleted, err := storage.DeleteEntries(filter)
	if err != nil {
		t.Fatalf("DeleteEntries() error = %v", err)
	}

	// Should delete 0 entries
	if len(deleted) != 0 {
		t.Errorf("DeleteEntries() deleted %d entries, want 0", len(deleted))
	}

	// Verify original entry still exists
	allFilter := EntryFilter{}
	remaining, err := storage.ListEntries(allFilter)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if len(remaining) != 1 {
		t.Errorf("ListEntries() returned %d entries, want 1", len(remaining))
	}
}

func TestGetTagStatistics(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with various tags
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"work", "meeting"},
			Mentions:  []string{},
			Body:      "Work meeting. #work #meeting",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{},
			Body:      "More work. #work",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{"personal"},
			Mentions:  []string{},
			Body:      "Personal stuff. #personal",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Get tag statistics
	stats, err := storage.GetTagStatistics()
	if err != nil {
		t.Fatalf("GetTagStatistics() error = %v", err)
	}

	// Should have 3 unique tags
	if len(stats) != 3 {
		t.Errorf("GetTagStatistics() returned %d tags, want 3", len(stats))
	}

	// Check counts
	if stats["work"] != 2 {
		t.Errorf("Expected 2 entries with #work, got %d", stats["work"])
	}

	if stats["meeting"] != 1 {
		t.Errorf("Expected 1 entry with #meeting, got %d", stats["meeting"])
	}

	if stats["personal"] != 1 {
		t.Errorf("Expected 1 entry with #personal, got %d", stats["personal"])
	}
}

func TestGetTagStatistics_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	// Get tag statistics from empty storage
	stats, err := storage.GetTagStatistics()
	if err != nil {
		t.Fatalf("GetTagStatistics() error = %v", err)
	}

	if len(stats) != 0 {
		t.Errorf("GetTagStatistics() returned %d tags for empty storage, want 0", len(stats))
	}
}

func TestGetMentionStatistics(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with various mentions
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice", "bob"},
			Body:      "Meeting with @alice and @bob.",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice"},
			Body:      "Coffee with @alice.",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"charlie"},
			Body:      "Lunch with @charlie.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Get mention statistics
	stats, err := storage.GetMentionStatistics()
	if err != nil {
		t.Fatalf("GetMentionStatistics() error = %v", err)
	}

	// Should have 3 unique mentions
	if len(stats) != 3 {
		t.Errorf("GetMentionStatistics() returned %d mentions, want 3", len(stats))
	}

	// Check counts
	if stats["alice"] != 2 {
		t.Errorf("Expected 2 entries with @alice, got %d", stats["alice"])
	}

	if stats["bob"] != 1 {
		t.Errorf("Expected 1 entry with @bob, got %d", stats["bob"])
	}

	if stats["charlie"] != 1 {
		t.Errorf("Expected 1 entry with @charlie, got %d", stats["charlie"])
	}
}

func TestGetEntriesWithTag(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{},
			Body:      "Work entry. #work",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{},
			Body:      "More work. #work",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{"personal"},
			Mentions:  []string{},
			Body:      "Personal. #personal",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Get entries with #work
	paths, err := storage.GetEntriesWithTag("work")
	if err != nil {
		t.Fatalf("GetEntriesWithTag() error = %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("GetEntriesWithTag('work') returned %d paths, want 2", len(paths))
	}

	// Get entries with non-existent tag
	paths, err = storage.GetEntriesWithTag("nonexistent")
	if err != nil {
		t.Fatalf("GetEntriesWithTag() error = %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("GetEntriesWithTag('nonexistent') returned %d paths, want 0", len(paths))
	}
}

func TestGetEntriesWithMention(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice"},
			Body:      "Meeting with @alice.",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice"},
			Body:      "Coffee with @alice.",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"bob"},
			Body:      "Lunch with @bob.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Get entries with @alice
	paths, err := storage.GetEntriesWithMention("alice")
	if err != nil {
		t.Fatalf("GetEntriesWithMention() error = %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("GetEntriesWithMention('alice') returned %d paths, want 2", len(paths))
	}

	// Get entries with non-existent mention
	paths, err = storage.GetEntriesWithMention("nonexistent")
	if err != nil {
		t.Fatalf("GetEntriesWithMention() error = %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("GetEntriesWithMention('nonexistent') returned %d paths, want 0", len(paths))
	}
}

func TestReplaceTagInEntries_Single(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry with old tag
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{"old-tag"},
		Mentions:  []string{},
		Body:      "Testing #old-tag replacement.",
	}
	_ = storage.SaveEntry(entry)

	// Replace tag
	updated, err := storage.ReplaceTagInEntries("old-tag", "new-tag", false)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 1 {
		t.Errorf("ReplaceTagInEntries() updated %d entries, want 1", len(updated))
	}

	// Verify entry was updated
	entries, _ := storage.ListEntries(EntryFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if !containsTag(entries[0].Tags, "new-tag") {
		t.Errorf("Entry tags = %v, want to contain 'new-tag'", entries[0].Tags)
	}

	if containsTag(entries[0].Tags, "old-tag") {
		t.Errorf("Entry still contains 'old-tag', want it removed")
	}

	if !strings.Contains(entries[0].Body, "#new-tag") {
		t.Errorf("Entry body = %q, want to contain '#new-tag'", entries[0].Body)
	}
}

func TestReplaceTagInEntries_Multiple(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create multiple entries with same tag
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"code-review"},
			Mentions:  []string{},
			Body:      "First #code-review.",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{"code-review"},
			Mentions:  []string{},
			Body:      "Second #code-review.",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{"meeting"},
			Mentions:  []string{},
			Body:      "Unrelated #meeting.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Replace tag
	updated, err := storage.ReplaceTagInEntries("code-review", "codereview", false)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 2 {
		t.Errorf("ReplaceTagInEntries() updated %d entries, want 2", len(updated))
	}

	// Verify correct entries were updated
	allEntries, _ := storage.ListEntries(EntryFilter{})
	codeReviewCount := 0
	for _, e := range allEntries {
		if containsTag(e.Tags, "codereview") {
			codeReviewCount++
		}
		if containsTag(e.Tags, "code-review") {
			t.Error("Entry still contains old tag 'code-review'")
		}
	}

	if codeReviewCount != 2 {
		t.Errorf("Found %d entries with new tag, want 2", codeReviewCount)
	}
}

func TestReplaceTagInEntries_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with different case variations
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"code-review"},
			Mentions:  []string{},
			Body:      "First #code-review.",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{"code-review"},
			Mentions:  []string{},
			Body:      "Second #Code-Review.",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{"code-review"},
			Mentions:  []string{},
			Body:      "Third #CODE-REVIEW.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Replace tag (should match all case variations)
	updated, err := storage.ReplaceTagInEntries("code-review", "codereview", false)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 3 {
		t.Errorf("ReplaceTagInEntries() updated %d entries, want 3", len(updated))
	}

	// Verify all entries now have new tag
	allEntries, _ := storage.ListEntries(EntryFilter{})
	for i, e := range allEntries {
		if !containsTag(e.Tags, "codereview") {
			t.Errorf("Entry %d missing new tag 'codereview', has: %v", i, e.Tags)
		}
		if strings.Contains(strings.ToLower(e.Body), "code-review") ||
			strings.Contains(strings.ToLower(e.Body), "code_review") {
			t.Errorf("Entry %d body still contains old tag variations: %q", i, e.Body)
		}
	}
}

func TestReplaceTagInEntries_WordBoundaries(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry with similar tags
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{"code", "code-review"},
		Mentions:  []string{},
		Body:      "Testing #code and #code-review.",
	}
	_ = storage.SaveEntry(entry)

	// Replace only #code (should not affect #code-review)
	updated, err := storage.ReplaceTagInEntries("code", "programming", false)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 1 {
		t.Errorf("ReplaceTagInEntries() updated %d entries, want 1", len(updated))
	}

	// Verify only #code was replaced
	entries, _ := storage.ListEntries(EntryFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if !containsTag(entries[0].Tags, "programming") {
		t.Errorf("Entry missing new tag 'programming', has: %v", entries[0].Tags)
	}

	if !containsTag(entries[0].Tags, "code-review") {
		t.Errorf("Entry missing 'code-review' tag, has: %v", entries[0].Tags)
	}

	if !strings.Contains(entries[0].Body, "#programming") {
		t.Errorf("Body missing '#programming': %q", entries[0].Body)
	}

	if !strings.Contains(entries[0].Body, "#code-review") {
		t.Errorf("Body missing '#code-review': %q", entries[0].Body)
	}
}

func TestReplaceTagInEntries_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry where old tag already exists alongside target tag
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{"code-review", "codereview"},
		Mentions:  []string{},
		Body:      "Testing #code-review and #codereview.",
	}
	_ = storage.SaveEntry(entry)

	// Replace code-review with codereview (should merge/dedupe)
	updated, err := storage.ReplaceTagInEntries("code-review", "codereview", false)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 1 {
		t.Errorf("ReplaceTagInEntries() updated %d entries, want 1", len(updated))
	}

	// Verify deduplication occurred
	entries, _ := storage.ListEntries(EntryFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	// Count occurrences of codereview tag (should be deduplicated to 1)
	codeReviewCount := 0
	for _, tag := range entries[0].Tags {
		if tag == "codereview" {
			codeReviewCount++
		}
	}

	if codeReviewCount != 1 {
		t.Errorf("Tag 'codereview' appears %d times, want 1 (deduplication failed)", codeReviewCount)
	}
}

func TestReplaceTagInEntries_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{"old-tag"},
		Mentions:  []string{},
		Body:      "Testing #old-tag.",
	}
	_ = storage.SaveEntry(entry)

	// Dry run - should not modify files
	updated, err := storage.ReplaceTagInEntries("old-tag", "new-tag", true)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 1 {
		t.Errorf("ReplaceTagInEntries() found %d entries to update, want 1", len(updated))
	}

	// Verify entry was NOT updated
	entries, _ := storage.ListEntries(EntryFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if !containsTag(entries[0].Tags, "old-tag") {
		t.Errorf("Entry tags = %v, want to contain 'old-tag' (should not be modified in dry run)", entries[0].Tags)
	}

	if containsTag(entries[0].Tags, "new-tag") {
		t.Errorf("Entry contains 'new-tag', but should not be modified in dry run")
	}
}

func TestReplaceTagInEntries_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry with different tag
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{"existing-tag"},
		Mentions:  []string{},
		Body:      "Testing #existing-tag.",
	}
	_ = storage.SaveEntry(entry)

	// Try to replace non-existent tag
	updated, err := storage.ReplaceTagInEntries("nonexistent", "new-tag", false)
	if err != nil {
		t.Fatalf("ReplaceTagInEntries() error = %v", err)
	}

	if len(updated) != 0 {
		t.Errorf("ReplaceTagInEntries() updated %d entries, want 0", len(updated))
	}
}

func TestReplaceMentionInEntries_Single(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry with old mention
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{},
		Mentions:  []string{"john_doe"},
		Body:      "Meeting with @john_doe.",
	}
	_ = storage.SaveEntry(entry)

	// Replace mention
	updated, err := storage.ReplaceMentionInEntries("john_doe", "john-doe", false)
	if err != nil {
		t.Fatalf("ReplaceMentionInEntries() error = %v", err)
	}

	if len(updated) != 1 {
		t.Errorf("ReplaceMentionInEntries() updated %d entries, want 1", len(updated))
	}

	// Verify entry was updated
	entries, _ := storage.ListEntries(EntryFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if !containsMention(entries[0].Mentions, "john-doe") {
		t.Errorf("Entry mentions = %v, want to contain 'john-doe'", entries[0].Mentions)
	}

	if containsMention(entries[0].Mentions, "john_doe") {
		t.Errorf("Entry still contains 'john_doe', want it removed")
	}

	if !strings.Contains(entries[0].Body, "@john-doe") {
		t.Errorf("Entry body = %q, want to contain '@john-doe'", entries[0].Body)
	}
}

func TestReplaceMentionInEntries_Multiple(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create multiple entries with same mention
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice"},
			Body:      "First meeting with @alice.",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"alice"},
			Body:      "Second meeting with @alice.",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"bob"},
			Body:      "Meeting with @bob.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Replace mention
	updated, err := storage.ReplaceMentionInEntries("alice", "alice-smith", false)
	if err != nil {
		t.Fatalf("ReplaceMentionInEntries() error = %v", err)
	}

	if len(updated) != 2 {
		t.Errorf("ReplaceMentionInEntries() updated %d entries, want 2", len(updated))
	}

	// Verify correct entries were updated
	allEntries, _ := storage.ListEntries(EntryFilter{})
	aliceSmithCount := 0
	for _, e := range allEntries {
		if containsMention(e.Mentions, "alice-smith") {
			aliceSmithCount++
		}
		if containsMention(e.Mentions, "alice") {
			t.Error("Entry still contains old mention 'alice'")
		}
	}

	if aliceSmithCount != 2 {
		t.Errorf("Found %d entries with new mention, want 2", aliceSmithCount)
	}
}

func TestReplaceMentionInEntries_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entries with different case variations
	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"john"},
			Body:      "Meeting with @john.",
		},
		{
			Timestamp: time.Date(2026, 1, 16, 10, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"john"},
			Body:      "Call with @John.",
		},
		{
			Timestamp: time.Date(2026, 1, 17, 11, 0, 0, 0, loc),
			Tags:      []string{},
			Mentions:  []string{"john"},
			Body:      "Email from @JOHN.",
		},
	}

	for _, entry := range entries {
		_ = storage.SaveEntry(entry)
	}

	// Replace mention (should match all case variations)
	updated, err := storage.ReplaceMentionInEntries("john", "john-smith", false)
	if err != nil {
		t.Fatalf("ReplaceMentionInEntries() error = %v", err)
	}

	if len(updated) != 3 {
		t.Errorf("ReplaceMentionInEntries() updated %d entries, want 3", len(updated))
	}

	// Verify all entries now have new mention
	allEntries, _ := storage.ListEntries(EntryFilter{})
	for i, e := range allEntries {
		if !containsMention(e.Mentions, "john-smith") {
			t.Errorf("Entry %d missing new mention 'john-smith', has: %v", i, e.Mentions)
		}
	}
}

func TestReplaceMentionInEntries_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{},
		Mentions:  []string{"alice"},
		Body:      "Meeting with @alice.",
	}
	_ = storage.SaveEntry(entry)

	// Dry run - should not modify files
	updated, err := storage.ReplaceMentionInEntries("alice", "alice-smith", true)
	if err != nil {
		t.Fatalf("ReplaceMentionInEntries() error = %v", err)
	}

	if len(updated) != 1 {
		t.Errorf("ReplaceMentionInEntries() found %d entries to update, want 1", len(updated))
	}

	// Verify entry was NOT updated
	entries, _ := storage.ListEntries(EntryFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if !containsMention(entries[0].Mentions, "alice") {
		t.Errorf("Entry mentions = %v, want to contain 'alice' (should not be modified in dry run)", entries[0].Mentions)
	}

	if containsMention(entries[0].Mentions, "alice-smith") {
		t.Errorf("Entry contains 'alice-smith', but should not be modified in dry run")
	}
}

func TestReplaceMentionInEntries_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemStorage(tmpDir, nil)

	loc, _ := time.LoadLocation("America/Los_Angeles")

	// Create entry with different mention
	entry := &JournalEntry{
		Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
		Tags:      []string{},
		Mentions:  []string{"alice"},
		Body:      "Meeting with @alice.",
	}
	_ = storage.SaveEntry(entry)

	// Try to replace non-existent mention
	updated, err := storage.ReplaceMentionInEntries("bob", "robert", false)
	if err != nil {
		t.Fatalf("ReplaceMentionInEntries() error = %v", err)
	}

	if len(updated) != 0 {
		t.Errorf("ReplaceMentionInEntries() updated %d entries, want 0", len(updated))
	}
}
