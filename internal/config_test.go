package internal

import (
	"reflect"
	"testing"
)

func TestParseEditorArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single argument",
			input:    "+startinsert",
			expected: []string{"+startinsert"},
		},
		{
			name:     "multiple arguments",
			input:    "+startinsert +call",
			expected: []string{"+startinsert", "+call"},
		},
		{
			name:     "arguments with spaces",
			input:    "+startinsert '+call cursor(3,1)'",
			expected: []string{"+startinsert", "+call cursor(3,1)"},
		},
		{
			name:     "double quoted arguments",
			input:    "+startinsert \"+call cursor(3,1)\"",
			expected: []string{"+startinsert", "+call cursor(3,1)"},
		},
		{
			name:     "multiple spaces between args",
			input:    "+startinsert    +call",
			expected: []string{"+startinsert", "+call"},
		},
		{
			name:     "complex vim arguments",
			input:    "+startinsert '+call cursor(3,1)' +set nocompatible",
			expected: []string{"+startinsert", "+call cursor(3,1)", "+set", "nocompatible"},
		},
		{
			name:     "mixed quotes",
			input:    "+start \"+call func()\" '+another command'",
			expected: []string{"+start", "+call func()", "+another command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEditorArgs(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseEditorArgs(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadConfig_EditorArgs(t *testing.T) {
	// Test with editor args set
	t.Setenv("JRNLG_EDITOR_ARGS", "+startinsert '+call cursor(3,1)'")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	expectedArgs := []string{"+startinsert", "+call cursor(3,1)"}
	if !reflect.DeepEqual(config.EditorArgs, expectedArgs) {
		t.Errorf("EditorArgs = %v, want %v", config.EditorArgs, expectedArgs)
	}
}

func TestLoadConfig_NoEditorArgs(t *testing.T) {
	// Ensure env var is not set
	t.Setenv("JRNLG_EDITOR_ARGS", "")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.EditorArgs) > 0 {
		t.Errorf("EditorArgs should be empty, got %v", config.EditorArgs)
	}
}
