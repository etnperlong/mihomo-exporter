// Package greeting builds human-readable greetings.
package greeting

import (
	"fmt"
	"strings"
)

// DefaultName is used when a greeting is requested without a name.
const DefaultName = "World"

// Greet returns a greeting for name. Blank or whitespace-only names fall back
// to DefaultName so callers never produce an empty salutation.
func Greet(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = DefaultName
	}

	return fmt.Sprintf("Hello, %s!", trimmed)
}
