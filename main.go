package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"

	"github.com/jashort/jrnlg/internal"
	"github.com/jashort/jrnlg/internal/cli"
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

	// Handle no-args case: default to add command
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"add"}
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
		kong.Vars{
			"version": buildVersionString(),
		},
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

func buildVersionString() string {
	var build, commitTime string
	if info, ok := debug.ReadBuildInfo(); ok {
		build = info.Main.Version
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				commitTime = setting.Value
			}
		}
	}
	return fmt.Sprintf("jrnlg %s (Commit Timestamp: %s)", build, commitTime)
}
