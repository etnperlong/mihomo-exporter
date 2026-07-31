# Concurrency and Resilience

Versions verified against the Go module proxy on 2026-07-31.
Re-resolve with `go list -m -versions <module>` before adding one.

## Principles

1. Do not add concurrency until there is a measured reason. Sequential code is
   easier to read, test, and debug.
2. **Every goroutine needs a defined exit.** If you cannot say how it stops,
   it is a leak.
3. Pass `context.Context` as the first parameter and honour cancellation.
4. Never store a `context.Context` in a struct field — the `containedctx`
   linter enforces this.
5. Do not use a bare `go func()` for work that must complete or whose error
   matters. Use `errgroup`.
6. Run `make test-race` for any concurrent code. A race that never fails
   locally will fail in production.

## Libraries

| Library | Version | When to choose it |
| --- | --- | --- |
| `sync`, `sync/atomic`, channels | stdlib | **Default.** |
| `golang.org/x/sync/errgroup` | v0.22.0 | **Recommended** for running several tasks and collecting the first error. Also provides `SetLimit` for bounded parallelism. |
| `golang.org/x/sync/semaphore` | v0.22.0 | Weighted concurrency limiting. |
| `golang.org/x/sync/singleflight` | v0.22.0 | Collapse duplicate concurrent work, e.g. cache stampede. |
| `github.com/sourcegraph/conc` | v0.3.0 | Panic-safe pools and structured helpers. Note: last release 2023-02-25; prefer `errgroup`. |
| `github.com/samber/lo` | v1.53.0 | Generic slice/map helpers, including parallel variants. Convenience, not concurrency primitives. Read `golang-samber-lo`. |

Prefer `errgroup`. It is maintained by the Go team, small, and covers most
real needs:

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8) // bound concurrency; unbounded fan-out is a failure mode

for _, item := range items {
    g.Go(func() error {
        return process(ctx, item)
    })
}

if err := g.Wait(); err != nil {
    return fmt.Errorf("process items: %w", err)
}
```

## Resilience

| Library | Version | When |
| --- | --- | --- |
| `github.com/cenkalti/backoff/v5` | v5.0.3 | Retry with exponential backoff and jitter. |
| `github.com/sony/gobreaker/v2` | v2.4.0 | Circuit breaker for a dependency that fails in bursts. |

Rules:

- **Only retry idempotent operations.** Retrying a non-idempotent write causes
  duplicates.
- Always bound retries and always apply jitter. Fixed-interval retries from many
  instances create a thundering herd.
- Never retry a context cancellation or a deterministic 4xx.
- Set a timeout on every outbound call. A missing timeout turns a slow
  dependency into an outage.

## Background work

For periodic work, use `time.Ticker` with `select` on `ctx.Done()`. Do not use
`time.Sleep` in a loop with no cancellation path.

Before adding a job queue or scheduler library, ask whether a ticker plus the
database is enough. Persistent queues are a significant operational commitment.

## Shutdown

Propagate shutdown from `signal.NotifyContext` (already wired in
`cmd/scaffold/main.go`), stop accepting new work, wait for in-flight work with a
bounded timeout, then close resources in reverse order of construction.

## Testing concurrent code

Use `testing/synctest` for deterministic control of time and goroutine
blocking. Never synchronise a test with `time.Sleep`.

## Rejected by default

- Unbounded goroutine fan-out over a slice of unknown size.
- `sync.Mutex` copied by value — pass a pointer. The `copylock` vet check
  catches common cases.
- A goroutine writing to a channel nobody reads.
- Sharing memory where a channel or a mutex-guarded struct would be clearer.
