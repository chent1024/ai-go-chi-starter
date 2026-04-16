#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "MIGRATION CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${MIGRATIONS_DIR:-}" || ! -d "${MIGRATIONS_DIR:-}" ]]; then
  echo "migration checks skipped: MIGRATIONS_DIR not configured"
  exit 0
fi

base_ref="${MIGRATION_BASE_REF:-HEAD}"
if git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
  base_ref="${MIGRATION_BASE_REF:-HEAD~1}"
fi

changed="$(git diff --name-status "$base_ref" -- "$MIGRATIONS_DIR" || true)"
if [[ -z "$changed" ]]; then
  exit 0
fi

violations=0
while IFS=$'\t' read -r status path _; do
  [[ -z "${status:-}" ]] && continue
  case "$status" in
    A)
      if [[ "${MIGRATION_FILE_REGEX:-}" != "" ]] && [[ ! "$path" =~ ${MIGRATION_FILE_REGEX} ]]; then
        echo "MIGRATION CHECK FAILED: unexpected migration filename: $path" >&2
        violations=1
      fi
      ;;
    M|D|R*)
      if [[ "${ALLOW_EDIT_EXISTING_MIGRATIONS:-false}" != "true" ]]; then
        echo "MIGRATION CHECK FAILED: existing migration changed: $path" >&2
        echo "Prefer forward-only migrations and fix-forward changes." >&2
        violations=1
      fi
      ;;
  esac
done <<<"$changed"

if [[ "$violations" -ne 0 ]]; then
  exit 1
fi
