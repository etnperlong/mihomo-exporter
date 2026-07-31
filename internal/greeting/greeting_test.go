package greeting_test

import (
	"testing"

	"github.com/etnperlong/go-scaffold/internal/greeting"
)

func TestGreet(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  string
	}{
		"named input":        {input: "Alice", want: "Hello, Alice!"},
		"empty input":        {input: "", want: "Hello, World!"},
		"whitespace input":   {input: "   ", want: "Hello, World!"},
		"surrounding spaces": {input: "  Bob  ", want: "Hello, Bob!"},
		"unicode input":      {input: "世界", want: "Hello, 世界!"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := greeting.Greet(tt.input); got != tt.want {
				t.Errorf("Greet(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
