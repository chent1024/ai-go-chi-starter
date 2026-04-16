#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: bash rules-kit/fixup.sh <target-repo> [--force]

actions:
  - copy Makefile.template to Makefile if Makefile is missing
  - append a small rules block to Makefile if possible
  - copy AGENTS.md template to AGENTS.md if AGENTS.md is missing

behavior:
  - conservative by default
  - does not overwrite existing files unless the action is an in-place placeholder replacement
  - with --force, placeholder/template-only files may be replaced and old unmanaged AGENTS content may be migrated into the user block
EOF
  exit 1
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
fi

target_repo="$1"
force="false"

if [[ $# -eq 2 ]]; then
  if [[ "$2" != "--force" ]]; then
    usage
  fi
  force="true"
fi

if [[ ! -d "$target_repo" ]]; then
  echo "fixup failed: target repository does not exist: $target_repo" >&2
  exit 1
fi

target_repo="$(cd "$target_repo" && pwd)"
env_file="$target_repo/.orch/rules/global.env"
if [[ ! -f "$env_file" && -f "$target_repo/.orch/rules/repo/rules.env" ]]; then
  env_file="$target_repo/.orch/rules/repo/rules.env"
fi

if [[ ! -f "$env_file" ]]; then
  echo "fixup failed: missing $env_file" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "$env_file"

say() {
  echo "$1"
}

copy_if_missing() {
  local src="$1"
  local dst="$2"
  if [[ -e "$dst" ]]; then
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  say "created: ${dst#$target_repo/}"
}

copy_agents_if_missing() {
  local src="$1"
  local dst="$2"
  if [[ -e "$dst" ]]; then
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  say "created: ${dst#$target_repo/}"
}

append_makefile_block_if_needed() {
  local makefile="$1"
  if [[ ! -f "$makefile" ]]; then
    return 0
  fi
  if rg -q '\.orch/rules/make/verify\.mk|verify-config-docs|verify-generated|verify-secrets' "$makefile"; then
    return 0
  fi
  cat >>"$makefile" <<'EOF'

# rules fixup block
.PHONY: verify-rules

RULES_ENV ?= .orch/rules/global.env

verify-rules:
	bash .orch/rules/scripts/verify-config-docs.sh $(RULES_ENV)
	bash .orch/rules/scripts/verify-generated.sh $(RULES_ENV)
	bash .orch/rules/scripts/verify-secrets.sh $(RULES_ENV)
EOF
  say "appended minimal rules block: ${makefile#$target_repo/}"
}

agents_template=""
if [[ -f "$target_repo/AGENTS.template.md" ]]; then
  agents_template="$target_repo/AGENTS.template.md"
elif [[ -f "$target_repo/rules-kit/AGENTS.base.md" ]]; then
  agents_template="$target_repo/rules-kit/AGENTS.base.md"
fi

if [[ -n "$agents_template" ]]; then
  copy_agents_if_missing "$agents_template" "$target_repo/AGENTS.md"
fi

if [[ ! -f "$target_repo/Makefile" && -f "$target_repo/Makefile.template" ]]; then
  cp "$target_repo/Makefile.template" "$target_repo/Makefile"
  say "created: Makefile"
elif [[ -f "$target_repo/Makefile" ]]; then
  append_makefile_block_if_needed "$target_repo/Makefile"
elif [[ "$force" == "true" && -f "$target_repo/Makefile.template" ]]; then
  cp "$target_repo/Makefile.template" "$target_repo/Makefile"
  say "created with --force: Makefile"
fi

say "fixup completed: $target_repo"
