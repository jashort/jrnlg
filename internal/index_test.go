package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIndex_Build(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	loc, _ := time.LoadLocation("America/Los_Angeles")

	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"work", "meeting"},
			Mentions:  []string{"alice"},
			Body:      "Meeting with @Alice about #work #meeting.",
		},
		{
			Timestamp: time.Date(2026, 1, 20, 14, 0, 0, 0, loc),
			Tags:      []string{"personal"},
			Mentions:  []string{"bob"},
			Body:      "Lunch with @Bob. #personal stuff.",
		},
	}

	// Write entries to files
	var files []string
	for i, entry := range entries {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("entry%d.md", i))
		content := SerializeEntry(entry)
		_ = os.WriteFile(filePath, []byte(content), 0644)
		files = append(files, filePath)
	}

	// Build index
	index := NewIndex()
	parseFunc := func(path string) (*JournalEntry, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return ParseEntry(string(content))
	}

	err := index.Build(files, 2, parseFunc)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Verify index size
	if index.Size() != 2 {
		t.Errorf("Index size = %d, want 2", index.Size())
	}

	// Verify tag index
	workEntries := index.tagIndex["work"]
	if len(workEntries) != 1 {
		t.Errorf("Tag 'work' has %d entries, want 1", len(workEntries))
	}

	personalEntries := index.tagIndex["personal"]
	if len(personalEntries) != 1 {
		t.Errorf("Tag 'personal' has %d entries, want 1", len(personalEntries))
	}

	// Verify mention index
	aliceEntries := index.mentionIndex["alice"]
	if len(aliceEntries) != 1 {
		t.Errorf("Mention 'alice' has %d entries, want 1", len(aliceEntries))
	}

	// Verify body map
	if len(index.bodyMap) != 2 {
		t.Errorf("Body map has %d entries, want 2", len(index.bodyMap))
	}
}

func TestIndex_SearchByTags_Single(t *testing.T) {
	index := createTestIndex(t)

	// Search for single tag
	results := index.SearchByTags([]string{"work"})

	if len(results) != 2 {
		t.Errorf("SearchByTags(['work']) returned %d results, want 2", len(results))
	}
}

func TestIndex_SearchByTags_Multiple(t *testing.T) {
	index := createTestIndex(t)

	// Search for multiple tags (AND logic)
	results := index.SearchByTags([]string{"work", "meeting"})

	// Only one entry has both tags
	if len(results) != 1 {
		t.Errorf("SearchByTags(['work', 'meeting']) returned %d results, want 1", len(results))
	}

	if len(results) > 0 && !contains(results[0].Tags, "work") {
		t.Error("Result should have 'work' tag")
	}
	if len(results) > 0 && !contains(results[0].Tags, "meeting") {
		t.Error("Result should have 'meeting' tag")
	}
}

func TestIndex_SearchByTags_NotFound(t *testing.T) {
	index := createTestIndex(t)

	// Search for non-existent tag
	results := index.SearchByTags([]string{"nonexistent"})

	if results != nil {
		t.Errorf("SearchByTags(['nonexistent']) returned %d results, want nil", len(results))
	}
}

func TestIndex_SearchByTags_CaseInsensitive(t *testing.T) {
	index := createTestIndex(t)

	// Search with different case
	results := index.SearchByTags([]string{"WORK"})

	if len(results) != 2 {
		t.Errorf("SearchByTags(['WORK']) returned %d results, want 2 (case-insensitive)", len(results))
	}
}

func TestIndex_SearchByMentions_Single(t *testing.T) {
	index := createTestIndex(t)

	// Search for single mention
	results := index.SearchByMentions([]string{"alice"})

	if len(results) != 1 {
		t.Errorf("SearchByMentions(['alice']) returned %d results, want 1", len(results))
	}
}

func TestIndex_SearchByMentions_Multiple(t *testing.T) {
	index := createTestIndex(t)

	// Search for multiple mentions (AND logic)
	results := index.SearchByMentions([]string{"alice", "bob"})

	// No entry has both mentions
	if len(results) != 0 {
		t.Errorf("SearchByMentions(['alice', 'bob']) returned %d results, want 0", len(results))
	}
}

func TestIndex_SearchByMentions_CaseInsensitive(t *testing.T) {
	index := createTestIndex(t)

	// Search with different case
	results := index.SearchByMentions([]string{"ALICE"})

	if len(results) != 1 {
		t.Errorf("SearchByMentions(['ALICE']) returned %d results, want 1 (case-insensitive)", len(results))
	}
}

func TestIndex_SearchByKeyword(t *testing.T) {
	index := createTestIndex(t)

	// Search for keyword in body
	results := index.SearchByKeyword("meeting")

	if len(results) != 1 {
		t.Errorf("SearchByKeyword('meeting') returned %d results, want 1", len(results))
	}
}

func TestIndex_SearchByKeyword_CaseInsensitive(t *testing.T) {
	index := createTestIndex(t)

	// Search with different case
	results := index.SearchByKeyword("MEETING")

	if len(results) != 1 {
		t.Errorf("SearchByKeyword('MEETING') returned %d results, want 1 (case-insensitive)", len(results))
	}
}

func TestIndex_SearchByKeyword_PartialMatch(t *testing.T) {
	index := createTestIndex(t)

	// Search for partial word
	results := index.SearchByKeyword("meet")

	if len(results) != 1 {
		t.Errorf("SearchByKeyword('meet') returned %d results, want 1 (partial match)", len(results))
	}
}

func TestIndex_SearchByKeyword_NotFound(t *testing.T) {
	index := createTestIndex(t)

	// Search for non-existent keyword
	results := index.SearchByKeyword("nonexistent")

	if len(results) != 0 {
		t.Errorf("SearchByKeyword('nonexistent') returned %d results, want 0", len(results))
	}
}

func TestIndex_GetBody(t *testing.T) {
	index := createTestIndex(t)

	// Get body for first entry
	if len(index.entries) == 0 {
		t.Fatal("Index has no entries")
	}

	filePath := index.entries[0].FilePath
	body := index.GetBody(filePath)

	if body == "" {
		t.Error("GetBody() returned empty string")
	}

	if !contains([]string{body}, "Meeting") && !contains([]string{body}, "Lunch") {
		t.Errorf("GetBody() returned unexpected body: %s", body)
	}
}

// createTestIndex builds a test index with sample data
func createTestIndex(t *testing.T) *Index {
	t.Helper()

	tmpDir := t.TempDir()
	loc, _ := time.LoadLocation("America/Los_Angeles")

	entries := []*JournalEntry{
		{
			Timestamp: time.Date(2026, 1, 15, 9, 30, 0, 0, loc),
			Tags:      []string{"work", "meeting"},
			Mentions:  []string{"alice"},
			Body:      "Meeting with @Alice about project. #work #meeting",
		},
		{
			Timestamp: time.Date(2026, 1, 20, 14, 0, 0, 0, loc),
			Tags:      []string{"work"},
			Mentions:  []string{"bob"},
			Body:      "Lunch with @Bob. #work",
		},
		{
			Timestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, loc),
			Tags:      []string{"personal"},
			Mentions:  []string{"charlie"},
			Body:      "Coffee with @Charlie. #personal",
		},
	}

	// Write entries to files
	var files []string
	for i, entry := range entries {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("entry%d.md", i))
		content := SerializeEntry(entry)
		_ = os.WriteFile(filePath, []byte(content), 0644)
		files = append(files, filePath)
	}

	// Build index
	index := NewIndex()
	parseFunc := func(path string) (*JournalEntry, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return ParseEntry(string(content))
	}

	err := index.Build(files, 2, parseFunc)
	if err != nil {
		t.Fatalf("Failed to build test index: %v", err)
	}

	return index
}

// contains checks if a slice contains a string (case-insensitive)
func contains(slice []string, str string) bool {
	str = strings.ToLower(str)
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), str) {
			return true
		}
	}
	return false
}
