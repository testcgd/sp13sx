#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

task_id="$1"
worktree_path="$2"
task_file=$(task_file_for "$task_id")
review_file="${3:-}"

prompt="Implement the task described in the attached files. Follow the repository instructions in .agent/agent.md. Work only in this worktree, complete the task end-to-end, run relevant validation, and create a git commit for the task if the implementation succeeds."

if [[ -n "$review_file" ]]; then
  prompt="Implement the task described in the attached files and address the attached review feedback. Follow the repository instructions in .agent/agent.md. Work only in this worktree, complete the task end-to-end, run relevant validation, and create a git commit for the task if the implementation succeeds."
  "$OPENCODE_BIN" run \
    --dir "$worktree_path" \
    --agent build \
    --dangerously-skip-permissions \
    -f "$ROOT_DIR/.agent/agent.md" \
    -f "$task_file" \
    -f "$review_file" \
    -- \
    "$prompt"
  exit $?
fi

"$OPENCODE_BIN" run \
  --dir "$worktree_path" \
  --agent build \
  --dangerously-skip-permissions \
  -f "$ROOT_DIR/.agent/agent.md" \
  -f "$task_file" \
  -- \
  "$prompt"
