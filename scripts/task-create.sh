#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
# shellcheck source=scripts/lib/workflow.sh
source "$ROOT_DIR/scripts/lib/workflow.sh"

task_id="${1:-}"
shift || true
title="${1:-}"

[[ -n "$task_id" ]] || die "usage: $0 <task-id> <title>"
[[ -n "$title" ]] || die "usage: $0 <task-id> <title>"

task_file="$TASK_DEFINITIONS_DIR/$task_id.md"
[[ ! -e "$task_file" ]] || die "task already exists: $task_file"

cat > "$task_file" <<EOF
---
id: "$task_id"
title: "$title"
status: "pending"
review_attempts: 0
max_review_attempts: 5
labels: []
dependencies: []
---

# $title

## Description

TODO: describe the task.

## Acceptance Criteria

- [ ] Define acceptance criteria

## Review Feedback

验收失败时，编排器会在这里追加反馈。
EOF

printf 'Created %s\n' "$task_file"
printf 'Append "%s" to docs/tasks/queue.yaml pending list to schedule it.\n' "$task_id"
