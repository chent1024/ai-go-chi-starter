#!/usr/bin/env bash
set -euo pipefail

env_file="${1:-.orch/rules/repo/rules.env}"

# shellcheck disable=SC1090
source "$env_file"

changed_files="$(git diff --name-only --diff-filter=ACMR HEAD 2>/dev/null || true)"
if [[ -z "$changed_files" ]]; then
  exit 0
fi

violations=0
while IFS= read -r path; do
  [[ -z "$path" || ! -f "$path" ]] && continue
  if head -n 5 "$path" | rg -q "$GENERATED_MARKERS"; then
    echo "GENERATED FILE CHECK FAILED: changed generated file: $path" >&2
    violations=1
  fi
done <<<"$changed_files"

if [[ "$violations" -ne 0 ]]; then
  exit 1
fi
