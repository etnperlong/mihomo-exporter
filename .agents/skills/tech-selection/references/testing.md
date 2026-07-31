# Testing

Versions verified against the Go module proxy on 2026-07-31.
Re-resolve with `go list -m -versions <module>` before adding one.

## Baseline

The standard `testing` package is the default. This scaffold uses table-driven
tests with no assertion library — see `internal/greeting/greeting_test.go`.

Commands (always via the Makefile):

```bash
make test        # fast feedback
make test-race   # race detector + shuffled order
make cover       # coverage summary
```

## Libraries

| Library | Version | When to choose it |
| --- | --- | --- |
| `testing` | stdlib | **Default.** |
| `github.com/google/go-cmp` | v0.7.0 | **Preferred for deep comparison.** Readable diffs; `cmp.Diff` beats hand-written struct comparison. |
| `github.com/stretchr/testify` | v1.11.1 | Team wants `require`/`assert` ergonomics across a large suite. Read `golang-stretchr-testify`. |
| `go.uber.org/mock` | v0.6.0 | Generated mocks for an interface at a real external boundary. |
| `github.com/testcontainers/testcontainers-go` | v0.43.0 | Integration tests needing a real database, broker, or cache. |

Preference order: stdlib, then `go-cmp` for comparisons. Testify is acceptable
but should be a deliberate, project-wide choice — do not mix testify assertions
and plain `if got != want` in the same package.

## Conventions used by this scaffold

- **Table-driven** tests with a `map[string]struct{...}` of named cases; the
  name becomes the subtest name.
- `t.Parallel()` in both the parent and each subtest when there is no shared
  mutable state.
- **External test package** (`package greeting_test`) for library packages, so
  tests exercise the exported API. Use the internal package only when testing
  an unexported function is genuinely necessary — as `cmd/scaffold` does for
  `run`.
- Failure messages state input, got, and want:
  `t.Errorf("Greet(%q) = %q, want %q", in, got, want)`.
- `t.Helper()` in any assertion helper.
- `t.Cleanup` for teardown, not `defer` in the test body.
- Use `t.Context()`, `t.TempDir()`, and `t.Setenv()` rather than hand-rolling
  equivalents.

## What to test

Test exported behaviour, not implementation details. Cover the happy path,
boundaries (empty, zero, single element, maximum), and error paths.

Verify wrapped errors with `errors.Is`/`errors.As`, never by string match.

## Mocking

Mock only at genuine boundaries: network, filesystem, clock, randomness,
external APIs. Do not mock your own domain types — construct them.

Define the interface in the **consumer** package and keep it narrow. A
one-method interface is easy to fake by hand and often needs no generator.

## Fuzzing and benchmarks

Add a fuzz target for parsers, decoders, and anything handling untrusted input.
Add a benchmark only when there is a performance question to answer; use
`for b.Loop()`. Read `golang-benchmark` before optimising.

## Coverage

Coverage is a diagnostic, not a target. This scaffold sets no global threshold
on purpose — a fixed percentage encourages tests that raise the number without
protecting behaviour. Look for untested error paths instead.

## Rejected by default

- Asserting on log output instead of behaviour.
- `time.Sleep` for synchronisation. Use channels, `sync.WaitGroup`, or
  `synctest`.
- Tests that depend on execution order — `make test-race` shuffles them.
- Mocking the database instead of testing against a real one.
- A test that reaches the public network.
