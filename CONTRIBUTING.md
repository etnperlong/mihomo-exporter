# Contributing to mihomo-exporter

[简体中文](CONTRIBUTING.zh.md)

Thank you for contributing. Bug reports, documentation corrections, tests, and
focused improvements are all welcome.

## Before You Start

- Search existing issues and pull requests before opening a new one.
- Open an issue first for a large feature or behavior change so its scope can
  be agreed before implementation.
- Keep changes focused. Do not combine unrelated refactors, dependency updates,
  and features in one pull request.

## Development Setup

Requirements: Go 1.26, Make, and Git. Node.js is optional and is only used to
install agent skills.

```bash
make setup
make test
```

All development commands are Make targets. Run `make help` to see the complete
list. Do not substitute ad-hoc command lines for repository checks.

## Making Changes

- Use the standard library before adding a dependency. Discuss new libraries or
  deployment tooling in an issue first.
- Keep `cmd/mihomo-exporter` limited to wiring and exit handling; put behavior
  in focused `internal/` packages.
- Pass dependencies explicitly. Avoid global mutable state.
- Keep Prometheus labels bounded. Do not add per-connection host or destination
  labels.
- Keep `Collect()` free of network and filesystem I/O. It must only emit from
  the latest cached snapshot.
- Never commit credentials, `.env` files, generated binaries, or real Mihomo
  API secrets.

## Testing And Verification

Add or update tests for behavior changes. Use named, table-driven cases where
appropriate; avoid `time.Sleep` in tests.

Before opening a pull request, run:

```bash
make check
make vuln
```

`make check` runs formatting validation, module tidiness, linting, race-enabled
tests, and a build. Do not disable linters or weaken assertions to make a check
pass.

## Commits And Pull Requests

Use Conventional Commits with the required matching emoji prefix. For example:

```text
✨ feat(collector): expose a new bounded metric
🐛 fix(mihomo): handle an API timeout
📚 docs: clarify Docker Compose deployment
```

Keep commit messages imperative, lowercase, and without a trailing period.
Each commit should build independently when practical.

Pull requests should explain the problem, the behavioral change, validation
performed, and any configuration, metric, or deployment impact. Update the
English and Chinese documentation together when user-facing behavior changes.

## Code Of Conduct

Be respectful, constructive, and considerate in issues, pull requests, and
reviews. Report unacceptable behavior privately to the repository maintainer.
