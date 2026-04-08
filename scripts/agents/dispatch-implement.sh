#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

task_id="$1"
worktree_path="$2"
task_file=$(task_file_for "$task_id")
review_file="${3:-}"

cmd=("$OPENCODE_BIN" run --config "$ROOT_DIR/.agent/agent.md" --cwd "$worktree_path" "Implement task from $task_file")

if [[ -n "$review_file" ]]; then
  cmd=("$OPENCODE_BIN" run --config "$ROOT_DIR/.agent/agent.md" --cwd "$worktree_path" "Implement task from $task_file and address feedback from $review_file")
fi

"${cmd[@]}"
