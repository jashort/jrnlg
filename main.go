package main

import (
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	storage := internal.NewFileSystemStorage(config.StoragePath, config)

	// Create CLI app
	app := cli.NewApp(storage, config)
	app.SetVersion(version)

	// Run with command-line arguments
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
