#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/rules_helpers.sh"

env_file="$(rules_default_env_file "${1:-}")"
if ! rules_load_target_env "$env_file" "CONFIG DOC CHECK FAILED"; then
  exit 1
fi

repo_root="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$repo_root"

if [[ -z "${CONFIG_SOURCES:-}" || -z "${ENV_EXAMPLE_FILE:-}" ]]; then
  echo "config doc checks skipped: CONFIG_SOURCES or ENV_EXAMPLE_FILE not configured"
  exit 0
fi

required_file="$(mktemp)"
documented_file="$(mktemp)"
trap 'rm -f "$required_file" "$documented_file"' EXIT

build_helper_regex() {
  local helpers="$1"
  local regex=""
  local helper escaped

  for helper in $helpers; do
    escaped="$(printf '%s' "$helper" | sed -E 's/[][(){}.^$*+?|\\]/\\&/g')"
    if [[ -n "$regex" ]]; then
      regex+="|"
    fi
    regex+="$escaped"
  done

  printf '%s' "$regex"
}

default_config_env_helpers="stringFromEnv boolFromEnv floatFromEnv intFromEnv int64FromEnv durationFromEnv uintFromEnv stringListFromEnv listFromEnv csvFromEnv os.Getenv os.LookupEnv LookupEnv"
default_config_env_second_arg_helpers="stringFromEnv boolFromEnv floatFromEnv intFromEnv int64FromEnv durationFromEnv uintFromEnv stringListFromEnv listFromEnv csvFromEnv"

config_env_helpers="${CONFIG_ENV_HELPERS:-$default_config_env_helpers}"
config_env_second_arg_helpers="${CONFIG_ENV_SECOND_ARG_HELPERS:-$default_config_env_second_arg_helpers}"
config_env_helper_regex="${CONFIG_ENV_HELPER_REGEX:-$(build_helper_regex "$config_env_helpers")}"
config_env_second_arg_helper_regex="${CONFIG_ENV_SECOND_ARG_HELPER_REGEX:-$(build_helper_regex "$config_env_second_arg_helpers")}"

{
  if [[ -n "$config_env_helper_regex" ]]; then
    rg -oN "(${config_env_helper_regex})\\(\\s*\"([A-Z][A-Z0-9_]*)\"" $CONFIG_SOURCES \
      | sed -E 's/.*\("([A-Z][A-Z0-9_]*)"/\1/' || true
  fi
  if [[ -n "$config_env_second_arg_helper_regex" ]]; then
    rg -oN "(${config_env_second_arg_helper_regex})\\([^,\n]+,\\s*\"([A-Z][A-Z0-9_]*)\"" $CONFIG_SOURCES \
      | sed -E 's/.*"([A-Z][A-Z0-9_]*)"/\1/' || true
  fi
  rg -oN 'process\.env\.([A-Z0-9_]+)' $CONFIG_SOURCES \
    | sed -E 's/.*process\.env\.([A-Z0-9_]+)/\1/' || true
  rg -oN 'process\.env\["([A-Z0-9_]+)"\]' $CONFIG_SOURCES \
    | sed -E 's/.*\["([A-Z0-9_]+)"\]/\1/' || true
} | sort -u >"$required_file"

rg -oN '^[A-Z][A-Z0-9_]*=' "$ENV_EXAMPLE_FILE" \
  | sed 's/=.*//' \
  | sort -u >"$documented_file"

missing_keys="$(comm -23 "$required_file" "$documented_file" || true)"
stale_keys="$(comm -13 "$required_file" "$documented_file" || true)"

if [[ -n "$missing_keys" ]]; then
  echo "CONFIG DOC CHECK FAILED: example env file is missing keys used by config sources" >&2
  echo "$missing_keys" >&2
  exit 1
fi

if [[ -n "$stale_keys" && "${ALLOW_STALE_ENV_EXAMPLE_KEYS:-false}" != "true" ]]; then
  echo "CONFIG DOC CHECK FAILED: example env file has keys no longer used by config sources" >&2
  echo "$stale_keys" >&2
  exit 1
fi
