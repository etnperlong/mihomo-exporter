# Logging, Metrics, Tracing, and Errors

Versions verified against the Go module proxy on 2026-07-31.
Re-resolve with `go list -m -versions <module>` before adding one.

## Logging

`log/slog` is in the standard library and is the default. Do not add a logging
dependency without a specific reason.

| Library | Version | When to choose it |
| --- | --- | --- |
| `log/slog` | stdlib | **Default.** Structured, levelled, JSON-capable. |
| `github.com/samber/slog-multi` | v1.8.0 | Fan-out, pipelines, or middleware over `slog` handlers while keeping the `slog` API. |
| `github.com/rs/zerolog` | v1.35.1 | Extreme allocation sensitivity in a proven hot path. |
| `go.uber.org/zap` | v1.28.0 | Team standard already exists, or an ecosystem requires it. |

Rules:

- Use structured key/value pairs, never `fmt.Sprintf` into a message.
- JSON in production; text is acceptable locally.
- Pass a `*slog.Logger` explicitly. Avoid `slog.Default()` in library code.
- **Never log secrets, tokens, passwords, or full request bodies.** Implement
  `LogValue()` on types holding sensitive data.
- Log an error once, at the boundary that handles it. Wrapping and returning is
  not logging — logging at every level produces duplicate noise.
- Do not use `log.Fatal` outside `main`; it skips deferred cleanup.

If you adopt any `samber/slog-*` package, read the installed
`golang-samber-slog` skill.

## Errors

Use the standard library: `errors.New`, `fmt.Errorf` with `%w`, `errors.Is`,
`errors.As`, `errors.Join`.

| Approach | Version | Notes |
| --- | --- | --- |
| `errors` + `fmt.Errorf("...: %w", err)` | stdlib | **Default.** |
| `github.com/samber/oops` | v1.23.0 | Want structured error context (tags, stacktrace, domain) carried to the log sink. Read `golang-samber-oops`. |

Rules:

- Wrap with context describing the operation: `fmt.Errorf("load user %d: %w", id, err)`.
- Do not prefix messages with "failed to" — the caller adds its own context.
- Define sentinel errors (`var ErrNotFound = errors.New("not found")`) or typed
  errors for conditions callers must branch on.
- Never compare error strings. Use `errors.Is`/`errors.As`.
- Never discard an error with `_` unless you write why in a comment.

The `err113`, `errorlint`, `errname`, and `wrapcheck` linters in
`.golangci.yml` enforce most of this.

## Metrics and tracing

| Library | Version | When |
| --- | --- | --- |
| `go.opentelemetry.io/otel` | v1.44.0 | **Default** for tracing, and for metrics in a vendor-neutral setup. |
| `github.com/prometheus/client_golang` | v1.24.1 | Prometheus scraping is the established metrics pipeline. |

Rules:

- Do not add instrumentation before there is somewhere to send it.
- **Never put unbounded values in metric labels** (user ID, request ID, raw
  path). This causes cardinality explosion.
- Propagate `context.Context` so spans nest correctly; a span that does not
  receive the parent context produces a broken trace.
- Record errors on the span and set its status.

## Health and readiness

Expose liveness and readiness separately. Readiness should check real
dependencies; liveness should stay cheap and must not fail because a downstream
service is slow.

## Rejected by default

- Two logging libraries in one module.
- `fmt.Println` for application logging.
- Logging and returning the same error at every layer.
- High-cardinality metric labels.
