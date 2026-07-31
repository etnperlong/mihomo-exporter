# Go Scaffold

A production-ready Go project scaffold with AI agent configuration for an
enhanced/vibe development workflow.

Minimal on purpose: no HTTP framework, no ORM, no DI container, no Dockerfile.
It establishes the engineering baseline — quality gates, reproducible tooling,
CI, and AI development rules — and leaves technology choices to your project.

## Stack

| Tool | Purpose |
| --- | --- |
| **Go 1.26** | Language and toolchain |
| **golangci-lint v2** | Linters and formatters (`gofumpt`, `gci`) |
| **`testing`** | Standard-library test runner |
| **govulncheck** | Reachable-vulnerability scanning |
| **Make** | Single entry point for every command |
| **Git hooks** | Conventional Commits with mandatory emoji |
| **GitHub Actions** | Lint, tidy, race tests, build, commit lint, security |
| **Dependabot** | Weekly updates for modules, tools, and actions |

Tooling lives in an isolated `tools/go.mod`, so linters never pollute your
application's dependency graph. This follows golangci-lint's official
recommendation for `go tool` usage.

## AI Components

- **`AGENTS.md`** — project rules for AI agents: verification gates, Go
  conventions, commit format, boundaries.
- **`.agents/skills/`** — skills managed by [dotagents](https://dotagents.sentry.dev/).
  - Local: `tech-selection`, `verify`, `commit`
  - Remote: 46 skills from [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang)

### tech-selection

Instead of baking a framework into the scaffold, technology choices are
delivered as an AI skill. Each area has a reference file recording the
recommended default, accepted alternatives with verified versions, and what is
rejected by default:

`http-and-api` · `cli` · `configuration` · `data-and-persistence` ·
`observability` · `testing` · `concurrency-and-resilience` · `architecture` ·
`build-and-release`

`AGENTS.md` requires agents to consult it before adding any dependency.

## Quick Start

```bash
make setup    # deps, AI skills, Git hooks
make run
```

Then talk to your AI coding agent:

```text
> Load skill: tech-selection
> I need to add an HTTP API to this service
```

The agent follows the skill's recommendations and the conventions in
`AGENTS.md`.

## Requirements

- Go 1.26 or newer
- Make
- Git
- Node.js (optional) — only for `npx dotagents`, which installs the remote
  agent skills. Everything else works without it.

No global Go tool installation is needed: `make setup` fetches pinned tools
from `tools/go.mod`.

Remote skills are gitignored by design (dotagents manages them), so a fresh
clone installs them via `make setup` or `make skills`.

On **Windows**, clone with symlink support (`git clone -c core.symlinks=true`,
or enable Developer Mode). `.claude/skills` is a symlink into `.agents/skills`;
without it Git writes a plain text stub and Claude Code finds no skills.

## Commands

| Command | Description |
| --- | --- |
| `make help` | List all targets |
| `make setup` | Install deps and skills, enable Git hooks |
| `make skills` | Install AI agent skills from `agents.toml` |
| `make run` | Run the application |
| `make build` | Compile into `bin/` with `-trimpath` |
| `make fmt` | Format and fix imports |
| `make fmt-check` | Report formatting differences |
| `make lint` | Run the linter suite |
| `make lint-fix` | Lint with autofixes |
| `make test` | Run tests |
| `make test-race` | Race detector, shuffled order |
| `make cover` | Coverage summary |
| `make vuln` | Scan for known vulnerabilities |
| `make tidy` | Tidy both modules |
| `make tidy-check` | Fail if modules are untidy |
| `make check` | Full gate: fmt, tidy, lint, race tests, build |
| `make commit-check` | Validate the last commit message |
| `make clean` | Remove `bin/`, coverage files, and the test cache |

## Layout

```text
cmd/scaffold/main.go   entry point: wiring and exit codes only
internal/greeting/     implementation packages
tools/go.mod           isolated module pinning dev tools
.githooks/             commit-msg and pre-commit hooks
.golangci.yml          linter and formatter configuration
Makefile               command contract shared by dev, hooks, and CI
```

`main` keeps a testable `run(ctx, args, stdout) error` and holds no logic.
There is no `pkg/` — `internal/` plus root packages is the Go convention.

## Testing

Table-driven tests with named subtests, parallel where safe, in an external
test package. See `internal/greeting/greeting_test.go`.

```bash
make test        # fast
make test-race   # race detector + shuffle
make cover
```

No coverage threshold is enforced. A fixed percentage encourages tests that
raise the number without protecting behaviour.

## CI

`.github/workflows/ci.yml` runs on every push and pull request to `main`:

- **Quality** — formatting, module tidiness, lint, race tests, build
- **Commit messages** — validates every commit in a PR with the same hook used locally
- **Security** — `govulncheck`

CI invokes `make` targets, so local and CI behaviour cannot drift.

## Commit Convention

[Conventional Commits](https://www.conventionalcommits.org/) with a mandatory
emoji prefix, enforced locally by `.githooks/commit-msg` and re-checked in CI.

```text
✨ feat: add new feature
🐛 fix(greeting): handle empty input
🔨 refactor(cmd)!: drop the legacy flag
```

| Type | Emoji | Type | Emoji |
| --- | --- | --- | --- |
| feat | ✨ | build | 📦 |
| fix | 🐛 | ci | 👷 |
| docs | 📚 | chore | 🔧 |
| style | 🎨 | revert | ⏪ |
| refactor | 🔨 | perf | 📈 |
| test | 🧪 | | |

The emoji must match the type. Description is imperative, lowercase, no
trailing period, header at most 100 characters. Dependabot's `deps:` /
`deps-dev:` and Git-generated `Merge`/`Revert`/`fixup!` headers are exempt.

Hooks are enabled by `make setup`, which sets `core.hooksPath` to `.githooks`.

## Using this template

Click **Use this template** on GitHub, or clone this repository, then:

1. **Module path** — replace `github.com/etnperlong/go-scaffold` in `go.mod`,
   `tools/go.mod`, and the two imports (`cmd/scaffold/main.go`,
   `internal/greeting/greeting_test.go`):
   ```bash
   grep -rl 'etnperlong/go-scaffold' --exclude-dir=.git . \
     | xargs sed -i 's|github.com/etnperlong/go-scaffold|your.host/you/yourapp|g'
   ```
2. **Binary name** — rename the `cmd/scaffold` directory, then update `BINARY`
   in the `Makefile` and `appName` in `main.go`. `appName` is the only place the
   program name is hardcoded, so error and usage output follow automatically.
3. **First package** — replace `internal/greeting` with your own.
4. **License** — update the copyright holder and year in `LICENSE`.
5. **Verify** — run `make setup`, then `make check`.

`.agents/skills/tech-selection/` cites this sample layout in its examples.
The guidance stays valid after renaming, but you may want to refresh the paths
once your own structure settles.

## License

MIT
