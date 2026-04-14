#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TAPES_DIR="$SCRIPT_DIR/tapes"
OUTPUT_DIR="$SCRIPT_DIR/output"

usage() {
    echo "Usage: $0 [command] [tape_name]"
    echo ""
    echo "Commands:"
    echo "  run <tape>    Run a single tape file"
    echo "  all           Run all tape files"
    echo "  list          List available tapes"
    echo "  clean         Clean output directory"
    echo "  install       Install vhs (requires go)"
    echo ""
    echo "Examples:"
    echo "  $0 run basic_chat"
    echo "  $0 all"
    echo "  $0 list"
}

check_vhs() {
    if ! command -v vhs &> /dev/null; then
        echo "Error: vhs is not installed."
        echo "Run '$0 install' to install it."
        exit 1
    fi
    if ! command -v ttyd &> /dev/null; then
        echo "Error: ttyd is not installed (required by vhs)."
        echo "Install it from: https://github.com/tsl0922/ttyd"
        echo ""
        echo "  Ubuntu/Debian: sudo apt install ttyd"
        echo "  macOS: brew install ttyd"
        exit 1
    fi
}

install_vhs() {
    echo "Installing vhs..."
    go install github.com/charmbracelet/vhs@latest
    echo "vhs installed successfully."
}

list_tapes() {
    echo "Available tapes:"
    echo ""
    for tape in "$TAPES_DIR"/*.tape; do
        if [ -f "$tape" ]; then
            name=$(basename "$tape" .tape)
            echo "  - $name"
        fi
    done
    echo ""
    echo "Note: real_llm tape requires PRIVATE_OPENAI_API_KEY environment variable."
}

run_tape() {
    local tape_name="$1"
    local tape_file="$TAPES_DIR/${tape_name}.tape"

    if [ ! -f "$tape_file" ]; then
        echo "Error: Tape file not found: $tape_file"
        exit 1
    fi

    check_vhs

    mkdir -p "$OUTPUT_DIR"

    echo "Running tape: $tape_name"
    cd "$PROJECT_ROOT"
    vhs < "$tape_file"

    echo "Output saved to: $OUTPUT_DIR/${tape_name}.gif"
}

run_all() {
    check_vhs

    mkdir -p "$OUTPUT_DIR"

    for tape in "$TAPES_DIR"/*.tape; do
        if [ -f "$tape" ]; then
            name=$(basename "$tape" .tape)
            echo ""
            echo "=== Running: $name ==="
            cd "$PROJECT_ROOT"
            vhs < "$tape" || echo "Warning: $name failed"
        fi
    done

    echo ""
    echo "All tapes completed. Outputs in: $OUTPUT_DIR"
}

clean() {
    rm -rf "$OUTPUT_DIR"/*
    echo "Output directory cleaned."
}

cd "$PROJECT_ROOT"

case "${1:-}" in
    run)
        if [ -z "${2:-}" ]; then
            echo "Error: Tape name required"
            usage
            exit 1
        fi
        run_tape "$2"
        ;;
    all)
        run_all
        ;;
    list)
        list_tapes
        ;;
    clean)
        clean
        ;;
    install)
        install_vhs
        ;;
    *)
        usage
        exit 1
        ;;
esac
