package internal

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Verify parsed entry matches original
	if !parsed.Timestamp.Equal(entry.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", parsed.Timestamp, entry.Timestamp)
	}
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
	if !retrieved.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", retrieved.Timestamp, original.Timestamp)
	}
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

		if !retrieved.Timestamp.Equal(original.Timestamp) {
			t.Errorf("Entry %d: timestamp mismatch", i)
		}
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

	// Verify timezone is preserved
	if retrieved.Timestamp.Location().String() != "America/New_York" {
		t.Errorf("Timezone not preserved: got %s, want America/New_York",
			retrieved.Timestamp.Location().String())
	}

	// Verify times are equal
	if !retrieved.Timestamp.Equal(entry.Timestamp) {
		t.Errorf("Timestamp not equal: got %v, want %v",
			retrieved.Timestamp, entry.Timestamp)
	}
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
