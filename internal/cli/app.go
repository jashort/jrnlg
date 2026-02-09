package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/jashort/jrnlg/internal"
)

// App coordinates CLI operations
type App struct {
	storage *internal.FileSystemStorage
	config  *internal.Config
	version string
}

// NewApp creates a new CLI application
func NewApp(storage *internal.FileSystemStorage, config *internal.Config) *App {
	return &App{
		storage: storage,
		config:  config,
		version: "dev",
	}
}

// SetVersion sets the version string for the application
func (a *App) SetVersion(version string) {
	a.version = version
}

// Run executes the CLI application with the given arguments
func (a *App) Run(args []string) error {
	// Handle no arguments → create entry
	if len(args) == 0 {
		return a.CreateEntry()
	}

	// Handle --help, --version flags
	if args[0] == "--help" || args[0] == "-h" {
		return a.ShowHelp()
	}
	if args[0] == "--version" || args[0] == "-v" {
		return a.ShowVersion()
	}

	// Handle explicit commands
	switch args[0] {
	case "search", "list":
		return a.Search(args[1:])
	default:
		// Treat as implicit search (backward compatibility)
		return a.Search(args)
	}
}

// ShowHelp displays usage information
func (a *App) ShowHelp() error {
	help := `jrnlg - A simple, fast journal application

USAGE:
    jrnlg                           Create new journal entry
    jrnlg search [terms] [flags]    Search journal entries
    jrnlg list [terms] [flags]      List journal entries (alias for search)

SEARCH TERMS:
    #tag                            Find entries with this tag
    @mention                        Find entries with this mention
    keyword                         Find entries containing keyword
    
    Multiple terms use AND logic (all must match)

FLAGS:
    -from <date>                    Show entries from this date onwards
    -to <date>                      Show entries up to this date
    -n, --limit <num>               Limit number of results
    --offset <num>                  Skip first N results
    -r, --reverse                   Show newest entries first
    --summary                       Show compact summary format
    --format <fmt>                  Output format: full, summary, json
    -h, --help                      Show this help
    -v, --version                   Show version

DATE FORMATS:
    ISO 8601:     2026-02-09, 2026-02-09T15:45:00
    Natural:      yesterday, today, tomorrow
                  "3 days ago", "1 week ago", "2 weeks ago"

EXAMPLES:
    jrnlg                           # Create new entry
    jrnlg search '#work'            # Find work entries (note: quotes required)
    jrnlg search '@alice'           # Find entries mentioning Alice
    jrnlg list -from yesterday      # Recent entries
    jrnlg list -from "3 days ago" -to today --summary
    jrnlg search '#work' '@alice' -from 2026-01-01 --summary

CONFIGURATION:
    JRNLG_STORAGE_PATH              Storage location (default: ~/.jrnlg/entries)
    JRNLG_EDITOR_ARGS               Additional arguments for editor (e.g., "+startinsert")
    VISUAL, EDITOR                  Editor to use (default: vim)

For more information: https://github.com/jashort/jrnlg
`
	fmt.Print(help)
	return nil
}

// ShowVersion displays version information
func (a *App) ShowVersion() error {
	var commit, commitTime, modified string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value
			} else if setting.Key == "vcs.modified" && setting.Value == "true" {
				modified = " Modified"
			} else if setting.Key == "vcs.time" {
				commitTime = setting.Value
			}
		}
	}
	fmt.Printf("Commit: %s (%s)%s\n", commit, commitTime, modified)
	fmt.Printf("jrnlg version %s\n", a.version)
	return nil
}
