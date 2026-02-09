package cli

import (
	"testing"
	"time"
)

func TestParseSearchArgs_NoArgs(t *testing.T) {
	args, err := parseSearchArgs([]string{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(args.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(args.Tags))
	}
	if len(args.Mentions) != 0 {
		t.Errorf("Expected 0 mentions, got %d", len(args.Mentions))
	}
	if len(args.Keywords) != 0 {
		t.Errorf("Expected 0 keywords, got %d", len(args.Keywords))
	}
	if args.Format != "full" {
		t.Errorf("Expected format 'full', got '%s'", args.Format)
	}
	if args.Limit != 0 {
		t.Errorf("Expected limit 0, got %d", args.Limit)
	}
}

func TestParseSearchArgs_Tags(t *testing.T) {
	args, err := parseSearchArgs([]string{"#work", "#meeting"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(args.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(args.Tags))
	}
	if args.Tags[0] != "work" || args.Tags[1] != "meeting" {
		t.Errorf("Expected tags [work, meeting], got %v", args.Tags)
	}
}

func TestParseSearchArgs_Mentions(t *testing.T) {
	args, err := parseSearchArgs([]string{"@alice", "@bob"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(args.Mentions) != 2 {
		t.Fatalf("Expected 2 mentions, got %d", len(args.Mentions))
	}
	if args.Mentions[0] != "alice" || args.Mentions[1] != "bob" {
		t.Errorf("Expected mentions [alice, bob], got %v", args.Mentions)
	}
}

func TestParseSearchArgs_Keywords(t *testing.T) {
	args, err := parseSearchArgs([]string{"deadline", "urgent"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(args.Keywords) != 2 {
		t.Fatalf("Expected 2 keywords, got %d", len(args.Keywords))
	}
	if args.Keywords[0] != "deadline" || args.Keywords[1] != "urgent" {
		t.Errorf("Expected keywords [deadline, urgent], got %v", args.Keywords)
	}
}

func TestParseSearchArgs_Mixed(t *testing.T) {
	args, err := parseSearchArgs([]string{"#work", "@alice", "deadline"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(args.Tags) != 1 || args.Tags[0] != "work" {
		t.Errorf("Expected tags [work], got %v", args.Tags)
	}
	if len(args.Mentions) != 1 || args.Mentions[0] != "alice" {
		t.Errorf("Expected mentions [alice], got %v", args.Mentions)
	}
	if len(args.Keywords) != 1 || args.Keywords[0] != "deadline" {
		t.Errorf("Expected keywords [deadline], got %v", args.Keywords)
	}
}

func TestParseSearchArgs_FromDate(t *testing.T) {
	args, err := parseSearchArgs([]string{"-from", "2026-02-01"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if args.FromDate == nil {
		t.Fatal("Expected FromDate to be set")
	}

	expected := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !args.FromDate.Equal(expected) {
		t.Errorf("Expected FromDate %v, got %v", expected, args.FromDate)
	}
}

func TestParseSearchArgs_ToDate(t *testing.T) {
	args, err := parseSearchArgs([]string{"-to", "2026-02-28"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if args.ToDate == nil {
		t.Fatal("Expected ToDate to be set")
	}

	// -to date is set to end of day (23:59:59.999...) for inclusive filtering
	expected := time.Date(2026, 2, 28, 23, 59, 59, 999999999, time.UTC)
	if !args.ToDate.Equal(expected) {
		t.Errorf("Expected ToDate %v, got %v", expected, args.ToDate)
	}
}

func TestParseSearchArgs_Limit(t *testing.T) {
	args, err := parseSearchArgs([]string{"-n", "10"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if args.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", args.Limit)
	}

	// Test --limit flag as well
	args, err = parseSearchArgs([]string{"--limit", "5"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if args.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", args.Limit)
	}
}

func TestParseSearchArgs_Offset(t *testing.T) {
	args, err := parseSearchArgs([]string{"--offset", "20"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if args.Offset != 20 {
		t.Errorf("Expected offset 20, got %d", args.Offset)
	}
}

func TestParseSearchArgs_Reverse(t *testing.T) {
	args, err := parseSearchArgs([]string{"-r"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !args.Reverse {
		t.Error("Expected Reverse to be true")
	}

	// Test --reverse flag as well
	args, err = parseSearchArgs([]string{"--reverse"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !args.Reverse {
		t.Error("Expected Reverse to be true")
	}
}

func TestParseSearchArgs_FormatSummary(t *testing.T) {
	args, err := parseSearchArgs([]string{"--summary"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if args.Format != "summary" {
		t.Errorf("Expected format 'summary', got '%s'", args.Format)
	}
}

func TestParseSearchArgs_FormatExplicit(t *testing.T) {
	testCases := []struct {
		name     string
		format   string
		expected string
	}{
		{"full format", "full", "full"},
		{"summary format", "summary", "summary"},
		{"json format", "json", "json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := parseSearchArgs([]string{"--format", tc.format})
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if args.Format != tc.expected {
				t.Errorf("Expected format '%s', got '%s'", tc.expected, args.Format)
			}
		})
	}
}

func TestParseSearchArgs_ComplexQuery(t *testing.T) {
	args, err := parseSearchArgs([]string{
		"#work", "#meeting", "@alice", "deadline",
		"-from", "2026-01-01",
		"-to", "2026-12-31",
		"-n", "50",
		"--offset", "10",
		"-r",
		"--format", "json",
	})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check all components
	if len(args.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(args.Tags))
	}
	if len(args.Mentions) != 1 {
		t.Errorf("Expected 1 mention, got %d", len(args.Mentions))
	}
	if len(args.Keywords) != 1 {
		t.Errorf("Expected 1 keyword, got %d", len(args.Keywords))
	}
	if args.FromDate == nil {
		t.Error("Expected FromDate to be set")
	}
	if args.ToDate == nil {
		t.Error("Expected ToDate to be set")
	}
	if args.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", args.Limit)
	}
	if args.Offset != 10 {
		t.Errorf("Expected offset 10, got %d", args.Offset)
	}
	if !args.Reverse {
		t.Error("Expected Reverse to be true")
	}
	if args.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", args.Format)
	}
}

// Test error cases
func TestParseSearchArgs_Errors(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{"missing from date", []string{"-from"}},
		{"missing to date", []string{"-to"}},
		{"missing limit value", []string{"-n"}},
		{"missing offset value", []string{"--offset"}},
		{"missing format value", []string{"--format"}},
		{"invalid limit", []string{"-n", "abc"}},
		{"negative limit", []string{"-n", "-5"}},
		{"invalid offset", []string{"--offset", "xyz"}},
		{"negative offset", []string{"--offset", "-10"}},
		{"invalid format", []string{"--format", "xml"}},
		{"invalid date", []string{"-from", "not-a-date"}},
		{"unknown flag", []string{"--unknown"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseSearchArgs(tc.args)
			if err == nil {
				t.Errorf("Expected error for args %v, got none", tc.args)
			}
		})
	}
}

// Test improved error messages with suggestions
func TestParseSearchArgs_ErrorMessages(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		errorContains string
	}{
		{
			name:          "typo in help flag",
			args:          []string{"--halp"},
			errorContains: "-h", // Short form is primary
		},
		{
			name:          "typo in summary flag",
			args:          []string{"--sumary"},
			errorContains: "--summary",
		},
		{
			name:          "typo in format flag",
			args:          []string{"--frmat", "json"},
			errorContains: "--format",
		},
		{
			name:          "invalid format with suggestion",
			args:          []string{"--format", "xml"},
			errorContains: "must be full, summary, or json",
		},
		{
			name:          "missing date with example",
			args:          []string{"-from"},
			errorContains: "Example:",
		},
		{
			name:          "invalid number with quote",
			args:          []string{"-n", "abc"},
			errorContains: "\"abc\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseSearchArgs(tc.args)
			if err == nil {
				t.Fatalf("Expected error for args %v, got none", tc.args)
			}
			if !containsString(err.Error(), tc.errorContains) {
				t.Errorf("Expected error to contain %q, got: %v", tc.errorContains, err)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestParseSearchArgs_EmptyTagAndMention(t *testing.T) {
	// Tags and mentions with just # or @ should be ignored
	args, err := parseSearchArgs([]string{"#", "@", "#work"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(args.Tags) != 1 || args.Tags[0] != "work" {
		t.Errorf("Expected only 'work' tag, got %v", args.Tags)
	}
	if len(args.Mentions) != 0 {
		t.Errorf("Expected 0 mentions, got %v", args.Mentions)
	}
}
