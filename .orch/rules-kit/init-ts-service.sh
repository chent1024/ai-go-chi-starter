#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/init-ts-service.sh <target-repo> <package-name> [--force]
  bash rules-kit/init-ts-service.sh <target-repo> <package-name> [--force]

examples:
  bash tools/rules-kit/init-ts-service.sh /tmp/my-service @acme/my-service
  bash rules-kit/init-ts-service.sh /tmp/my-service @acme/my-service
  bash tools/rules-kit/init-ts-service.sh ./sandbox/my-service my-service --force
  bash rules-kit/init-ts-service.sh ./sandbox/my-service my-service --force
EOF
  exit 1
}

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage
fi

target_repo="$1"
package_name="$2"
force="false"

if [[ $# -eq 3 ]]; then
  if [[ "$3" != "--force" ]]; then
    usage
  fi
  force="true"
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

install_args=("$target_repo" "--mode" "new" "--layout" "single" "--targets" "app:ts:." "--yes")
if [[ "$force" == "true" ]]; then
  install_args+=("--force")
fi
bash "$script_dir/install.sh" "${install_args[@]}" >/dev/null

target_repo="$(cd "$target_repo" && pwd)"

python3 - "$target_repo/package.json" "$package_name" <<'PY'
from pathlib import Path
import json
import sys

path = Path(sys.argv[1])
package_name = sys.argv[2]
data = json.loads(path.read_text())
data["name"] = package_name
path.write_text(json.dumps(data, indent=2) + "\n")
PY

echo "initialized TypeScript service skeleton: $target_repo"
echo "package: $package_name"
echo "next: cd $target_repo && npm install && bash rules-kit/doctor.sh . --strict"
