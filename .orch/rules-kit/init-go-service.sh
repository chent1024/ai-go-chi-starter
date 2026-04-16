#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/init-go-service.sh <target-repo> <module-path> [--force]
  bash rules-kit/init-go-service.sh <target-repo> <module-path> [--force]

examples:
  bash tools/rules-kit/init-go-service.sh /tmp/my-service github.com/acme/my-service
  bash rules-kit/init-go-service.sh /tmp/my-service github.com/acme/my-service
  bash tools/rules-kit/init-go-service.sh ./sandbox/my-service github.com/acme/my-service --force
  bash rules-kit/init-go-service.sh ./sandbox/my-service github.com/acme/my-service --force
EOF
  exit 1
}

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage
fi

target_repo="$1"
module_path="$2"
force="false"

if [[ $# -eq 3 ]]; then
  if [[ "$3" != "--force" ]]; then
    usage
  fi
  force="true"
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

install_args=("$target_repo" "--mode" "new" "--layout" "single" "--targets" "app:go:." "--yes")
if [[ "$force" == "true" ]]; then
  install_args+=("--force")
fi
bash "$script_dir/install.sh" "${install_args[@]}" >/dev/null

target_repo="$(cd "$target_repo" && pwd)"

python3 - "$target_repo/go.mod" "$module_path" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
module = sys.argv[2]
text = path.read_text()
lines = text.splitlines()
if lines:
    lines[0] = f"module {module}"
path.write_text("\n".join(lines) + "\n")
PY

echo "initialized Go service skeleton: $target_repo"
echo "module: $module_path"
echo "next: cd $target_repo && bash rules-kit/doctor.sh . --strict"
