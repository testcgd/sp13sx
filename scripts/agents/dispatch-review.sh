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
  --config "$ROOT_DIR/.agent/review.md" \
  --cwd "$worktree_path" \
  "Review task $(task_file_for "$task_id") and write report to $review_file"

printf '%s\n' "$review_file"
