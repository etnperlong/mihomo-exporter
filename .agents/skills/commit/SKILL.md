---
name: commit
description: Create Conventional Commits with a mandatory emoji prefix for this repository. Use when committing changes, staging files, or when the user says commit, save changes, or push.
---

# Commit

## Quick Start

Review the change, pick the type, prefix the required emoji, commit.

## Format

```text
<emoji> <type>[(scope)][!]: <description>
```

| Type | Emoji | Use for |
| --- | --- | --- |
| feat | ✨ | New feature |
| fix | 🐛 | Bug fix |
| docs | 📚 | Documentation only |
| style | 🎨 | Formatting, no behaviour change |
| refactor | 🔨 | Restructuring without behaviour change |
| perf | 📈 | Performance improvement |
| test | 🧪 | Adding or fixing tests |
| build | 📦 | Build system or dependencies |
| ci | 👷 | CI configuration |
| chore | 🔧 | Maintenance |
| revert | ⏪ | Reverting a commit |

The emoji and type must match. `✨ fix:` is rejected.

Rules enforced by `.githooks/commit-msg`:

- emoji is mandatory and must correspond to the type;
- scope is optional, lowercase, `[a-z0-9._/-]`;
- `!` before the colon marks a breaking change;
- description is imperative mood, starts lowercase, no trailing period;
- header is at most 100 characters;
- `deps:` / `deps-dev:` (Dependabot) and Git-generated `Merge`/`Revert`/`fixup!`
  headers bypass the emoji rule.

## Workflow

1. Run `git status` and `git diff --staged` to review what is staged.
2. If nothing is staged, stage the relevant files. **Never stage secrets,
   `.env`, credentials, or `bin/`.**
3. Verify before committing:
   ```bash
   make fmt-check && make lint && make test
   ```
   Do not commit with any of these failing.
4. Choose the type from the actual change, and take its emoji from the table.
5. Write the description: imperative, lowercase, no period, under 72 characters
   where practical.
6. Commit:
   ```bash
   git commit -m "✨ feat(greeting): support unicode names"
   ```
7. If the hook rejects the message, read its output, fix the message, and retry.
   **Never use `--no-verify` to bypass it.**

## Examples

```text
✨ feat(greeting): support unicode names
🐛 fix: handle whitespace-only input
🔨 refactor(cmd)!: remove the legacy output flag
🧪 test(greeting): cover surrounding whitespace
👷 ci: cache the tool module
📚 docs: document the make targets
```

## Rules

- Split unrelated changes into separate commits.
- English for all commit messages.
- One logical change per commit; do not mix a refactor with a behaviour change.
- Do not amend or force-push a commit that is already shared, unless the user
  asks for it.
- The body explains **why**, not what — the diff already shows what.

## Verifying the last commit

```bash
make commit-check
```

CI re-runs the same hook against every commit in a pull request, so a message
that bypasses the local hook still fails there.
