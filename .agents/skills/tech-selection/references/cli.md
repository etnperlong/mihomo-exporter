# Command-Line Interfaces

Versions verified against the Go module proxy on 2026-07-31.
Re-resolve with `go list -m -versions <module>` before adding one.

## Start with the standard library

This scaffold uses `flag` in `cmd/scaffold/main.go`. For a tool with a handful
of options and no subcommands, `flag` is sufficient and adds no dependency.

Keep the pattern already in the scaffold:

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
        fmt.Fprintf(os.Stderr, "scaffold: %v\n", err)
        os.Exit(1)
    }
}
```

`run` takes args and an `io.Writer` so it stays testable, and `main` holds no
logic. Keep this shape even after adopting a CLI library.

## Libraries

| Library | Version | When to choose it |
| --- | --- | --- |
| `flag` | stdlib | **Default.** Single-purpose tool, few flags, no subcommands. |
| `github.com/spf13/cobra` | v1.10.2 | Nested subcommands, generated help, shell completion, man pages. The de facto standard for large CLIs. |
| `github.com/urfave/cli/v3` | v3.10.1 | Subcommands with less scaffolding than Cobra. |
| `github.com/alecthomas/kong` | v1.16.0 | Prefer declaring the CLI as a struct with tags rather than building a command tree. |

Adopt Cobra when the tool genuinely has subcommands (`app create`, `app list`).
Do not adopt it for a single-command binary; it adds a large dependency and a
lot of boilerplate for no benefit.

If you adopt Cobra, read the installed `golang-spf13-cobra` skill.

## Exit codes and output

- Human-readable errors go to stderr; primary output goes to stdout.
- Exit non-zero on failure. `os.Exit(1)` is fine for a general error.
- Never call `os.Exit` from library code or from deep inside a call tree; return
  an error and let `main` decide.
- Do not use `log.Fatal` outside `main`. It bypasses deferred cleanup.
- Support `--json` output only when a machine actually consumes it.

## Interactive output

Only add a TUI or spinner library when the tool is genuinely interactive.
Detect whether stdout is a terminal before emitting colour or progress
animation; CI logs and pipes must stay readable.

## Rejected by default

- Cobra plus Viper in a tool that reads no configuration file.
- A TUI framework for output that is one line of text.
- Colour output with no terminal detection.
