package cli

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/jashort/jrnlg/internal/cli/color"
)

// CLI defines the command-line interface structure
type CLI struct {
	// Global flags
	Color   string           `enum:"auto,always,never" default:"auto" help:"Color mode"`
	Version kong.VersionFlag `short:"v" help:"Show version"`

	// Commands
	Add      AddCmd      `cmd:"" aliases:"create" help:"Add new journal entry"`
	Search   SearchCmd   `cmd:"" help:"Search journal entries"`
	List     SearchCmd   `cmd:"" hidden:"" help:"Alias for search"`
	Edit     EditCmd     `cmd:"" help:"Edit an entry"`
	Delete   DeleteCmd   `cmd:"" aliases:"rm" help:"Delete entries"`
	Tags     TagsCmd     `cmd:"" help:"Manage tags"`
	Mentions MentionsCmd `cmd:"" help:"Manage mentions"`
	Stats    StatsCmd    `cmd:"" help:"Show journal statistics"`
}

// AddCmd creates a new journal entry
type AddCmd struct {
	Message []string `arg:"" optional:"" help:"Entry message (opens editor if not provided)"`
}

// SearchCmd searches journal entries
type SearchCmd struct {
	Terms   []string     `arg:"" optional:"" help:"Search terms (#tag, @mention, keyword)"`
	From    *NaturalDate `help:"Show entries from this date onwards"`
	To      *NaturalDate `help:"Show entries up to this date"`
	Limit   int          `short:"n" help:"Limit number of results"`
	Offset  int          `help:"Skip first N results"`
	Reverse bool         `short:"r" help:"Show newest entries first"`
	Summary bool         `help:"Show compact summary format"`
	Format  string       `enum:"full,summary,json" default:"full" help:"Output format"`
}

// EditCmd edits an entry
type EditCmd struct {
	Selector string `arg:"" optional:"" help:"Entry selector (timestamp or date like 'yesterday')"`
}

// DeleteCmd deletes entries
type DeleteCmd struct {
	Selector string       `arg:"" optional:"" help:"Entry to delete (timestamp)"`
	From     *NaturalDate `help:"Delete entries from this date"`
	To       *NaturalDate `help:"Delete entries up to this date"`
	Force    bool         `short:"f" help:"Skip confirmation"`
}

// TagsCmd manages tags
type TagsCmd struct {
	List   TagsListCmd   `cmd:"" default:"1" help:"List all tags"`
	Rename TagsRenameCmd `cmd:"" help:"Rename a tag"`
}

// TagsListCmd lists all tags
type TagsListCmd struct {
	Orphaned bool `help:"Show only tags used once"`
}

// TagsRenameCmd renames a tag
type TagsRenameCmd struct {
	Old    string `arg:"" help:"Old tag name"`
	New    string `arg:"" help:"New tag name"`
	DryRun bool   `help:"Preview changes without applying"`
	Force  bool   `short:"f" help:"Skip confirmation"`
}

// MentionsCmd manages mentions
type MentionsCmd struct {
	List   MentionsListCmd   `cmd:"" default:"1" help:"List all mentions"`
	Rename MentionsRenameCmd `cmd:"" help:"Rename a mention"`
}

// MentionsListCmd lists all mentions
type MentionsListCmd struct {
	Orphaned bool `help:"Show only mentions used once"`
}

// MentionsRenameCmd renames a mention
type MentionsRenameCmd struct {
	Old    string `arg:"" help:"Old mention name"`
	New    string `arg:"" help:"New mention name"`
	DryRun bool   `help:"Preview changes without applying"`
	Force  bool   `short:"f" help:"Skip confirmation"`
}

// StatsCmd shows journal statistics
type StatsCmd struct {
	All      bool         `help:"Show all-time statistics"`
	From     *NaturalDate `help:"Start date (e.g., 'yesterday', '30 days ago', '2024-01-01')"`
	To       *NaturalDate `help:"End date"`
	Tag      string       `help:"Filter by tag" xor:"filter"`
	Mention  string       `help:"Filter by mention" xor:"filter"`
	Format   string       `enum:"default,json,detailed" default:"default" help:"Output format"`
	Detailed bool         `help:"Show detailed breakdown"`
}

// Run implementations for each command

func (c *AddCmd) Run(ctx *Context) error {
	// If message provided, check if it's all whitespace
	if len(c.Message) > 0 {
		msg := strings.Join(c.Message, " ")
		// If the message is only whitespace, open editor instead
		if strings.TrimSpace(msg) == "" {
			return ctx.App.CreateEntry()
		}
		return ctx.App.CreateEntryWithMessage(msg)
	}
	return ctx.App.CreateEntry()
}

func (c *SearchCmd) Run(ctx *Context) error {
	// Convert Kong struct to SearchArgs
	colorMode, err := color.ParseMode(ctx.CLI.Color)
	if err != nil {
		return err
	}

	args := SearchArgs{
		Tags:      []string{},
		Mentions:  []string{},
		Keywords:  []string{},
		FromDate:  c.From.Ptr(),
		ToDate:    c.To.Ptr(),
		Limit:     c.Limit,
		Offset:    c.Offset,
		Format:    c.Format,
		Reverse:   c.Reverse,
		ColorMode: colorMode,
	}

	if c.Summary {
		args.Format = "summary"
	}

	// Parse search terms
	for _, term := range c.Terms {
		if strings.HasPrefix(term, "#") {
			args.Tags = append(args.Tags, strings.ToLower(term[1:]))
		} else if strings.HasPrefix(term, "@") {
			args.Mentions = append(args.Mentions, strings.ToLower(term[1:]))
		} else {
			args.Keywords = append(args.Keywords, term)
		}
	}

	return ctx.App.executeSearch(args)
}

func (c *EditCmd) Run(ctx *Context) error {
	return ctx.App.executeEdit(c.Selector)
}

func (c *DeleteCmd) Run(ctx *Context) error {
	return ctx.App.executeDelete(c.Selector, c.From.Ptr(), c.To.Ptr(), c.Force)
}

func (c *TagsListCmd) Run(ctx *Context) error {
	return ctx.App.listTags(c.Orphaned)
}

func (c *TagsRenameCmd) Run(ctx *Context) error {
	return ctx.App.renameTags(c.Old, c.New, c.DryRun, c.Force)
}

func (c *MentionsListCmd) Run(ctx *Context) error {
	return ctx.App.listMentions(c.Orphaned)
}

func (c *MentionsRenameCmd) Run(ctx *Context) error {
	return ctx.App.renameMentions(c.Old, c.New, c.DryRun, c.Force)
}

func (c *StatsCmd) Run(ctx *Context) error {
	// Apply detailed flag
	format := c.Format
	if c.Detailed {
		format = "detailed"
	}

	// Validate flag combinations (xor should handle tag/mention, but check all/from/to)
	if c.All && c.From != nil {
		return fmt.Errorf("cannot use --all with --from")
	}
	if c.All && c.To != nil {
		return fmt.Errorf("cannot use --all with --to")
	}

	opts := &statsOptions{
		All:      c.All,
		FromDate: c.From.Ptr(),
		ToDate:   c.To.Ptr(),
		Tag:      strings.ToLower(strings.TrimPrefix(c.Tag, "#")),
		Mention:  strings.ToLower(strings.TrimPrefix(c.Mention, "@")),
		Format:   format,
		Detailed: c.Detailed,
	}

	return ctx.App.executeStats(opts)
}

// Context provides access to CLI and App for command execution
type Context struct {
	CLI *CLI
	App *App
}
