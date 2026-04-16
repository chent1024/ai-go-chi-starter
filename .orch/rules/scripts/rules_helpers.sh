#!/usr/bin/env bash

rules_local_variant_path() {
  local path="$1"
  case "$path" in
    */rules.env)
      printf '%s/local.env\n' "${path%/rules.env}"
      ;;
    *.env)
      printf '%s.local.env\n' "${path%.env}"
      ;;
    */arch.rules)
      printf '%s/local.arch.rules\n' "${path%/arch.rules}"
      ;;
    *.arch.rules)
      printf '%s.local.arch.rules\n' "${path%.arch.rules}"
      ;;
    */docsync.rules)
      printf '%s/local.docsync.rules\n' "${path%/docsync.rules}"
      ;;
    *.docsync.rules)
      printf '%s.local.docsync.rules\n' "${path%.docsync.rules}"
      ;;
    *)
      return 1
      ;;
  esac
}

rules_default_env_file() {
  local requested="${1:-}"
  if [[ -n "$requested" ]]; then
    printf '%s\n' "$requested"
    return
  fi
  if [[ -f ".orch/rules/global.env" ]]; then
    printf '.orch/rules/global.env\n'
    return
  fi
  if [[ -f ".orch/rules/repo/rules.env" ]]; then
    printf '.orch/rules/repo/rules.env\n'
    return
  fi
  printf '.orch/rules/global.env\n'
}

rules_load_target_env() {
  local env_file="$1"
  local failure_prefix="$2"

  if [[ ! -f "$env_file" ]]; then
    echo "$failure_prefix: missing env file: $env_file" >&2
    return 1
  fi

  # shellcheck disable=SC1090
  source "$env_file"

  local local_env_file=""
  if local_env_file="$(rules_local_variant_path "$env_file" 2>/dev/null)"; then
    if [[ -f "$local_env_file" ]]; then
      # shellcheck disable=SC1090
      source "$local_env_file"
    fi
  fi

  RULES_LOCAL_ENV_FILE="$local_env_file"
  export RULES_LOCAL_ENV_FILE
}

rules_list_rule_files() {
  local base_file="$1"
  printf '%s\n' "$base_file"

  local local_file=""
  if local_file="$(rules_local_variant_path "$base_file" 2>/dev/null)"; then
    if [[ -f "$local_file" ]]; then
      printf '%s\n' "$local_file"
    fi
  fi
}
