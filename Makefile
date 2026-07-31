# Single source of truth for every development command.
# README, CI, and Git hooks must call these targets instead of duplicating flags.

BINARY      := mihomo-exporter
CMD_PACKAGE := ./cmd/$(BINARY)
BIN_DIR     := bin
TOOLS       := -modfile=tools/go.mod
COVER_FILE  := coverage.out

GO          ?= go
GOLANGCILINT := $(GO) tool $(TOOLS) golangci-lint
GOVULNCHECK  := $(GO) tool $(TOOLS) govulncheck

.DEFAULT_GOAL := help

## help: list available targets
.PHONY: help
help:
	@awk 'BEGIN { FS = ": " } /^## / { sub(/^## /, ""); split($$0, p, ": "); printf "  \033[36m%-16s\033[0m %s\n", p[1], p[2] }' $(MAKEFILE_LIST)

## setup: install tool dependencies, agent skills, and enable Git hooks
.PHONY: setup
setup: skills
	$(GO) mod download
	$(GO) mod download $(TOOLS)
	git config core.hooksPath .githooks
	@echo "Git hooks enabled from .githooks"

## skills: install AI agent skills declared in agents.toml
.PHONY: skills
skills:
	@if command -v npx >/dev/null 2>&1; then \
		npx --yes @sentry/dotagents@latest install; \
	else \
		echo "npx not found; skipping agent skills (see agents.toml)"; \
	fi

## run: run the application
.PHONY: run
run:
	$(GO) run $(CMD_PACKAGE)

## build: compile the binary into bin/
.PHONY: build
build:
	$(GO) build -trimpath -o $(BIN_DIR)/$(BINARY) $(CMD_PACKAGE)

## fmt: format and fix imports in place
.PHONY: fmt
fmt:
	$(GOLANGCILINT) fmt ./...

## fmt-check: verify formatting without writing changes
.PHONY: fmt-check
fmt-check:
	$(GOLANGCILINT) fmt --diff ./...

## lint: run the full linter suite
.PHONY: lint
lint:
	$(GOLANGCILINT) run ./...

## lint-fix: run linters and apply available autofixes
.PHONY: lint-fix
lint-fix:
	$(GOLANGCILINT) run --fix ./...

## test: run tests
.PHONY: test
test:
	$(GO) test ./...

## test-race: run tests with the race detector and shuffled order
.PHONY: test-race
test-race:
	$(GO) test -race -shuffle=on ./...

## cover: run tests and report total coverage
.PHONY: cover
cover:
	$(GO) test -coverprofile=$(COVER_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVER_FILE) | tail -1

## vuln: scan dependencies for known vulnerabilities
.PHONY: vuln
vuln:
	$(GOVULNCHECK) ./...

## tidy: tidy the main and tool modules
.PHONY: tidy
tidy:
	$(GO) mod tidy
	cd tools && $(GO) mod tidy

## tidy-check: fail if either module is not tidy
.PHONY: tidy-check
tidy-check:
	@set -e; \
	for dir in . tools; do \
		cd "$(CURDIR)/$$dir"; \
		cp go.mod go.mod.bak; \
		[ -f go.sum ] && cp go.sum go.sum.bak || true; \
		trap 'mv -f go.mod.bak go.mod 2>/dev/null || true; \
		      mv -f go.sum.bak go.sum 2>/dev/null || true' EXIT INT TERM; \
		$(GO) mod tidy; \
		if ! cmp -s go.mod go.mod.bak || \
		   { [ -f go.sum.bak ] && ! cmp -s go.sum go.sum.bak; }; then \
			echo "$$dir/go.mod or go.sum is not tidy; run 'make tidy'"; \
			exit 1; \
		fi; \
		trap - EXIT INT TERM; \
		mv -f go.mod.bak go.mod; \
		[ -f go.sum.bak ] && mv -f go.sum.bak go.sum || true; \
	done; \
	echo "modules are tidy"

## check: full verification gate used by CI
.PHONY: check
check: fmt-check tidy-check lint test-race build
	@echo "All checks passed"

## commit-check: validate the most recent commit message
.PHONY: commit-check
commit-check:
	@git log -1 --pretty=%B | .githooks/commit-msg /dev/stdin

## clean: remove build and coverage artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(COVER_FILE)
	$(GO) clean -testcache
