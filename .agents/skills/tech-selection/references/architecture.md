# Package Layout, Dependency Injection, and Service Structure

Versions verified against the Go module proxy on 2026-07-31.

## Layout this scaffold establishes

```text
cmd/scaffold/main.go        entry point: wiring + exit codes only
internal/greeting/          implementation packages
```

Rules:

- `cmd/<name>/` holds one entry point per binary. Keep it thin: parse config,
  construct dependencies, call into packages, map errors to exit codes.
- `internal/` is the default home for implementation. The compiler prevents
  other modules from importing it, which keeps your API surface deliberate.
- Add a package at the module root or a public subdirectory **only** when you
  intend to support external importers.
- **Do not create `/pkg`.** It is not a Go convention and usually becomes a
  dumping ground. `internal/` plus root packages is enough.
- Do not add empty placeholder directories (`api/`, `configs/`, `deployments/`)
  before there is content for them.

Read the installed `golang-project-layout` skill before restructuring.

## Package design

- Name packages for what they provide: `greeting`, `billing`, `storage`. Never
  `utils`, `helpers`, `common`, `base`, or `misc` — these attract unrelated code.
- Avoid stutter: `greeting.Greet`, not `greeting.GreetingGreet`.
- A package should have one clear responsibility and a small exported surface.
  Unexport anything callers do not need.
- Keep the dependency graph acyclic and pointing inward: transport depends on
  domain, never the reverse.
- Define interfaces in the **consumer** package, not next to the
  implementation. Keep them narrow — one or two methods is typical Go.
- Accept interfaces, return concrete types.

## Dependency injection

Constructor injection with plain functions is the Go default and is sufficient
for the large majority of services:

```go
func NewServer(cfg Config, logger *slog.Logger, store Store) *Server
```

Wire everything explicitly in `main`. This keeps the graph readable and needs no
framework.

| Approach | Version | When to choose it |
| --- | --- | --- |
| manual constructor wiring | stdlib | **Default.** Correct for nearly all services. |
| `github.com/google/wire` | see `golang-google-wire` | Compile-time generated wiring when `main` becomes genuinely unwieldy. No runtime reflection. |
| `github.com/samber/do/v2` | v2.1.0 | Runtime container with generics and lifecycle hooks. Read `golang-samber-do`. |
| `go.uber.org/fx` | see `golang-uber-fx` | Large application already standardised on Fx. Significant framework commitment. |
| `go.uber.org/dig` | see `golang-uber-dig` | Reflection-based container underlying Fx. |

Do not introduce a DI container to avoid writing twenty lines of wiring. The
cost is indirection that makes call paths hard to follow. Prefer `wire` over a
runtime container if you need generation, because errors surface at compile
time. Read `golang-dependency-injection` before deciding.

## Global state

Avoid it. No mutable package-level variables, no `init()` performing I/O or
registration with side effects, no singleton accessor for a database or logger.
Pass dependencies explicitly — it is what makes tests possible.

Package-level `const` and sentinel `var ErrX = errors.New(...)` are fine.

## When to split a service

Do not split a module into multiple services or a multi-module repository until
there is a concrete reason: independent deployment, separate scaling, or a hard
ownership boundary. A single module with clear packages is easier to refactor.

## Rejected by default

- `/pkg`, `utils`, `helpers`, `common`.
- Interfaces with one implementation created "for testability" when a concrete
  type would do.
- A DI framework in a service with fewer than a dozen dependencies.
- Layered directories (`controllers/`, `services/`, `repositories/`) imposed
  before the domain justifies them. Organise by domain, not by technical role.
- Import cycles resolved by creating a shared `types` package.
