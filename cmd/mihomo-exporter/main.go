// Command mihomo-exporter exports Mihomo metrics for Prometheus.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

const appName = "mihomo-exporter"

var errUnexpectedArguments = errors.New("unexpected arguments")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", appName, err)
		os.Exit(1)
	}
}

// run holds the real program so it stays testable and keeps main free of logic.
func run(ctx context.Context, args []string, stdout io.Writer) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancelled before start: %w", err)
	}

	if len(args) != 0 {
		return fmt.Errorf("%w: %v", errUnexpectedArguments, args)
	}

	_ = stdout // Wiring is added with the HTTP server in Phase 6.

	return nil
}
