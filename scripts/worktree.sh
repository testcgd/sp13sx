#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(dirname "$0")
ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

wt_path() {
  local task_id="$1"
  printf '%s/%s\n' "$WORKTREE_BASE" "$task_id"
}

wt_branch() {
  local task_id="$1"
  printf 'task/%s\n' "$task_id"
}

wt_create() {
  local task_id="$1"
  local path
  path=$(wt_path "$task_id")
  local branch
  branch=$(wt_branch "$task_id")
  git worktree add "$path" -b "$branch" "$MASTER_BRANCH"
}

wt_remove() {
  local task_id="$1"
  local path
  path=$(wt_path "$task_id")
  local branch
  branch=$(wt_branch "$task_id")
  if [[ -d "$path" ]]; then
    git worktree remove "$path" --force
  fi
  git branch -D "$branch" >/dev/null 2>&1 || true
}

case "${1:-}" in
  create)
    wt_create "$2"
    ;;
  remove)
    wt_remove "$2"
    ;;
  list)
    git worktree list
    ;;
  *)
    printf 'usage: %s {create|remove|list} [task-id]\n' "$0" >&2
    exit 1
    ;;
esac
