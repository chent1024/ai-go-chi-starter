#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "ARCH CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${ARCH_IMPORT_RULES_FILE:-}" ]]; then
  echo "architecture checks skipped: ARCH_IMPORT_RULES_FILE not configured"
  exit 0
fi

if [[ ! -f "$ARCH_IMPORT_RULES_FILE" ]]; then
  echo "ARCH CHECK FAILED: rules file not found: $ARCH_IMPORT_RULES_FILE" >&2
  exit 1
fi

arch_backend="${ARCH_BACKEND:-text}"

if [[ "$arch_backend" == "go_imports" ]]; then
  exec python3 "$(dirname "$0")/verify-arch-go.py" "$env_file"
fi

if ! command -v rg >/dev/null 2>&1; then
  echo "ARCH CHECK FAILED: rg is required" >&2
  exit 1
fi

failures=0
while IFS= read -r rules_file; do
  while IFS='|' read -r scope pattern detail guidance; do
    [[ -z "${scope:-}" || "${scope:0:1}" == "#" ]] && continue
    matches="$(rg -n "$pattern" "$scope" || true)"
    if [[ -n "$matches" ]]; then
      echo "ARCH CHECK FAILED: $detail" >&2
      if [[ -n "${guidance:-}" ]]; then
        echo "recommended fix: $guidance" >&2
      fi
      echo "$matches" >&2
      echo >&2
      failures=1
    fi
  done <"$rules_file"
done < <(rules_list_rule_files "$ARCH_IMPORT_RULES_FILE")

if [[ "$failures" -ne 0 ]]; then
  exit 1
fi

echo "architecture checks passed"
