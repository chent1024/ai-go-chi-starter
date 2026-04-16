#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash rules-kit/preflight.sh [target-repo] [--target <name>]
EOF
  exit 1
}

target_repo="."
target_name=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target)
      target_name="${2:-}"
      shift 2
      ;;
    -*)
      usage
      ;;
    *)
      target_repo="$1"
      shift
      ;;
  esac
done

if [[ ! -d "$target_repo" ]]; then
  echo "preflight failed: target repository does not exist: $target_repo" >&2
  exit 1
fi

target_repo="$(cd "$target_repo" && pwd)"
manifest_file="$target_repo/.orch/rules/manifest.json"
legacy_manifest_file="$target_repo/.orch/rules/repo/manifest.json"
global_env_file="$target_repo/.orch/rules/global.env"
legacy_env_file="$target_repo/.orch/rules/repo/rules.env"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

run_target_env() {
  local env_file="$1"
  bash "$script_dir/scripts/verify-openspec-required.sh" "$env_file"
  bash "$script_dir/scripts/verify-change-scope.sh" "$env_file"
}

manifest_path=""
if [[ -f "$manifest_file" ]]; then
  manifest_path="$manifest_file"
elif [[ -f "$legacy_manifest_file" ]]; then
  manifest_path="$legacy_manifest_file"
fi

if [[ -n "$manifest_path" ]]; then
  matched_target="false"
  while IFS=$'\t' read -r name env_rel; do
    [[ -z "$name" || -z "$env_rel" ]] && continue
    if [[ -n "$target_name" && "$name" != "$target_name" ]]; then
      continue
    fi
    matched_target="true"
    run_target_env "$target_repo/$env_rel"
  done < <(
    python3 - "$manifest_path" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
for target in payload.get("targets", []):
    print(target.get("name", ""), target.get("env_file", ""), sep="\t")
PY
  )
  if [[ -n "$target_name" && "$matched_target" != "true" ]]; then
    echo "preflight failed: target not found: $target_name" >&2
    exit 1
  fi
  exit 0
fi

if [[ -f "$global_env_file" ]]; then
  run_target_env "$global_env_file"
  exit 0
fi

if [[ -f "$legacy_env_file" ]]; then
  run_target_env "$legacy_env_file"
  exit 0
fi

echo "preflight failed: no manifest or global env found under $target_repo/.orch/rules" >&2
exit 1
