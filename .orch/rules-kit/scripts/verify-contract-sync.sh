#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "CONTRACT CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${CONTRACT_PATHS:-}" || -z "${API_SURFACE_PATHS:-}" ]]; then
  echo "contract sync checks skipped: CONTRACT_PATHS or API_SURFACE_PATHS not configured"
  exit 0
fi

base_ref="${CONTRACT_SYNC_BASE_REF:-HEAD}"
if git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
  base_ref="${CONTRACT_SYNC_BASE_REF:-HEAD~1}"
fi

api_changed=0
contract_changed=0

path_has_non_test_go_changes() {
  local base_ref="$1"
  local path="$2"
  local changed_files=""

  changed_files="$(git diff --name-only "$base_ref" -- "$path" || true)"
  if [[ -z "$changed_files" ]]; then
    return 1
  fi

  if printf '%s\n' "$changed_files" | rg -qv '(^|/)[^/]+_test\.go$'; then
    return 0
  fi

  return 1
}

for path in $API_SURFACE_PATHS; do
  if path_has_non_test_go_changes "$base_ref" "$path"; then
    api_changed=1
    break
  fi
done

for path in $CONTRACT_PATHS; do
  if ! git diff --quiet "$base_ref" -- "$path"; then
    contract_changed=1
    break
  fi
done

if [[ "$api_changed" -eq 1 && "$contract_changed" -eq 0 ]]; then
  echo "CONTRACT CHECK FAILED: API surface changed but contract files did not." >&2
  echo "Review whether the change should update: $CONTRACT_PATHS" >&2
  exit 1
fi
