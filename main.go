package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli"
)

// Version information (set via ldflags during build)
var (
	version = "dev"
)

func main() {
	// Load configuration
	config, err := internal.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	storage := internal.NewFileSystemStorage(config.StoragePath, config)

	// Create CLI app
	app := cli.NewApp(storage, config)
	app.SetVersion(version)

	// Handle no-args case: default to create command
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"create"}
	}

	// Handle version flag early (before Kong parsing)
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			_ = app.ShowVersion()
			return
		}
	}

	// Parse CLI with Kong
	var cliStruct cli.CLI
	parser, err := kong.New(&cliStruct,
		kong.Name("jrnlg"),
		kong.Description("A simple, fast journal application"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		parser.Errorf("%s", err)
		os.Exit(1)
	}

	// Run the command
	cmdCtx := &cli.Context{
		CLI: &cliStruct,
		App: app,
	}

	if err := ctx.Run(cmdCtx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
