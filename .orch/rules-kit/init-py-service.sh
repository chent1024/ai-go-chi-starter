#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/init-py-service.sh <target-repo> <package-name> [--force]
  bash rules-kit/init-py-service.sh <target-repo> <package-name> [--force]

examples:
  bash tools/rules-kit/init-py-service.sh /tmp/my-app acme_demo
  bash rules-kit/init-py-service.sh /tmp/my-app acme_demo
  bash tools/rules-kit/init-py-service.sh ./sandbox/my-app acme-demo --force
  bash rules-kit/init-py-service.sh ./sandbox/my-app acme-demo --force
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

install_args=("$target_repo" "--mode" "new" "--layout" "single" "--targets" "app:py:." "--yes")
if [[ "$force" == "true" ]]; then
  install_args+=("--force")
fi
bash "$script_dir/install.sh" "${install_args[@]}" >/dev/null

target_repo="$(cd "$target_repo" && pwd)"

python3 - "$target_repo" "$package_name" <<'PY'
from pathlib import Path
import re
import sys

repo = Path(sys.argv[1])
package_name = sys.argv[2]
module_name = re.sub(r"[^a-zA-Z0-9_]+", "_", package_name.strip().replace("-", "_")).strip("_").lower() or "app"

pyproject = repo / "pyproject.toml"
pyproject.write_text(pyproject.read_text().replace('name = "app"', f'name = "{package_name}"'))

default_package = repo / "src" / "app"
desired_package = repo / "src" / module_name
if default_package.exists() and default_package != desired_package:
    desired_package.parent.mkdir(parents=True, exist_ok=True)
    default_package.rename(desired_package)
PY

echo "initialized Python project skeleton: $target_repo"
echo "package: $package_name"
echo "next: cd $target_repo && python3 -m pip install -e . && bash rules-kit/doctor.sh . --strict"
