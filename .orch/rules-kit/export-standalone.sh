#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/export-standalone.sh <destination-dir>
  bash rules-kit/export-standalone.sh <destination-dir>

exports:
  - a self-contained rules-kit directory at the exact destination path
EOF
  exit 1
}

if [[ $# -ne 1 ]]; then
  usage
fi

dest_dir="$1"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "$(dirname "$dest_dir")"
rm -rf "$dest_dir"
cp -R "$script_dir" "$dest_dir"

echo "exported standalone rules kit to: $dest_dir"
