#!/usr/bin/env bash
# Record 模式启动脚本 - 录制真实 LLM 交互
set -euo pipefail

OUTPUT="${1:-./test/recordings/$(date +%Y-%m-%d_%H-%M%S).jsonl}"

# 确保输出目录存在
mkdir -p "$(dirname "$OUTPUT")"

SP13SX_TEST_MODE=record \
SP13SX_RECORDING_OUTPUT="$OUTPUT" \
SP13SX_CONFIG="${SP13SX_CONFIG:-$(pwd)/config.local.yml}" \
go run ./cmd/sp13sx

echo "Recording saved to: $OUTPUT"
