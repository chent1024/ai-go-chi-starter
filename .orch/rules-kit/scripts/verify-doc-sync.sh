#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

trim() {
  printf '%s' "$1" | xargs
}

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "DOC SYNC CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${DOC_SYNC_RULES_FILE:-}" ]]; then
  echo "doc sync checks skipped: DOC_SYNC_RULES_FILE not configured"
  exit 0
fi

if [[ ! -f "$DOC_SYNC_RULES_FILE" ]]; then
  echo "DOC SYNC CHECK FAILED: rules file not found: $DOC_SYNC_RULES_FILE" >&2
  exit 1
fi

default_base_ref="HEAD"
if git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
  default_base_ref="HEAD~1"
fi
base_ref="${DOC_SYNC_BASE_REF:-$default_base_ref}"
if ! git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  base_ref="$default_base_ref"
fi

failures=0

while IFS= read -r rules_file; do
  while IFS='|' read -r code_paths doc_paths detail; do
    [[ -z "${code_paths:-}" || "${code_paths:0:1}" == "#" ]] && continue

    code_changed=0
    docs_changed=0

    IFS=',' read -r -a code_paths_arr <<<"$code_paths"
    for raw_path in "${code_paths_arr[@]}"; do
      path="$(trim "$raw_path")"
      [[ -z "$path" ]] && continue
      if ! git diff --quiet "$base_ref" -- "$path"; then
        code_changed=1
        break
      fi
    done

    if [[ "$code_changed" -eq 0 ]]; then
      continue
    fi

    IFS=',' read -r -a doc_paths_arr <<<"$doc_paths"
    for raw_path in "${doc_paths_arr[@]}"; do
      path="$(trim "$raw_path")"
      [[ -z "$path" ]] && continue
      if ! git diff --quiet "$base_ref" -- "$path"; then
        docs_changed=1
        break
      fi
    done

    if [[ "$docs_changed" -eq 0 ]]; then
      if [[ -n "${detail:-}" ]]; then
        echo "DOC SYNC CHECK FAILED: $detail" >&2
      else
        echo "DOC SYNC CHECK FAILED: code paths changed without synchronized docs update" >&2
      fi
      echo "Expected one of these doc paths to change: $doc_paths" >&2
      echo "Triggered by code paths: $code_paths" >&2
      echo >&2
      failures=1
    fi
  done <"$rules_file"
done < <(rules_list_rule_files "$DOC_SYNC_RULES_FILE")

if [[ "$failures" -ne 0 ]]; then
  exit 1
fi

echo "doc sync checks passed"
