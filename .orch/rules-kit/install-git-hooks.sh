#!/usr/bin/env bash
set -euo pipefail

repo_arg="${1:-.}"
repo_root="$(cd "$repo_arg" && pwd)"

if [[ ! -f "$repo_root/.githooks/pre-commit" || ! -f "$repo_root/.githooks/pre-push" ]]; then
  echo "git hook install failed: .githooks assets are missing; run rules install . --manifest .orch/rules/manifest.json --yes first" >&2
  exit 1
fi

git -C "$repo_root" config core.hooksPath .githooks
chmod +x "$repo_root/.githooks/pre-commit" "$repo_root/.githooks/pre-push"

echo "Configured core.hooksPath=.githooks"
echo "pre-commit -> make verify"
echo "pre-push   -> make verify-strict"
