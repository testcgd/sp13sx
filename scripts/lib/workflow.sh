#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
TASKS_DIR="$ROOT_DIR/docs/tasks"
TASK_DEFINITIONS_DIR="$TASKS_DIR/definitions"
TASK_REVIEWS_DIR="$TASKS_DIR/reviews"
TASK_COMPLETED_DIR="$TASKS_DIR/completed"
QUEUE_FILE="$TASKS_DIR/queue.yaml"
WORKTREE_BASE="$ROOT_DIR/.worktrees"
LOG_DIR="$WORKTREE_BASE/.logs"
PAUSE_FILE="$WORKTREE_BASE/.pause"
INTEGRATION_WORKTREE="$WORKTREE_BASE/_integration"
INTEGRATION_BRANCH="integration/master"
MASTER_BRANCH="${MASTER_BRANCH:-master}"
OPENCODE_BIN="${OPENCODE_BIN:-opencode}"

log() {
  printf '[%s] %s\n' "$(date -Iseconds)" "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

ensure_layout() {
  mkdir -p "$TASK_DEFINITIONS_DIR" "$TASK_REVIEWS_DIR" "$TASK_COMPLETED_DIR" "$WORKTREE_BASE" "$LOG_DIR"
  [[ -f "$QUEUE_FILE" ]] || die "missing queue file: $QUEUE_FILE"
}

ensure_integration_worktree() {
  if [[ ! -d "$INTEGRATION_WORKTREE/.git" ]]; then
    git worktree add "$INTEGRATION_WORKTREE" -b "$INTEGRATION_BRANCH" "$MASTER_BRANCH"
  fi
}

task_file_for() {
  local task_id="$1"
  printf '%s/%s.md\n' "$TASK_DEFINITIONS_DIR" "$task_id"
}

task_title() {
  local task_file="$1"
  awk -F': ' '/^title:/ {gsub(/"/, "", $2); print $2; exit}' "$task_file"
}

task_review_attempts() {
  local task_file="$1"
  awk -F': ' '/^review_attempts:/ {print $2; exit}' "$task_file"
}

task_max_review_attempts() {
  local task_file="$1"
  awk -F': ' '/^max_review_attempts:/ {print $2; exit}' "$task_file"
}

queue_first_pending() {
  awk '
    /^pending:/ {in_pending=1; next}
    in_pending && /^history:/ {exit}
    in_pending && /^[[:space:]]*-[[:space:]]*"/ {
      gsub(/^[[:space:]]*-[[:space:]]*"/, "")
      gsub(/"[[:space:]]*$/, "")
      print
      exit
    }
  ' "$QUEUE_FILE"
}

queue_pending_count() {
  awk '
    /^pending:/ {in_pending=1; next}
    in_pending && /^history:/ {printed=1; print count+0; exit}
    in_pending && /^[[:space:]]*-[[:space:]]*"/ {count++}
    END {if (in_pending && !printed) print count+0}
  ' "$QUEUE_FILE"
}

queue_pop_first_pending() {
  local tmp
  tmp=$(mktemp)
  awk '
    /^pending:/ {print; in_pending=1; next}
    in_pending && !removed && /^[[:space:]]*-[[:space:]]*"/ {removed=1; next}
    in_pending && /^history:/ {in_pending=0}
    {print}
  ' "$QUEUE_FILE" > "$tmp"
  mv "$tmp" "$QUEUE_FILE"
}

queue_set_current() {
  local task_id="$1"
  local worktree_path="$2"
  local started_at
  started_at=$(date -Iseconds)
  local tmp
  tmp=$(mktemp)
  awk -v task_id="$task_id" -v started_at="$started_at" -v worktree_path="$worktree_path" '
    /^current:/ {
      print
      getline; print "  id: \"" task_id "\""
      getline; print "  started_at: \"" started_at "\""
      getline; print "  worktree: \"" worktree_path "\""
      next
    }
    {print}
  ' "$QUEUE_FILE" > "$tmp"
  mv "$tmp" "$QUEUE_FILE"
}

queue_clear_current() {
  local tmp
  tmp=$(mktemp)
  awk '
    /^current:/ {
      print
      getline; print "  id: \"\""
      getline; print "  started_at: \"\""
      getline; print "  worktree: \"\""
      next
    }
    {print}
  ' "$QUEUE_FILE" > "$tmp"
  mv "$tmp" "$QUEUE_FILE"
}

queue_append_history() {
  local task_id="$1"
  local status="$2"
  local completed_at
  completed_at=$(date -Iseconds)
  printf '  - id: "%s"\n    status: %s\n    completed_at: "%s"\n' "$task_id" "$status" "$completed_at" >> "$QUEUE_FILE"
}

append_review_feedback() {
  local task_file="$1"
  local review_file="$2"
  local attempt="$3"
  {
    printf '\n### Review Attempt %s\n\n' "$attempt"
    awk 'f{print} /^## Findings/{f=1; next}' "$review_file"
  } >> "$task_file"
}

set_task_status() {
  local task_file="$1"
  local new_status="$2"
  local tmp
  tmp=$(mktemp)
  awk -F': ' -v status="$new_status" '
    /^status:/ {$0 = "status: \"" status "\""}
    {print}
  ' "$task_file" > "$tmp"
  mv "$tmp" "$task_file"
}

increment_task_review_attempts() {
  local task_file="$1"
  local current
  current=$(task_review_attempts "$task_file")
  local next=$((current + 1))
  local tmp
  tmp=$(mktemp)
  awk -F': ' -v next="$next" '
    /^review_attempts:/ {$0 = "review_attempts: " next}
    {print}
  ' "$task_file" > "$tmp"
  mv "$tmp" "$task_file"
}

review_status() {
  local review_file="$1"
  awk -F': ' '/^status:/ {print $2; exit}' "$review_file"
}

gofmt_file_list() {
  git -C "$ROOT_DIR" ls-files '*.go'
}

require_clean_gofmt_repo() {
  local files
  files=$(gofmt_file_list)
  if [[ -z "$files" ]]; then
    return 0
  fi
  local out
  out=$(printf '%s\n' "$files" | xargs gofmt -l)
  [[ -z "$out" ]]
}
