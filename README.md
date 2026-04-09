# sp13sx

`sp13sx` is a Go TUI agent scaffold with:

- OpenAI Responses backend abstraction
- JSONL conversation persistence
- YAML configuration
- standard `SKILL.md` skill discovery
- MCP client management scaffolding
- built-in local tools

## Autonomous Workflow

This repo also contains a local autonomous development workflow built on tasks, agents, and `git worktree`.

### Start The Workflow

Run the orchestrator from the repo root:

```bash
./scripts/orchestrator.sh
```

Useful commands:

```bash
./scripts/task-queue.sh status
./scripts/task-queue.sh validate
./scripts/orchestrator.sh pause
./scripts/orchestrator.sh resume
```

### Add A Task

Create a new task definition:

```bash
./scripts/task-create.sh 002 add-config-tests
```

Then append the task id to `docs/tasks/queue.yaml` under `pending`:

```yaml
pending:
  - "001-bootstrap-autonomous-workflow"
  - "002"
```

You can also create the Markdown file directly under `docs/tasks/definitions/` using `docs/tasks/definitions/template.md` as the template.

### Main Flow

1. The orchestrator reads `docs/tasks/queue.yaml`.
2. It creates a dedicated worktree for the next task.
3. An implementation agent executes the task.
4. A new review agent validates the result.
5. If review fails, feedback is appended back into the task document and implementation restarts.
6. If review passes, the task branch is merged locally into `master` through the integration worktree.
7. The merged result is validated again with formatting and tests.
8. If integration validation fails, a fix agent runs and retries integration.
9. After success, the task is archived and the next queued task starts.

More detail:

- `docs/workflow/README.md`
- `docs/tasks/README.md`

## Layout

- `cmd/sp13sx`: entrypoint
- `internal`: application implementation
- `docs`: design documents
- `examples`: example config and skills

## Status

This is a first implementation scaffold. The TUI, config, JSONL store, skills loader, tool registry, and backend interfaces are in place. OpenAI and MCP integrations are wired as runtime components, with MCP transport intentionally left at manager/config level until SDK binding is added.
