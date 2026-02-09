package patterns

import "regexp"

// Shared regex patterns used across the application

var (
	// Tag matches: #letter followed by alphanumeric/underscore/hyphen
	// Hyphens are preserved as part of the tag
	// Example: #work, #machine-learning, #project_alpha
	Tag = regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9_-]*)`)

	// Mention matches: @letter followed by alphanumeric/underscore
	// The @ must not be preceded by alphanumeric (excludes emails)
	// Example: @alice, @bob_smith (but not bob@example.com)
	Mention = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])@([a-zA-Z][a-zA-Z0-9_]*)`)
)
