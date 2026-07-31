package main

import (
	"context"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args    []string
		wantErr bool
	}{
		"no arguments":         {args: nil},
		"unexpected arguments": {args: []string{"-nope"}, wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var out strings.Builder
			err := run(context.Background(), tt.args, &out)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("run(%q) = nil error, want error", tt.args)
				}

				return
			}

			if err != nil {
				t.Fatalf("run(%q) returned unexpected error: %v", tt.args, err)
			}

			if got := out.String(); got != "" {
				t.Errorf("run(%q) wrote %q, want empty output", tt.args, got)
			}
		})
	}
}

func TestRunCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out strings.Builder
	if err := run(ctx, nil, &out); err == nil {
		t.Fatal("run with cancelled context = nil error, want error")
	}
}
