#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  bash tools/rules-kit/validate-manifest.sh <manifest.json>
  bash rules-kit/validate-manifest.sh <manifest.json>

Validates the repository manifest used by rules-kit/install.sh and rules-kit/doctor.sh.
EOF
  exit 1
}

if [[ $# -ne 1 ]]; then
  usage
fi

manifest_file="$1"
if [[ ! -f "$manifest_file" ]]; then
  echo "manifest file does not exist: $manifest_file" >&2
  exit 1
fi

python3 - "$manifest_file" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
manifest = json.loads(path.read_text())

errors = []

if manifest.get("version") != 1:
    errors.append("version must be 1")

repo_mode = manifest.get("repo_mode")
if repo_mode not in {"existing", "new"}:
    errors.append("repo_mode must be one of existing or new")

layout = manifest.get("layout")
if layout not in {"single", "monorepo"}:
    errors.append("layout must be one of single or monorepo")

targets = manifest.get("targets")
if not isinstance(targets, list) or not targets:
    errors.append("targets must be a non-empty array")
    targets = []

seen_names = set()
seen_roots = set()

for idx, target in enumerate(targets):
    prefix = f"targets[{idx}]"
    if not isinstance(target, dict):
        errors.append(f"{prefix} must be an object")
        continue

    name = target.get("name")
    language = target.get("language")
    root = target.get("root")
    env_file = target.get("env_file")

    if not isinstance(name, str) or not name.strip():
        errors.append(f"{prefix}.name must be a non-empty string")
    if language not in {"go", "ts", "py"}:
        errors.append(f"{prefix}.language must be one of go, ts, or py")
    if not isinstance(root, str) or not root.strip():
        errors.append(f"{prefix}.root must be a non-empty string")
    if env_file is not None and (not isinstance(env_file, str) or not env_file.strip()):
        errors.append(f"{prefix}.env_file must be a non-empty string when provided")

    if isinstance(name, str) and name.strip():
        if name in seen_names:
            errors.append(f"duplicate target name: {name}")
        seen_names.add(name)
    if isinstance(root, str) and root.strip():
        normalized_root = "." if root in {".", "./"} else root.lstrip("./")
        if normalized_root in seen_roots:
            errors.append(f"duplicate target root: {normalized_root}")
        seen_roots.add(normalized_root)

if layout == "single":
    if len(targets) != 1:
        errors.append("single layout requires exactly one target")
    elif targets and targets[0].get("root") not in {".", "./"}:
        errors.append("single layout requires target root '.'")

if errors:
    for error in errors:
        print(f"MANIFEST CHECK FAILED: {error}", file=sys.stderr)
    sys.exit(1)
PY
