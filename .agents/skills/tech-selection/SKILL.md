---
name: tech-selection
description: Choose libraries and architecture for this Go project. Use before adding any third-party dependency, when starting a new feature area (HTTP API, CLI, database, config, logging, observability, messaging, testing), or when the user asks what library to use, how to structure a service, or which framework to pick.
---

# Tech Selection

This scaffold ships no framework on purpose. It is a minimal, runnable baseline.
Technology choices belong to the project being built, not to the scaffold.

This skill records the **recommended defaults** so each new feature area does
not restart the same debate, and so an agent never silently invents a stack.

## Non-negotiable rules

1. **Standard library first.** Add a dependency only when the standard library
   genuinely cannot do the job, or when doing it by hand would be materially
   worse. `net/http`, `log/slog`, `database/sql`, `encoding/json`, `context`,
   `errors`, and `testing` cover a large share of real work.
2. **Never add a dependency silently.** State what you are adding, which
   alternative you rejected, and why. Get the user's agreement for anything
   that shapes architecture (HTTP router, DI, ORM, messaging).
3. **Verify the version.** Resolve the current version before writing it into
   `go.mod`. Never guess a version number from memory.
   ```bash
   go list -m -versions <module>          # available versions
   go get <module>@latest                 # resolve and record
   ```
4. **One concern, one library.** Do not add two routers, two logging
   libraries, or two assertion libraries to the same module.
5. **Prefer boring and maintained.** Check the last release date and open
   issue trend before adopting anything. An unmaintained library is a future
   migration.
6. **A framework is not a substitute for design.** Decide package boundaries
   first; pick libraries second.

## Workflow

1. Identify the feature area (HTTP, CLI, config, database, logging,
   observability, testing, messaging, resilience).
2. Read the matching reference file below. It contains the recommended default
   and the accepted alternatives.
3. If the standard library suffices, stop. Do not add a dependency.
4. If a dependency is needed:
   - confirm the current version with `go list -m -versions`;
   - add it with `go get`;
   - run `make tidy`;
   - state the decision and the rejected alternative in your summary.
5. If the area has no reference file, apply the rules above and say
   explicitly that you are making a judgement call.

## References

Read only the file matching the current task.

| Area | Reference |
| --- | --- |
| HTTP servers, routing, middleware, REST APIs | [references/http-and-api.md](references/http-and-api.md) |
| CLI tools, flags, subcommands | [references/cli.md](references/cli.md) |
| Configuration and secrets | [references/configuration.md](references/configuration.md) |
| Databases, SQL, migrations, caching | [references/data-and-persistence.md](references/data-and-persistence.md) |
| Logging, metrics, tracing, errors | [references/observability.md](references/observability.md) |
| Testing, mocking, fixtures, integration tests | [references/testing.md](references/testing.md) |
| Concurrency, resilience, background work | [references/concurrency-and-resilience.md](references/concurrency-and-resilience.md) |
| Package layout, dependency injection, service structure | [references/architecture.md](references/architecture.md) |
| Build, release, containers, deployment | [references/build-and-release.md](references/build-and-release.md) |

## Deeper Go expertise

The `samber/cc-skills-golang` skills are installed in this project and are the
preferred source for idiomatic Go detail. Load the relevant one alongside this
skill:

- `golang-popular-libraries` — ecosystem survey
- `golang-project-layout` — module and package structure
- `golang-error-handling`, `golang-concurrency`, `golang-context`
- `golang-testing`, `golang-benchmark`, `golang-performance`
- `golang-security`, `golang-safety`, `golang-observability`
- `golang-database`, `golang-cli`, `golang-grpc`, `golang-graphql`
- `golang-dependency-injection`, `golang-design-patterns`
- `golang-spf13-cobra`, `golang-spf13-viper`, `golang-stretchr-testify`
- `golang-samber-lo`, `golang-samber-do`, `golang-samber-oops`, `golang-samber-slog`
- `golang-stay-updated` — verify current ecosystem state

This skill decides **what to adopt**. Those skills explain **how to use it well**.

## Reporting format

When a technology decision is made, report it as:

```
Decision: <library or standard-library approach>
Area:     <HTTP / CLI / database / ...>
Version:  <resolved version, or n/a for stdlib>
Why:      <one or two sentences>
Rejected: <alternative> because <reason>
```
