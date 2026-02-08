package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jashort/jrnlg/internal"
)

// TestSearchIntegration tests the full search workflow with temporary storage
func TestSearchIntegration(t *testing.T) {
	// Create temporary directory for test storage
	tempDir, err := os.MkdirTemp("", "jrnlg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test config
	config := &internal.Config{
		StoragePath:     tempDir,
		MaxParseWorkers: 4,
		ParallelParse:   true,
		IndexCacheSize:  1000,
	}

	// Create storage
	storage := internal.NewFileSystemStorage(tempDir, config)

	// Create test entries
	entries := []*internal.JournalEntry{
		{
			Timestamp: time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC),
			Tags:      []string{"feature", "planning"},
			Mentions:  []string{"bob"},
			Body:      "Started planning new #feature with @bob. This is exciting!",
		},
		{
			Timestamp: time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC),
			Tags:      []string{"meeting", "work"},
			Mentions:  []string{"alice"},
			Body:      "Had #meeting with @alice about #work project. Great discussion.",
		},
		{
			Timestamp: time.Date(2026, 2, 9, 15, 30, 0, 0, time.UTC),
			Tags:      []string{"work", "coding"},
			Mentions:  []string{"alice"},
			Body:      "Implemented #work features. Will demo to @alice tomorrow.",
		},
		{
			Timestamp: time.Date(2026, 2, 10, 9, 0, 0, 0, time.UTC),
			Tags:      []string{"review"},
			Mentions:  []string{},
			Body:      "Code #review session. Everything looks good!",
		},
	}

	// Save all entries
	for _, entry := range entries {
		err := storage.SaveEntry(entry)
		if err != nil {
			t.Fatalf("Failed to save entry: %v", err)
		}
	}

	// Create CLI app
	app := NewApp(storage, config)

	// Helper to capture stdout
	captureOutput := func(fn func() error) (string, error) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fn()

		_ = w.Close()
		os.Stdout = oldStdout

		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		return buf.String(), err
	}

	// Test 1: List all entries
	t.Run("list all", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 4 entries") {
			t.Error("Expected to find 4 entries")
		}
	})

	// Test 2: Search by tag
	t.Run("search by tag #work", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"#work"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 2 entries") {
			t.Errorf("Expected to find 2 entries with #work")
		}
	})

	// Test 3: Search by mention
	t.Run("search by mention @alice", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"@alice"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 2 entries") {
			t.Errorf("Expected to find 2 entries with @alice")
		}
	})

	// Test 4: Search with AND logic
	t.Run("search #work AND @alice", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"#work", "@alice"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 2 entries") {
			t.Errorf("Expected to find 2 entries with both #work AND @alice")
		}
	})

	// Test 5: Search with date filter
	t.Run("search with date filter", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"-from", "2026-02-09"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 3 entries") {
			t.Errorf("Expected to find 3 entries from 2026-02-09 onwards")
		}
	})

	// Test 6: Search with limit
	t.Run("search with limit", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"-n", "2"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 2 entries") {
			t.Errorf("Expected to find 2 entries with limit")
		}
	})

	// Test 7: Search with summary format
	t.Run("search with summary format", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"--summary"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		// Summary format should have timestamp | preview format
		if !strings.Contains(output, " | ") {
			t.Error("Expected summary format with pipe separator")
		}
		if !strings.Contains(output, "2026-02-") {
			t.Error("Expected timestamp in summary format")
		}
	})

	// Test 8: Search with JSON format
	t.Run("search with json format", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"--format", "json"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		// JSON format should be valid JSON
		if !strings.HasPrefix(strings.TrimSpace(output), "[") {
			t.Error("Expected JSON array output")
		}
		if !strings.Contains(output, "\"timestamp\"") {
			t.Error("Expected timestamp field in JSON")
		}
		if !strings.Contains(output, "\"tags\"") {
			t.Error("Expected tags field in JSON")
		}
	})

	// Test 9: Search with reverse order
	t.Run("search with reverse order", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"-r"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		// Check order - should have 2026-02-10 before 2026-02-08
		idx10 := strings.Index(output, "2026-02-10")
		idx08 := strings.Index(output, "2026-02-08")

		if idx10 == -1 || idx08 == -1 {
			t.Error("Expected both dates in output")
		}
		if idx10 > idx08 {
			t.Error("Expected reverse chronological order (newest first)")
		}
	})

	// Test 10: Search with keyword
	t.Run("search by keyword", func(t *testing.T) {
		output, err := captureOutput(func() error {
			return app.Search([]string{"demo"})
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if !strings.Contains(output, "Found 1 entries") {
			t.Errorf("Expected to find 1 entry with keyword 'demo'")
		}
	})
}

// TestCreateEntry tests entry creation with temporary storage
func TestCreateEntry(t *testing.T) {
	// Create temporary directory for test storage
	tempDir, err := os.MkdirTemp("", "jrnlg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	// Create test config
	config := &internal.Config{
		StoragePath:     tempDir,
		MaxParseWorkers: 4,
		ParallelParse:   true,
		IndexCacheSize:  1000,
	}

	// Create storage
	storage := internal.NewFileSystemStorage(tempDir, config)

	// Create a temporary file with entry content
	entryContent := `## Monday 2026-02-09 2:30 PM UTC

This is a test #entry with @mention.

Multiple paragraphs work too!`

	tmpFile, err := os.CreateTemp("", "entry-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile.Name())

	_, err = tmpFile.WriteString(entryContent)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Simulate CreateEntry by reading the file and saving it
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	entry, err := internal.ParseEntry(string(content))
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	err = storage.SaveEntry(entry)
	if err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Build expected path
	expectedPath := filepath.Join(tempDir, "2026", "02", "2026-02-09-14-30-00.md")

	// Verify file exists
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Entry file was not created at %s", expectedPath)
	}

	// Verify we can retrieve it
	retrieved, err := storage.GetEntry(entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}

	if retrieved.Body != entry.Body {
		t.Errorf("Retrieved entry body doesn't match. Expected %q, got %q", entry.Body, retrieved.Body)
	}

	if len(retrieved.Tags) != 1 || retrieved.Tags[0] != "entry" {
		t.Errorf("Expected tag 'entry', got %v", retrieved.Tags)
	}

	if len(retrieved.Mentions) != 1 || retrieved.Mentions[0] != "mention" {
		t.Errorf("Expected mention 'mention', got %v", retrieved.Mentions)
	}
}
