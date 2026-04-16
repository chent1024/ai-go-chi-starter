#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "OPENSPEC REQUIRED CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

exec python3 "$(dirname "$0")/verify-openspec-required.py" "$env_file"
