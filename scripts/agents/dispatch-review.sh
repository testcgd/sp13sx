#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

task_id="$1"
worktree_path="$2"
review_number="$3"
task_review_dir="$TASK_REVIEWS_DIR/$task_id"
review_file="$task_review_dir/review-$review_number.md"

mkdir -p "$task_review_dir"

"$OPENCODE_BIN" run \
  --dir "$worktree_path" \
  --agent build \
  --dangerously-skip-permissions \
  -f "$ROOT_DIR/.agent/review.md" \
  -f "$(task_file_for "$task_id")" \
  -- \
  "Review the attached task implementation. Follow .agent/review.md exactly, run the required checks, and write the review report to $review_file." >&2

printf '%s\n' "$review_file"
