#!/usr/bin/env bash
set -euo pipefail

SP13SX_CONFIG="${SP13SX_CONFIG:-$(pwd)/examples/config.example.yml}" go run ./cmd/sp13sx
