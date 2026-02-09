package cli

import (
	"testing"

	"github.com/jashort/jrnlg/internal"
)

func TestApp_Run_HelpFlags(t *testing.T) {
	tmpDir := t.TempDir()
	config := &internal.Config{
		StoragePath: tmpDir,
	}
	storage := internal.NewFileSystemStorage(tmpDir, nil)
	app := NewApp(storage, config)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "global --help",
			args: []string{"--help"},
		},
		{
			name: "global -h",
			args: []string{"-h"},
		},
		{
			name: "list --help",
			args: []string{"list", "--help"},
		},
		{
			name: "list -h",
			args: []string{"list", "-h"},
		},
		{
			name: "search --help",
			args: []string{"search", "--help"},
		},
		{
			name: "search -h",
			args: []string{"search", "-h"},
		},
		// NEW: Test help flag with other flags (this was broken before)
		{
			name: "list with -from and --help",
			args: []string{"list", "-from", "today", "--help"},
		},
		{
			name: "search with terms and --help",
			args: []string{"search", "#work", "@alice", "--help"},
		},
		{
			name: "list with multiple flags and -h",
			args: []string{"list", "-from", "yesterday", "-to", "today", "-h"},
		},
		{
			name: "help flag in middle of args",
			args: []string{"list", "-from", "today", "--help", "-to", "tomorrow"},
		},
		{
			name: "implicit search with --help",
			args: []string{"#work", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.Run(tt.args)
			if err != nil {
				t.Errorf("Run(%v) returned error: %v, expected nil (help should not error)", tt.args, err)
			}
		})
	}
}

func TestApp_Run_VersionFlags(t *testing.T) {
	tmpDir := t.TempDir()
	config := &internal.Config{
		StoragePath: tmpDir,
	}
	storage := internal.NewFileSystemStorage(tmpDir, nil)
	app := NewApp(storage, config)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "--version",
			args: []string{"--version"},
		},
		{
			name: "-v",
			args: []string{"-v"},
		},
		// NEW: Test version flag with other flags
		{
			name: "list with --version",
			args: []string{"list", "--version"},
		},
		{
			name: "list with -from and -v",
			args: []string{"list", "-from", "today", "-v"},
		},
		{
			name: "search with terms and --version",
			args: []string{"search", "#work", "--version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.Run(tt.args)
			if err != nil {
				t.Errorf("Run(%v) returned error: %v, expected nil (version should not error)", tt.args, err)
			}
		})
	}
}
