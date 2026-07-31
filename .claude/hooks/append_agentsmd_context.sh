#!/bin/bash
# Load every AGENTS.md in the project into the session context at startup.
set -eu

find "${CLAUDE_PROJECT_DIR:?CLAUDE_PROJECT_DIR must be set}" \
	-name AGENTS.md \
	-type f \
	-not -path '*/.agents/*' \
	-not -path '*/node_modules/*' \
	-exec cat {} +
