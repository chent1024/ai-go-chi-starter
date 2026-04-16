#!/usr/bin/env bash
set -euo pipefail

env_file="${1:-.orch/rules/repo/rules.env}"

# shellcheck disable=SC1090
source "$env_file"

exclude_args=()
for pattern in ${SECRET_SCAN_EXCLUDE_GLOBS:-}; do
  exclude_args+=(--glob "!$pattern")
done

default_pattern='(?i)(api[_-]?key|secret|token|private[_-]?key|signing[_-]?secret)\s*[:=]\s*["'"'"'][^"'"'"']{12,}["'"'"']'
matches="$(rg -n "${SECRET_SCAN_REGEX:-$default_pattern}" ${SECRET_SCAN_PATHS:-.} "${exclude_args[@]}" || true)"
if [[ -n "$matches" ]]; then
  echo "SECRET CHECK FAILED: possible hard-coded secrets detected" >&2
  echo "$matches" >&2
  exit 1
fi
