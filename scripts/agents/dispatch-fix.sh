#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

task_id="$1"
worktree_path="$2"
failure_log="$3"

"$OPENCODE_BIN" run \
  --dir "$worktree_path" \
  --agent build \
  --dangerously-skip-permissions \
  -f "$ROOT_DIR/.agent/agent.md" \
  -f "$(task_file_for "$task_id")" \
  -f "$failure_log" \
  -- \
  "Fix the attached post-merge validation failure for task $task_id. Follow .agent/agent.md, update the implementation in this worktree, run validation, and create a git commit if you make changes."
