#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "CHANGE SCOPE CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

backend="${CHANGE_SCOPE_BACKEND:-}"
if [[ -z "$backend" ]]; then
  echo "change scope checks skipped: CHANGE_SCOPE_BACKEND not configured"
  exit 0
fi

if [[ "$backend" == "openspec" ]]; then
  exec python3 "$(dirname "$0")/verify-change-scope-openspec.py" "$env_file"
fi

echo "CHANGE SCOPE CHECK FAILED: unsupported backend: $backend" >&2
exit 1
