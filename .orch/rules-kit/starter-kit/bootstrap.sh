#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <output-dir>" >&2
  exit 1
fi

output_dir="$1"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
template_root="$(cd "$script_dir/.." && pwd)"

mkdir -p "$output_dir/rules-kit"
mkdir -p "$output_dir/.githooks"

cp "$script_dir/AGENTS.md.template" "$output_dir/AGENTS.md"
cp "$script_dir/Makefile.template" "$output_dir/Makefile.template"

cp -R "$template_root/ci" "$output_dir/rules-kit/"
cp -R "$template_root/githooks" "$output_dir/rules-kit/"
cp -R "$template_root/lint" "$output_dir/rules-kit/"
cp -R "$template_root/make" "$output_dir/rules-kit/"
cp -R "$template_root/repo" "$output_dir/rules-kit/"
cp -R "$template_root/review" "$output_dir/rules-kit/"
cp -R "$template_root/scripts" "$output_dir/rules-kit/"
cp "$template_root/AGENTS.base.md" "$output_dir/rules-kit/AGENTS.base.md"
cp "$template_root/README.md" "$output_dir/rules-kit/README.md"
cp "$template_root/install.sh" "$output_dir/rules-kit/install.sh"
cp "$template_root/install_lib.sh" "$output_dir/rules-kit/install_lib.sh"
cp "$template_root/validate-manifest.sh" "$output_dir/rules-kit/validate-manifest.sh"
cp "$template_root/doctor.sh" "$output_dir/rules-kit/doctor.sh"
cp "$template_root/fixup.sh" "$output_dir/rules-kit/fixup.sh"
cp "$template_root/init-go-service.sh" "$output_dir/rules-kit/init-go-service.sh"
cp "$template_root/init-ts-service.sh" "$output_dir/rules-kit/init-ts-service.sh"

cp "$template_root/githooks/pre-commit" "$output_dir/.githooks/pre-commit"
cp "$template_root/githooks/pre-push" "$output_dir/.githooks/pre-push"

echo "starter kit exported to: $output_dir"
