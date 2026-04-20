#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/install.sh <target-repo> [options]
  bash rules-kit/install.sh <target-repo> [options]
  bash tools/rules-kit/install.sh <output-dir> <legacy-profile>
  bash rules-kit/install.sh <output-dir> <legacy-profile>

new install options:
  --manifest <manifest.json>
  --mode existing|new
  --layout single|monorepo
  --targets "<name>:<go|ts|py>:<root>[,<name>:<go|ts|py>:<root>...]"
  --yes
  --force
  --dry-run

legacy profiles:
  minimal
  starter
  go-starter
  ts-starter

examples:
  bash tools/rules-kit/install.sh /repo --mode existing --targets "backend:go:app,frontend:ts:web"
  bash rules-kit/install.sh /repo --mode existing --targets "backend:go:app,frontend:ts:web"
  bash tools/rules-kit/install.sh /repo --mode new --targets "app:go:."
  bash rules-kit/install.sh /repo --mode new --targets "app:go:."
  bash tools/rules-kit/install.sh /repo --manifest ./tools/rules-kit/repo/manifest.example.json
  bash rules-kit/install.sh /repo --manifest ./rules-kit/repo/manifest.example.json
  bash tools/rules-kit/install.sh /repo --dry-run --mode new --layout monorepo --targets "backend:go:app,frontend:ts:web"
  bash rules-kit/install.sh /repo --dry-run --mode new --layout monorepo --targets "backend:go:app,frontend:ts:web"
  bash tools/rules-kit/install.sh /repo starter
  bash rules-kit/install.sh /repo starter
EOF
  exit 1
}

if [[ $# -lt 1 ]]; then
  usage
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir="$script_dir"

# shellcheck disable=SC1091
source "$root_dir/install_lib.sh"

validate_manifest_script="$root_dir/validate-manifest.sh"
detect_repo_script="$root_dir/detect-repo.py"

resolve_path() {
  python3 - "$1" <<'PY'
from pathlib import Path
import sys

print(Path(sys.argv[1]).resolve())
PY
}

validate_manifest_file() {
  local manifest_file="$1"
  bash "$validate_manifest_script" "$manifest_file"
}

load_manifest_value() {
  local manifest_file="$1"
  local key="$2"
  python3 - "$manifest_file" "$key" <<'PY'
import json
import sys
from pathlib import Path

manifest = json.loads(Path(sys.argv[1]).read_text())
value = manifest.get(sys.argv[2], "")
if isinstance(value, str):
    print(value)
PY
}

load_manifest_targets() {
  local manifest_file="$1"
  python3 - "$manifest_file" <<'PY'
import json
import sys
from pathlib import Path

manifest = json.loads(Path(sys.argv[1]).read_text())
items = []
for target in manifest.get("targets", []):
    name = target.get("name", "")
    language = target.get("language", "")
    root = target.get("root", "")
    if name and language and root:
        items.append(f"{name}:{language}:{root}")
print(",".join(items))
PY
}

copy_githooks_dir() {
  local output_dir="$1"
  rules_copy_dir "$root_dir/githooks" "$output_dir/.githooks"
}

cleanup_deprecated_managed_paths() {
  local repo="$1"
  if [[ "$force" != "true" || "${RULES_DRY_RUN:-false}" == "true" ]]; then
    return
  fi

  rm -rf \
    "$repo/.github" \
    "$repo/rules-kit" \
    "$repo/.orch/rules/ci" \
    "$repo/.orch/rules/make" \
    "$repo/.orch/rules/review"
  rm -f \
    "$repo/.orch/rules/AGENTS.base.md" \
    "$repo/.orch/rules/GO_STARTER_README.md" \
    "$repo/.orch/rules/TS_STARTER_README.md" \
    "$repo/.orch/rules/lint/gitleaks.toml" \
    "$repo/.orch/rules/lint/semgrep.yml"
}

install_minimal_legacy() {
  local output_dir="$1"
  local overwrite="$2"
  rules_mkdir_p "$output_dir/.orch/rules"
  rules_mkdir_p "$output_dir/.githooks"
  rules_mkdir_p "$output_dir/.orch/rules/repo"
  rules_mkdir_p "$output_dir/.orch/rules/scripts"

  rules_upsert_agents_file "$root_dir/minimal/AGENTS.base.md" "$output_dir/AGENTS.md" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/minimal/Makefile.snippet" "$output_dir/Makefile.template" "$overwrite"
  copy_githooks_dir "$output_dir"
  rules_copy_file_if_allowed "$root_dir/RULES_README.md" "$output_dir/.orch/rules/README.md" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/minimal/repo/rules.env" "$output_dir/.orch/rules/repo/rules.env" "$overwrite"
  rules_copy_dir "$root_dir/minimal/scripts" "$output_dir/.orch/rules/scripts"
}

install_starter_legacy() {
  local output_dir="$1"
  local overwrite="$2"
  rules_mkdir_p "$output_dir/.orch/rules"
  rules_mkdir_p "$output_dir/.githooks"
  rules_mkdir_p "$output_dir/.orch/rules/lint"
  rules_mkdir_p "$output_dir/.orch/rules/repo"
  rules_mkdir_p "$output_dir/.orch/rules/scripts"

  rules_upsert_agents_file "$root_dir/starter-kit/AGENTS.md.template" "$output_dir/AGENTS.md" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/starter-kit/Makefile.template" "$output_dir/Makefile.template" "$overwrite"
  copy_githooks_dir "$output_dir"
  rules_copy_file_if_allowed "$root_dir/RULES_README.md" "$output_dir/.orch/rules/README.md" "$overwrite"

  rules_copy_dir "$root_dir/scripts" "$output_dir/.orch/rules/scripts"
  rules_copy_file_if_allowed "$root_dir/lint/.golangci.base.yml" "$output_dir/.orch/rules/lint/.golangci.base.yml" "$overwrite"
}

install_go_starter_legacy() {
  local output_dir="$1"
  local overwrite="$2"
  install_starter_legacy "$output_dir" "$overwrite"
  rules_upsert_agents_file "$root_dir/go-starter/AGENTS.md.template" "$output_dir/AGENTS.md" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/go-starter/Makefile.template" "$output_dir/Makefile.template" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/go-starter/rules.env" "$output_dir/.orch/rules/repo/rules.env" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/go-starter/.golangci.yml" "$output_dir/.orch/rules/lint/.golangci.base.yml" "$overwrite"
}

install_ts_starter_legacy() {
  local output_dir="$1"
  local overwrite="$2"
  install_starter_legacy "$output_dir" "$overwrite"
  rules_upsert_agents_file "$root_dir/ts-starter/AGENTS.md.template" "$output_dir/AGENTS.md" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/ts-starter/Makefile.template" "$output_dir/Makefile.template" "$overwrite"
  rules_copy_file_if_allowed "$root_dir/ts-starter/rules.env" "$output_dir/.orch/rules/repo/rules.env" "$overwrite"
}

legacy_install() {
  local output_dir="$1"
  local profile="$2"
  local overwrite="${3:-false}"
  case "$profile" in
    minimal)
      install_minimal_legacy "$output_dir" "$overwrite"
      ;;
    starter)
      install_starter_legacy "$output_dir" "$overwrite"
      ;;
    go-starter)
      install_go_starter_legacy "$output_dir" "$overwrite"
      ;;
    ts-starter)
      install_ts_starter_legacy "$output_dir" "$overwrite"
      ;;
    *)
      echo "unknown legacy profile: $profile" >&2
      usage
      ;;
  esac
  echo "rules installed to: $output_dir"
  echo "profile: $profile"
}

if [[ $# -ge 2 && "${2:0:1}" != "-" ]]; then
  case "$2" in
    minimal|starter|go-starter|ts-starter)
      legacy_install "$1" "$2"
      exit 0
      ;;
  esac
fi

target_repo="$1"
shift

mode=""
layout=""
targets_spec=""
manifest_input=""
yes="false"
force="false"
dry_run="false"
mode_explicit="false"
layout_explicit="false"
targets_explicit="false"
detected_mode=""
detected_layout=""
detected_targets_spec=""
detected_confidence=""
detected_ambiguous="false"
detected_reasons=""
detected_warnings=""
printed_detection_summary="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --manifest)
      manifest_input="${2:-}"
      shift 2
      ;;
    --mode)
      mode="${2:-}"
      mode_explicit="true"
      shift 2
      ;;
    --layout)
      layout="${2:-}"
      layout_explicit="true"
      shift 2
      ;;
    --targets)
      targets_spec="${2:-}"
      targets_explicit="true"
      shift 2
      ;;
    --yes)
      yes="true"
      shift
      ;;
    --force)
      force="true"
      shift
      ;;
    --dry-run)
      dry_run="true"
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      ;;
  esac
done

export RULES_DRY_RUN="$dry_run"
export RULES_ASSUME_YES="$yes"
target_repo="$(resolve_path "$target_repo")"

if [[ -n "$manifest_input" ]]; then
  if [[ ! -f "$manifest_input" ]]; then
    echo "manifest file does not exist: $manifest_input" >&2
    exit 1
  fi
  validate_manifest_file "$manifest_input"
  if [[ -z "$mode" ]]; then
    mode="$(load_manifest_value "$manifest_input" repo_mode)"
  fi
  if [[ -z "$layout" ]]; then
    layout="$(load_manifest_value "$manifest_input" layout)"
  fi
  if [[ -z "$targets_spec" ]]; then
    targets_spec="$(load_manifest_targets "$manifest_input")"
  fi
fi

if [[ -z "$mode" || -z "$layout" || -z "$targets_spec" ]]; then
  detected_mode="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "repo_mode")"
  detected_layout="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "layout")"
  detected_targets_spec="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "targets_spec")"
  detected_confidence="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "confidence")"
  detected_ambiguous="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "ambiguous")"
  detected_reasons="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "reasons_text")"
  detected_warnings="$(rules_detect_repo_field "$detect_repo_script" "$target_repo" "warnings_text")"

  if [[ -z "$mode" && -n "$detected_mode" ]]; then
    mode="$detected_mode"
    printed_detection_summary="true"
  fi
  if [[ -z "$layout" && -n "$detected_layout" ]]; then
    layout="$detected_layout"
    printed_detection_summary="true"
  fi
  if [[ -z "$targets_spec" && -n "$detected_targets_spec" ]]; then
    targets_spec="$detected_targets_spec"
    printed_detection_summary="true"
  fi
fi

if [[ "$printed_detection_summary" == "true" ]]; then
  rules_print_detection_summary "$detected_mode" "$detected_layout" "$detected_targets_spec" "$detected_reasons" "$detected_warnings"
fi

if [[ "$targets_explicit" != "true" && "$detected_ambiguous" == "true" ]]; then
  echo "auto-detect is ambiguous; pass --targets explicitly" >&2
  if [[ -n "$detected_reasons" ]]; then
    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      printf 'reason: %s\n' "$line" >&2
    done <<<"$detected_reasons"
  fi
  if [[ -n "$detected_warnings" ]]; then
    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      printf 'warning: %s\n' "$line" >&2
    done <<<"$detected_warnings"
  fi
  exit 1
fi

rules_prompt_if_needed mode "Repository mode (existing/new)" "existing"
rules_prompt_if_needed layout "Repository layout (single/monorepo)" "single"

case "$mode" in
  existing|new) ;;
  *)
    echo "unsupported mode: $mode" >&2
    exit 1
    ;;
esac

case "$layout" in
  single|monorepo) ;;
  *)
    echo "unsupported layout: $layout" >&2
    exit 1
    ;;
esac

if [[ -z "$targets_spec" ]]; then
  if [[ "$yes" == "true" || ! -t 0 ]]; then
    if [[ "$layout" == "single" ]]; then
      targets_spec="app:go:."
    else
      echo "cannot infer targets for monorepo; pass --targets" >&2
      exit 1
    fi
  else
    if [[ "$layout" == "single" ]]; then
      read -r -p "Target language (go/ts/py) [go]: " lang
      lang="${lang:-go}"
      read -r -p "Target name [app]: " tname
      tname="${tname:-app}"
      read -r -p "Target root [.]: " troot
      troot="${troot:-.}"
      targets_spec="${tname}:${lang}:${troot}"
    else
      read -r -p 'Targets (<name>:<go|ts|py>:<root>, comma-separated): ' targets_spec
    fi
  fi
fi

target_names=()
target_langs=()
target_roots=()

parse_targets_spec() {
  local entries=()
  local seen_names=" "
  local seen_roots=" "
  IFS=',' read -r -a entries <<<"$targets_spec"

  local entry tname tlang troot extra normalized_root
  for entry in "${entries[@]}"; do
    [[ -z "$entry" ]] && continue
    IFS=':' read -r tname tlang troot extra <<<"$entry"
    if [[ -z "${tname:-}" || -z "${tlang:-}" || -z "${troot:-}" || -n "${extra:-}" ]]; then
      echo "invalid target spec: $entry" >&2
      exit 1
    fi
    if [[ "$tlang" != "go" && "$tlang" != "ts" && "$tlang" != "py" ]]; then
      echo "unsupported target language for $tname: $tlang" >&2
      exit 1
    fi
    normalized_root="$(rules_normalize_root "$troot")"
    if [[ " $seen_names " == *" $tname "* ]]; then
      echo "duplicate target name: $tname" >&2
      exit 1
    fi
    if [[ " $seen_roots " == *" $normalized_root "* ]]; then
      echo "duplicate target root: $normalized_root" >&2
      exit 1
    fi
    seen_names+=" $tname"
    seen_roots+=" $normalized_root"
    target_names+=("$tname")
    target_langs+=("$tlang")
    target_roots+=("$normalized_root")
  done

  if [[ "${#target_names[@]}" -eq 0 ]]; then
    echo "no targets resolved from --targets" >&2
    exit 1
  fi
  if [[ "$layout" == "single" ]]; then
    if [[ "${#target_names[@]}" -ne 1 ]]; then
      echo "single layout requires exactly one target" >&2
      exit 1
    fi
    if [[ "${target_roots[0]}" != "." ]]; then
      echo "single layout requires target root '.'" >&2
      exit 1
    fi
  fi
}

parse_targets_spec

render_go_layout_required_dirs() {
  local prefix="$1"
  printf '%s' "${prefix}cmd ${prefix}internal/config ${prefix}internal/runtime ${prefix}internal/transport/httpapi ${prefix}internal/service ${prefix}internal/infra ${prefix}openapi ${prefix}db/migrations"
}

render_go_layout_forbidden_dirs() {
  local prefix="$1"
  printf '%s' "${prefix}internal/handlers ${prefix}internal/repository ${prefix}internal/storage ${prefix}internal/upload ${prefix}internal/misc ${prefix}internal/tmp"
}

render_arch_rules_content() {
  local lang="$1"
  local root="${2:-.}"
  local prefix
  prefix="$(rules_path_prefix_for_root "$root")"
  if [[ "$lang" == "go" ]]; then
    cat <<EOF
# format:
# <scope>|<forbidden_scope>|<message>[|<recommended fix>]
#
# default layered Go service rules:
# Add repository-specific extensions in ${prefix}.orch/rules/<target>/local.arch.rules.
${prefix}internal/service|${prefix}internal/transport|service must not depend on transport
${prefix}internal/transport|${prefix}internal/infra/store|transport must not depend on concrete stores
${prefix}internal/transport|${prefix}internal/infra/objectstore|transport must not depend on object store adapters
${prefix}internal/transport|${prefix}internal/infra/provider|transport must not depend on provider adapters
${prefix}internal/runtime|${prefix}internal/transport|runtime must not depend on transport
${prefix}internal/infra|${prefix}internal/transport|infrastructure must not depend on transport
EOF
    return
  fi

  cat <<'EOF'
# format:
# <scope>|<regex>|<message>[|<recommended fix>]
#
# examples:
# src/domain|src/http|domain must not depend on transport
# src/http|src/store|transport must not depend on concrete stores
EOF
}

render_doc_sync_rules_content() {
  local lang="$1"
  local root="${2:-.}"
  local prefix
  prefix="$(rules_path_prefix_for_root "$root")"

  if [[ "$lang" == "go" ]]; then
    cat <<EOF
# Optional code-to-doc sync rules.
# Format:
# <code path>[,<code path>...]|<doc path>[,<doc path>...]|<message>
# Add repository-specific extensions in ${prefix}.orch/rules/<target>/local.docsync.rules.
#
# Examples:
# ${prefix}internal/config,${prefix}.env.example|docs/config.md|config changes must update docs
# ${prefix}internal/transport/httpapi,${prefix}openapi|docs/api.md|api or contract changes must update docs
EOF
    return
  fi

  cat <<EOF
# Optional code-to-doc sync rules.
# Format:
# <code path>[,<code path>...]|<doc path>[,<doc path>...]|<message>
#
# Examples:
# ${prefix}src/config|docs/config.md|config changes must update docs
# ${prefix}src/http,${prefix}openapi|docs/api.md|api or contract changes must update docs
EOF
}

render_target_env_content() {
  local name="$1"
  local lang="$2"
  local root="$3"
  local prefix config_sources env_example api_paths contract_paths migrations_dir arch_file doc_sync_file generated_markers arch_backend
  local go_layout_profile layout_required_dirs layout_forbidden_dirs

  prefix="$(rules_path_prefix_for_root "$root")"
  env_example="${prefix}.env.example"
  contract_paths="${prefix}openapi ${prefix}openapi/openapi.yaml ${prefix}api"
  migrations_dir="${prefix}db/migrations"
  arch_file=".orch/rules/$name/arch.rules"
  doc_sync_file=".orch/rules/$name/docsync.rules"
  generated_markers='Code generated|DO NOT'" EDIT|AUTO-GENERATED FILE"

  if [[ "$lang" == "go" ]]; then
    config_sources="${prefix}internal/config/config.go"
    if [[ -f "$target_repo/deploy/.env.dev.example" ]]; then
      env_example="deploy/.env.dev.example"
    fi
    api_paths="${prefix}internal/httpapi ${prefix}internal/transport/httpapi/router.go ${prefix}internal/transport/httpapi/public ${prefix}internal/transport/httpapi/internal ${prefix}internal/transport/httpapi/httpx ${prefix}internal/transport/httpapi/security ${prefix}internal/api"
    arch_backend="go_imports"
    go_layout_profile="service_layered"
    if [[ "$mode" == "new" ]]; then
      layout_required_dirs="$(render_go_layout_required_dirs "$prefix")"
      layout_forbidden_dirs="$(render_go_layout_forbidden_dirs "$prefix")"
    else
      layout_required_dirs=""
      layout_forbidden_dirs=""
    fi
  elif [[ "$lang" == "ts" ]]; then
    config_sources="${prefix}src/config/env.ts"
    api_paths="${prefix}src/http ${prefix}src/routes ${prefix}src/api"
    arch_backend="text"
    go_layout_profile=""
    layout_required_dirs=""
    layout_forbidden_dirs=""
  else
    config_sources="${prefix}pyproject.toml"
    env_example=""
    api_paths=""
    contract_paths=""
    migrations_dir=""
    arch_backend="text"
    go_layout_profile=""
    layout_required_dirs=""
    layout_forbidden_dirs=""
  fi

  cat <<EOF
REPO_ROOT="."
TARGET_NAME="$name"
TARGET_LANGUAGE="$lang"
TARGET_ROOT="$root"

VERIFY_CMD="make verify-$name"
STRICT_VERIFY_CMD="make verify-strict-$name"

CONFIG_SOURCES="$config_sources"
ENV_EXAMPLE_FILE="$env_example"
ALLOW_STALE_ENV_EXAMPLE_KEYS="false"

API_SURFACE_PATHS="$api_paths"
CONTRACT_PATHS="$contract_paths"
CONTRACT_SYNC_BASE_REF="HEAD~1"
CONTRACT_PARITY_BACKEND=""
CONTRACT_PARITY_ROUTE_MANIFEST_FILE=""
CONTRACT_PARITY_CONTRACT_MANIFEST_FILE=""
CHANGE_SCOPE_BACKEND=""
CHANGE_SCOPE_MODE="strict"
CHANGE_SCOPE_BASE_REF="HEAD~1"
CHANGE_SCOPE_ROOT="openspec/changes"
CHANGE_SCOPE_ACTIVE_CHANGE=""
CHANGE_SCOPE_METADATA_FILE=".openspec.yaml"
CHANGE_SCOPE_CODE_PATHS=""
CHANGE_SCOPE_ALLOW_PATHS=""
OPENSPEC_REQUIRED_MODE="smart"
OPENSPEC_REQUIRED_PATHS=""
DOC_SYNC_RULES_FILE="$doc_sync_file"
DOC_SYNC_BASE_REF="HEAD~1"
# Optional local override file loaded automatically by verify scripts:
# .orch/rules/$name/local.env

GENERATED_MARKERS="$generated_markers"

SECRET_SCAN_PATHS="."
SECRET_SCAN_EXCLUDE_GLOBS=".git/** vendor/** node_modules/** dist/** build/** coverage/** .orch/runtime/**"

MIGRATIONS_DIR="$migrations_dir"
MIGRATION_BASE_REF="HEAD~1"
MIGRATION_FILE_REGEX='^[^[:space:]]+/[0-9]{3,}.*\.(sql|up\.sql|down\.sql)$'
ALLOW_EDIT_EXISTING_MIGRATIONS="false"

ARCH_IMPORT_RULES_FILE="$arch_file"
ARCH_BACKEND="$arch_backend"
GO_LAYOUT_PROFILE="$go_layout_profile"
LAYOUT_REQUIRED_DIRS="$layout_required_dirs"
LAYOUT_FORBIDDEN_DIRS="$layout_forbidden_dirs"
EOF
}

render_global_env_content() {
  local generated_markers
  generated_markers='Code generated|DO NOT'" EDIT|AUTO-GENERATED FILE"
  cat <<'EOF'
REPO_ROOT="."
SECRET_SCAN_PATHS="."
SECRET_SCAN_EXCLUDE_GLOBS=".git/** vendor/** node_modules/** dist/** build/** coverage/** .orch/runtime/**"
EOF
  printf 'GENERATED_MARKERS="%s"\n' "$generated_markers"
}

render_manifest_content() {
  local names langs roots
  names="$(IFS=$'\t'; printf '%s' "${target_names[*]}")"
  langs="$(IFS=$'\t'; printf '%s' "${target_langs[*]}")"
  roots="$(IFS=$'\t'; printf '%s' "${target_roots[*]}")"
  python3 - "$mode" "$layout" "$names" "$langs" "$roots" <<'PY'
import json
import sys

mode = sys.argv[1]
layout = sys.argv[2]
names = sys.argv[3].split("\t") if sys.argv[3] else []
langs = sys.argv[4].split("\t") if sys.argv[4] else []
roots = sys.argv[5].split("\t") if sys.argv[5] else []
targets = []
for name, lang, root in zip(names, langs, roots):
    if not name:
        continue
    targets.append(
        {
            "name": name,
            "language": lang,
            "root": root,
            "env_file": f".orch/rules/{name}/rules.env",
        }
    )
payload = {
    "version": 1,
    "repo_mode": mode,
    "layout": layout,
    "targets": targets,
}
sys.stdout.write(json.dumps(payload, ensure_ascii=True, indent=2))
PY
}

render_makefile_rules_content() {
  local has_go="false"
  local i
  for i in "${!target_names[@]}"; do
    if [[ "${target_langs[$i]}" == "go" ]]; then
      has_go="true"
      break
    fi
  done

  {
    echo '# Generated by rules-kit/install.sh. Prefer editing .orch/rules/manifest.json and re-running install.'
    if [[ "$has_go" == "true" ]]; then
      echo 'TOOLS_BIN ?= $(CURDIR)/.tools/bin'
      echo 'GOLANGCI_LINT ?= $(TOOLS_BIN)/golangci-lint'
      echo 'GOLANGCI_LINT_VERSION ?= v1.64.8'
      echo 'GOLANGCI_LINT_ARGS ?= --tests=false'
      echo
      echo '$(GOLANGCI_LINT):'
      echo '	mkdir -p $(TOOLS_BIN)'
      echo '	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)'
      echo
    fi
    echo '.PHONY: verify verify-strict verify-perf verify-common'
    for i in "${!target_names[@]}"; do
      printf '.PHONY: verify-%s verify-strict-%s verify-perf-%s\n' "${target_names[$i]}" "${target_names[$i]}" "${target_names[$i]}"
    done
    echo
    printf 'verify: verify-common'
    for i in "${!target_names[@]}"; do
      printf ' verify-%s' "${target_names[$i]}"
    done
    echo
    printf 'verify-strict: verify'
    for i in "${!target_names[@]}"; do
      printf ' verify-strict-%s' "${target_names[$i]}"
    done
    echo
    printf 'verify-perf: verify-common'
    for i in "${!target_names[@]}"; do
      printf ' verify-perf-%s' "${target_names[$i]}"
    done
    echo
    echo
    echo 'verify-common:'
    echo '	bash .orch/rules/scripts/verify-generated.sh .orch/rules/global.env'
    echo '	bash .orch/rules/scripts/verify-secrets.sh .orch/rules/global.env'
    echo

    for i in "${!target_names[@]}"; do
      local name="${target_names[$i]}"
      local lang="${target_langs[$i]}"
      local root="${target_roots[$i]}"
      local root_cmd="cd ${root} &&"
      local npm_prefix="npm --prefix ${root}"
      if [[ "$root" == "." ]]; then
        root_cmd="cd . &&"
        npm_prefix="npm"
      fi

      if [[ "$lang" == "go" ]]; then
        echo "verify-$name: \$(GOLANGCI_LINT)"
        echo "	@out=\$\$(${root_cmd} gofmt -l \$\$(find . -type f -name '*.go' -not -path './.tools/*')); \\"
        echo "	if [ -n \"\$\$out\" ]; then \\"
        echo "		echo \"gofmt required for:\"; \\"
        echo "		echo \"\$\$out\"; \\"
        echo "		exit 1; \\"
        echo "	fi"
        echo "	${root_cmd} \$(GOLANGCI_LINT) run \$(GOLANGCI_LINT_ARGS) --config \$(CURDIR)/.orch/rules/lint/.golangci.base.yml ./..."
        echo "	${root_cmd} go test ./..."
        echo "	bash .orch/rules/scripts/verify-config-docs.sh .orch/rules/$name/rules.env"
        echo
        echo "verify-strict-$name: verify-$name"
        echo "	${root_cmd} go test -race ./..."
      elif [[ "$lang" == "ts" ]]; then
        echo "verify-$name:"
        echo "	${npm_prefix} run build"
        echo "	${npm_prefix} run lint"
        echo "	${npm_prefix} run typecheck"
        echo "	${npm_prefix} run test"
        echo "	bash .orch/rules/scripts/verify-config-docs.sh .orch/rules/$name/rules.env"
        echo
        echo "verify-strict-$name: verify-$name"
        echo
        echo "verify-perf-$name: verify-$name"
      else
        echo "verify-$name:"
        echo "	bash .orch/rules/scripts/verify-config-docs.sh .orch/rules/$name/rules.env"
        echo
        echo "verify-strict-$name: verify-$name"
        echo
        echo "verify-perf-$name: verify-$name"
      fi
      echo "	bash .orch/rules/scripts/verify-doc-sync.sh .orch/rules/$name/rules.env"
      echo "	bash .orch/rules/scripts/verify-layout.sh .orch/rules/$name/rules.env"
      echo "	bash .orch/rules/scripts/verify-contract-parity.sh .orch/rules/$name/rules.env"
      echo "	bash .orch/rules/scripts/verify-openspec-required.sh .orch/rules/$name/rules.env"
      echo "	bash .orch/rules/scripts/verify-change-scope.sh .orch/rules/$name/rules.env"
      echo "	bash .orch/rules/scripts/verify-migrations.sh .orch/rules/$name/rules.env"
      echo "	bash .orch/rules/scripts/verify-arch.sh .orch/rules/$name/rules.env"
      if [[ "$lang" == "go" ]]; then
        echo
        echo "verify-perf-$name: verify-$name"
        printf '\t%s go test -run '\''^$$'\'' -bench . -benchmem ./...\n' "$root_cmd"
      fi
      echo
    done
  }
}

ensure_root_makefile() {
  local makefile="$target_repo/Makefile"
  local include_block=$'include Makefile.rules\n'
  local append_block=$'\n# rules install block\ninclude Makefile.rules\n'

  if [[ ! -f "$makefile" ]]; then
    rules_write_text "$makefile" "$include_block" "$force"
    return
  fi
  if rg -q '^include Makefile\.rules$' "$makefile"; then
    return
  fi
  if [[ "$dry_run" == "true" ]]; then
    printf 'DRY RUN: append to %s\n' "$makefile"
    return
  fi
  printf '%s' "$append_block" >>"$makefile"
}

validate_generated_manifest_content() {
  local manifest_content="$1"
  local tmp_manifest
  tmp_manifest="$(mktemp)"
  printf '%s' "$manifest_content" >"$tmp_manifest"
  validate_manifest_file "$tmp_manifest"
  rm -f "$tmp_manifest"
}

copy_generic_base() {
  local repo="$1"
  local overwrite="$2"

  rules_mkdir_p "$repo/.orch/rules"
  rules_mkdir_p "$repo/.githooks"
  rules_mkdir_p "$repo/.orch/rules/lint"
  rules_mkdir_p "$repo/.orch/rules"

  copy_githooks_dir "$repo"
  rules_copy_file_if_allowed "$root_dir/RULES_README.md" "$repo/.orch/rules/README.md" "$overwrite"
  rules_copy_dir "$root_dir/scripts" "$repo/.orch/rules/scripts"
  rules_copy_file_if_allowed "$root_dir/lint/.golangci.base.yml" "$repo/.orch/rules/lint/.golangci.base.yml" "$overwrite"
  rules_upsert_agents_file "$root_dir/starter-kit/AGENTS.md.template" "$repo/AGENTS.md" "$overwrite"
}

write_target_file() {
  local path="$1"
  local content="$2"
  rules_write_text "$path" "$content" "$force"
}

write_generated_file() {
  local path="$1"
  local content="$2"
  rules_write_text "$path" "$content" "true"
}

init_go_target() {
  local root="$1"
  local name="$2"
  local target_dir="$target_repo/$root"
  local module_path="example.com/replace-me/${name}"

  write_target_file "$target_dir/go.mod" "module $module_path

go 1.23.0
"
  write_target_file "$target_dir/internal/config/config.go" 'package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv string
}

func Load() Config {
	return Config{
		AppEnv: stringFromEnv("APP_ENV", "dev"),
	}
}

func stringFromEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
'
  write_target_file "$target_dir/.env.example" 'APP_ENV=dev
'
  write_target_file "$target_dir/openapi/openapi.yaml" 'openapi: 3.1.0
info:
  title: Example API
  version: 0.1.0
paths: {}
'
  write_target_file "$target_dir/db/migrations/001_init.sql" '-- Write your first forward-only migration here.
'
  rules_mkdir_p "$target_dir/cmd/api"
  rules_mkdir_p "$target_dir/cmd/worker"
  rules_mkdir_p "$target_dir/internal/runtime"
  rules_mkdir_p "$target_dir/internal/httpapi"
  rules_mkdir_p "$target_dir/internal/transport/httpapi/public"
  rules_mkdir_p "$target_dir/internal/transport/httpapi/internal"
  rules_mkdir_p "$target_dir/internal/transport/httpapi/httpx"
  rules_mkdir_p "$target_dir/internal/transport/httpapi/security"
  rules_mkdir_p "$target_dir/internal/service/shared"
  rules_mkdir_p "$target_dir/internal/infra/store"
  rules_mkdir_p "$target_dir/internal/infra/objectstore"
  rules_mkdir_p "$target_dir/internal/infra/provider"
}

init_ts_target() {
  local root="$1"
  local name="$2"
  local target_dir="$target_repo/$root"
  local package_name="@replace-me/${name}"

  write_target_file "$target_dir/package.json" '{
  "name": "'"$package_name"'",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "build": "tsc -p tsconfig.json",
    "typecheck": "tsc --noEmit -p tsconfig.json",
    "lint": "npm run typecheck",
    "test": "node --test"
  },
  "devDependencies": {
    "@types/node": "^24.0.0",
    "typescript": "^5.7.0"
  }
}
'
  write_target_file "$target_dir/tsconfig.json" '{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "skipLibCheck": true,
    "types": ["node"]
  },
  "include": ["src/**/*.ts"]
}
'
  write_target_file "$target_dir/src/config/env.ts" 'export type AppConfig = {
  appEnv: string;
};

export function loadConfig(env: NodeJS.ProcessEnv = process.env): AppConfig {
  return {
    appEnv: stringFromEnv(env, "APP_ENV", "dev"),
  };
}

function stringFromEnv(
  env: NodeJS.ProcessEnv,
  key: string,
  fallback: string,
): string {
  const value = env[key]?.trim();
  return value && value.length > 0 ? value : fallback;
}
'
  write_target_file "$target_dir/src/index.ts" 'import { loadConfig } from "./config/env.js";

const config = loadConfig();

console.log(`service booted in ${config.appEnv}`);
'
  write_target_file "$target_dir/.env.example" 'APP_ENV=dev
'
  write_target_file "$target_dir/openapi/openapi.yaml" 'openapi: 3.1.0
info:
  title: Example API
  version: 0.1.0
paths: {}
'
  write_target_file "$target_dir/db/migrations/001_init.sql" '-- Write your first forward-only migration here.
'
  rules_mkdir_p "$target_dir/src/http"
  rules_mkdir_p "$target_dir/src/routes"
}

init_py_target() {
  local root="$1"
  local name="$2"
  local target_dir="$target_repo/$root"
  local module_name="${name//-/_}"

  write_target_file "$target_dir/pyproject.toml" "[build-system]
requires = [\"setuptools>=69\"]
build-backend = \"setuptools.build_meta\"

[project]
name = \"$name\"
version = \"0.1.0\"
description = \"Example Python project managed by rules-kit\"
requires-python = \">=3.10\"
dependencies = []
"
  write_target_file "$target_dir/src/$module_name/__init__.py" "__all__ = [\"main\"]

def main() -> None:
    print(\"hello from rules-kit python target\")
"
}

copy_generic_base "$target_repo" "$force"
cleanup_deprecated_managed_paths "$target_repo"
rules_append_root_gitignore_once "$target_repo" $'.tools/\n.orch/runtime/\nnode_modules/\ndist/\ncoverage/\n.env\n.env.*\n!.env.example\n' '^\.tools/$'

write_generated_file "$target_repo/.orch/rules/global.env" "$(render_global_env_content)"

for i in "${!target_names[@]}"; do
  target_name="${target_names[$i]}"
  target_lang="${target_langs[$i]}"
  target_root="${target_roots[$i]}"
  rules_mkdir_p "$target_repo/.orch/rules/$target_name"
  arch_rules_file="$target_repo/.orch/rules/$target_name/arch.rules"
  doc_sync_rules_file="$target_repo/.orch/rules/$target_name/docsync.rules"

  write_generated_file "$target_repo/.orch/rules/$target_name/rules.env" \
    "$(render_target_env_content "$target_name" "$target_lang" "$target_root")"
  write_target_file "$arch_rules_file" "$(render_arch_rules_content "$target_lang" "$target_root")"
  write_target_file "$doc_sync_rules_file" "$(render_doc_sync_rules_content "$target_lang" "$target_root")"

  if [[ "$mode" == "new" ]]; then
    if [[ "$target_lang" == "go" ]]; then
      init_go_target "$target_root" "$target_name"
    elif [[ "$target_lang" == "ts" ]]; then
      init_ts_target "$target_root" "$target_name"
    else
      init_py_target "$target_root" "$target_name"
    fi
  fi
done

manifest_content="$(render_manifest_content)"
validate_generated_manifest_content "$manifest_content"
write_generated_file "$target_repo/.orch/rules/manifest.json" "$manifest_content"
write_generated_file "$target_repo/Makefile.rules" "$(render_makefile_rules_content)"
rules_export_toolkit "$root_dir" "$target_repo/.orch/rules-kit" "$force"
ensure_root_makefile

echo "rules installed to: $target_repo"
echo "mode: $mode"
echo "layout: $layout"
echo "targets: $targets_spec"
if [[ "$dry_run" == "true" ]]; then
  rules_print_summary "$target_repo" "$mode" "$layout" "$targets_spec"
else
  echo "next: cd $target_repo && rules doctor . --strict"
fi
