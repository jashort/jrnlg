package cli

import (
	"testing"
	"time"
)

func TestParseDate_ISO8601_DateOnly(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			"simple date",
			"2026-02-09",
			time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			"first day of year",
			"2026-01-01",
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"last day of year",
			"2026-12-31",
			time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			"leap year date",
			"2024-02-29",
			time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseDate_ISO8601_DateTime(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			"date with time",
			"2026-02-09T14:30:00Z",
			time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
		},
		{
			"midnight",
			"2026-02-09T00:00:00Z",
			time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			"end of day",
			"2026-02-09T23:59:59Z",
			time.Date(2026, 2, 9, 23, 59, 59, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseDate_ISO8601_WithTimezone(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			"with +00:00",
			"2026-02-09T14:30:00+00:00",
			time.Date(2026, 2, 9, 14, 30, 0, 0, time.UTC),
		},
		{
			"with -08:00 (PST)",
			"2026-02-09T14:30:00-08:00",
			time.Date(2026, 2, 9, 22, 30, 0, 0, time.UTC), // 14:30 PST = 22:30 UTC
		},
		{
			"with +05:30 (IST)",
			"2026-02-09T14:30:00+05:30",
			time.Date(2026, 2, 9, 9, 0, 0, 0, time.UTC), // 14:30 IST = 09:00 UTC
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseDate_NaturalLanguage(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name           string
		input          string
		validateResult func(time.Time) bool
	}{
		{
			"today",
			"today",
			func(result time.Time) bool {
				// Should be today's date (ignoring time)
				return result.Year() == now.Year() &&
					result.Month() == now.Month() &&
					result.Day() == now.Day()
			},
		},
		{
			"yesterday",
			"yesterday",
			func(result time.Time) bool {
				yesterday := now.AddDate(0, 0, -1)
				return result.Year() == yesterday.Year() &&
					result.Month() == yesterday.Month() &&
					result.Day() == yesterday.Day()
			},
		},
		{
			"tomorrow",
			"tomorrow",
			func(result time.Time) bool {
				tomorrow := now.AddDate(0, 0, 1)
				return result.Year() == tomorrow.Year() &&
					result.Month() == tomorrow.Month() &&
					result.Day() == tomorrow.Day()
			},
		},
		{
			"1 day ago",
			"1 day ago",
			func(result time.Time) bool {
				expected := now.AddDate(0, 0, -1)
				return result.Year() == expected.Year() &&
					result.Month() == expected.Month() &&
					result.Day() == expected.Day()
			},
		},
		{
			"2 days ago",
			"2 days ago",
			func(result time.Time) bool {
				expected := now.AddDate(0, 0, -2)
				return result.Year() == expected.Year() &&
					result.Month() == expected.Month() &&
					result.Day() == expected.Day()
			},
		},
		{
			"3 days ago",
			"3 days ago",
			func(result time.Time) bool {
				expected := now.AddDate(0, 0, -3)
				return result.Year() == expected.Year() &&
					result.Month() == expected.Month() &&
					result.Day() == expected.Day()
			},
		},
		{
			"5 days ago",
			"5 days ago",
			func(result time.Time) bool {
				expected := now.AddDate(0, 0, -5)
				return result.Year() == expected.Year() &&
					result.Month() == expected.Month() &&
					result.Day() == expected.Day()
			},
		},
		{
			"1 week ago",
			"1 week ago",
			func(result time.Time) bool {
				expected := now.AddDate(0, 0, -7)
				return result.Year() == expected.Year() &&
					result.Month() == expected.Month() &&
					result.Day() == expected.Day()
			},
		},
		{
			"2 weeks ago",
			"2 weeks ago",
			func(result time.Time) bool {
				expected := now.AddDate(0, 0, -14)
				return result.Year() == expected.Year() &&
					result.Month() == expected.Month() &&
					result.Day() == expected.Day()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !tc.validateResult(result) {
				t.Errorf("Validation failed for input '%s', got result: %v", tc.input, result)
			}
		})
	}
}

func TestParseDate_NaturalLanguage_WithWhitespace(t *testing.T) {
	// Test that whitespace is handled properly
	result1, err1 := ParseDate("  yesterday  ")
	if err1 != nil {
		t.Fatalf("Expected no error for whitespace-padded input, got: %v", err1)
	}

	result2, err2 := ParseDate("yesterday")
	if err2 != nil {
		t.Fatalf("Expected no error, got: %v", err2)
	}

	// Both should give the same date
	if result1.Year() != result2.Year() || result1.Month() != result2.Month() || result1.Day() != result2.Day() {
		t.Errorf("Whitespace-padded and non-padded inputs should give same date")
	}
}

func TestParseDate_Errors(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"not a date", "hello world"},
		{"partial date", "2026"},
		{"letters in date", "2026-ab-cd"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDate(tc.input)
			if err == nil {
				t.Errorf("Expected error for input '%s', got none", tc.input)
			}
		})
	}
}

// TestParseDate_InvalidISO8601 tests that invalid ISO dates that can't be
// parsed as natural language still return errors
func TestParseDate_InvalidISO8601(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"missing day", "2026-02"},
		{"wrong separator", "2026.02.09"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDate(tc.input)
			if err == nil {
				t.Errorf("Expected error for input '%s', got none", tc.input)
			}
		})
	}
}

func TestParseDate_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			"year 2000",
			"2000-01-01",
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"year 2099",
			"2099-12-31",
			time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			"leap year Feb 29",
			"2024-02-29",
			time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
