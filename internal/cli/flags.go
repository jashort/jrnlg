package cli

import (
	"fmt"
	"strings"
)

// FlagDef defines a command-line flag
type FlagDef struct {
	Names       []string // e.g., ["-n", "--limit"]
	Description string   // Help text
	HasValue    bool     // Does flag take an argument?
	Global      bool     // Available in all contexts?
}

// AllFlags is the registry of all supported flags
var AllFlags = []FlagDef{
	// Global flags
	{[]string{"-h", "--help"}, "Show help message", false, true},
	{[]string{"-v", "--version"}, "Show version information", false, true},

	// Search/list flags
	{[]string{"-from", "--from"}, "Show entries from this date onwards", true, false},
	{[]string{"-to", "--to"}, "Show entries up to this date", true, false},
	{[]string{"-n", "--limit"}, "Limit number of results", true, false},
	{[]string{"--offset"}, "Skip first N results", true, false},
	{[]string{"-r", "--reverse"}, "Show newest entries first", false, false},
	{[]string{"--summary"}, "Show compact summary format", false, false},
	{[]string{"--format"}, "Output format: full, summary, json", true, false},
	{[]string{"--color"}, "When to use colors: auto, always, never (default: auto)", true, false},

	// Delete flags
	{[]string{"-f", "--force"}, "Skip confirmation prompt", false, false},
}

// isKnownFlag checks if a flag is registered
func isKnownFlag(flag string) bool {
	for _, def := range AllFlags {
		for _, name := range def.Names {
			if name == flag {
				return true
			}
		}
	}
	return false
}

// flagRequiresValue checks if a flag requires a value argument
func flagRequiresValue(flag string) bool {
	for _, def := range AllFlags {
		for _, name := range def.Names {
			if name == flag {
				return def.HasValue
			}
		}
	}
	return false
}

// getFlagHelp returns the help text for a flag
func getFlagHelp(flag string) string {
	for _, def := range AllFlags {
		for _, name := range def.Names {
			if name == flag {
				return def.Description
			}
		}
	}
	return ""
}

// findSimilarFlags finds flags similar to the given input (for suggestions)
func findSimilarFlags(input string) []string {
	var similar []string
	input = strings.ToLower(input)

	// Remove leading dashes for comparison
	inputClean := strings.TrimLeft(input, "-")

	for _, def := range AllFlags {
		for _, name := range def.Names {
			nameClean := strings.TrimLeft(strings.ToLower(name), "-")

			// Check for prefix match or contains
			if strings.HasPrefix(nameClean, inputClean) || strings.Contains(nameClean, inputClean) {
				// Add the primary name (first in list)
				similar = append(similar, def.Names[0])
				break
			}

			// Check for small edit distance (Levenshtein-like simple check)
			if levenshteinDistance(inputClean, nameClean) <= 2 {
				similar = append(similar, def.Names[0])
				break
			}
		}
	}

	return similar
}

// unknownFlagError creates a helpful error message for unknown flags
func unknownFlagError(flag string) error {
	msg := fmt.Sprintf("unknown flag: %s", flag)

	// Find similar flags
	similar := findSimilarFlags(flag)
	if len(similar) > 0 {
		msg += "\n\nDid you mean one of these?"
		for _, s := range similar {
			help := getFlagHelp(s)
			msg += fmt.Sprintf("\n  %s", s)
			if help != "" {
				msg += fmt.Sprintf(" - %s", help)
			}
		}
	}

	return fmt.Errorf("%s", msg)
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = minInts(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func minInts(is ...int) int {
	m := is[0]
	for _, i := range is[1:] {
		if i < m {
			m = i
		}
	}
	return m
}
