package cli

import (
	"testing"
	"time"
)

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		maxLen   int
		expected string
	}{
		{
			name:     "short text",
			body:     "Hello world",
			maxLen:   20,
			expected: "Hello world",
		},
		{
			name:     "exact length",
			body:     "Hello world",
			maxLen:   11,
			expected: "Hello world",
		},
		{
			name:     "needs truncation",
			body:     "This is a very long text that needs to be truncated",
			maxLen:   20,
			expected: "This is a very long ...",
		},
		{
			name:     "multi-line shows first line only",
			body:     "First line\nSecond line\nThird line",
			maxLen:   50,
			expected: "First line",
		},
		{
			name:     "multi-line with truncation",
			body:     "This is a very long first line\nSecond line",
			maxLen:   15,
			expected: "This is a very ...",
		},
		{
			name:     "empty string",
			body:     "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "whitespace only",
			body:     "   \n\n  ",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "first line has leading/trailing spaces",
			body:     "  Hello world  \nSecond line",
			maxLen:   50,
			expected: "Hello world",
		},
		{
			name:     "very small maxLen",
			body:     "Hello world",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "maxLen of 1",
			body:     "Hello",
			maxLen:   1,
			expected: "H...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateBody(tt.body, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateBody(%q, %d) = %q, want %q", tt.body, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestIsTimestampFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid timestamp",
			input:    "2026-02-09-14-30-45",
			expected: true,
		},
		{
			name:     "valid timestamp with zeros",
			input:    "2020-01-01-00-00-00",
			expected: true,
		},
		{
			name:     "valid timestamp end of year",
			input:    "2025-12-31-23-59-59",
			expected: true,
		},
		{
			name:     "too short",
			input:    "2026-02-09-14-30",
			expected: false,
		},
		{
			name:     "too long",
			input:    "2026-02-09-14-30-45-00",
			expected: false,
		},
		{
			name:     "wrong separator at position 4",
			input:    "2026/02-09-14-30-45",
			expected: false,
		},
		{
			name:     "wrong separator at position 7",
			input:    "2026-02/09-14-30-45",
			expected: false,
		},
		{
			name:     "wrong separator at position 10",
			input:    "2026-02-09/14-30-45",
			expected: false,
		},
		{
			name:     "wrong separator at position 13",
			input:    "2026-02-09-14/30-45",
			expected: false,
		},
		{
			name:     "wrong separator at position 16",
			input:    "2026-02-09-14-30/45",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "random text",
			input:    "not a timestamp",
			expected: false,
		},
		{
			name:     "date only",
			input:    "2026-02-09",
			expected: false,
		},
		{
			name:     "invalid characters but correct format",
			input:    "abcd-ef-gh-ij-kl-mn",
			expected: true, // Format check only, not validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimestampFormat(tt.input)
			if result != tt.expected {
				t.Errorf("IsTimestampFormat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Time
		expectErr bool
	}{
		{
			name:      "valid timestamp",
			input:     "2026-02-09-14-30-45",
			expected:  time.Date(2026, 2, 9, 14, 30, 45, 0, time.UTC),
			expectErr: false,
		},
		{
			name:      "midnight",
			input:     "2020-01-01-00-00-00",
			expected:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectErr: false,
		},
		{
			name:      "end of day",
			input:     "2025-12-31-23-59-59",
			expected:  time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expectErr: false,
		},
		{
			name:      "leap year date",
			input:     "2024-02-29-12-00-00",
			expected:  time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			expectErr: false,
		},
		{
			name:      "invalid format - wrong separators",
			input:     "2026/02/09-14-30-45",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "invalid format - too short",
			input:     "2026-02-09",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "invalid date - month 13",
			input:     "2026-13-09-14-30-45",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "invalid date - day 32",
			input:     "2026-02-32-14-30-45",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "invalid time - hour 24",
			input:     "2026-02-09-24-30-45",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "invalid time - minute 60",
			input:     "2026-02-09-14-60-45",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "invalid time - second 60",
			input:     "2026-02-09-14-30-60",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name:      "random text",
			input:     "not a timestamp",
			expected:  time.Time{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimestamp(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseTimestamp(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseTimestamp(%q) unexpected error: %v", tt.input, err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("ParseTimestamp(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}
