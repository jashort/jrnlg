package internal

import (
	"sort"
	"time"
)

// Statistics represents computed journal statistics
type Statistics struct {
	Period      PeriodInfo
	Summary     SummaryStats
	Tags        []TagStat     // All tags (for filtered view)
	TopTags     []TagStat     // Top 5 only
	Mentions    []MentionStat // All mentions
	TopMentions []MentionStat // Top 5 only
	Patterns    ActivityPatterns
	FilteredBy  *FilterInfo // nil if not filtered
}

// FilterInfo describes what filter was applied
type FilterInfo struct {
	Type  string // "tag" or "mention"
	Value string // The tag/mention being filtered
}

// PeriodInfo describes the time period analyzed
type PeriodInfo struct {
	StartDate time.Time
	EndDate   time.Time
	TotalDays int
	IsAllTime bool // True when --all is used
}

// SummaryStats contains aggregate statistics
type SummaryStats struct {
	TotalEntries    int
	ActiveDays      int     // Days with at least one entry
	ActiveDaysPct   float64 // Percentage of total days
	AvgPerActiveDay float64 // Total / active days
	LongestStreak   int     // Most consecutive active days
	CurrentStreak   int     // Current consecutive active days
	LongestGap      GapInfo // Longest period without entries
}

// GapInfo describes a gap in journal entries
type GapInfo struct {
	Days      int
	StartDate time.Time
	EndDate   time.Time
}

// TagStat contains statistics for a single tag
type TagStat struct {
	Name  string
	Count int
}

// MentionStat contains statistics for a single mention
type MentionStat struct {
	Name  string
	Count int
}

// ActivityPatterns describes temporal usage patterns
type ActivityPatterns struct {
	DayOfWeek          map[time.Weekday]int // Monday->12, Tuesday->5
	TimeOfDay          map[string]int       // "morning"->15, "afternoon"->20
	BusiestDay         time.Weekday
	BusiestDayCount    int // Count for busiest day (0 = no data)
	BusiestTime        string
	HourlyDistribution map[int]int // Hour (0-23) -> count
}

// Time of day categories
const (
	TimeMorning   = "morning"   // 5:00 AM - 11:59 AM
	TimeAfternoon = "afternoon" // 12:00 PM - 4:59 PM
	TimeEvening   = "evening"   // 5:00 PM - 8:59 PM
	TimeNight     = "night"     // 9:00 PM - 4:59 AM
)

// CalculateStatistics computes comprehensive statistics for the given entries
func CalculateStatistics(entries []*IndexedEntry, startDate, endDate time.Time, isAllTime bool) *Statistics {
	stats := &Statistics{
		Period: PeriodInfo{
			StartDate: startDate,
			EndDate:   endDate,
			TotalDays: calculateTotalDays(startDate, endDate),
			IsAllTime: isAllTime,
		},
	}

	// If no entries, return empty stats
	if len(entries) == 0 {
		stats.Summary = SummaryStats{}
		stats.TopTags = []TagStat{}
		stats.TopMentions = []MentionStat{}
		stats.Patterns = ActivityPatterns{
			DayOfWeek:          make(map[time.Weekday]int),
			TimeOfDay:          make(map[string]int),
			HourlyDistribution: make(map[int]int),
		}
		return stats
	}

	// Calculate summary statistics
	stats.Summary = calculateSummaryStats(entries, startDate, endDate)

	// Aggregate tags and mentions
	stats.Tags = aggregateTags(entries)
	stats.TopTags = topN(stats.Tags, 5)
	stats.Mentions = aggregateMentions(entries)
	stats.TopMentions = topNMentions(stats.Mentions, 5)

	// Calculate activity patterns
	stats.Patterns = calculateActivityPatterns(entries)

	return stats
}

// CalculateFilteredStatistics computes statistics for entries filtered by tag or mention
func CalculateFilteredStatistics(allEntries []*IndexedEntry, startDate, endDate time.Time, isAllTime bool, filterType, filterValue string) *Statistics {
	// Filter entries
	var filteredEntries []*IndexedEntry
	switch filterType {
	case "tag":
		filteredEntries = filterByTag(allEntries, filterValue)
	case "mention":
		filteredEntries = filterByMention(allEntries, filterValue)
	default:
		filteredEntries = allEntries
	}

	// Calculate base statistics
	stats := CalculateStatistics(filteredEntries, startDate, endDate, isAllTime)

	// Add filter info
	stats.FilteredBy = &FilterInfo{
		Type:  filterType,
		Value: filterValue,
	}

	return stats
}

// calculateTotalDays returns the number of days between start and end (inclusive)
func calculateTotalDays(start, end time.Time) int {
	// Truncate to day boundary
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	duration := endDay.Sub(startDay)
	days := int(duration.Hours()/24) + 1 // +1 to include both start and end days

	return days
}

// calculateSummaryStats computes aggregate summary statistics
func calculateSummaryStats(entries []*IndexedEntry, startDate, endDate time.Time) SummaryStats {
	totalEntries := len(entries)

	// Calculate active days
	activeDays := calculateActiveDays(entries)
	activeDayCount := len(activeDays)

	// Calculate percentages and averages
	totalDays := calculateTotalDays(startDate, endDate)
	activeDaysPct := 0.0
	if totalDays > 0 {
		activeDaysPct = float64(activeDayCount) / float64(totalDays) * 100.0
	}

	avgPerActiveDay := 0.0
	if activeDayCount > 0 {
		avgPerActiveDay = float64(totalEntries) / float64(activeDayCount)
	}

	// Calculate streaks
	longestStreak, currentStreak := calculateStreaks(activeDays, endDate)

	// Calculate longest gap
	longestGap := calculateLongestGap(activeDays)

	return SummaryStats{
		TotalEntries:    totalEntries,
		ActiveDays:      activeDayCount,
		ActiveDaysPct:   activeDaysPct,
		AvgPerActiveDay: avgPerActiveDay,
		LongestStreak:   longestStreak,
		CurrentStreak:   currentStreak,
		LongestGap:      longestGap,
	}
}

// calculateActiveDays returns a sorted list of dates that have at least one entry
func calculateActiveDays(entries []*IndexedEntry) []time.Time {
	dayMap := make(map[string]time.Time)

	for _, entry := range entries {
		// Normalize to day boundary (midnight)
		day := time.Date(entry.Timestamp.Year(), entry.Timestamp.Month(), entry.Timestamp.Day(),
			0, 0, 0, 0, entry.Timestamp.Location())
		dayKey := day.Format("2006-01-02")
		dayMap[dayKey] = day
	}

	// Convert to sorted slice
	days := make([]time.Time, 0, len(dayMap))
	for _, day := range dayMap {
		days = append(days, day)
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})

	return days
}

// calculateStreaks computes longest and current streaks of consecutive active days
func calculateStreaks(activeDays []time.Time, endDate time.Time) (longest, current int) {
	if len(activeDays) == 0 {
		return 0, 0
	}

	// Calculate longest streak
	longestStreak := 1
	currentStreakCount := 1

	for i := 1; i < len(activeDays); i++ {
		prevDay := activeDays[i-1]
		currDay := activeDays[i]

		// Check if consecutive (exactly 1 day apart)
		diff := currDay.Sub(prevDay)
		if diff.Hours() <= 24*1.5 { // Allow some flexibility for DST
			currentStreakCount++
		} else {
			if currentStreakCount > longestStreak {
				longestStreak = currentStreakCount
			}
			currentStreakCount = 1
		}
	}

	// Check final streak
	if currentStreakCount > longestStreak {
		longestStreak = currentStreakCount
	}

	// Calculate current streak (working backwards from endDate)
	endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())
	currentStreakDays := 0

	for i := len(activeDays) - 1; i >= 0; i-- {
		day := activeDays[i]

		// Calculate expected day (counting backwards from end)
		expectedDay := endDay.AddDate(0, 0, -currentStreakDays)

		// Check if this active day matches the expected day
		if day.Format("2006-01-02") == expectedDay.Format("2006-01-02") {
			currentStreakDays++
		} else if day.Before(expectedDay) {
			// Found a gap, stop counting
			break
		}
	}

	return longestStreak, currentStreakDays
}

// calculateLongestGap finds the longest period without entries
func calculateLongestGap(activeDays []time.Time) GapInfo {
	if len(activeDays) <= 1 {
		return GapInfo{Days: 0}
	}

	longestGap := GapInfo{Days: 0}

	// Check gaps between active days
	for i := 1; i < len(activeDays); i++ {
		prevDay := activeDays[i-1]
		currDay := activeDays[i]

		diff := currDay.Sub(prevDay)
		gapDays := int(diff.Hours()/24) - 1 // -1 because we count days between, not including boundaries

		if gapDays > longestGap.Days {
			longestGap = GapInfo{
				Days:      gapDays,
				StartDate: prevDay,
				EndDate:   currDay,
			}
		}
	}

	return longestGap
}

// aggregateTags counts occurrences of each tag across entries
func aggregateTags(entries []*IndexedEntry) []TagStat {
	tagCounts := make(map[string]int)

	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tagCounts[tag]++
		}
	}

	// Convert to slice and sort by count (descending), then name (ascending)
	stats := make([]TagStat, 0, len(tagCounts))
	for name, count := range tagCounts {
		stats = append(stats, TagStat{Name: name, Count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Count > stats[j].Count
	})

	return stats
}

// aggregateMentions counts occurrences of each mention across entries
func aggregateMentions(entries []*IndexedEntry) []MentionStat {
	mentionCounts := make(map[string]int)

	for _, entry := range entries {
		for _, mention := range entry.Mentions {
			mentionCounts[mention]++
		}
	}

	// Convert to slice and sort by count (descending), then name (ascending)
	stats := make([]MentionStat, 0, len(mentionCounts))
	for name, count := range mentionCounts {
		stats = append(stats, MentionStat{Name: name, Count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Count > stats[j].Count
	})

	return stats
}

// topN returns the top N items from a sorted list
func topN(stats []TagStat, n int) []TagStat {
	if len(stats) <= n {
		return stats
	}
	return stats[:n]
}

// topNMentions returns the top N mentions from a sorted list
func topNMentions(stats []MentionStat, n int) []MentionStat {
	if len(stats) <= n {
		return stats
	}
	return stats[:n]
}

// calculateActivityPatterns analyzes temporal patterns in entries
func calculateActivityPatterns(entries []*IndexedEntry) ActivityPatterns {
	patterns := ActivityPatterns{
		DayOfWeek:          make(map[time.Weekday]int),
		TimeOfDay:          make(map[string]int),
		HourlyDistribution: make(map[int]int),
	}

	// Count entries by day of week and time of day
	for _, entry := range entries {
		// Day of week
		patterns.DayOfWeek[entry.Timestamp.Weekday()]++

		// Time of day category
		category := categorizeTimeOfDay(entry.Timestamp)
		patterns.TimeOfDay[category]++

		// Hour of day
		hour := entry.Timestamp.Hour()
		patterns.HourlyDistribution[hour]++
	}

	// Find busiest day of week
	maxDayCount := 0
	for day, count := range patterns.DayOfWeek {
		if count > maxDayCount {
			maxDayCount = count
			patterns.BusiestDay = day
		}
	}
	patterns.BusiestDayCount = maxDayCount

	// Find busiest time of day
	maxTimeCount := 0
	for timeCategory, count := range patterns.TimeOfDay {
		if count > maxTimeCount {
			maxTimeCount = count
			patterns.BusiestTime = timeCategory
		}
	}

	return patterns
}

// categorizeTimeOfDay returns the time of day category for a given timestamp
func categorizeTimeOfDay(t time.Time) string {
	hour := t.Hour()

	if hour >= 5 && hour < 12 {
		return TimeMorning
	} else if hour >= 12 && hour < 17 {
		return TimeAfternoon
	} else if hour >= 17 && hour < 21 {
		return TimeEvening
	}
	return TimeNight
}

// filterByTag returns only entries that have the specified tag
func filterByTag(entries []*IndexedEntry, tag string) []*IndexedEntry {
	var filtered []*IndexedEntry

	for _, entry := range entries {
		for _, entryTag := range entry.Tags {
			if entryTag == tag {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}

// filterByMention returns only entries that have the specified mention
func filterByMention(entries []*IndexedEntry, mention string) []*IndexedEntry {
	var filtered []*IndexedEntry

	for _, entry := range entries {
		for _, entryMention := range entry.Mentions {
			if entryMention == mention {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}
