# sp13sx

`sp13sx` is a Go TUI agent scaffold with:

- OpenAI Responses backend abstraction
- JSONL conversation persistence
- YAML configuration
- standard `SKILL.md` skill discovery
- MCP client management scaffolding
- built-in local tools

## Layout

- `cmd/sp13sx`: entrypoint
- `internal`: application implementation
- `docs`: design documents
- `examples`: example config and skills

## Status

This is a first implementation scaffold. The TUI, config, JSONL store, skills loader, tool registry, and backend interfaces are in place. OpenAI and MCP integrations are wired as runtime components, with MCP transport intentionally left at manager/config level until SDK binding is added.
