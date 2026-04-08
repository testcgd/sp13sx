#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

ensure_layout
ensure_integration_worktree

wt_path_for() {
  printf '%s/%s\n' "$WORKTREE_BASE" "$1"
}

wt_branch_for() {
  printf 'task/%s\n' "$1"
}

run_pre_merge_validation() {
  local worktree_path="$1"
  (
    cd "$worktree_path"
    local files
    files=$(git ls-files '*.go')
    if [[ -n "$files" ]] && printf '%s\n' "$files" | xargs gofmt -l | grep -q .; then
      return 1
    fi
    go test ./...
  )
}

run_post_merge_validation() {
  local integration_path="$INTEGRATION_WORKTREE"
  local failure_log="$LOG_DIR/post-merge-failure.log"
  : > "$failure_log"
  local files
  files=$(git -C "$integration_path" ls-files '*.go')
  if [[ -n "$files" ]] && (cd "$integration_path" && printf '%s\n' "$files" | xargs gofmt -l) | tee -a "$failure_log" | grep -q .; then
    return 1
  fi
  if ! (cd "$integration_path" && go test ./... 2>&1) | tee -a "$failure_log"; then
    return 1
  fi
  return 0
}

integrate_task_branch() {
  local task_id="$1"
  local task_file
  task_file=$(task_file_for "$task_id")
  local title
  title="$task_id: $(task_title "$task_file")"
  local branch
  branch=$(wt_branch_for "$task_id")

  git -C "$INTEGRATION_WORKTREE" checkout "$INTEGRATION_BRANCH" >/dev/null 2>&1 || true
  git -C "$INTEGRATION_WORKTREE" reset --hard "$MASTER_BRANCH" >/dev/null
  git -C "$INTEGRATION_WORKTREE" clean -fd >/dev/null
  git -C "$INTEGRATION_WORKTREE" merge --squash "$branch"
  git -C "$INTEGRATION_WORKTREE" commit -m "$title"
}

advance_master_branch() {
  local merged_sha
  merged_sha=$(git -C "$INTEGRATION_WORKTREE" rev-parse HEAD)
  git update-ref "refs/heads/$MASTER_BRANCH" "$merged_sha"
}

process_task() {
  local task_id="$1"
  local task_file
  task_file=$(task_file_for "$task_id")
  [[ -f "$task_file" ]] || die "missing task definition: $task_file"

  local worktree_path
  worktree_path=$(wt_path_for "$task_id")
  local branch
  branch=$(wt_branch_for "$task_id")

  if [[ ! -d "$worktree_path" ]]; then
    git worktree add "$worktree_path" -b "$branch" "$MASTER_BRANCH"
  fi

  queue_set_current "$task_id" "$worktree_path"
  set_task_status "$task_file" "in_progress"

  local review_round=0
  local max_reviews
  max_reviews=$(task_max_review_attempts "$task_file")
  local latest_review=""

  while true; do
    review_round=$((review_round + 1))
    log "implementation round $review_round for $task_id"
    "$ROOT_DIR/scripts/agents/dispatch-implement.sh" "$task_id" "$worktree_path" "$latest_review" | tee "$LOG_DIR/$task_id-implement-$review_round.log"

    run_pre_merge_validation "$worktree_path"

    set_task_status "$task_file" "review"
    latest_review=$("$ROOT_DIR/scripts/agents/dispatch-review.sh" "$task_id" "$worktree_path" "$review_round")
    local status
    status=$(review_status "$latest_review")

    if [[ "$status" == "PASSED" ]]; then
      set_task_status "$task_file" "passed"
      break
    fi

    increment_task_review_attempts "$task_file"
    append_review_feedback "$task_file" "$latest_review" "$review_round"
    set_task_status "$task_file" "failed"

    if [[ "$review_round" -ge "$max_reviews" ]]; then
      die "task $task_id exceeded max review attempts"
    fi

    set_task_status "$task_file" "in_progress"
  done

  integrate_task_branch "$task_id"

  if ! run_post_merge_validation; then
    log "post-merge validation failed for $task_id"
    "$ROOT_DIR/scripts/agents/dispatch-fix.sh" "$task_id" "$worktree_path" "$LOG_DIR/post-merge-failure.log" | tee "$LOG_DIR/$task_id-fix.log"
    integrate_task_branch "$task_id"
    run_post_merge_validation
  fi

  advance_master_branch

  mv "$task_file" "$TASK_COMPLETED_DIR/"
  queue_append_history "$task_id" completed
  queue_clear_current
  git worktree remove "$worktree_path" --force
  git branch -D "$branch" >/dev/null 2>&1 || true
}

loop() {
  while true; do
    if [[ -f "$PAUSE_FILE" ]]; then
      sleep 5
      continue
    fi

    local task_id
    task_id=$(queue_first_pending || true)
    if [[ -z "$task_id" ]]; then
      log "queue empty"
      exit 0
    fi

    queue_pop_first_pending
    process_task "$task_id"
  done
}

case "${1:-run}" in
  run)
    loop
    ;;
  status)
    "$ROOT_DIR/scripts/task-queue.sh" status
    ;;
  pause)
    mkdir -p "$WORKTREE_BASE"
    touch "$PAUSE_FILE"
    ;;
  resume)
    rm -f "$PAUSE_FILE"
    ;;
  *)
    die "usage: $0 [run|status|pause|resume]"
    ;;
esac
