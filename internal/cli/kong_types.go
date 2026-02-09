package cli

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"
)

// NaturalDate wraps time.Time to provide natural language date parsing
type NaturalDate struct {
	time.Time
}

// Decode implements kong.MapperValue to parse dates using our ParseDate function
func (d *NaturalDate) Decode(ctx *kong.DecodeContext) error {
	var str string
	if err := ctx.Scan.PopValueInto("date", &str); err != nil {
		return err
	}

	parsed, err := ParseDate(str)
	if err != nil {
		return fmt.Errorf("invalid date: %w\n\nExamples: today, yesterday, \"3 days ago\", 2024-01-01", err)
	}

	d.Time = parsed
	return nil
}

// Ptr returns a *time.Time for optional date fields
func (d *NaturalDate) Ptr() *time.Time {
	if d == nil {
		return nil
	}
	return &d.Time
}
