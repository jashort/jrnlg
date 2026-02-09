package cli

import (
	"strings"
	"testing"
)

func TestIsKnownFlag(t *testing.T) {
	testCases := []struct {
		name     string
		flag     string
		expected bool
	}{
		{"help short", "-h", true},
		{"help long", "--help", true},
		{"version short", "-v", true},
		{"version long", "--version", true},
		{"from", "-from", true},
		{"to", "-to", true},
		{"limit short", "-n", true},
		{"limit long", "--limit", true},
		{"offset", "--offset", true},
		{"reverse short", "-r", true},
		{"reverse long", "--reverse", true},
		{"summary", "--summary", true},
		{"format", "--format", true},
		{"unknown", "--unknown", false},
		{"typo", "--halp", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isKnownFlag(tc.flag)
			if result != tc.expected {
				t.Errorf("isKnownFlag(%q) = %v, expected %v", tc.flag, result, tc.expected)
			}
		})
	}
}

func TestFlagRequiresValue(t *testing.T) {
	testCases := []struct {
		name     string
		flag     string
		expected bool
	}{
		{"help", "--help", false},
		{"version", "--version", false},
		{"reverse", "-r", false},
		{"summary", "--summary", false},
		{"from", "-from", true},
		{"to", "-to", true},
		{"limit", "-n", true},
		{"offset", "--offset", true},
		{"format", "--format", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flagRequiresValue(tc.flag)
			if result != tc.expected {
				t.Errorf("flagRequiresValue(%q) = %v, expected %v", tc.flag, result, tc.expected)
			}
		})
	}
}

func TestGetFlagHelp(t *testing.T) {
	testCases := []struct {
		name     string
		flag     string
		contains string
	}{
		{"help", "--help", "help"},
		{"version", "--version", "version"},
		{"from", "-from", "from this date"},
		{"to", "-to", "up to this date"},
		{"limit", "-n", "Limit number"},
		{"reverse", "-r", "newest"},
		{"summary", "--summary", "summary"},
		{"format", "--format", "format"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getFlagHelp(tc.flag)
			if result == "" {
				t.Errorf("getFlagHelp(%q) returned empty string", tc.flag)
			}
			if !strings.Contains(strings.ToLower(result), strings.ToLower(tc.contains)) {
				t.Errorf("getFlagHelp(%q) = %q, expected to contain %q", tc.flag, result, tc.contains)
			}
		})
	}
}

func TestFindSimilarFlags(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		shouldFind string
	}{
		{"typo help", "--halp", "-h"}, // Primary name is short form
		{"typo summary", "--sumary", "--summary"},
		{"typo format", "--frmat", "--format"},
		{"typo reverse", "--revers", "-r"}, // Primary name is short form
		{"prefix match", "--sum", "--summary"},
		{"single char off", "--ofset", "--offset"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			similar := findSimilarFlags(tc.input)
			found := false
			for _, flag := range similar {
				if flag == tc.shouldFind {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("findSimilarFlags(%q) = %v, expected to include %q", tc.input, similar, tc.shouldFind)
			}
		})
	}
}

func TestUnknownFlagError(t *testing.T) {
	testCases := []struct {
		name          string
		flag          string
		shouldContain []string
	}{
		{
			name:          "typo in help",
			flag:          "--halp",
			shouldContain: []string{"unknown flag", "--halp", "-h"}, // Short form shown
		},
		{
			name:          "typo in summary",
			flag:          "--sumary",
			shouldContain: []string{"unknown flag", "--sumary", "--summary"},
		},
		{
			name:          "completely unknown",
			flag:          "--foobar",
			shouldContain: []string{"unknown flag", "--foobar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := unknownFlagError(tc.flag)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			errMsg := err.Error()
			for _, expected := range tc.shouldContain {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("Error message %q should contain %q", errMsg, expected)
				}
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	testCases := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"hello", "", 5},
		{"", "world", 5},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"help", "halp", 1},
		{"summary", "sumary", 1},
		{"format", "frmat", 1},
		{"kitten", "sitting", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.s1+"_"+tc.s2, func(t *testing.T) {
			result := levenshteinDistance(tc.s1, tc.s2)
			if result != tc.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", tc.s1, tc.s2, result, tc.expected)
			}
		})
	}
}
