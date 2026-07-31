# Build, Release, and Deployment

Versions verified on 2026-07-31 from the GitHub Releases API.

## Building

`go build` is the whole toolchain. `make build` already applies `-trimpath`,
which removes local filesystem paths from the binary.

Since Go 1.24, `go build` stamps VCS revision into the binary automatically;
read it with `runtime/debug.ReadBuildInfo`. Prefer this over hand-passing a
commit hash via `-ldflags`.

To inject a human-facing version string:

```bash
go build -trimpath -ldflags "-s -w -X main.version=$(git describe --tags --always)" ./cmd/scaffold
```

`-s -w` strips debug information and shrinks the binary. Omit it when you need
usable stack traces or a debugger.

Cross-compilation needs no toolchain install:

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o bin/scaffold-linux-arm64 ./cmd/scaffold
```

Set `CGO_ENABLED=0` for a static binary unless a dependency genuinely requires
cgo. Static binaries make container images far simpler.

## Release automation

| Tool | Version | When |
| --- | --- | --- |
| `make build` | n/a | **Default.** Sufficient until you publish artifacts. |
| `goreleaser/goreleaser` | v2.17.1 | Publishing multi-platform binaries, checksums, archives, and changelogs on a tag. |

Do not add release automation before the project actually publishes anything.

## Containers

Use a multi-stage build. Compile in the Go image, then copy the binary into a
minimal runtime image.

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/scaffold ./cmd/scaffold

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/scaffold /scaffold
USER nonroot:nonroot
ENTRYPOINT ["/scaffold"]
```

Rules:

- Copy `go.mod`/`go.sum` and run `go mod download` before copying source, so the
  dependency layer caches.
- **Never run as root.** Use a `nonroot` user.
- Prefer `distroless` or `scratch` for static binaries; use `alpine` only if you
  need a shell for debugging.
- Never bake secrets into an image or pass them as build args.
- Pin the base image and let Dependabot update it.

This scaffold ships no Dockerfile on purpose — deployment target is a project
decision.

## CI

`.github/workflows/ci.yml` already provides three jobs: quality
(format, tidy, lint, race tests, build), commit message validation, and a
security scan (`govulncheck`).

Principles to preserve:

- CI calls `make` targets. Never duplicate command flags into the workflow.
- `go-version-file: go.mod` keeps CI on the version the module declares.
- Cache using both `go.sum` and `tools/go.sum`.
- Keep the vulnerability scan in its own job so an external database outage is
  distinguishable from a code failure.
- Grant least privilege: `permissions: contents: read` unless a job needs more.
- Pin actions to a major version tag and let Dependabot bump them.

Read the installed `golang-continuous-integration` skill before extending CI.

## Supply chain

- Commit `go.sum`. Never bypass checksum verification with `GOFLAGS=-mod=mod`
  plus `GONOSUMDB`/`GONOSUMCHECK`.
- Keep `make vuln` (govulncheck) in CI. It reports only vulnerabilities actually
  reachable from your code.
- Review new direct dependencies: maintenance activity, transitive dependency
  count, and licence.

## Rejected by default

- A Dockerfile or Kubernetes manifests in a scaffold with no deployment target.
- `latest` base image tags.
- Release workflows for a project that publishes nothing.
- Building inside CI with a Go version different from `go.mod`.
