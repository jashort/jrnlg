package color

import (
	"testing"
)

func TestNew_Always(t *testing.T) {
	c := New(Always)
	if !c.Enabled() {
		t.Error("Always mode should enable colors")
	}
}

func TestNew_Never(t *testing.T) {
	c := New(Never)
	if c.Enabled() {
		t.Error("Never mode should disable colors")
	}
}

func TestNew_Auto_NoColor(t *testing.T) {
	// Set NO_COLOR env var
	t.Setenv("NO_COLOR", "1")

	c := New(Auto)
	if c.Enabled() {
		t.Error("Auto mode should respect NO_COLOR env var")
	}
}

func TestColorizer_Timestamp(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		input   string
		want    string
	}{
		{
			name:    "enabled",
			enabled: true,
			input:   "2024-02-09",
			want:    cyan + "2024-02-09" + reset,
		},
		{
			name:    "disabled",
			enabled: false,
			input:   "2024-02-09",
			want:    "2024-02-09",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Colorizer{enabled: tt.enabled}
			got := c.Timestamp(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorizer_Tag(t *testing.T) {
	c := &Colorizer{enabled: true}
	got := c.Tag("#work")
	want := green + "#work" + reset
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Test disabled
	c2 := &Colorizer{enabled: false}
	got2 := c2.Tag("#work")
	if got2 != "#work" {
		t.Errorf("disabled colorizer should return plain text, got %q", got2)
	}
}

func TestColorizer_Mention(t *testing.T) {
	c := &Colorizer{enabled: true}
	got := c.Mention("@alice")
	want := yellow + "@alice" + reset
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Test disabled
	c2 := &Colorizer{enabled: false}
	got2 := c2.Mention("@alice")
	if got2 != "@alice" {
		t.Errorf("disabled colorizer should return plain text, got %q", got2)
	}
}

func TestColorizer_Dim(t *testing.T) {
	c := &Colorizer{enabled: true}
	got := c.Dim("metadata")
	want := gray + "metadata" + reset
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Test disabled
	c2 := &Colorizer{enabled: false}
	got2 := c2.Dim("metadata")
	if got2 != "metadata" {
		t.Errorf("disabled colorizer should return plain text, got %q", got2)
	}
}

func TestColorizer_Bold(t *testing.T) {
	c := &Colorizer{enabled: true}
	got := c.Bold("header")
	want := bold + "header" + reset
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Test disabled
	c2 := &Colorizer{enabled: false}
	got2 := c2.Bold("header")
	if got2 != "header" {
		t.Errorf("disabled colorizer should return plain text, got %q", got2)
	}
}

func TestColorizer_Separator(t *testing.T) {
	c := &Colorizer{enabled: true}
	got := c.Separator("---")
	want := gray + "---" + reset
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		input   string
		want    Mode
		wantErr bool
	}{
		{"auto", Auto, false},
		{"always", Always, false},
		{"never", Never, false},
		{"invalid", Auto, true},
		{"", Auto, true},
		{"AUTO", Auto, true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
