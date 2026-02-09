package patterns

import (
	"testing"
)

func TestTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single tag",
			input: "#work",
			want:  []string{"work"},
		},
		{
			name:  "tag with hyphen",
			input: "#machine-learning",
			want:  []string{"machine-learning"},
		},
		{
			name:  "tag with underscore",
			input: "#project_alpha",
			want:  []string{"project_alpha"},
		},
		{
			name:  "multiple tags",
			input: "#work #personal #project",
			want:  []string{"work", "personal", "project"},
		},
		{
			name:  "no tags",
			input: "no tags here",
			want:  []string{},
		},
		{
			name:  "tag in sentence",
			input: "Working on #project today",
			want:  []string{"project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := Tag.FindAllStringSubmatch(tt.input, -1)
			var got []string
			for _, match := range matches {
				if len(match) > 1 {
					got = append(got, match[1])
				}
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %v matches, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("match[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestMention(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single mention",
			input: "@alice",
			want:  []string{"alice"},
		},
		{
			name:  "mention with underscore",
			input: "@bob_smith",
			want:  []string{"bob_smith"},
		},
		{
			name:  "multiple mentions",
			input: "@alice @bob @charlie",
			want:  []string{"alice", "bob", "charlie"},
		},
		{
			name:  "mention in sentence",
			input: "Hi @alice how are you",
			want:  []string{"alice"},
		},
		{
			name:  "email should not match",
			input: "Email bob@example.com",
			want:  []string{},
		},
		{
			name:  "mention after space",
			input: "Meeting with @alice",
			want:  []string{"alice"},
		},
		{
			name:  "no mentions",
			input: "no mentions here",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := Mention.FindAllStringSubmatch(tt.input, -1)
			var got []string
			for _, match := range matches {
				if len(match) > 1 {
					got = append(got, match[1])
				}
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %v matches, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("match[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
