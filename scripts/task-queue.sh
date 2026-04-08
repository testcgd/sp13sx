#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

cmd="${1:-status}"

show_status() {
  local current_id
  current_id=$(awk -F'"' '/^  id: "/ {print $2; exit}' "$QUEUE_FILE")
  printf 'Queue file: %s\n' "$QUEUE_FILE"
  printf 'Current: %s\n' "${current_id:-<none>}"
  printf 'Pending count: %s\n' "$(queue_pending_count)"
  printf 'Next: %s\n' "$(queue_first_pending || true)"
}

list_pending() {
  awk '
    /^pending:/ {in_pending=1; next}
    in_pending && /^history:/ {exit}
    in_pending && /^[[:space:]]*-[[:space:]]*"/ {
      gsub(/^[[:space:]]*-[[:space:]]*"/, "")
      gsub(/"[[:space:]]*$/, "")
      print
    }
  ' "$QUEUE_FILE"
}

validate_queue() {
  local next
  next=$(queue_first_pending || true)
  if [[ -n "$next" ]]; then
    [[ -f "$(task_file_for "$next")" ]] || die "missing definition for queued task: $next"
  fi
  printf 'Queue looks valid.\n'
}

case "$cmd" in
  status)
    show_status
    ;;
  list)
    list_pending
    ;;
  next)
    queue_first_pending
    ;;
  validate)
    validate_queue
    ;;
  *)
    die "unknown command: $cmd"
    ;;
esac
