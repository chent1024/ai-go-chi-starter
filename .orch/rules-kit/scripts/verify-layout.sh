#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "LAYOUT CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${LAYOUT_REQUIRED_DIRS:-}" && -z "${LAYOUT_FORBIDDEN_DIRS:-}" ]]; then
  echo "layout checks skipped: LAYOUT_REQUIRED_DIRS and LAYOUT_FORBIDDEN_DIRS not configured"
  exit 0
fi

failures=0

for path in ${LAYOUT_REQUIRED_DIRS:-}; do
  if [[ ! -e "$path" ]]; then
    echo "LAYOUT CHECK FAILED: required path is missing: $path" >&2
    failures=1
  fi
done

for path in ${LAYOUT_FORBIDDEN_DIRS:-}; do
  if [[ -e "$path" ]]; then
    echo "LAYOUT CHECK FAILED: forbidden path exists: $path" >&2
    failures=1
  fi
done

if [[ "$failures" -ne 0 ]]; then
  exit 1
fi

echo "layout checks passed"
