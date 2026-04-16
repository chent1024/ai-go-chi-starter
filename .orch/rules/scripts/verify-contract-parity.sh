#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "CONTRACT PARITY CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

backend="${CONTRACT_PARITY_BACKEND:-}"
if [[ -z "$backend" ]]; then
  echo "contract parity checks skipped: CONTRACT_PARITY_BACKEND not configured"
  exit 0
fi

if [[ "$backend" == "manifest" ]]; then
  exec python3 "$(dirname "$0")/verify-contract-parity-manifest.py" "$env_file"
fi

echo "CONTRACT PARITY CHECK FAILED: unsupported backend: $backend" >&2
exit 1
