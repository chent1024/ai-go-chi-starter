#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/install-into-repo.sh <target-repo> [profile] [--force]
  bash rules-kit/install-into-repo.sh <target-repo> [profile] [--force]

profiles:
  minimal
  starter
  go-starter
  ts-starter

behavior:
  - installs files directly into the target repository
  - refuses to overwrite existing files unless --force is provided
EOF
  exit 1
}

if [[ $# -lt 1 || $# -gt 3 ]]; then
  usage
fi

target_repo="$1"
profile="starter"
force="false"

for arg in "${@:2}"; do
    case "$arg" in
    minimal|starter|go-starter|ts-starter)
      profile="$arg"
      ;;
    --force)
      force="true"
      ;;
    *)
      echo "unknown argument: $arg" >&2
      usage
      ;;
  esac
done

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ ! -d "$target_repo" ]]; then
  echo "target repository does not exist: $target_repo" >&2
  exit 1
fi

target_repo="$(cd "$target_repo" && pwd)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

bash "$script_dir/install.sh" "$tmpdir" "$profile" >/dev/null

conflicts=()
while IFS= read -r path; do
  rel="${path#$tmpdir/}"
  dest="$target_repo/$rel"
  if [[ -e "$dest" && "$force" != "true" ]]; then
    conflicts+=("$rel")
  fi
done < <(find "$tmpdir" -type f | sort)

if [[ "${#conflicts[@]}" -gt 0 ]]; then
  echo "install aborted: target repo already has files that would be overwritten" >&2
  printf '%s\n' "${conflicts[@]}" >&2
  echo "re-run with --force to overwrite these files" >&2
  exit 1
fi

while IFS= read -r path; do
  rel="${path#$tmpdir/}"
  dest="$target_repo/$rel"
  mkdir -p "$(dirname "$dest")"
  cp "$path" "$dest"
done < <(find "$tmpdir" -type f | sort)

echo "rules installed into repo: $target_repo"
echo "profile: $profile"
if [[ "$force" == "true" ]]; then
  echo "overwrite mode: enabled"
fi
