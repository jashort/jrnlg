package cli

import (
	"fmt"
	"sort"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli/color"
	"github.com/jashort/jrnlg/internal/cli/format"
)

// Search executes a search query and displays results (legacy entry point)
func (a *App) Search(args []string) error {
	// Parse search arguments
	searchArgs, err := parseSearchArgs(args)
	if err != nil {
		return err
	}

	return a.executeSearch(searchArgs)
}

// executeSearch performs the actual search logic
func (a *App) executeSearch(searchArgs SearchArgs) error {
	// Build entry filter for date ranges and pagination
	filter := internal.EntryFilter{
		Limit:  searchArgs.Limit,
		Offset: searchArgs.Offset,
	}

	if searchArgs.FromDate != nil {
		filter.StartDate = searchArgs.FromDate
	}
	if searchArgs.ToDate != nil {
		filter.EndDate = searchArgs.ToDate
	}

	// Collect result sets for each search term type
	var resultSets [][]*internal.JournalEntry

	// Search by tags
	if len(searchArgs.Tags) > 0 {
		results, err := a.storage.SearchByTags(searchArgs.Tags, filter)
		if err != nil {
			return fmt.Errorf("tag search failed: %w", err)
		}
		resultSets = append(resultSets, results)
	}

	// Search by mentions
	if len(searchArgs.Mentions) > 0 {
		results, err := a.storage.SearchByMentions(searchArgs.Mentions, filter)
		if err != nil {
			return fmt.Errorf("mention search failed: %w", err)
		}
		resultSets = append(resultSets, results)
	}

	// Search by keywords
	for _, keyword := range searchArgs.Keywords {
		results, err := a.storage.SearchByKeyword(keyword, filter)
		if err != nil {
			return fmt.Errorf("keyword search failed: %w", err)
		}
		resultSets = append(resultSets, results)
	}

	// If no search terms, just list all entries in date range
	var finalResults []*internal.JournalEntry
	if len(resultSets) == 0 {
		results, err := a.storage.ListEntries(filter)
		if err != nil {
			return fmt.Errorf("listing entries failed: %w", err)
		}
		finalResults = results
	} else {
		// Intersect all result sets (AND logic)
		finalResults = intersectResults(resultSets)
	}

	// Apply reverse sort if requested (newest first)
	if searchArgs.Reverse {
		sort.Slice(finalResults, func(i, j int) bool {
			return finalResults[i].Timestamp.After(finalResults[j].Timestamp)
		})
	}

	// Create colorizer based on color mode
	colorizer := color.New(searchArgs.ColorMode)

	// Format and display results
	formatter := format.GetFormatter(searchArgs.Format)
	output := formatter.Format(finalResults, colorizer)
	fmt.Print(output)

	return nil
}

// intersectResults returns only entries present in ALL result sets
// Uses timestamp as the unique key for comparison
func intersectResults(sets [][]*internal.JournalEntry) []*internal.JournalEntry {
	if len(sets) == 0 {
		return []*internal.JournalEntry{}
	}

	if len(sets) == 1 {
		return sets[0]
	}

	// Build a map of timestamp -> entry for the first set
	candidates := make(map[int64]*internal.JournalEntry)
	for _, entry := range sets[0] {
		candidates[entry.Timestamp.Unix()] = entry
	}

	// For each subsequent set, keep only entries that exist in candidates
	for i := 1; i < len(sets); i++ {
		found := make(map[int64]bool)
		for _, entry := range sets[i] {
			ts := entry.Timestamp.Unix()
			if _, exists := candidates[ts]; exists {
				found[ts] = true
			}
		}

		// Remove entries not found in this set
		for ts := range candidates {
			if !found[ts] {
				delete(candidates, ts)
			}
		}

		// Early exit if no candidates remain
		if len(candidates) == 0 {
			return []*internal.JournalEntry{}
		}
	}

	// Convert map back to slice and sort by timestamp
	result := make([]*internal.JournalEntry, 0, len(candidates))
	for _, entry := range candidates {
		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result
}
