// Command scaffold greets a name. Rename appName when reusing this template.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/etnperlong/go-scaffold/internal/greeting"
)

// appName labels error and usage output. It is the only place the program name
// is hardcoded, so renaming this binary means changing exactly one line.
const appName = "scaffold"

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
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)
	fs.SetOutput(stdout)
	name := fs.String("name", greeting.DefaultName, "name to greet")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancelled before start: %w", err)
	}

	if _, err := fmt.Fprintln(stdout, greeting.Greet(*name)); err != nil {
		return fmt.Errorf("write greeting: %w", err)
	}

	return nil
}
