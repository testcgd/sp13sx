#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

task_id="$1"
worktree_path="$2"
failure_log="$3"

"$OPENCODE_BIN" run \
  --config "$ROOT_DIR/.agent/agent.md" \
  --cwd "$worktree_path" \
  "Fix post-merge validation failure for task $task_id using log $failure_log"
