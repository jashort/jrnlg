# jrnlg

A fast, minimal CLI journal application written in Go. Inspired by [jrnl](https://jrnl.sh/en/stable/)

## Features
- **Fast & Simple**: One Markdown file per entry, no database required
- **Natural Language Dates**: Search using "yesterday", "3 days ago", "last week"
- **Rich Metadata**: Automatic extraction of #tags, @mentions, and timestamps
- **Tag & Mention Management**: List and rename tags/mentions across all entries
- **Edit & Delete**: Edit existing entries or delete by date range with confirmation
- **Colorized Output**: Beautiful syntax highlighting for timestamps, tags, and mentions with smart terminal detection
- **Multiple Output Formats**: Full, summary, or JSON output
- **Timezone Aware**: Preserves original timezone abbreviation (PST, EST, etc.) in entry content
- **Search**: Filter by tags, mentions, keywords, and date ranges
- **Editor Integration**: Uses your preferred editor (VISUAL/EDITOR environment variables)

## Installation

### From Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/jashort/jrnlg.git
cd jrnlg
go build -o jrnlg
sudo mv jrnlg /usr/local/bin/  # Optional: install system-wide
```

### Quick Start

```bash
# Create your first entry
jrnlg

# List all entries
jrnlg list

# Search for entries
jrnlg search '#work'
jrnlg list --from yesterday
```

## Usage

### Creating Entries

Simply run `jrnlg` to open your editor:

```bash
jrnlg
```

Write your entry in Markdown format. Tags (`#work`, `#personal`) and mentions (`@alice`, `@bob`) are automatically extracted.

**Example Entry:**
```markdown
## Monday 2024-02-09 2:30 PM PST

Had a productive meeting with @alice about the #project deadline. 
Need to focus on #development this week.
```

### Listing Entries

```bash
# List all entries (full format)
jrnlg list

# List with summary format (one line per entry)
jrnlg list --summary

# List as JSON
jrnlg list --format json

# Limit number of results
jrnlg list -n 10

# Show newest entries first
jrnlg list -r
```

### Editing Entries

```bash
# Edit the most recent entry
jrnlg edit

# Edit a specific entry by timestamp (from filename)
jrnlg edit 2024-02-09-14-30-00

# Edit an entry from yesterday
# (shows picker if multiple entries found)
jrnlg edit yesterday

# Edit from a specific date
jrnlg edit "3 days ago"
```

**Note:** Timestamps cannot be changed during editing to maintain data integrity.

### Deleting Entries

```bash
# Delete a specific entry (with confirmation)
jrnlg delete 2024-02-09-14-30-00

# Delete entries from a date range (with confirmation)
jrnlg delete --from yesterday --to today

# Delete without confirmation prompt
jrnlg delete 2024-02-09-14-30-00 --force

# Delete all entries from last week
jrnlg delete --from "1 week ago" --to yesterday
```

**Warning:** Deletion is permanent and cannot be undone.

### Searching Entries

```bash
# Search by tag (quotes required in bash because # is a comment character)
jrnlg search '#work'

# Search by mention
jrnlg search '@alice'

# Search by keyword
jrnlg search deadline

# Multiple search terms (AND logic - all must match)
jrnlg search '#work' '@alice' deadline

# Search with date ranges
jrnlg search '#work' --from 2024-01-01 --to 2024-12-31
```

### Managing Tags and Mentions

Over time, you might accumulate inconsistent tags (e.g., `#code_review`, `#Code_Review`, `#code-review`). The tags and mentions commands help you find and merge these variations.

**List all tags with usage counts:**

```bash
# List all tags alphabetically
jrnlg tags

# List only tags used once (orphaned tags)
jrnlg tags list --orphaned
```

**Rename tags (case-insensitive, merges all variations):**

```bash
# Rename a tag - will match #code_review, #Code_Review, etc.
jrnlg tags rename code_review code-review

# Preview changes without applying them
jrnlg tags rename code_review code-review --dry-run

# Skip confirmation prompt
jrnlg tags rename code_review code-review --force
```

**Manage mentions:**

```bash
# List all mentions
jrnlg mentions

# Rename a mention (case-insensitive)
jrnlg mentions rename john_doe john-smith

# Preview changes
jrnlg mentions rename john_doe john-smith --dry-run
```

**Key features:**
- **Case-insensitive**: Matches and renames `#Code_Review`, `#CODE_REVIEW`, `#code-review`
- **Automatic deduplication**: Merges duplicate tags after renaming
- **Preview before changes**: Shows affected entries before applying changes
- **Safe by default**: Asks for confirmation unless `--force` is used
- **Warning on merges**: Shows a warning if the target tag/mention already exists

### Natural Language Date Filters

```bash
# Entries from yesterday
jrnlg list --from yesterday

# Entries from the last 3 days
jrnlg list --from "3 days ago"

# Entries from last week until yesterday
jrnlg list --from "1 week ago" --to yesterday

# ISO 8601 dates also work
jrnlg list --from 2024-01-01 --to 2024-12-31
```

### Shell Quoting Important Note

When using tags or mentions in bash/zsh, you **must quote them** because `#` starts a comment:

```bash
jrnlg search '#work'    # ✅ Correct
jrnlg search #work      # ❌ Wrong - # treated as comment
```

## Configuration

### Environment Variables

- `JRNLG_STORAGE_PATH` - Storage location (default: `~/.jrnlg/entries`)
- `VISUAL` or `EDITOR` - Editor to use (default: vim → vi → nano)
- `JRNLG_EDITOR_ARGS` - Additional arguments passed to the editor (optional)
- `NO_COLOR` - Set to any value to disable colored output (follows [no-color.org](https://no-color.org/) standard)

### Color Output

jrnlg uses colors to make output more readable by highlighting timestamps, tags, and mentions:

```bash
# Auto mode (default) - colors when outputting to terminal, plain when piped
jrnlg list

# Always use colors (even when piped)
jrnlg list --color always

# Never use colors
jrnlg list --color never

# Disable colors via environment variable
NO_COLOR=1 jrnlg list
```

**Color scheme:**
- Timestamps: cyan
- Tags (#work): green  
- Mentions (@alice): yellow
- Metadata/separators: dim gray

Colors are automatically disabled when output is redirected or piped, and respect the `NO_COLOR` environment variable.

### Editor Configuration

You can customize how your editor opens by setting `JRNLG_EDITOR_ARGS`. Arguments are passed to the editor before the filename.

**Vim/Neovim Examples:**
```bash
# Start in insert mode
export JRNLG_EDITOR_ARGS="+startinsert"

# Start at line 3 in insert mode
export JRNLG_EDITOR_ARGS="+startinsert +call cursor(3,1)"

# Start at end of file
export JRNLG_EDITOR_ARGS="+normal G"
```

**VSCode Example:**
```bash
# Wait for file to close before continuing
export EDITOR="code"
export JRNLG_EDITOR_ARGS="--wait"
```

**Emacs Example:**
```bash
# Open in terminal mode
export EDITOR="emacs"
export JRNLG_EDITOR_ARGS="-nw"
```

**Note:** Arguments with spaces should be quoted:
```bash
export JRNLG_EDITOR_ARGS='"+startinsert" "+call cursor(3,1)"'
```

### Storage Format

Entries are stored as individual Markdown files:

```
~/.jrnlg/entries/
├── 2024/
│   ├── 01/
│   │   ├── 2024-01-15-09-30-00.md
│   │   └── 2024-01-15-14-45-00.md
│   └── 02/
│       └── 2024-02-09-14-30-00.md
```

**File naming:**
- UTC timestamps in filenames (e.g., `2024-02-09-14-30-00.md`)
- Original timezone preserved in entry content
- Collision handling: adds `-01`, `-02` suffix if needed

## Command Reference

### Global Options

```
--help, -h       Show help message
--version, -v    Show version information
```

### Create Command

```
jrnlg

Opens your editor to create a new journal entry.
```

### Edit Command

```
jrnlg edit [selector]

Selectors:
  (none)                      Edit most recent entry
  YYYY-MM-DD-HH-MM-SS         Edit specific entry by timestamp
  yesterday, "3 days ago"     Edit entry from date (picker if multiple)

Note: Timestamps cannot be changed during editing.
```

### Delete Command

```
jrnlg delete [selector] [options]

Selectors:
  YYYY-MM-DD-HH-MM-SS         Delete specific entry
  --from <date> --to <date>   Delete entries in date range

Options:
  -f, --force                 Skip confirmation prompt
  --from <date>               Start date
  --to <date>                 End date

Warning: Deletion is permanent and cannot be undone.
```

### List Command

```
jrnlg list [options]

Options:
  -n <number>          Limit number of results
  --offset <number>    Skip first N results
  -r                   Reverse order (newest first)
  --summary            Use summary format (one line per entry)
  --format <format>    Output format: full, summary, json (default: full)
  --from <date>        Start date (ISO 8601 or natural language)
  --to <date>          End date (ISO 8601 or natural language)
  --color <mode>       Color mode: auto, always, never (default: auto)
```

### Search Command

```
jrnlg search <terms...> [options]

Search terms can be:
  #tag         Search by tag
  @mention     Search by mention
  keyword      Search by keyword

Multiple terms use AND logic (all must match).

Options:
  Same as list command
```

### Tags Command

```
jrnlg tags [command] [options]

Commands:
  (none), list            List all tags with usage counts
  rename OLD NEW          Rename tag across all entries (case-insensitive)

List Options:
  --orphaned              Show only tags used once

Rename Options:
  --dry-run               Preview changes without applying
  --force                 Skip confirmation prompt

Examples:
  jrnlg tags                              # List all tags
  jrnlg tags list --orphaned              # Show tags used only once
  jrnlg tags rename code_review code-review   # Merge tag variations
  jrnlg tags rename old new --dry-run     # Preview changes
  jrnlg tags rename old new --force       # Skip confirmation

Note: Rename is case-insensitive. "code_review" matches #code_review, 
#Code_Review, #CODE_REVIEW, etc.
```

### Mentions Command

```
jrnlg mentions [command] [options]

Commands:
  (none), list            List all mentions with usage counts
  rename OLD NEW          Rename mention across all entries (case-insensitive)

List Options:
  --orphaned              Show only mentions used once

Rename Options:
  --dry-run               Preview changes without applying
  --force                 Skip confirmation prompt

Examples:
  jrnlg mentions                          # List all mentions
  jrnlg mentions list --orphaned          # Show mentions used only once
  jrnlg mentions rename john_doe john-smith  # Rename mention
  jrnlg mentions rename old new --dry-run # Preview changes

Note: Rename is case-insensitive and matches all variations.
```

## Examples

### Daily Journaling

```bash
# Morning journal
jrnlg
# Write: "Starting the day with #planning and coffee ☕"

# Evening reflection
jrnlg
# Write: "Completed #work tasks. Meeting with @team was productive."

# Review today's entries
jrnlg list --from today --summary
```

### Project Tracking

```bash
# Log project work
jrnlg
# Write: "Made progress on #project-alpha. Fixed bug in @authentication module."

# Review all project entries
jrnlg search '#project-alpha' --from "1 week ago"

# Find specific person's mentions
jrnlg search '@alice' '#project-alpha'
```

### Weekly Reviews

```bash
# See last week's work
jrnlg list --from "1 week ago" --to yesterday --summary

# Export to JSON for analysis
jrnlg list --from "1 week ago" --format json > weekly_journal.json

# Edit yesterday's entry
jrnlg edit yesterday

# Delete old test entries
jrnlg delete --from "1 month ago" --to "2 weeks ago" --force
```

### Tag Management

```bash
# After months of journaling, you notice inconsistent tags
jrnlg tags
# Output shows:
#   #code-review (15 entries)
#   #code_review (8 entries)
#   #Code_Review (3 entries)

# Merge all variations into one consistent format
jrnlg tags rename code_review code-review
# Shows preview and asks for confirmation
# After confirming: ✓ Updated 11 entries

# Verify the merge
jrnlg tags
# Output now shows:
#   #code-review (26 entries)

# Find tags you only used once (might be typos)
jrnlg tags list --orphaned
# Output:
#   #meetting (1 entry)    # Typo!
#   #wrok (1 entry)        # Another typo!

# Fix the typos
jrnlg tags rename meetting meeting
jrnlg tags rename wrok work

# Preview changes before applying
jrnlg tags rename old-name new-name --dry-run

# Batch rename without confirmation (use carefully!)
jrnlg tags rename temporary-tag permanent-tag --force
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Check test coverage
make test-coverage
```

### Building

```bash
# Build for current platform
make build
```

## Design Decisions

### Why One File Per Entry?

- **Atomic writes**: No corruption risk from partial writes
- **Simple**: No database to maintain or migrate
- **Git-friendly**: Easy to version control and sync
- **Fast**: Direct filesystem access with lazy indexing

### Why UTC for Filenames?

- **Consistency**: Works across timezone changes
- **Sorting**: Lexicographic sort = chronological sort
- **Preservation**: Original timezone saved in entry content

### Why AND Logic for Search?

- **Precision**: Multiple terms narrow results
- **Intuitive**: Matches most user expectations
- **Flexible**: Single terms still work for broad searches

## Troubleshooting

### Editor not opening

Check your environment variables:
```bash
echo $VISUAL
echo $EDITOR
```

Set one if needed:
```bash
export EDITOR=vim
# or
export VISUAL=code
```

### Entries not appearing in search

Search is case-insensitive, but requires exact matches:
- Tags must include `#`: use `'#work'` not `'work'`
- Mentions must include `@`: use `'@alice'` not `'alice'`
- Remember to quote in bash: `'#tag'` not `#tag`

### Timezone issues

Entries are stored with UTC timestamps in filenames but preserve your original timezone in the content header. If you see unexpected times:
- Check your system timezone: `date`
- Verify entry content header shows correct timezone
- Filenames will always show UTC time

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Natural language date parsing powered by [olebedev/when](https://github.com/olebedev/when)
- Inspired by [jrnl](https://jrnl.sh/) and other CLI journaling tools
