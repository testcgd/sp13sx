#!/usr/bin/env bash
# Mock 模式启动脚本 - 使用场景脚本模拟 LLM 响应
set -euo pipefail

SCENARIO="${1:-basic_chat}"

SP13SX_TEST_MODE=mock \
SP13SX_SCENARIO="$SCENARIO" \
SP13SX_CONFIG="${SP13SX_CONFIG:-$(pwd)/config.local.yml}" \
go run ./cmd/sp13sx
