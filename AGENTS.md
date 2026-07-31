# AGENTS.md

Project rules for AI coding agents working in this repository.

## Project

A minimal, production-ready Go scaffold. Go 1.26, one module, no framework.

```text
cmd/scaffold/main.go        entry point: wiring and exit codes only
internal/greeting/          implementation packages
tools/go.mod                isolated module pinning dev tools
.githooks/                  commit-msg and pre-commit hooks
.agents/skills/             AI skills (managed by dotagents)
```

## Verification gates

Every check is a Makefile target. **Always use them; never invent ad-hoc `go`
command lines.**

```bash
make check      # fmt-check + tidy-check + lint + test-race + build
make test       # fast tests
make lint       # golangci-lint
make fmt        # format and fix imports
make vuln       # govulncheck
```

Run `make check` before declaring any task complete. Never weaken a check to
make it pass — do not disable a linter, delete an assertion, or add a blanket
exclusion. If something fails and you cannot fix it, stop and report.

Load the `verify` skill for failure-by-failure guidance.

## Technology selection

This scaffold deliberately ships no HTTP framework, CLI framework, ORM,
logger, DI container, or Dockerfile. Those are project decisions.

**Before adding any third-party dependency or starting a new feature area
(HTTP, CLI, config, database, logging, observability, testing, concurrency,
deployment), load the `tech-selection` skill** and read the matching reference.
It records the recommended default and the accepted alternatives.

Non-negotiable:

1. Standard library first. `net/http`, `log/slog`, `database/sql`, `errors`,
   `context`, and `testing` cover most work.
2. Never add a dependency silently. State what you added, what you rejected,
   and why.
3. Verify the version before writing it into `go.mod`:
   `go list -m -versions <module>`. Never guess a version from memory.
4. One concern, one library. Never two routers or two logging libraries.

## Go conventions

- Formatting is `gofumpt` via `make fmt`. Never hand-format.
- Imports are grouped standard / third-party / local, enforced by `gci`.
- Wrap errors with context: `fmt.Errorf("load user %d: %w", id, err)`.
  Do not prefix with "failed to".
- Compare errors with `errors.Is`/`errors.As`, never by string.
- Never discard an error with `_` unless a comment explains why.
- `context.Context` is the first parameter and is always propagated. Never
  store it in a struct field.
- Define interfaces in the consumer package; keep them narrow. Accept
  interfaces, return concrete types.
- Every exported symbol has a doc comment starting with its name. Every
  package has a package comment.
- No global mutable state. No `init()` doing I/O. Pass dependencies explicitly.
- `main` holds no logic: keep the `run(ctx, args, stdout) error` pattern so it
  stays testable.
- Never use `log.Fatal` or `os.Exit` outside `main`.
- Do not create `pkg/`, `utils`, `helpers`, or `common` packages.

## Testing

- Table-driven tests with named cases in a `map[string]struct{...}`.
- `t.Parallel()` in parent and subtests when there is no shared state.
- External test package (`package foo_test`) for library packages.
- Failure messages state input, got, and want.
- Standard `testing` plus `go-cmp` for deep comparison. No assertion library
  unless `tech-selection` says the project adopted one.
- Mock only real boundaries: network, filesystem, clock, randomness.
- Never synchronise tests with `time.Sleep`.

## Commits

Conventional Commits with a **mandatory** emoji prefix, enforced by
`.githooks/commit-msg` and re-checked in CI.

```text
<emoji> <type>[(scope)][!]: <description>
```

| feat ✨ | fix 🐛 | docs 📚 | style 🎨 | refactor 🔨 | perf 📈 |
| test 🧪 | build 📦 | ci 👷 | chore 🔧 | revert ⏪ | |

The emoji must match the type. Description is imperative, lowercase, no
trailing period, header at most 100 characters. Never bypass the hook with
`--no-verify`. Load the `commit` skill for the full workflow.

## Skill usage

Skills are managed by `dotagents` via `agents.toml`. Never copy skills into
`.claude/skills` by hand and never edit `agents.lock`.

Project skills:

| Skill | Load when |
| --- | --- |
| `tech-selection` | Choosing a library or structuring a new area |
| `verify` | Running or debugging the verification gates |
| `commit` | Creating a commit |

The `samber/cc-skills-golang` collection is installed and is the preferred
source for idiomatic Go detail — `golang-error-handling`, `golang-concurrency`,
`golang-context`, `golang-testing`, `golang-project-layout`,
`golang-security`, `golang-performance`, `golang-popular-libraries`, and more.
Load the relevant one alongside the project skill.

## Boundaries

- Do not add features, refactors, or abstractions beyond what was asked.
- Do not add a dependency, framework, or Dockerfile without agreement.
- Do not upgrade pinned tool versions in `tools/go.mod` as a side effect.
- Do not commit secrets, `.env`, credentials, or build artifacts.
- Do not edit generated state: `agents.lock`, `.agents/.gitignore`.
- Prefer editing an existing file over creating a new one.
- Do not create documentation files unless asked.
