---
name: verify
description: Run this project's verification gates in the right order. Use after writing or changing Go code, before committing, when a build or test fails, or when the user asks to check, validate, verify, or run tests.
---

# Verify

## Quick Start

Every check is a Makefile target. Never invent ad-hoc `go` command lines —
the Makefile is the contract shared by local development, hooks, and CI.

```bash
make check
```

That runs, in order: `fmt-check`, `tidy-check`, `lint`, `test-race`, `build`.

## Targets

| Command | Purpose |
| --- | --- |
| `make fmt` | Format and fix imports in place |
| `make fmt-check` | Report formatting differences without writing |
| `make lint` | Full golangci-lint suite |
| `make lint-fix` | Lint with available autofixes applied |
| `make test` | Fast test run |
| `make test-race` | Race detector, shuffled order |
| `make cover` | Coverage summary |
| `make vuln` | govulncheck against the vulnerability database |
| `make tidy` | Tidy the main and tool modules |
| `make tidy-check` | Fail if `go.mod`/`go.sum` are untidy |
| `make check` | Everything CI enforces |
| `make build` | Compile into `bin/` |

## Workflow while developing

1. After an edit, run the narrowest useful check first:
   ```bash
   go test ./internal/greeting/     # one package
   go test -run TestGreet ./...     # one test
   ```
2. Before considering the work done, run `make check`.
3. Run `make vuln` when dependencies changed.

## Interpreting failures

**`fmt-check` fails** — run `make fmt`. Never hand-format to satisfy it.

**`lint` fails** — read the linter name in the message and fix the cause.
Do not add `//nolint` to silence a finding. If a suppression is genuinely
correct, it must name the linter and carry an explanation, because
`nolintlint` requires both:
```go
//nolint:gosec // path is validated against an allowlist above
```

**`tidy-check` fails** — run `make tidy` and commit the module changes.

**`test-race` fails but `test` passes** — a real data race. Do not retry hoping
it passes. Read the race report: it names both conflicting accesses.

**A test fails only under `-shuffle=on`** — tests share state or depend on
order. Fix the test, do not disable shuffling.

**`vuln` reports a finding** — upgrade the affected module. Only if no fix
exists, assess whether the vulnerable path is reachable and report it.

## Rules

- Never weaken a check to make it pass: do not disable a linter, lower a
  setting, delete an assertion, or add a blanket exclusion.
- Never commit with a failing gate.
- If a check fails for a reason you cannot fix, stop and report what failed and
  what you tried, instead of routing around it.
- Tool versions are pinned in `tools/go.mod`. Do not upgrade them as a
  side effect of unrelated work.
