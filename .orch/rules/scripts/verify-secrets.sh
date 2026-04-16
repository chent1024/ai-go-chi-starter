#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "SECRET CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if ! command -v rg >/dev/null 2>&1; then
  echo "SECRET CHECK FAILED: rg is required" >&2
  exit 1
fi

paths="${SECRET_SCAN_PATHS:-.}"
exclude_args=()
for pattern in ${SECRET_SCAN_EXCLUDE_GLOBS:-}; do
  exclude_args+=(--glob "!$pattern")
done

default_pattern='(?i)(api[_-]?key|secret|token|private[_-]?key|signing[_-]?secret)\s*[:=]\s*["'"'"'][^"'"'"']{12,}["'"'"']'
pattern="${SECRET_SCAN_REGEX:-$default_pattern}"

matches="$(rg -n "$pattern" $paths "${exclude_args[@]}" || true)"
if [[ -n "$matches" ]]; then
  echo "SECRET CHECK FAILED: possible hard-coded secrets detected" >&2
  echo "$matches" >&2
  exit 1
fi
