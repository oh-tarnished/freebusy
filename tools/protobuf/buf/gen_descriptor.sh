#!/usr/bin/env bash
set -euo pipefail

OUT_DIR="generated/descriptor"
OUT_FILE="${OUT_DIR}/api_descriptors.pb"

mkdir -p "${OUT_DIR}"

echo "==> Generating proto descriptors with buf…"
buf build -o "${OUT_FILE}"
