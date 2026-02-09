package cli

import (
	"testing"
	"time"

	"github.com/jashort/jrnlg/internal"
)

func TestEntrySelector_SelectEntry_MostRecent(t *testing.T) {
	// Setup test storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Create test entries
	// Note: Use distinct minute precision since the markdown format only stores minutes
	base := time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC)
	entry1 := &internal.JournalEntry{
		Timestamp: base,
		Body:      "Old entry",
	}
	entry2 := &internal.JournalEntry{
		Timestamp: base.Add(1 * time.Minute),
		Body:      "Recent entry",
	}
	entry3 := &internal.JournalEntry{
		Timestamp: base.Add(2 * time.Minute),
		Body:      "Most recent entry",
	}

	if err := storage.SaveEntry(entry1); err != nil {
		t.Fatalf("Failed to save entry1: %v", err)
	}
	if err := storage.SaveEntry(entry2); err != nil {
		t.Fatalf("Failed to save entry2: %v", err)
	}
	if err := storage.SaveEntry(entry3); err != nil {
		t.Fatalf("Failed to save entry3: %v", err)
	}

	// Test selector
	selector := NewEntrySelector(storage)
	entry, filePath, err := selector.SelectEntry("")
	if err != nil {
		t.Fatalf("SelectEntry(\"\") returned error: %v", err)
	}

	if entry == nil {
		t.Fatal("SelectEntry(\"\") returned nil entry")
	}

	if entry.Body != "Most recent entry" {
		t.Errorf("SelectEntry(\"\") returned wrong entry. Got body %q, want %q", entry.Body, "Most recent entry")
	}

	if filePath == "" {
		t.Error("SelectEntry(\"\") returned empty filePath")
	}
}

func TestEntrySelector_SelectEntry_MostRecent_NoEntries(t *testing.T) {
	// Setup empty storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Test selector
	selector := NewEntrySelector(storage)
	entry, filePath, err := selector.SelectEntry("")
	if err == nil {
		t.Fatal("SelectEntry(\"\") should return error when no entries exist")
	}

	if entry != nil {
		t.Error("SelectEntry(\"\") should return nil entry on error")
	}

	if filePath != "" {
		t.Error("SelectEntry(\"\") should return empty filePath on error")
	}
}

func TestEntrySelector_SelectEntry_ByTimestamp(t *testing.T) {
	// Setup test storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Create test entry with known timestamp
	timestamp := time.Date(2026, 2, 9, 14, 30, 45, 0, time.UTC)
	entry := &internal.JournalEntry{
		Timestamp: timestamp,
		Body:      "Specific entry",
	}

	if err := storage.SaveEntry(entry); err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Test selector
	selector := NewEntrySelector(storage)
	selectedEntry, filePath, err := selector.SelectEntry("2026-02-09-14-30-45")
	if err != nil {
		t.Fatalf("SelectEntry(timestamp) returned error: %v", err)
	}

	if selectedEntry == nil {
		t.Fatal("SelectEntry(timestamp) returned nil entry")
	}

	if selectedEntry.Body != "Specific entry" {
		t.Errorf("SelectEntry(timestamp) returned wrong entry. Got body %q, want %q", selectedEntry.Body, "Specific entry")
	}

	if filePath == "" {
		t.Error("SelectEntry(timestamp) returned empty filePath")
	}
}

func TestEntrySelector_SelectEntry_ByTimestamp_NotFound(t *testing.T) {
	// Setup empty storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Test selector
	selector := NewEntrySelector(storage)
	entry, filePath, err := selector.SelectEntry("2026-02-09-14-30-45")
	if err == nil {
		t.Fatal("SelectEntry(timestamp) should return error when entry doesn't exist")
	}

	if entry != nil {
		t.Error("SelectEntry(timestamp) should return nil entry on error")
	}

	if filePath != "" {
		t.Error("SelectEntry(timestamp) should return empty filePath on error")
	}
}

func TestEntrySelector_SelectEntry_ByTimestamp_InvalidFormat(t *testing.T) {
	// Setup storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Test selector with invalid timestamp
	selector := NewEntrySelector(storage)
	entry, filePath, err := selector.SelectEntry("2026-13-09-14-30-45") // Invalid month
	if err == nil {
		t.Fatal("SelectEntry(invalid timestamp) should return error")
	}

	if entry != nil {
		t.Error("SelectEntry(invalid timestamp) should return nil entry on error")
	}

	if filePath != "" {
		t.Error("SelectEntry(invalid timestamp) should return empty filePath on error")
	}
}

func TestEntrySelector_SelectEntry_ByDate_SingleEntry(t *testing.T) {
	// Setup test storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Create entry for a specific date
	targetDate := time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC)
	entry := &internal.JournalEntry{
		Timestamp: targetDate,
		Body:      "Entry for Feb 9",
	}

	if err := storage.SaveEntry(entry); err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Test selector with ISO date
	selector := NewEntrySelector(storage)
	selectedEntry, filePath, err := selector.SelectEntry("2026-02-09")
	if err != nil {
		t.Fatalf("SelectEntry(date) returned error: %v", err)
	}

	if selectedEntry == nil {
		t.Fatal("SelectEntry(date) returned nil entry")
	}

	if selectedEntry.Body != "Entry for Feb 9" {
		t.Errorf("SelectEntry(date) returned wrong entry. Got body %q, want %q", selectedEntry.Body, "Entry for Feb 9")
	}

	if filePath == "" {
		t.Error("SelectEntry(date) returned empty filePath")
	}
}

func TestEntrySelector_SelectEntry_ByDate_NoEntries(t *testing.T) {
	// Setup empty storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Test selector
	selector := NewEntrySelector(storage)
	entry, filePath, err := selector.SelectEntry("2026-02-09")
	if err == nil {
		t.Fatal("SelectEntry(date) should return error when no entries found for date")
	}

	if entry != nil {
		t.Error("SelectEntry(date) should return nil entry on error")
	}

	if filePath != "" {
		t.Error("SelectEntry(date) should return empty filePath on error")
	}
}

func TestEntrySelector_SelectEntry_ByDate_InvalidDate(t *testing.T) {
	// Setup storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Test selector with invalid date
	selector := NewEntrySelector(storage)
	entry, filePath, err := selector.SelectEntry("invalid-date")
	if err == nil {
		t.Fatal("SelectEntry(invalid date) should return error")
	}

	if entry != nil {
		t.Error("SelectEntry(invalid date) should return nil entry on error")
	}

	if filePath != "" {
		t.Error("SelectEntry(invalid date) should return empty filePath on error")
	}
}

func TestEntrySelector_SelectEntries(t *testing.T) {
	// Setup test storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Create test entries with distinct minute precision
	base := time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC)
	entry1 := &internal.JournalEntry{
		Timestamp: base,
		Body:      "Entry 1",
	}
	entry2 := &internal.JournalEntry{
		Timestamp: base.Add(1 * time.Minute),
		Body:      "Entry 2",
	}
	entry3 := &internal.JournalEntry{
		Timestamp: base.Add(2 * time.Minute),
		Body:      "Entry 3",
	}

	if err := storage.SaveEntry(entry1); err != nil {
		t.Fatalf("Failed to save entry1: %v", err)
	}
	if err := storage.SaveEntry(entry2); err != nil {
		t.Fatalf("Failed to save entry2: %v", err)
	}
	if err := storage.SaveEntry(entry3); err != nil {
		t.Fatalf("Failed to save entry3: %v", err)
	}

	// Test selector
	selector := NewEntrySelector(storage)
	entries, err := selector.SelectEntries(internal.EntryFilter{})
	if err != nil {
		t.Fatalf("SelectEntries() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("SelectEntries() returned %d entries, want 3", len(entries))
	}
}

func TestEntrySelector_SelectEntries_WithFilter(t *testing.T) {
	// Setup test storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Create test entries on different dates
	date1 := time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC)
	date3 := time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC)

	entry1 := &internal.JournalEntry{Timestamp: date1, Body: "Feb 8"}
	entry2 := &internal.JournalEntry{Timestamp: date2, Body: "Feb 9"}
	entry3 := &internal.JournalEntry{Timestamp: date3, Body: "Feb 10"}

	if err := storage.SaveEntry(entry1); err != nil {
		t.Fatalf("Failed to save entry1: %v", err)
	}
	if err := storage.SaveEntry(entry2); err != nil {
		t.Fatalf("Failed to save entry2: %v", err)
	}
	if err := storage.SaveEntry(entry3); err != nil {
		t.Fatalf("Failed to save entry3: %v", err)
	}

	// Filter for Feb 9 only
	startDate := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 2, 9, 23, 59, 59, 999999999, time.UTC)
	filter := internal.EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	// Test selector
	selector := NewEntrySelector(storage)
	entries, err := selector.SelectEntries(filter)
	if err != nil {
		t.Fatalf("SelectEntries(filter) returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("SelectEntries(filter) returned %d entries, want 1", len(entries))
	}

	if len(entries) > 0 && entries[0].Body != "Feb 9" {
		t.Errorf("SelectEntries(filter) returned wrong entry. Got body %q, want %q", entries[0].Body, "Feb 9")
	}
}

func TestEntrySelector_SelectEntries_EmptyResult(t *testing.T) {
	// Setup empty storage
	tmpDir := t.TempDir()
	storage := internal.NewFileSystemStorage(tmpDir, nil)

	// Test selector
	selector := NewEntrySelector(storage)
	entries, err := selector.SelectEntries(internal.EntryFilter{})
	if err != nil {
		t.Fatalf("SelectEntries() returned error: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("SelectEntries() returned %d entries, want 0", len(entries))
	}
}
