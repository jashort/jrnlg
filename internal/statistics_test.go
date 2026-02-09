package internal

import (
	"testing"
	"time"
)

// Helper function to create test entries
func makeTestEntry(timestamp time.Time, tags, mentions []string) *IndexedEntry {
	return &IndexedEntry{
		FilePath:  "/test/path.md",
		Timestamp: timestamp,
		Tags:      tags,
		Mentions:  mentions,
	}
}

// TestCalculateStatistics_Empty tests statistics with no entries
func TestCalculateStatistics_Empty(t *testing.T) {
	loc := time.UTC
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, loc)

	stats := CalculateStatistics([]*IndexedEntry{}, startDate, endDate, false)

	if stats.Summary.TotalEntries != 0 {
		t.Errorf("TotalEntries = %d, want 0", stats.Summary.TotalEntries)
	}
	if stats.Summary.ActiveDays != 0 {
		t.Errorf("ActiveDays = %d, want 0", stats.Summary.ActiveDays)
	}
	if len(stats.TopTags) != 0 {
		t.Errorf("TopTags length = %d, want 0", len(stats.TopTags))
	}
	if len(stats.TopMentions) != 0 {
		t.Errorf("TopMentions length = %d, want 0", len(stats.TopMentions))
	}
}

// TestCalculateStatistics_SingleEntry tests statistics with one entry
func TestCalculateStatistics_SingleEntry(t *testing.T) {
	loc := time.UTC
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, loc)

	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 15, 10, 0, 0, 0, loc), []string{"work"}, []string{"alice"}),
	}

	stats := CalculateStatistics(entries, startDate, endDate, false)

	if stats.Summary.TotalEntries != 1 {
		t.Errorf("TotalEntries = %d, want 1", stats.Summary.TotalEntries)
	}
	if stats.Summary.ActiveDays != 1 {
		t.Errorf("ActiveDays = %d, want 1", stats.Summary.ActiveDays)
	}
	if stats.Summary.LongestStreak != 1 {
		t.Errorf("LongestStreak = %d, want 1", stats.Summary.LongestStreak)
	}
	if stats.Summary.CurrentStreak != 0 {
		t.Errorf("CurrentStreak = %d, want 0 (entry not on end date)", stats.Summary.CurrentStreak)
	}
	if stats.Summary.AvgPerActiveDay != 1.0 {
		t.Errorf("AvgPerActiveDay = %f, want 1.0", stats.Summary.AvgPerActiveDay)
	}
}

// TestCalculateStatistics_BasicStats tests basic statistics calculations
func TestCalculateStatistics_BasicStats(t *testing.T) {
	loc := time.UTC
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2024, 1, 10, 23, 59, 59, 0, loc)

	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), []string{"work", "meeting"}, []string{"alice"}),
		makeTestEntry(time.Date(2024, 1, 1, 14, 0, 0, 0, loc), []string{"work"}, []string{"bob"}),
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), []string{"personal"}, []string{"alice"}),
		makeTestEntry(time.Date(2024, 1, 5, 16, 0, 0, 0, loc), []string{"work"}, []string{"alice"}),
	}

	stats := CalculateStatistics(entries, startDate, endDate, false)

	// Total entries
	if stats.Summary.TotalEntries != 4 {
		t.Errorf("TotalEntries = %d, want 4", stats.Summary.TotalEntries)
	}

	// Active days: Jan 1, 2, 5 = 3 days
	if stats.Summary.ActiveDays != 3 {
		t.Errorf("ActiveDays = %d, want 3", stats.Summary.ActiveDays)
	}

	// Total days: Jan 1-10 = 10 days
	if stats.Period.TotalDays != 10 {
		t.Errorf("TotalDays = %d, want 10", stats.Period.TotalDays)
	}

	// Active days percentage: 3/10 = 30%
	expectedPct := 30.0
	if stats.Summary.ActiveDaysPct != expectedPct {
		t.Errorf("ActiveDaysPct = %f, want %f", stats.Summary.ActiveDaysPct, expectedPct)
	}

	// Average per active day: 4 entries / 3 days = 1.333...
	expectedAvg := 4.0 / 3.0
	if stats.Summary.AvgPerActiveDay != expectedAvg {
		t.Errorf("AvgPerActiveDay = %f, want %f", stats.Summary.AvgPerActiveDay, expectedAvg)
	}

	// Tags: work=3, meeting=1, personal=1
	if len(stats.Tags) != 3 {
		t.Errorf("Tags count = %d, want 3", len(stats.Tags))
	}
	if stats.Tags[0].Name != "work" || stats.Tags[0].Count != 3 {
		t.Errorf("Top tag = %s:%d, want work:3", stats.Tags[0].Name, stats.Tags[0].Count)
	}

	// Mentions: alice=3, bob=1
	if len(stats.Mentions) != 2 {
		t.Errorf("Mentions count = %d, want 2", len(stats.Mentions))
	}
	if stats.Mentions[0].Name != "alice" || stats.Mentions[0].Count != 3 {
		t.Errorf("Top mention = %s:%d, want alice:3", stats.Mentions[0].Name, stats.Mentions[0].Count)
	}
}

// TestCalculateActiveDays tests active day calculation
func TestCalculateActiveDays(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), nil, nil),
		makeTestEntry(time.Date(2024, 1, 1, 14, 0, 0, 0, loc), nil, nil), // Same day
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), nil, nil),
		makeTestEntry(time.Date(2024, 1, 5, 16, 0, 0, 0, loc), nil, nil),
	}

	activeDays := calculateActiveDays(entries)

	// Should have 3 unique days: Jan 1, 2, 5
	if len(activeDays) != 3 {
		t.Errorf("Active days count = %d, want 3", len(activeDays))
	}

	// Verify dates
	expected := []string{"2024-01-01", "2024-01-02", "2024-01-05"}
	for i, day := range activeDays {
		actual := day.Format("2006-01-02")
		if actual != expected[i] {
			t.Errorf("Active day[%d] = %s, want %s", i, actual, expected[i])
		}
	}
}

// TestCalculateActiveDays_MultipleSameDay ensures deduplication works
func TestCalculateActiveDays_MultipleSameDay(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 8, 0, 0, 0, loc), nil, nil),
		makeTestEntry(time.Date(2024, 1, 1, 12, 0, 0, 0, loc), nil, nil),
		makeTestEntry(time.Date(2024, 1, 1, 18, 0, 0, 0, loc), nil, nil),
	}

	activeDays := calculateActiveDays(entries)

	if len(activeDays) != 1 {
		t.Errorf("Active days count = %d, want 1 (same day deduplication)", len(activeDays))
	}
}

// TestCalculateStreaks_NoStreak tests with no entries
func TestCalculateStreaks_NoStreak(t *testing.T) {
	loc := time.UTC
	activeDays := []time.Time{}
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, loc)

	longest, current := calculateStreaks(activeDays, endDate)

	if longest != 0 || current != 0 {
		t.Errorf("Streaks = (%d, %d), want (0, 0)", longest, current)
	}
}

// TestCalculateStreaks_SingleDay tests with a single entry
func TestCalculateStreaks_SingleDay(t *testing.T) {
	loc := time.UTC
	activeDays := []time.Time{
		time.Date(2024, 1, 15, 0, 0, 0, 0, loc),
	}
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, loc)

	longest, current := calculateStreaks(activeDays, endDate)

	if longest != 1 {
		t.Errorf("Longest streak = %d, want 1", longest)
	}
	if current != 0 {
		t.Errorf("Current streak = %d, want 0 (not on end date)", current)
	}
}

// TestCalculateStreaks_CurrentStreak tests current streak calculation
func TestCalculateStreaks_CurrentStreak(t *testing.T) {
	loc := time.UTC
	// Entries: Jan 1, Jan 2, gap, Jan 10, Jan 11, Jan 12
	activeDays := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 2, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 10, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 11, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 12, 0, 0, 0, 0, loc),
	}
	endDate := time.Date(2024, 1, 12, 23, 59, 59, 0, loc)

	longest, current := calculateStreaks(activeDays, endDate)

	if longest != 3 {
		t.Errorf("Longest streak = %d, want 3", longest)
	}
	if current != 3 {
		t.Errorf("Current streak = %d, want 3", current)
	}
}

// TestCalculateStreaks_LongestStreak tests finding longest streak among multiple
func TestCalculateStreaks_LongestStreak(t *testing.T) {
	loc := time.UTC
	// Entries: Jan 1-2 (2 days), gap, Jan 5-9 (5 days), gap, Jan 15-16 (2 days)
	activeDays := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 2, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 5, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 6, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 7, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 8, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 9, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 15, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 16, 0, 0, 0, 0, loc),
	}
	endDate := time.Date(2024, 1, 20, 23, 59, 59, 0, loc)

	longest, current := calculateStreaks(activeDays, endDate)

	if longest != 5 {
		t.Errorf("Longest streak = %d, want 5", longest)
	}
	if current != 0 {
		t.Errorf("Current streak = %d, want 0 (no recent activity)", current)
	}
}

// TestCalculateStreaks_CurrentIsLongest tests when current is also longest
func TestCalculateStreaks_CurrentIsLongest(t *testing.T) {
	loc := time.UTC
	// Entries: Jan 1 (single), gap, Jan 5-10 (6 days continuing to end)
	activeDays := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 5, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 6, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 7, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 8, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 9, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 10, 0, 0, 0, 0, loc),
	}
	endDate := time.Date(2024, 1, 10, 23, 59, 59, 0, loc)

	longest, current := calculateStreaks(activeDays, endDate)

	if longest != 6 {
		t.Errorf("Longest streak = %d, want 6", longest)
	}
	if current != 6 {
		t.Errorf("Current streak = %d, want 6", current)
	}
}

// TestCalculateLongestGap_NoGap tests when there are no gaps
func TestCalculateLongestGap_NoGap(t *testing.T) {
	loc := time.UTC
	// Consecutive days: Jan 1, 2, 3, 4
	activeDays := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 2, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 3, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 4, 0, 0, 0, 0, loc),
	}

	gap := calculateLongestGap(activeDays)

	if gap.Days != 0 {
		t.Errorf("Longest gap = %d days, want 0", gap.Days)
	}
}

// TestCalculateLongestGap_WithGaps tests finding longest gap
func TestCalculateLongestGap_WithGaps(t *testing.T) {
	loc := time.UTC
	// Jan 1, gap of 2 days, Jan 4, gap of 5 days, Jan 10
	activeDays := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
		time.Date(2024, 1, 4, 0, 0, 0, 0, loc),  // 2-day gap (Jan 2-3)
		time.Date(2024, 1, 10, 0, 0, 0, 0, loc), // 5-day gap (Jan 5-9)
	}

	gap := calculateLongestGap(activeDays)

	if gap.Days != 5 {
		t.Errorf("Longest gap = %d days, want 5", gap.Days)
	}
	expectedStart := "2024-01-04"
	expectedEnd := "2024-01-10"
	if gap.StartDate.Format("2006-01-02") != expectedStart {
		t.Errorf("Gap start = %s, want %s", gap.StartDate.Format("2006-01-02"), expectedStart)
	}
	if gap.EndDate.Format("2006-01-02") != expectedEnd {
		t.Errorf("Gap end = %s, want %s", gap.EndDate.Format("2006-01-02"), expectedEnd)
	}
}

// TestCalculateLongestGap_SingleEntry tests with single entry
func TestCalculateLongestGap_SingleEntry(t *testing.T) {
	loc := time.UTC
	activeDays := []time.Time{
		time.Date(2024, 1, 15, 0, 0, 0, 0, loc),
	}

	gap := calculateLongestGap(activeDays)

	if gap.Days != 0 {
		t.Errorf("Longest gap = %d days, want 0 (single entry has no gaps)", gap.Days)
	}
}

// TestAggregateTags tests tag counting and sorting
func TestAggregateTags(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), []string{"work", "meeting"}, nil),
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), []string{"work", "code"}, nil),
		makeTestEntry(time.Date(2024, 1, 3, 11, 0, 0, 0, loc), []string{"work"}, nil),
		makeTestEntry(time.Date(2024, 1, 4, 12, 0, 0, 0, loc), []string{"meeting", "planning"}, nil),
	}

	tags := aggregateTags(entries)

	// Expected: work=3, meeting=2, code=1, planning=1
	if len(tags) != 4 {
		t.Errorf("Tag count = %d, want 4", len(tags))
	}

	// Verify sorting (by count desc, then name asc)
	expected := []struct {
		name  string
		count int
	}{
		{"work", 3},
		{"meeting", 2},
		{"code", 1},
		{"planning", 1},
	}

	for i, exp := range expected {
		if i >= len(tags) {
			t.Errorf("Missing tag at index %d", i)
			continue
		}
		if tags[i].Name != exp.name || tags[i].Count != exp.count {
			t.Errorf("Tag[%d] = %s:%d, want %s:%d", i, tags[i].Name, tags[i].Count, exp.name, exp.count)
		}
	}
}

// TestAggregateMentions tests mention counting and sorting
func TestAggregateMentions(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), nil, []string{"alice", "bob"}),
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), nil, []string{"alice"}),
		makeTestEntry(time.Date(2024, 1, 3, 11, 0, 0, 0, loc), nil, []string{"alice", "charlie"}),
		makeTestEntry(time.Date(2024, 1, 4, 12, 0, 0, 0, loc), nil, []string{"bob"}),
	}

	mentions := aggregateMentions(entries)

	// Expected: alice=3, bob=2, charlie=1
	if len(mentions) != 3 {
		t.Errorf("Mention count = %d, want 3", len(mentions))
	}

	expected := []struct {
		name  string
		count int
	}{
		{"alice", 3},
		{"bob", 2},
		{"charlie", 1},
	}

	for i, exp := range expected {
		if i >= len(mentions) {
			t.Errorf("Missing mention at index %d", i)
			continue
		}
		if mentions[i].Name != exp.name || mentions[i].Count != exp.count {
			t.Errorf("Mention[%d] = %s:%d, want %s:%d", i, mentions[i].Name, mentions[i].Count, exp.name, exp.count)
		}
	}
}

// TestTopN tests top N selection
func TestTopN(t *testing.T) {
	tags := []TagStat{
		{"tag1", 10},
		{"tag2", 8},
		{"tag3", 6},
		{"tag4", 4},
		{"tag5", 2},
		{"tag6", 1},
	}

	// Test with n=5 (should return first 5)
	top5 := topN(tags, 5)
	if len(top5) != 5 {
		t.Errorf("topN(tags, 5) length = %d, want 5", len(top5))
	}

	// Test with n > length (should return all)
	top10 := topN(tags, 10)
	if len(top10) != 6 {
		t.Errorf("topN(tags, 10) length = %d, want 6", len(top10))
	}

	// Test with n=0 (should return empty)
	top0 := topN(tags, 0)
	if len(top0) != 0 {
		t.Errorf("topN(tags, 0) length = %d, want 0", len(top0))
	}
}

// TestCategorizeTimeOfDay tests time categorization
func TestCategorizeTimeOfDay(t *testing.T) {
	loc := time.UTC
	tests := []struct {
		hour     int
		expected string
	}{
		{0, TimeNight},
		{4, TimeNight},
		{5, TimeMorning},
		{8, TimeMorning},
		{11, TimeMorning},
		{12, TimeAfternoon},
		{14, TimeAfternoon},
		{16, TimeAfternoon},
		{17, TimeEvening},
		{19, TimeEvening},
		{20, TimeEvening},
		{21, TimeNight},
		{23, TimeNight},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			timestamp := time.Date(2024, 1, 1, tt.hour, 0, 0, 0, loc)
			category := categorizeTimeOfDay(timestamp)
			if category != tt.expected {
				t.Errorf("categorizeTimeOfDay(%d:00) = %s, want %s", tt.hour, category, tt.expected)
			}
		})
	}
}

// TestCalculateActivityPatterns tests pattern calculation
func TestCalculateActivityPatterns(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		// Monday morning
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), nil, nil),
		makeTestEntry(time.Date(2024, 1, 1, 10, 0, 0, 0, loc), nil, nil),
		// Tuesday afternoon
		makeTestEntry(time.Date(2024, 1, 2, 14, 0, 0, 0, loc), nil, nil),
		// Wednesday evening
		makeTestEntry(time.Date(2024, 1, 3, 18, 0, 0, 0, loc), nil, nil),
		// Thursday night
		makeTestEntry(time.Date(2024, 1, 4, 22, 0, 0, 0, loc), nil, nil),
		// Another Monday morning
		makeTestEntry(time.Date(2024, 1, 8, 8, 0, 0, 0, loc), nil, nil),
	}

	patterns := calculateActivityPatterns(entries)

	// Day of week: Monday=3, Tuesday=1, Wednesday=1, Thursday=1
	if patterns.DayOfWeek[time.Monday] != 3 {
		t.Errorf("Monday count = %d, want 3", patterns.DayOfWeek[time.Monday])
	}
	if patterns.BusiestDay != time.Monday {
		t.Errorf("Busiest day = %s, want Monday", patterns.BusiestDay)
	}

	// Time of day: morning=3, afternoon=1, evening=1, night=1
	if patterns.TimeOfDay[TimeMorning] != 3 {
		t.Errorf("Morning count = %d, want 3", patterns.TimeOfDay[TimeMorning])
	}
	if patterns.BusiestTime != TimeMorning {
		t.Errorf("Busiest time = %s, want %s", patterns.BusiestTime, TimeMorning)
	}

	// Hourly distribution
	if patterns.HourlyDistribution[9] != 1 {
		t.Errorf("Hour 9 count = %d, want 1", patterns.HourlyDistribution[9])
	}
	if patterns.HourlyDistribution[10] != 1 {
		t.Errorf("Hour 10 count = %d, want 1", patterns.HourlyDistribution[10])
	}
}

// TestFilterByTag tests tag filtering
func TestFilterByTag(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), []string{"work", "meeting"}, nil),
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), []string{"personal"}, nil),
		makeTestEntry(time.Date(2024, 1, 3, 11, 0, 0, 0, loc), []string{"work"}, nil),
	}

	filtered := filterByTag(entries, "work")

	if len(filtered) != 2 {
		t.Errorf("Filtered entries = %d, want 2", len(filtered))
	}

	// Verify correct entries were selected
	for _, entry := range filtered {
		hasTag := false
		for _, tag := range entry.Tags {
			if tag == "work" {
				hasTag = true
				break
			}
		}
		if !hasTag {
			t.Errorf("Filtered entry missing 'work' tag")
		}
	}
}

// TestFilterByMention tests mention filtering
func TestFilterByMention(t *testing.T) {
	loc := time.UTC
	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), nil, []string{"alice", "bob"}),
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), nil, []string{"charlie"}),
		makeTestEntry(time.Date(2024, 1, 3, 11, 0, 0, 0, loc), nil, []string{"alice"}),
	}

	filtered := filterByMention(entries, "alice")

	if len(filtered) != 2 {
		t.Errorf("Filtered entries = %d, want 2", len(filtered))
	}

	// Verify correct entries were selected
	for _, entry := range filtered {
		hasMention := false
		for _, mention := range entry.Mentions {
			if mention == "alice" {
				hasMention = true
				break
			}
		}
		if !hasMention {
			t.Errorf("Filtered entry missing 'alice' mention")
		}
	}
}

// TestCalculateTotalDays tests day counting
func TestCalculateTotalDays(t *testing.T) {
	loc := time.UTC
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			"Same day",
			time.Date(2024, 1, 1, 9, 0, 0, 0, loc),
			time.Date(2024, 1, 1, 17, 0, 0, 0, loc),
			1,
		},
		{
			"Two days",
			time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
			time.Date(2024, 1, 2, 23, 59, 59, 0, loc),
			2,
		},
		{
			"Full month",
			time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
			time.Date(2024, 1, 31, 23, 59, 59, 0, loc),
			31,
		},
		{
			"Leap year February",
			time.Date(2024, 2, 1, 0, 0, 0, 0, loc),
			time.Date(2024, 2, 29, 23, 59, 59, 0, loc),
			29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days := calculateTotalDays(tt.start, tt.end)
			if days != tt.expected {
				t.Errorf("calculateTotalDays() = %d, want %d", days, tt.expected)
			}
		})
	}
}

// TestCalculateFilteredStatistics tests filtering with statistics
func TestCalculateFilteredStatistics(t *testing.T) {
	loc := time.UTC
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2024, 1, 10, 23, 59, 59, 0, loc)

	entries := []*IndexedEntry{
		makeTestEntry(time.Date(2024, 1, 1, 9, 0, 0, 0, loc), []string{"work", "meeting"}, []string{"alice"}),
		makeTestEntry(time.Date(2024, 1, 2, 10, 0, 0, 0, loc), []string{"personal"}, []string{"bob"}),
		makeTestEntry(time.Date(2024, 1, 3, 11, 0, 0, 0, loc), []string{"work"}, []string{"alice"}),
	}

	// Test tag filtering
	stats := CalculateFilteredStatistics(entries, startDate, endDate, false, "tag", "work")

	if stats.FilteredBy == nil {
		t.Fatal("FilteredBy is nil, expected filter info")
	}
	if stats.FilteredBy.Type != "tag" || stats.FilteredBy.Value != "work" {
		t.Errorf("FilteredBy = %v, want tag:work", stats.FilteredBy)
	}
	if stats.Summary.TotalEntries != 2 {
		t.Errorf("Filtered total entries = %d, want 2", stats.Summary.TotalEntries)
	}

	// Test mention filtering
	stats = CalculateFilteredStatistics(entries, startDate, endDate, false, "mention", "alice")

	if stats.FilteredBy.Type != "mention" || stats.FilteredBy.Value != "alice" {
		t.Errorf("FilteredBy = %v, want mention:alice", stats.FilteredBy)
	}
	if stats.Summary.TotalEntries != 2 {
		t.Errorf("Filtered total entries = %d, want 2", stats.Summary.TotalEntries)
	}
}
