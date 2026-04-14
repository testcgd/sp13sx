# VHS TUI Testing

This directory contains [vhs](https://github.com/charmbracelet/vhs) tape files for automated TUI testing.

## Prerequisites

1. Install ttyd (required by vhs):
   ```bash
   # Ubuntu/Debian
   sudo apt install ttyd
   
   # macOS
   brew install ttyd
   
   # Or build from source: https://github.com/tsl0922/ttyd
   ```

2. Install vhs:
   ```bash
   go install github.com/charmbracelet/vhs@latest
   # or
   ./run.sh install
   ```

## Usage

### Run a single tape

```bash
./test/vhs/run.sh run basic_chat
```

### Run all tapes

```bash
./test/vhs/run.sh all
```

### List available tapes

```bash
./test/vhs/run.sh list
```

### Clean outputs

```bash
./test/vhs/run.sh clean
```

## Available Tapes

| Tape | Description |
|------|-------------|
| `basic_chat` | Basic chat interaction with real LLM |
| `mock_basic` | Basic chat with mock backend |
| `help_command` | Test /help command |

## Writing New Tapes

Tape files use TOML syntax. Example:

```toml
Output "output/my_test.gif"
Set FontSize 14
Set Width 120
Set Height 30

Type "./sp13sx -f ./config.local.yml"
Enter
Wait "Ask"

Type "your input"
Enter
Sleep 2s
```

See [vhs documentation](https://github.com/charmbracelet/vhs#commands) for all available commands.

## CI Integration

Add to your CI pipeline:

```yaml
- name: Install vhs
  run: go install github.com/charmbracelet/vhs@latest

- name: Run VHS tests
  run: ./test/vhs/run.sh all
```

## Output

Generated GIFs are saved to `test/vhs/output/`.
