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

	// Pre-scan for global flags (--help, --version) anywhere in args
	// This allows: ./jrnlg list -from today --help
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return a.ShowHelp()
		}
		if arg == "--version" || arg == "-v" {
			return a.ShowVersion()
		}
	}

	// Handle explicit commands
	switch args[0] {
	case "search", "list":
		return a.Search(args[1:])
	case "edit":
		return a.EditEntry(args[1:])
	case "delete", "rm":
		return a.DeleteEntries(args[1:])
	case "tags":
		return a.HandleTagsCommand(args[1:])
	case "mentions":
		return a.HandleMentionsCommand(args[1:])
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
    jrnlg edit [selector]           Edit an entry
    jrnlg delete [selector] [flags] Delete entries
    jrnlg tags [command]            Manage tags
    jrnlg mentions [command]        Manage mentions

EDIT SELECTORS:
    jrnlg edit                      Edit most recent entry
    jrnlg edit 2026-02-08-21-32-00  Edit specific entry by timestamp
    jrnlg edit yesterday            Edit entry from date (picker if multiple)

DELETE SELECTORS:
    jrnlg delete 2026-02-08-21-32-00       Delete specific entry
    jrnlg delete --from yesterday          Delete entries from date range
    jrnlg delete --from yesterday --force  Skip confirmation

TAG COMMANDS:
    jrnlg tags                      List all tags with usage counts
    jrnlg tags list                 List all tags (same as above)
    jrnlg tags list --orphaned      Show only tags used once
    jrnlg tags rename OLD NEW       Rename tag (case-insensitive)
    jrnlg tags rename OLD NEW --dry-run    Preview changes without applying
    jrnlg tags rename OLD NEW --force      Skip confirmation prompt

MENTION COMMANDS:
    jrnlg mentions                  List all mentions with usage counts
    jrnlg mentions list             List all mentions (same as above)
    jrnlg mentions list --orphaned  Show only mentions used once
    jrnlg mentions rename OLD NEW   Rename mention (case-insensitive)
    jrnlg mentions rename OLD NEW --dry-run    Preview changes
    jrnlg mentions rename OLD NEW --force      Skip confirmation

SEARCH TERMS:
    #tag                            Find entries with this tag
    @mention                        Find entries with this mention
    keyword                         Find entries containing keyword
    
    Multiple terms use AND logic (all must match)

FLAGS:
    -from, --from <date>            Show entries from this date onwards
    -to, --to <date>                Show entries up to this date
    -n, --limit <num>               Limit number of results
    --offset <num>                  Skip first N results
    -r, --reverse                   Show newest entries first
    --summary                       Show compact summary format
    --format <fmt>                  Output format: full, summary, json
    --color <mode>                  Color mode: auto, always, never
    -f, --force                     Skip confirmation (delete/tags/mentions)
    -h, --help                      Show this help
    -v, --version                   Show version

DATE FORMATS:
    ISO 8601:     2026-02-09, 2026-02-09T15:45:00
    Natural:      yesterday, today, tomorrow
                  "3 days ago", "1 week ago", "2 weeks ago"

EXAMPLES:
    jrnlg                           # Create new entry
    jrnlg edit                      # Edit most recent entry
    jrnlg edit yesterday            # Edit yesterday's entry
    jrnlg delete 2026-02-08-21-32-00 # Delete specific entry
    jrnlg delete --from "last week" --to yesterday  # Delete range
    jrnlg search '#work'            # Find work entries (note: quotes required)
    jrnlg search '@alice'           # Find entries mentioning Alice
    jrnlg list --from yesterday     # Recent entries
    jrnlg list --from "3 days ago" --to today --summary
    jrnlg search '#work' '@alice' --from 2026-01-01 --summary
    jrnlg tags                      # List all tags
    jrnlg tags rename code_review code-review    # Merge tag variations
    jrnlg mentions rename alice alice-smith      # Rename mention

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
