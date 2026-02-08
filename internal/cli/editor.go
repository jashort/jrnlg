package cli

import (
	"errors"
	"os"
	"os/exec"
)

// OpenEditor opens an editor with the given initial content and returns the edited content
func OpenEditor(initialContent string) (string, error) {
	// 1. Get editor command
	editor := getEditorCommand()

	// 2. Create temp file
	tmpFile, err := os.CreateTemp("", "jrnlg-*.md")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// 3. Write initial content
	if _, err := tmpFile.Write([]byte(initialContent)); err != nil {
		err := errors.Join(err, tmpFile.Close())
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	// 4. Launch editor
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// 5. Read edited content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// getEditorCommand returns the editor to use
// Priority: VISUAL > EDITOR > vim (with fallbacks to vi, nano)
func getEditorCommand() string {
	// Try VISUAL environment variable first
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// Try EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Try fallbacks in order: vim → vi → nano
	for _, cmd := range []string{"vim", "vi", "nano"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}

	// Last resort (will fail if vim not available)
	return "vim"
}
