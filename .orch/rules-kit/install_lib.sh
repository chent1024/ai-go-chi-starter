#!/usr/bin/env bash

rules_bool_env() {
  local primary_name="$1"
  printf '%s' "${!primary_name:-false}"
}

rules_paths_identical() {
  local src="$1"
  local dst="$2"
  if [[ ! -e "$src" || ! -e "$dst" ]]; then
    return 1
  fi
  [[ "$src" -ef "$dst" ]]
}

rules_mkdir_p() {
  local dir="$1"
  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    printf 'DRY RUN: mkdir -p %s\n' "$dir"
    return
  fi
  mkdir -p "$dir"
}

rules_copy_dir() {
  local src="$1"
  local dst="$2"
  if rules_paths_identical "$src" "$dst"; then
    return 0
  fi
  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    printf 'DRY RUN: copy dir %s -> %s\n' "$src" "$dst"
    return
  fi
  mkdir -p "$dst"
  cp -R "$src"/. "$dst"/
}

rules_copy_file_if_allowed() {
  local src="$1"
  local dst="$2"
  local force="$3"
  if rules_paths_identical "$src" "$dst"; then
    return 0
  fi
  if [[ -e "$dst" && "$force" != "true" ]]; then
    return 0
  fi
  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    printf 'DRY RUN: copy file %s -> %s\n' "$src" "$dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
}

rules_write_text() {
  local dest="$1"
  local content="$2"
  local force="$3"
  if [[ -e "$dest" && "$force" != "true" ]]; then
    return
  fi
  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    printf 'DRY RUN: write file %s\n' "$dest"
    return
  fi
  mkdir -p "$(dirname "$dest")"
  if [[ "$content" == *$'\n' ]]; then
    printf '%s' "$content" >"$dest"
  else
    printf '%s\n' "$content" >"$dest"
  fi
}

rules_upsert_agents_file() {
  local src="$1"
  local dst="$2"
  local force="$3"

  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    printf 'DRY RUN: write file %s\n' "$dst"
    return 0
  fi

  mkdir -p "$(dirname "$dst")"
  python3 - "$src" "$dst" "$force" <<'PY'
from pathlib import Path
import sys

MANAGED_START = "<!-- rules:managed:start -->"
MANAGED_END = "<!-- rules:managed:end -->"
USER_START = "<!-- rules:user:start -->"
USER_END = "<!-- rules:user:end -->"
DEFAULT_USER_BODY = "<!-- Add repository-specific instructions below this line. -->"


def normalize(content: str) -> str:
    return content.rstrip("\n")


def render(managed: str, user: str = DEFAULT_USER_BODY) -> str:
    return "\n".join(
        [
            MANAGED_START,
            normalize(managed),
            MANAGED_END,
            "",
            USER_START,
            normalize(user),
            USER_END,
            "",
        ]
    )


src = Path(sys.argv[1])
dst = Path(sys.argv[2])
force = sys.argv[3] == "true"
managed = src.read_text(encoding="utf-8")

if not dst.exists():
    dst.write_text(render(managed), encoding="utf-8")
    raise SystemExit(0)

current = dst.read_text(encoding="utf-8")
start = current.find(MANAGED_START)
end = current.find(MANAGED_END)
if start != -1 and end != -1 and start < end:
    end += len(MANAGED_END)
    replacement = render(managed).split(USER_START, 1)[0].rstrip("\n")
    updated = current[:start] + replacement + current[end:]
    if updated != current:
        dst.write_text(updated, encoding="utf-8")
    raise SystemExit(0)

if not force:
    raise SystemExit(10)

dst.write_text(render(managed, current.rstrip("\n")), encoding="utf-8")
PY
  local status="$?"
  if [[ "$status" == "10" ]]; then
    printf 'AGENTS.md exists without managed markers; keeping %s unchanged\n' "$dst"
    return 0
  fi
  return "$status"
}

rules_export_toolkit() {
  local src="$1"
  local dst="$2"
  local force="$3"

  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    printf 'DRY RUN: export toolkit -> %s\n' "$dst"
    return 0
  fi

  if [[ -f "$dst" ]]; then
    echo "refusing to export into an unsafe destination: $dst" >&2
    return 1
  fi
  if [[ -d "$dst" && "$force" != "true" ]]; then
    if find "$dst" -mindepth 1 -maxdepth 1 | read -r _ && [[ ! -f "$dst/.rules-kit-export" ]]; then
      echo "refusing to overwrite a non-rules-kit directory: $dst; remove it first, choose an empty path, or pass --force" >&2
      return 1
    fi
  fi

  rm -rf "$dst"
  mkdir -p "$(dirname "$dst")"
  cp -R "$src" "$dst"
  printf 'rules-kit export marker\n' >"$dst/.rules-kit-export"
}

rules_prompt_if_needed() {
  local var_name="$1"
  local prompt_text="$2"
  local default_value="$3"
  local current_value="${!var_name}"
  if [[ -n "$current_value" || "$(rules_bool_env RULES_ASSUME_YES)" == "true" || ! -t 0 ]]; then
    if [[ -z "$current_value" ]]; then
      printf -v "$var_name" '%s' "$default_value"
    fi
    return
  fi
  read -r -p "$prompt_text [$default_value]: " current_value
  current_value="${current_value:-$default_value}"
  printf -v "$var_name" '%s' "$current_value"
}

rules_detect_repo_field() {
  local detect_script="$1"
  local repo="$2"
  local field="$3"
  python3 "$detect_script" "$repo" --field "$field"
}

rules_print_detection_summary() {
  local mode="$1"
  local layout="$2"
  local targets_spec="$3"
  local reasons_text="$4"
  local warnings_text="$5"

  [[ -n "$mode" ]] && printf 'detected repo_mode: %s\n' "$mode"
  [[ -n "$layout" ]] && printf 'detected layout: %s\n' "$layout"
  [[ -n "$targets_spec" ]] && printf 'detected targets: %s\n' "$targets_spec"

  if [[ -n "$reasons_text" ]]; then
    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      printf 'detect reason: %s\n' "$line"
    done <<<"$reasons_text"
  fi

  if [[ -n "$warnings_text" ]]; then
    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      printf 'detect warning: %s\n' "$line"
    done <<<"$warnings_text"
  fi
}

rules_normalize_root() {
  local root="$1"
  if [[ "$root" == "." || "$root" == "./" ]]; then
    printf '.'
    return
  fi
  printf '%s' "${root#./}"
}

rules_path_prefix_for_root() {
  local root="$1"
  if [[ "$root" == "." ]]; then
    printf ''
  else
    printf '%s/' "$root"
  fi
}

rules_append_root_gitignore_once() {
  local repo="$1"
  local block="$2"
  local key_pattern="$3"
  local dest="$repo/.gitignore"

  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    if [[ ! -f "$dest" ]]; then
      printf 'DRY RUN: write file %s\n' "$dest"
      return
    fi
    if rg -q "$key_pattern" "$dest"; then
      return
    fi
    printf 'DRY RUN: append to %s\n' "$dest"
    return
  fi

  if [[ ! -f "$dest" ]]; then
    printf '%s' "$block" >"$dest"
    return
  fi
  if rg -q "$key_pattern" "$dest"; then
    return
  fi
  printf '\n%s' "$block" >>"$dest"
}

rules_replace_in_file() {
  local file="$1"
  local from="$2"
  local to="$3"
  if [[ ! -f "$file" ]]; then
    return
  fi
  if [[ "$(rules_bool_env RULES_DRY_RUN)" == "true" ]]; then
    if grep -Fq "$from" "$file"; then
      printf 'DRY RUN: replace text in %s\n' "$file"
    fi
    return
  fi
  python3 - "$file" "$from" "$to" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
old = sys.argv[2]
new = sys.argv[3]
text = path.read_text()
path.write_text(text.replace(old, new))
PY
}

rules_print_summary() {
  local target_repo="$1"
  local mode="$2"
  local layout="$3"
  local targets_spec="$4"
  printf 'DRY RUN: no files will be written\n'
  printf 'target repo: %s\n' "$target_repo"
  printf 'mode: %s\n' "$mode"
  printf 'layout: %s\n' "$layout"
  printf 'targets: %s\n' "$targets_spec"
}
