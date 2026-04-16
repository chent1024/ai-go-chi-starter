#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash rules-kit/scaffold-local-overrides.sh <target-repo> [--target <name>] [--yes]

behavior:
  - creates missing .orch/rules/<target>/local.env / local.arch.rules / local.docsync.rules files
  - copies selected override-friendly env keys from the current base env into .orch/rules/<target>/local.env
  - does not rewrite or delete base generated files
EOF
  exit 1
}

if [[ $# -lt 1 ]]; then
  usage
fi

target_repo="$1"
shift

target_name=""
yes="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target)
      target_name="${2:-}"
      shift 2
      ;;
    --yes)
      yes="true"
      shift
      ;;
    *)
      usage
      ;;
  esac
done

if [[ ! -d "$target_repo" ]]; then
  echo "scaffold-local-overrides failed: target repository does not exist: $target_repo" >&2
  exit 1
fi

target_repo="$(cd "$target_repo" && pwd)"
manifest_file="$target_repo/.orch/rules/manifest.json"
legacy_manifest_file="$target_repo/.orch/rules/repo/manifest.json"
legacy_env_file="$target_repo/.orch/rules/global.env"
legacy_legacy_env_file="$target_repo/.orch/rules/repo/rules.env"

confirm() {
  local prompt="$1"
  if [[ "$yes" == "true" ]]; then
    return 0
  fi
  read -r -p "$prompt [y/N] " reply
  [[ "$reply" == "y" || "$reply" == "Y" ]]
}

copy_override_key_if_present() {
  local env_file="$1"
  local key="$2"
  python3 - "$env_file" "$key" <<'PY'
from pathlib import Path
import sys

env_file = Path(sys.argv[1])
key = sys.argv[2]
found = ""

for raw_line in env_file.read_text(encoding="utf-8").splitlines():
    line = raw_line.strip()
    if not line or line.startswith("#") or "=" not in line:
        continue
    current_key, value = line.split("=", 1)
    if current_key.strip() != key:
        continue
    value = value.strip()
    if value:
        found = f"{key}={value}"

if found:
    print(found)
PY
}

ensure_local_env() {
  local env_file="$1"
  local local_env_file
  if [[ "$(basename "$env_file")" == "rules.env" ]]; then
    local_env_file="$(dirname "$env_file")/local.env"
  else
    local_env_file="${env_file%.env}.local.env"
  fi
  local appended="false"

  if [[ ! -f "$local_env_file" ]]; then
    cat >"$local_env_file" <<'EOF'
# Repository-specific local overrides preserved across `rules install --force --yes`.
# Keep generated defaults in the base rules.env file and add only repository
# customizations here.
EOF
    appended="true"
  fi

  local key line
  for key in \
    API_SURFACE_PATHS \
    CONTRACT_PATHS \
    CHANGE_SCOPE_BACKEND \
    CHANGE_SCOPE_MODE \
    CHANGE_SCOPE_BASE_REF \
    CHANGE_SCOPE_ROOT \
    CHANGE_SCOPE_ACTIVE_CHANGE \
    CHANGE_SCOPE_METADATA_FILE \
    CHANGE_SCOPE_CODE_PATHS \
    CHANGE_SCOPE_ALLOW_PATHS \
    CONTRACT_PARITY_BACKEND \
    CONTRACT_PARITY_ROUTE_MANIFEST_FILE \
    CONTRACT_PARITY_CONTRACT_MANIFEST_FILE \
    LAYOUT_REQUIRED_DIRS \
    LAYOUT_FORBIDDEN_DIRS
  do
    if rg -q "^${key}=" "$local_env_file"; then
      continue
    fi
    line="$(copy_override_key_if_present "$env_file" "$key")"
    if [[ -n "$line" ]]; then
      if [[ "$appended" != "true" ]]; then
        printf '\n' >>"$local_env_file"
        appended="true"
      fi
      printf '%s\n' "$line" >>"$local_env_file"
    fi
  done

  printf '%s\n' "$local_env_file"
}

ensure_local_rules() {
  local base_file="$1"
  local local_file="$2"
  local header="$3"
  if [[ -f "$local_file" ]]; then
    printf '%s\n' "$local_file"
    return 0
  fi
  cat >"$local_file" <<EOF
$header
# Base file: ${base_file#$target_repo/}
EOF
  printf '%s\n' "$local_file"
}

process_target_env() {
  local env_file="$1"
  local label="$2"
  local arch_file docsync_file local_env_file

  # shellcheck disable=SC1090
  source "$env_file"

  printf 'target: %s\n' "$label"
  local_env_file="$(ensure_local_env "$env_file")"
  printf '  local env: %s\n' "${local_env_file#$target_repo/}"

  if [[ -n "${ARCH_IMPORT_RULES_FILE:-}" ]]; then
    arch_file="$target_repo/$ARCH_IMPORT_RULES_FILE"
    local local_arch_file
    if [[ "$(basename "$arch_file")" == "arch.rules" ]]; then
      local_arch_file="$(dirname "$arch_file")/local.arch.rules"
    else
      local_arch_file="${arch_file%.arch.rules}.local.arch.rules"
    fi
    ensure_local_rules \
      "$arch_file" \
      "$local_arch_file" \
      "# Repository-specific ARCH rule extensions preserved across upgrade." >/dev/null
    printf '  local arch rules: %s\n' "${local_arch_file#$target_repo/}"
  fi

  if [[ -n "${DOC_SYNC_RULES_FILE:-}" ]]; then
    docsync_file="$target_repo/$DOC_SYNC_RULES_FILE"
    local local_docsync_file
    if [[ "$(basename "$docsync_file")" == "docsync.rules" ]]; then
      local_docsync_file="$(dirname "$docsync_file")/local.docsync.rules"
    else
      local_docsync_file="${docsync_file%.docsync.rules}.local.docsync.rules"
    fi
    ensure_local_rules \
      "$docsync_file" \
      "$local_docsync_file" \
      "# Repository-specific DOC_SYNC rule extensions preserved across upgrade." >/dev/null
    printf '  local doc sync rules: %s\n' "${local_docsync_file#$target_repo/}"
  fi
}

if ! confirm "Create or update local override files under $target_repo?"; then
  echo "aborted"
  exit 1
fi

if [[ -f "$manifest_file" || -f "$legacy_manifest_file" ]]; then
  if [[ ! -f "$manifest_file" ]]; then
    manifest_file="$legacy_manifest_file"
  fi
  while IFS=$'\t' read -r name env_rel; do
    [[ -z "$name" ]] && continue
    if [[ -n "$target_name" && "$name" != "$target_name" ]]; then
      continue
    fi
    process_target_env "$target_repo/$env_rel" "target[$name]"
  done < <(
    python3 - "$manifest_file" <<'PY'
import json
import sys
from pathlib import Path

manifest = json.loads(Path(sys.argv[1]).read_text())
for target in manifest.get("targets", []):
    print(target.get("name", ""), target.get("env_file", ""), sep="\t")
PY
  )
elif [[ -f "$legacy_env_file" ]]; then
  process_target_env "$legacy_env_file" "legacy"
elif [[ -f "$legacy_legacy_env_file" ]]; then
  process_target_env "$legacy_legacy_env_file" "legacy"
else
  echo "scaffold-local-overrides failed: no manifest or legacy env found under $target_repo/.orch/rules" >&2
  exit 1
fi

echo "local override scaffolding completed"
