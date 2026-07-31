# Databases and Persistence

Versions verified against the Go module proxy on 2026-07-31.
Re-resolve with `go list -m -versions <module>` before adding one.

## Principles

1. `database/sql` plus a driver is the baseline. It handles pooling,
   transactions, and context cancellation.
2. Always pass `context.Context` — use `QueryContext`, `ExecContext`,
   `BeginTx`. A query with no context cannot be cancelled or timed out.
3. **Never build SQL with string concatenation or `fmt.Sprintf`.** Use
   placeholders. This is the single most common source of SQL injection.
4. Close everything: `rows.Close()`, and always check `rows.Err()` after the
   iteration loop. A missed `rows.Err()` silently hides truncated results.
5. Configure the pool explicitly: `SetMaxOpenConns`, `SetMaxIdleConns`,
   `SetConnMaxLifetime`. Defaults are not production settings.
6. An ORM is a decision about the whole codebase. Do not introduce one to
   avoid writing a single query.

## Drivers and access layers

| Library | Version | When to choose it |
| --- | --- | --- |
| `database/sql` | stdlib | **Default** access layer. |
| `github.com/jackc/pgx/v5` | v5.10.0 | **Recommended for PostgreSQL.** Best-in-class driver; usable through `database/sql` or its faster native API. |
| `github.com/jmoiron/sqlx` | v1.4.0 | Thin `database/sql` extension for struct scanning. Low risk, no ORM semantics. Note: last release 2024-04-15. |
| `github.com/sqlc-dev/sqlc` | v1.31.1 | **Preferred typed approach.** Generates Go from SQL, so queries stay reviewable SQL. A build-time tool, not a runtime dependency. |
| `entgo.io/ent` | v0.14.6 | Schema-as-code with a strong typed graph API. Real ORM commitment. |
| `gorm.io/gorm` | v1.31.2 | Only when the team explicitly wants a full ORM and accepts its query semantics and reflection cost. |

Recommended order: `database/sql` + `pgx` for simple access, add `sqlc` when
queries grow numerous, and only adopt Ent or GORM as a deliberate, stated
architectural decision.

## Migrations

| Library | Version | Notes |
| --- | --- | --- |
| `github.com/pressly/goose/v3` | v3.27.3 | Migrations as SQL or Go, embeddable. Actively released. |
| `github.com/golang-migrate/migrate/v4` | v4.19.1 | Widely used, many database backends. |

Rules: migrations are versioned and committed; they run as an explicit
deploy step, never implicitly on application start in production; every
migration must be reviewed for whether it locks a large table.

## Caching

| Library | Version | Notes |
| --- | --- | --- |
| in-process `map` + `sync.RWMutex` | stdlib | Fine for small, bounded caches. |
| `github.com/redis/go-redis/v9` | v9.21.0 | Shared cache across instances. |

An unbounded in-process cache is a memory leak. Bound the size or set a TTL.

## Transactions

Keep transactions short. Do not perform network calls inside one. Always
guarantee rollback:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin transaction: %w", err)
}
defer func() { _ = tx.Rollback() }() // no-op after a successful commit

// ... work ...

if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit transaction: %w", err)
}
```

## Testing

Prefer a real database over a mocked driver — SQL behaviour is the thing worth
testing. Use `github.com/testcontainers/testcontainers-go` (v0.43.0) and guard
those tests with `testing.Short()` so unit runs stay fast.

## Rejected by default

- String-concatenated SQL. No exceptions.
- Adding both `sqlx` and an ORM.
- Auto-migrating on application startup in production.
- A global `*sql.DB` package variable; inject it instead.
