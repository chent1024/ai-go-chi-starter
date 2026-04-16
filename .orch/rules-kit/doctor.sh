#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/doctor.sh <target-repo> [--strict]
  bash rules-kit/doctor.sh <target-repo> [--strict]

checks:
  - required constraint files exist
  - Makefile wiring exists
  - Git hooks are installed in the repository
  - manifest or legacy env paths point at valid files
  - each configured target is checked independently

exit codes:
  0  all required checks passed
  1  required checks failed, or warnings exist in --strict mode
EOF
  exit 1
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
fi

target_repo="$1"
strict="false"
if [[ $# -eq 2 ]]; then
  if [[ "$2" != "--strict" ]]; then
    usage
  fi
  strict="true"
fi

if [[ ! -d "$target_repo" ]]; then
  echo "doctor failed: target repository does not exist: $target_repo" >&2
  exit 1
fi

target_repo="$(cd "$target_repo" && pwd)"
manifest_file="$target_repo/.orch/rules/manifest.json"
legacy_manifest_file="$target_repo/.orch/rules/repo/manifest.json"
legacy_env_file="$target_repo/.orch/rules/global.env"
legacy_legacy_env_file="$target_repo/.orch/rules/repo/rules.env"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
manifest_validator="$script_dir/validate-manifest.sh"

errors=0
warnings=0
color_pass=""
color_warn=""
color_fail=""
color_reset=""

if [[ -t 1 && -z "${NO_COLOR:-}" && "${TERM:-}" != "dumb" ]]; then
  color_pass=$'\033[32m'
  color_warn=$'\033[33m'
  color_fail=$'\033[31m'
  color_reset=$'\033[0m'
fi

print_status() {
  local stream="$1"
  local color="$2"
  local label="$3"
  local message="$4"
  if [[ -n "$color" ]]; then
    printf '%b%s%b: %s\n' "$color" "$label" "$color_reset" "$message" >&"$stream"
    return
  fi
  printf '%s: %s\n' "$label" "$message" >&"$stream"
}

pass() {
  print_status 1 "$color_pass" "PASS" "$1"
}

warn() {
  print_status 1 "$color_warn" "WARN" "$1"
  warnings=1
}

fail() {
  print_status 2 "$color_fail" "FAIL" "$1"
  errors=1
}

check_file() {
  local rel="$1"
  if [[ -e "$target_repo/$rel" ]]; then
    pass "$rel exists"
  else
    fail "$rel is missing"
  fi
}

path_exists_any() {
  local base="$1"
  shift
  local path
  for path in "$@"; do
    [[ -z "$path" ]] && continue
    if [[ -e "$base/$path" ]]; then
      return 0
    fi
  done
  return 1
}

check_path_list_any() {
  local label="$1"
  local base="$2"
  local required="$3"
  shift 3
  local paths=("$@")
  if [[ "${#paths[@]}" -eq 0 ]]; then
    if [[ "$required" == "true" ]]; then
      fail "$label is not configured"
    else
      warn "$label is not configured"
    fi
    return
  fi
  if path_exists_any "$base" "${paths[@]}"; then
    pass "$label points to an existing path"
  else
    if [[ "$required" == "true" ]]; then
      fail "$label does not point to an existing path: ${paths[*]}"
    else
      warn "$label does not point to an existing path: ${paths[*]}"
    fi
  fi
}

check_common_repo() {
  if command -v orch >/dev/null 2>&1; then
    pass "orch command available on PATH"
  else
    pass "orch command not found on PATH (optional when invoking bundled scripts directly)"
  fi

  check_file "AGENTS.md"
  check_file ".githooks/pre-commit"
  check_file ".githooks/pre-push"

  if [[ -f "$target_repo/Makefile" ]]; then
    if rg -q 'Makefile\.rules|\.orch/rules/make/verify\.mk|verify-config-docs|verify-generated|verify-secrets' "$target_repo/Makefile"; then
      pass "Makefile references rules verification"
    else
      fail "Makefile does not reference rules verification"
    fi
  elif [[ -f "$target_repo/Makefile.rules" ]]; then
    warn "Makefile is missing; Makefile.rules exists but is not wired"
  else
    fail "Makefile and generated rules makefile are both missing"
  fi

  if [[ -f "$target_repo/Makefile.rules" ]]; then
    pass "Makefile.rules exists"
  fi
}

check_target_env() {
  local env_file="$1"
  local target_label="$2"
  local env_dir env_name local_env_file local_arch_rules_file local_doc_sync_rules_file

  if [[ ! -f "$env_file" ]]; then
    fail "$target_label env file is missing: ${env_file#$target_repo/}"
    return
  fi

  # shellcheck disable=SC1090
  source "$env_file"
  env_dir="$(dirname "$env_file")"
  env_name="$(basename "$env_file")"
  if [[ "$env_name" == "rules.env" ]]; then
    local_env_file="$env_dir/local.env"
  else
    local_env_file="$env_dir/${env_name%.env}.local.env"
  fi
  if [[ -f "$local_env_file" ]]; then
    # shellcheck disable=SC1090
    source "$local_env_file"
    pass "$target_label local env override exists"
  else
    pass "$target_label local env override not configured (optional)"
  fi

  read -r -a config_sources_arr <<<"${CONFIG_SOURCES:-}"
  check_path_list_any "$target_label CONFIG_SOURCES" "$target_repo" "true" "${config_sources_arr[@]}"

  if [[ -n "${ENV_EXAMPLE_FILE:-}" ]]; then
    if [[ -e "$target_repo/$ENV_EXAMPLE_FILE" ]]; then
      pass "$target_label ENV_EXAMPLE_FILE exists"
    else
      fail "$target_label ENV_EXAMPLE_FILE is missing: $ENV_EXAMPLE_FILE"
    fi
  elif [[ "${TARGET_LANGUAGE:-}" == "py" ]]; then
    pass "$target_label ENV_EXAMPLE_FILE not configured (optional for python targets)"
  else
    fail "$target_label ENV_EXAMPLE_FILE is not configured"
  fi

  if [[ -n "${API_SURFACE_PATHS:-}" ]]; then
    read -r -a api_paths_arr <<<"${API_SURFACE_PATHS:-}"
    check_path_list_any "$target_label API_SURFACE_PATHS" "$target_repo" "false" "${api_paths_arr[@]}"
    if [[ -f "$local_env_file" ]] && [[ "${API_SURFACE_PATHS:-}" != "" ]]; then
      pass "$target_label effective API_SURFACE_PATHS=${API_SURFACE_PATHS}"
    fi
  fi

  if [[ -n "${CONTRACT_PATHS:-}" ]]; then
    read -r -a contract_paths_arr <<<"${CONTRACT_PATHS:-}"
    check_path_list_any "$target_label CONTRACT_PATHS" "$target_repo" "false" "${contract_paths_arr[@]}"
    if [[ -f "$local_env_file" ]] && [[ "${CONTRACT_PATHS:-}" != "" ]]; then
      pass "$target_label effective CONTRACT_PATHS=${CONTRACT_PATHS}"
    fi
  fi

  if [[ -n "${MIGRATIONS_DIR:-}" ]]; then
    if [[ -d "$target_repo/$MIGRATIONS_DIR" ]]; then
      pass "$target_label MIGRATIONS_DIR exists"
    else
      warn "$target_label MIGRATIONS_DIR is missing: $MIGRATIONS_DIR"
    fi
  fi

  if [[ -n "${ARCH_IMPORT_RULES_FILE:-}" ]]; then
    if [[ -f "$target_repo/$ARCH_IMPORT_RULES_FILE" ]]; then
      pass "$target_label ARCH_IMPORT_RULES_FILE exists"
    else
      fail "$target_label ARCH_IMPORT_RULES_FILE is configured but missing: $ARCH_IMPORT_RULES_FILE"
    fi
    if [[ "$(basename "$ARCH_IMPORT_RULES_FILE")" == "arch.rules" ]]; then
      local_arch_rules_file="$(dirname "$target_repo/$ARCH_IMPORT_RULES_FILE")/local.arch.rules"
    else
      local_arch_rules_file="$target_repo/${ARCH_IMPORT_RULES_FILE%.arch.rules}.local.arch.rules"
    fi
    if [[ -f "$local_arch_rules_file" ]]; then
      pass "$target_label local ARCH rules override exists"
    else
      pass "$target_label local ARCH rules override not configured (optional)"
    fi
  else
    warn "$target_label ARCH_IMPORT_RULES_FILE is empty"
  fi

  if [[ -n "${DOC_SYNC_RULES_FILE:-}" ]]; then
    if [[ -f "$target_repo/$DOC_SYNC_RULES_FILE" ]]; then
      pass "$target_label DOC_SYNC_RULES_FILE exists"
    else
      fail "$target_label DOC_SYNC_RULES_FILE is configured but missing: $DOC_SYNC_RULES_FILE"
    fi
    if [[ "$(basename "$DOC_SYNC_RULES_FILE")" == "docsync.rules" ]]; then
      local_doc_sync_rules_file="$(dirname "$target_repo/$DOC_SYNC_RULES_FILE")/local.docsync.rules"
    else
      local_doc_sync_rules_file="$target_repo/${DOC_SYNC_RULES_FILE%.docsync.rules}.local.docsync.rules"
    fi
    if [[ -f "$local_doc_sync_rules_file" ]]; then
      pass "$target_label local DOC_SYNC rules override exists"
    else
      pass "$target_label local DOC_SYNC rules override not configured (optional)"
    fi
  else
    pass "$target_label DOC_SYNC_RULES_FILE not configured (optional)"
  fi

  if [[ -n "${CONTRACT_PARITY_BACKEND:-}" ]]; then
    pass "$target_label CONTRACT_PARITY_BACKEND configured: ${CONTRACT_PARITY_BACKEND}"
    if [[ -n "${CONTRACT_PARITY_ROUTE_MANIFEST_FILE:-}" ]]; then
      if [[ -f "$target_repo/$CONTRACT_PARITY_ROUTE_MANIFEST_FILE" ]]; then
        pass "$target_label CONTRACT_PARITY_ROUTE_MANIFEST_FILE exists"
      else
        fail "$target_label CONTRACT_PARITY_ROUTE_MANIFEST_FILE is configured but missing: $CONTRACT_PARITY_ROUTE_MANIFEST_FILE"
      fi
    fi
    if [[ -n "${CONTRACT_PARITY_CONTRACT_MANIFEST_FILE:-}" ]]; then
      if [[ -f "$target_repo/$CONTRACT_PARITY_CONTRACT_MANIFEST_FILE" ]]; then
        pass "$target_label CONTRACT_PARITY_CONTRACT_MANIFEST_FILE exists"
      else
        fail "$target_label CONTRACT_PARITY_CONTRACT_MANIFEST_FILE is configured but missing: $CONTRACT_PARITY_CONTRACT_MANIFEST_FILE"
      fi
    fi
  else
    pass "$target_label CONTRACT_PARITY_BACKEND not configured (optional)"
  fi

  if [[ -n "${CHANGE_SCOPE_BACKEND:-}" ]]; then
    pass "$target_label CHANGE_SCOPE_BACKEND configured: ${CHANGE_SCOPE_BACKEND}"
    pass "$target_label CHANGE_SCOPE_MODE=${CHANGE_SCOPE_MODE:-strict}"
    if [[ -n "${CHANGE_SCOPE_CODE_PATHS:-}" ]]; then
      pass "$target_label CHANGE_SCOPE_CODE_PATHS=${CHANGE_SCOPE_CODE_PATHS}"
    fi
    if [[ -n "${CHANGE_SCOPE_ALLOW_PATHS:-}" ]]; then
      pass "$target_label CHANGE_SCOPE_ALLOW_PATHS=${CHANGE_SCOPE_ALLOW_PATHS}"
    fi
  else
    pass "$target_label CHANGE_SCOPE_BACKEND not configured (optional)"
  fi

  if [[ -n "${LAYOUT_REQUIRED_DIRS:-}" ]]; then
    pass "$target_label LAYOUT_REQUIRED_DIRS=${LAYOUT_REQUIRED_DIRS}"
  fi
  if [[ -n "${LAYOUT_FORBIDDEN_DIRS:-}" ]]; then
    pass "$target_label LAYOUT_FORBIDDEN_DIRS=${LAYOUT_FORBIDDEN_DIRS}"
  fi

  case "${TARGET_LANGUAGE:-}" in
    go)
      if [[ -f "$target_repo/${TARGET_ROOT:-.}/go.mod" || ("${TARGET_ROOT:-.}" == "." && -f "$target_repo/go.mod") ]]; then
        pass "$target_label Go marker detected"
      else
        warn "$target_label expected go.mod but did not find one under target root"
      fi
      ;;
    ts)
      if [[ -f "$target_repo/${TARGET_ROOT:-.}/package.json" || ("${TARGET_ROOT:-.}" == "." && -f "$target_repo/package.json") ]]; then
        pass "$target_label TypeScript marker detected"
      else
        warn "$target_label expected package.json but did not find one under target root"
      fi
      ;;
    py)
      if [[ -f "$target_repo/${TARGET_ROOT:-.}/pyproject.toml" || -f "$target_repo/${TARGET_ROOT:-.}/requirements.txt" || -f "$target_repo/${TARGET_ROOT:-.}/setup.py" || -f "$target_repo/${TARGET_ROOT:-.}/setup.cfg" || ("${TARGET_ROOT:-.}" == "." && ( -f "$target_repo/pyproject.toml" || -f "$target_repo/requirements.txt" || -f "$target_repo/setup.py" || -f "$target_repo/setup.cfg" )) ]]; then
        pass "$target_label Python marker detected"
      else
        warn "$target_label expected pyproject.toml/requirements.txt/setup.py under target root"
      fi
      ;;
  esac
}

check_manifest_mode() {
  if [[ "$manifest_file" == "$target_repo/.orch/rules/manifest.json" ]]; then
    check_file ".orch/rules/manifest.json"
  else
    check_file ".orch/rules/repo/manifest.json"
  fi
  check_common_repo
  if [[ -f "$target_repo/.orch/rules/global.env" ]]; then
    check_file ".orch/rules/global.env"
  else
    check_file ".orch/rules/repo/global.env"
  fi

  if [[ -f "$manifest_validator" ]]; then
    if bash "$manifest_validator" "$manifest_file" >/dev/null; then
      pass "manifest schema is valid"
    else
      fail "manifest schema validation failed"
      return
    fi
  else
    fail "bundled validate-manifest.sh is missing"
    return
  fi

  while IFS=$'\t' read -r name language root env_file; do
    [[ -z "${name:-}" ]] && continue
    pass "target configured: $name ($language at $root)"
    check_target_env "$target_repo/$env_file" "target[$name]"
  done < <(
    python3 - "$manifest_file" <<'PY'
import json
import sys
from pathlib import Path

manifest = json.loads(Path(sys.argv[1]).read_text())
for target in manifest.get("targets", []):
    print(
        target.get("name", ""),
        target.get("language", ""),
        target.get("root", ""),
        target.get("env_file", ""),
        sep="\t",
    )
PY
  )
}

check_legacy_mode() {
  if [[ -f "$legacy_env_file" ]]; then
    check_file ".orch/rules/global.env"
    check_target_env "$legacy_env_file" "legacy"
  else
    check_file ".orch/rules/repo/rules.env"
    check_target_env "$legacy_legacy_env_file" "legacy"
  fi
  check_common_repo

  if [[ -f "$target_repo/.orch/rules/lint/.golangci.base.yml" ]]; then
    warn "lint base exists but starter marker file is absent; verify profile selection manually"
  fi
}

if [[ -f "$manifest_file" || -f "$legacy_manifest_file" ]]; then
  if [[ ! -f "$manifest_file" ]]; then
    manifest_file="$legacy_manifest_file"
  fi
  check_manifest_mode
elif [[ -f "$legacy_env_file" || -f "$legacy_legacy_env_file" ]]; then
  check_legacy_mode
else
  fail "neither .orch/rules/manifest.json nor a legacy rules env exists"
fi

if [[ "$errors" -ne 0 ]]; then
  exit 1
fi

if [[ "$warnings" -ne 0 && "$strict" == "true" ]]; then
  echo "doctor strict mode failed due to warnings" >&2
  exit 1
fi

echo "doctor completed"
if [[ "$warnings" -ne 0 ]]; then
  echo "doctor completed with warnings"
fi
