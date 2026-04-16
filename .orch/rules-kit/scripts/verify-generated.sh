#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "GENERATED FILE CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${GENERATED_MARKERS:-}" ]]; then
  echo "generated file checks skipped: GENERATED_MARKERS not configured"
  exit 0
fi

changed_files="$(git diff --name-only --diff-filter=ACMR HEAD 2>/dev/null || true)"
if [[ -z "$changed_files" ]]; then
  exit 0
fi

violations=0

while IFS= read -r path; do
  [[ -z "$path" || ! -f "$path" ]] && continue
  case "$path" in
    .orch/rules/*|tools/rules-kit/*|Makefile.rules)
      continue
      ;;
  esac
  if head -n 5 "$path" | rg -q "$GENERATED_MARKERS"; then
    echo "GENERATED FILE CHECK FAILED: changed generated file: $path" >&2
    echo "Update generated files through their source inputs or generation command." >&2
    echo >&2
    violations=1
  fi
done <<<"$changed_files"

if [[ "$violations" -ne 0 ]]; then
  exit 1
fi
