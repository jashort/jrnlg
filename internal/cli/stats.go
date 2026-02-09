package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
)

// statsOptions holds parsed command line arguments for stats command
type statsOptions struct {
	FromDate *time.Time
	ToDate   *time.Time
	All      bool   // --all flag
	Tag      string // --tag filter
	Mention  string // --mention filter
	Format   string // "default", "json", "detailed"
	Detailed bool   // --detailed flag (sets format to "detailed")
}

// executeStats performs the actual stats logic
func (a *App) executeStats(opts *statsOptions) error {
	// Fetch entries based on options
	entries, startDate, endDate, isAllTime, err := a.fetchEntriesForStats(opts)
	if err != nil {
		return err
	}

	// Calculate statistics
	var stats *internal.Statistics
	if opts.Tag != "" {
		stats = internal.CalculateFilteredStatistics(entries, startDate, endDate, isAllTime, "tag", opts.Tag)
	} else if opts.Mention != "" {
		stats = internal.CalculateFilteredStatistics(entries, startDate, endDate, isAllTime, "mention", opts.Mention)
	} else {
		stats = internal.CalculateStatistics(entries, startDate, endDate, isAllTime)
	}

	// Handle empty results
	if stats.Summary.TotalEntries == 0 {
		if opts.Tag != "" {
			return fmt.Errorf("no entries found with tag #%s", opts.Tag)
		} else if opts.Mention != "" {
			return fmt.Errorf("no entries found with mention @%s", opts.Mention)
		}
		fmt.Println("No entries found in the specified date range.")
		return nil
	}

	// Display statistics
	switch opts.Format {
	case "json":
		output := displayStatsJSON(stats)
		fmt.Println(output)
	case "detailed":
		output := displayStatsDetailed(stats)
		fmt.Print(output)
	default:
		output := displayStatsDefault(stats)
		fmt.Print(output)
	}

	return nil
}

// fetchEntriesForStats fetches entries based on options and returns entries, date range, and isAllTime flag
func (a *App) fetchEntriesForStats(opts *statsOptions) ([]*internal.IndexedEntry, time.Time, time.Time, bool, error) {
	// Determine date range
	var startDate, endDate time.Time
	isAllTime := opts.All

	now := time.Now()

	if opts.All {
		// Build index for all entries
		filter := internal.EntryFilter{} // No filters = all entries
		index, err := a.storage.GetIndex(filter)
		if err != nil {
			return nil, time.Time{}, time.Time{}, false, fmt.Errorf("failed to build index: %w", err)
		}

		// Get all entries
		allEntries := index.GetAllEntries()
		if len(allEntries) == 0 {
			return nil, time.Time{}, time.Time{}, false, fmt.Errorf("no entries found")
		}

		// Find earliest and latest timestamps
		startDate = allEntries[0].Timestamp
		endDate = allEntries[0].Timestamp

		for _, entry := range allEntries {
			if entry.Timestamp.Before(startDate) {
				startDate = entry.Timestamp
			}
			if entry.Timestamp.After(endDate) {
				endDate = entry.Timestamp
			}
		}

		// Normalize to day boundaries
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

		return allEntries, startDate, endDate, isAllTime, nil
	}

	// Use specified date range or default to last 30 days
	if opts.FromDate != nil {
		startDate = *opts.FromDate
		// Normalize to start of day
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	} else {
		// Default: 30 days ago
		startDate = now.AddDate(0, 0, -30)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	}

	if opts.ToDate != nil {
		endDate = *opts.ToDate
		// Normalize to end of day
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	} else {
		// Default: today
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	}

	// Build index for date range
	filter := internal.EntryFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	index, err := a.storage.GetIndex(filter)
	if err != nil {
		return nil, time.Time{}, time.Time{}, false, fmt.Errorf("failed to build index: %w", err)
	}

	// Fetch entries in range
	entries := index.GetEntriesInRange(startDate, endDate)

	return entries, startDate, endDate, isAllTime, nil
}

// displayStatsDefault renders statistics in default format with colors
func displayStatsDefault(stats *internal.Statistics) string {
	var sb strings.Builder

	// Header
	if stats.FilteredBy != nil {
		if stats.FilteredBy.Type == "tag" {
			sb.WriteString(color.Cyan(fmt.Sprintf("Journal Statistics for #%s", stats.FilteredBy.Value)))
		} else {
			sb.WriteString(color.Cyan(fmt.Sprintf("Journal Statistics for @%s", stats.FilteredBy.Value)))
		}
	} else {
		sb.WriteString(color.Cyan("Journal Statistics"))
	}

	// Date range
	dateFormat := "Jan 2, 2006"
	sb.WriteString(fmt.Sprintf(" (%s - %s):\n\n",
		stats.Period.StartDate.Format(dateFormat),
		stats.Period.EndDate.Format(dateFormat)))

	// Summary section
	sb.WriteString(color.Cyan("Summary:\n"))
	sb.WriteString(fmt.Sprintf("  Total entries:        %s\n", color.Green(fmt.Sprintf("%d", stats.Summary.TotalEntries))))
	sb.WriteString(fmt.Sprintf("  Active days:          %s/%d (%s)\n",
		color.Green(fmt.Sprintf("%d", stats.Summary.ActiveDays)),
		stats.Period.TotalDays,
		color.Green(fmt.Sprintf("%.1f%%", stats.Summary.ActiveDaysPct))))
	sb.WriteString(fmt.Sprintf("  Avg per active day:   %s entries\n",
		color.Green(fmt.Sprintf("%.1f", stats.Summary.AvgPerActiveDay))))
	sb.WriteString(fmt.Sprintf("  Longest streak:       %s days\n",
		color.Green(fmt.Sprintf("%d", stats.Summary.LongestStreak))))
	sb.WriteString(fmt.Sprintf("  Current streak:       %s days\n",
		color.Green(fmt.Sprintf("%d", stats.Summary.CurrentStreak))))

	if stats.Summary.LongestGap.Days > 0 {
		sb.WriteString(fmt.Sprintf("  Longest gap:          %s days (%s - %s)\n",
			color.Green(fmt.Sprintf("%d", stats.Summary.LongestGap.Days)),
			stats.Summary.LongestGap.StartDate.Format(dateFormat),
			stats.Summary.LongestGap.EndDate.Format(dateFormat)))
	} else {
		sb.WriteString(fmt.Sprintf("  Longest gap:          %s\n", color.Green("0 days")))
	}

	sb.WriteString("\n")

	// Tags section
	if len(stats.TopTags) > 0 {
		if stats.FilteredBy != nil && stats.FilteredBy.Type == "tag" {
			sb.WriteString(color.Cyan("Co-occurring Tags:\n"))
		} else {
			sb.WriteString(color.Cyan("Top Tags:\n"))
		}
		for _, tag := range stats.TopTags {
			sb.WriteString(fmt.Sprintf("  #%-20s %s entries\n", tag.Name, color.Green(fmt.Sprintf("%d", tag.Count))))
		}
		sb.WriteString("\n")
	}

	// Mentions section
	if len(stats.TopMentions) > 0 {
		if stats.FilteredBy != nil {
			sb.WriteString(color.Cyan(fmt.Sprintf("Top Mentions (in %s entries):\n", formatFilterType(stats.FilteredBy))))
		} else {
			sb.WriteString(color.Cyan("Top Mentions:\n"))
		}
		for _, mention := range stats.TopMentions {
			sb.WriteString(fmt.Sprintf("  @%-20s %s entries\n", mention.Name, color.Green(fmt.Sprintf("%d", mention.Count))))
		}
		sb.WriteString("\n")
	}

	// Activity patterns
	sb.WriteString(color.Cyan("Activity Patterns:\n"))
	if stats.Patterns.BusiestDayCount > 0 {
		sb.WriteString(fmt.Sprintf("  Busiest day:          %s (%s entries)\n",
			stats.Patterns.BusiestDay.String(),
			color.Green(fmt.Sprintf("%d", stats.Patterns.DayOfWeek[stats.Patterns.BusiestDay]))))
	}
	if stats.Patterns.BusiestTime != "" {
		sb.WriteString(fmt.Sprintf("  Busiest time:         %s (%s entries)\n",
			formatTimeCategory(stats.Patterns.BusiestTime),
			color.Green(fmt.Sprintf("%d", stats.Patterns.TimeOfDay[stats.Patterns.BusiestTime]))))
	}

	return sb.String()
}

// displayStatsDetailed renders statistics in detailed format
func displayStatsDetailed(stats *internal.Statistics) string {
	var sb strings.Builder

	// Start with default output
	sb.WriteString(displayStatsDefault(stats))

	// Add detailed sections
	sb.WriteString("\n")
	sb.WriteString(color.Cyan("Day of Week Breakdown:\n"))

	// Sort weekdays
	weekdays := []time.Weekday{
		time.Monday, time.Tuesday, time.Wednesday, time.Thursday,
		time.Friday, time.Saturday, time.Sunday,
	}

	totalEntries := stats.Summary.TotalEntries
	maxCount := 0
	for _, count := range stats.Patterns.DayOfWeek {
		if count > maxCount {
			maxCount = count
		}
	}

	for _, day := range weekdays {
		count := stats.Patterns.DayOfWeek[day]
		pct := 0.0
		if totalEntries > 0 {
			pct = float64(count) / float64(totalEntries) * 100
		}

		arrow := ""
		if day == stats.Patterns.BusiestDay {
			arrow = " ←"
		}

		sb.WriteString(fmt.Sprintf("  %-10s %s entries (%s)%s\n",
			day.String()+":",
			color.Green(fmt.Sprintf("%d", count)),
			color.Green(fmt.Sprintf("%.1f%%", pct)),
			arrow))
	}

	sb.WriteString("\n")
	sb.WriteString(color.Cyan("Time of Day Breakdown:\n"))

	// Order: morning, afternoon, evening, night
	timeCategories := []string{internal.TimeMorning, internal.TimeAfternoon, internal.TimeEvening, internal.TimeNight}
	for _, category := range timeCategories {
		count := stats.Patterns.TimeOfDay[category]
		pct := 0.0
		if totalEntries > 0 {
			pct = float64(count) / float64(totalEntries) * 100
		}

		arrow := ""
		if category == stats.Patterns.BusiestTime {
			arrow = " ←"
		}

		sb.WriteString(fmt.Sprintf("  %-25s %s entries (%s)%s\n",
			formatTimeCategory(category)+":",
			color.Green(fmt.Sprintf("%d", count)),
			color.Green(fmt.Sprintf("%.1f%%", pct)),
			arrow))
	}

	// Hourly distribution (top 5 hours)
	if len(stats.Patterns.HourlyDistribution) > 0 {
		sb.WriteString("\n")
		sb.WriteString(color.Cyan("Hourly Distribution (top 5 hours):\n"))

		// Sort hours by count
		type hourCount struct {
			hour  int
			count int
		}
		var hours []hourCount
		for hour, count := range stats.Patterns.HourlyDistribution {
			hours = append(hours, hourCount{hour, count})
		}
		sort.Slice(hours, func(i, j int) bool {
			if hours[i].count == hours[j].count {
				return hours[i].hour < hours[j].hour
			}
			return hours[i].count > hours[j].count
		})

		// Show top 5
		limit := 5
		if len(hours) < limit {
			limit = len(hours)
		}
		for i := 0; i < limit; i++ {
			h := hours[i]
			sb.WriteString(fmt.Sprintf("  %02d:00-%02d:00          %s entries\n",
				h.hour, (h.hour+1)%24,
				color.Green(fmt.Sprintf("%d", h.count))))
		}
	}

	return sb.String()
}

// displayStatsJSON renders statistics as JSON
// nolint:gofmt
func displayStatsJSON(stats *internal.Statistics) string {
	output := map[string]any{
		"period": map[string]any{
			"start_date":  stats.Period.StartDate.Format(time.RFC3339),
			"end_date":    stats.Period.EndDate.Format(time.RFC3339),
			"total_days":  stats.Period.TotalDays,
			"is_all_time": stats.Period.IsAllTime,
		},
		"summary": map[string]any{
			"total_entries":      stats.Summary.TotalEntries,
			"active_days":        stats.Summary.ActiveDays,
			"active_days_pct":    roundFloat(stats.Summary.ActiveDaysPct, 1),
			"avg_per_active_day": roundFloat(stats.Summary.AvgPerActiveDay, 1),
			"longest_streak":     stats.Summary.LongestStreak,
			"current_streak":     stats.Summary.CurrentStreak,
		},
	}

	// Add longest gap if present
	if stats.Summary.LongestGap.Days > 0 {
		output["summary"].(map[string]any)["longest_gap"] = map[string]any{
			"days":       stats.Summary.LongestGap.Days,
			"start_date": stats.Summary.LongestGap.StartDate.Format("2006-01-02"),
			"end_date":   stats.Summary.LongestGap.EndDate.Format("2006-01-02"),
		}
	} else {
		output["summary"].(map[string]any)["longest_gap"] = map[string]any{
			"days": 0,
		}
	}

	// Add tags
	topTags := make([]map[string]any, len(stats.TopTags))
	for i, tag := range stats.TopTags {
		topTags[i] = map[string]any{
			"name":  tag.Name,
			"count": tag.Count,
		}
	}
	output["top_tags"] = topTags

	// Add mentions
	topMentions := make([]map[string]any, len(stats.TopMentions))
	for i, mention := range stats.TopMentions {
		topMentions[i] = map[string]any{
			"name":  mention.Name,
			"count": mention.Count,
		}
	}
	output["top_mentions"] = topMentions

	// Add patterns
	dayOfWeek := make(map[string]int)
	for day, count := range stats.Patterns.DayOfWeek {
		dayOfWeek[day.String()] = count
	}

	output["patterns"] = map[string]any{
		"day_of_week":  dayOfWeek,
		"time_of_day":  stats.Patterns.TimeOfDay,
		"busiest_day":  stats.Patterns.BusiestDay.String(),
		"busiest_time": stats.Patterns.BusiestTime,
	}

	// Add filter info if present
	if stats.FilteredBy != nil {
		output["filtered_by"] = map[string]any{
			"type":  stats.FilteredBy.Type,
			"value": stats.FilteredBy.Value,
		}
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes)
}

// Helper functions

func formatTimeCategory(category string) string {
	switch category {
	case internal.TimeMorning:
		return "Morning (5am-12pm)"
	case internal.TimeAfternoon:
		return "Afternoon (12pm-5pm)"
	case internal.TimeEvening:
		return "Evening (5pm-9pm)"
	case internal.TimeNight:
		return "Night (9pm-5am)"
	default:
		return category
	}
}

func formatFilterType(filter *internal.FilterInfo) string {
	if filter.Type == "tag" {
		return fmt.Sprintf("#%s", filter.Value)
	}
	return fmt.Sprintf("@%s", filter.Value)
}

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}
