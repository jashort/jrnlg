# jrnlg

A fast, minimal CLI journal application written in Go. Inspired by [jrnl](https://jrnl.sh/en/stable/)

## Features
- **Fast & Simple**: One Markdown file per entry, no database required
- **Natural Language Dates**: Search using "yesterday", "3 days ago", "last week"
- **Rich Metadata**: Automatic extraction of #tags, @mentions, and timestamps
- **Multiple Output Formats**: Full, summary, or JSON output
- **Timezone Aware**: Preserves original timezone while using UTC for storage
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
jrnlg list -from yesterday
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
## Monday 2024-02-09 2:30 PM UTC

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
jrnlg search '#work' -from 2024-01-01 -to 2024-12-31
```

### Natural Language Date Filters

```bash
# Entries from yesterday
jrnlg list -from yesterday

# Entries from the last 3 days
jrnlg list -from "3 days ago"

# Entries from last week until yesterday
jrnlg list -from "1 week ago" -to yesterday

# ISO 8601 dates also work
jrnlg list -from 2024-01-01 -to 2024-12-31
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

### List Command

```
jrnlg list [options]

Options:
  -n <number>          Limit number of results
  --offset <number>    Skip first N results
  -r                   Reverse order (newest first)
  --summary            Use summary format (one line per entry)
  --format <format>    Output format: full, summary, json (default: full)
  -from <date>         Start date (ISO 8601 or natural language)
  -to <date>           End date (ISO 8601 or natural language)
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
jrnlg list -from today --summary
```

### Project Tracking

```bash
# Log project work
jrnlg
# Write: "Made progress on #project-alpha. Fixed bug in @authentication module."

# Review all project entries
jrnlg search '#project-alpha' -from "1 week ago"

# Find specific person's mentions
jrnlg search '@alice' '#project-alpha'
```

### Weekly Reviews

```bash
# See last week's work
jrnlg list -from "1 week ago" -to yesterday --summary

# Export to JSON for analysis
jrnlg list -from "1 week ago" --format json > weekly_journal.json
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
