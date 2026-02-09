package internal

import (
	"strings"
	"sync"
	"time"
)

// Index provides fast lookup of journal entries by tags, mentions, and keywords
type Index struct {
	entries      []*IndexedEntry
	tagIndex     map[string][]*IndexedEntry // tag -> entries with that tag
	mentionIndex map[string][]*IndexedEntry // mention -> entries with that mention
	bodyMap      map[string]string          // filePath -> body text for keyword search
	mu           sync.RWMutex
}

// IndexedEntry contains metadata about a journal entry for indexing
type IndexedEntry struct {
	FilePath  string
	Timestamp time.Time
	Tags      []string
	Mentions  []string
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		entries:      make([]*IndexedEntry, 0),
		tagIndex:     make(map[string][]*IndexedEntry),
		mentionIndex: make(map[string][]*IndexedEntry),
		bodyMap:      make(map[string]string),
	}
}

// Build constructs the index from a list of files
// Uses parallel parsing for performance
func (idx *Index) Build(files []string, maxWorkers int, parseFunc func(string) (*JournalEntry, error)) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if maxWorkers < 1 {
		maxWorkers = 1
	}

	// Parse files in parallel
	type result struct {
		filePath string
		entry    *JournalEntry
		err      error
	}

	results := make(chan result, len(files))
	jobs := make(chan string, len(files))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				entry, err := parseFunc(filePath)
				results <- result{filePath: filePath, entry: entry, err: err}
			}
		}()
	}

	// Send jobs
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and build index
	for res := range results {
		if res.err != nil {
			// Skip invalid files
			continue
		}

		indexed := &IndexedEntry{
			FilePath:  res.filePath,
			Timestamp: res.entry.Timestamp,
			Tags:      res.entry.Tags,
			Mentions:  res.entry.Mentions,
		}

		idx.entries = append(idx.entries, indexed)
		idx.bodyMap[res.filePath] = res.entry.Body

		// Build tag index
		for _, tag := range res.entry.Tags {
			idx.tagIndex[tag] = append(idx.tagIndex[tag], indexed)
		}

		// Build mention index
		for _, mention := range res.entry.Mentions {
			idx.mentionIndex[mention] = append(idx.mentionIndex[mention], indexed)
		}
	}

	return nil
}

// SearchByTags finds entries that have ALL the specified tags (AND logic)
func (idx *Index) SearchByTags(tags []string) []*IndexedEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(tags) == 0 {
		return nil
	}

	// Normalize tags to lowercase
	normalizedTags := make([]string, len(tags))
	for i, tag := range tags {
		normalizedTags[i] = strings.ToLower(tag)
	}

	// Start with entries that have the first tag
	candidates := idx.tagIndex[normalizedTags[0]]
	if len(candidates) == 0 {
		return nil
	}

	// Filter to entries that have ALL tags
	var results []*IndexedEntry
	for _, entry := range candidates {
		if idx.hasAllTags(entry, normalizedTags) {
			results = append(results, entry)
		}
	}

	return results
}

// SearchByMentions finds entries that have ALL the specified mentions (AND logic)
func (idx *Index) SearchByMentions(mentions []string) []*IndexedEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(mentions) == 0 {
		return nil
	}

	// Normalize mentions to lowercase
	normalizedMentions := make([]string, len(mentions))
	for i, mention := range mentions {
		normalizedMentions[i] = strings.ToLower(mention)
	}

	// Start with entries that have the first mention
	candidates := idx.mentionIndex[normalizedMentions[0]]
	if len(candidates) == 0 {
		return nil
	}

	// Filter to entries that have ALL mentions
	var results []*IndexedEntry
	for _, entry := range candidates {
		if idx.hasAllMentions(entry, normalizedMentions) {
			results = append(results, entry)
		}
	}

	return results
}

// SearchByKeyword finds entries whose body contains the keyword (case-insensitive)
func (idx *Index) SearchByKeyword(keyword string) []*IndexedEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if keyword == "" {
		return nil
	}

	keyword = strings.ToLower(keyword)
	var results []*IndexedEntry

	for filePath, body := range idx.bodyMap {
		if strings.Contains(strings.ToLower(body), keyword) {
			// Find the entry with this file path
			for _, entry := range idx.entries {
				if entry.FilePath == filePath {
					results = append(results, entry)
					break
				}
			}
		}
	}

	return results
}

// GetBody retrieves the body text for a given file path
func (idx *Index) GetBody(filePath string) string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.bodyMap[filePath]
}

// Size returns the number of indexed entries
func (idx *Index) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.entries)
}

// hasAllTags checks if an entry has all the specified tags
func (idx *Index) hasAllTags(entry *IndexedEntry, tags []string) bool {
	entryTags := make(map[string]bool)
	for _, tag := range entry.Tags {
		entryTags[tag] = true
	}

	for _, tag := range tags {
		if !entryTags[tag] {
			return false
		}
	}

	return true
}

// hasAllMentions checks if an entry has all the specified mentions
func (idx *Index) hasAllMentions(entry *IndexedEntry, mentions []string) bool {
	entryMentions := make(map[string]bool)
	for _, mention := range entry.Mentions {
		entryMentions[mention] = true
	}

	for _, mention := range mentions {
		if !entryMentions[mention] {
			return false
		}
	}

	return true
}

// TagStatistics returns a map of tag -> count of entries with that tag
func (idx *Index) TagStatistics() map[string]int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	stats := make(map[string]int)
	for tag, entries := range idx.tagIndex {
		stats[tag] = len(entries)
	}

	return stats
}

// MentionStatistics returns a map of mention -> count of entries with that mention
func (idx *Index) MentionStatistics() map[string]int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	stats := make(map[string]int)
	for mention, entries := range idx.mentionIndex {
		stats[mention] = len(entries)
	}

	return stats
}

// GetEntriesForTag returns all entries with the specified tag
func (idx *Index) GetEntriesForTag(tag string) []*IndexedEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	normalized := strings.ToLower(tag)
	return idx.tagIndex[normalized]
}

// GetEntriesForMention returns all entries with the specified mention
func (idx *Index) GetEntriesForMention(mention string) []*IndexedEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	normalized := strings.ToLower(mention)
	return idx.mentionIndex[normalized]
}
