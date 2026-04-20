#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path


def parse_env_file(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
            value = value[1:-1]
        values[key.strip()] = value
    return values


def local_variant_path(path: Path) -> Path | None:
    name = path.name
    if name == "rules.env":
        return path.with_name("local.env")
    if name.endswith(".env"):
        return path.with_name(f"{name[:-4]}.local.env")
    return None


def load_config(path: Path) -> dict[str, str]:
    values = parse_env_file(path)
    local_path = local_variant_path(path)
    if local_path is not None and local_path.exists():
        values.update(parse_env_file(local_path))
    return values


def rule_files(path: Path) -> list[Path]:
    files = [path]
    if path.name == "arch.rules":
        local_path = path.with_name("local.arch.rules")
    else:
        local_path = path.with_name(path.name.replace(".arch.rules", ".local.arch.rules"))
    if local_path.exists():
        files.append(local_path)
    return files


def rel_to_repo(path: str, repo_root: Path) -> str:
    return Path(path).resolve().relative_to(repo_root.resolve()).as_posix()


def parse_rule_line(line: str) -> tuple[str, str, str, str]:
    parts = [part.strip() for part in line.split("|", 3)]
    if len(parts) == 3:
        scope, forbidden_scope, message = parts
        return scope, forbidden_scope, message, ""
    if len(parts) == 4:
        scope, forbidden_scope, message, guidance = parts
        return scope, forbidden_scope, message, guidance
    raise ValueError(line)


def main() -> int:
    env_file = Path(sys.argv[1] if len(sys.argv) > 1 else ".orch/rules/global.env")
    if not env_file.exists():
        print(f"ARCH CHECK FAILED: missing env file: {env_file}", file=sys.stderr)
        return 1

    config = load_config(env_file)
    repo_root = Path(config.get("REPO_ROOT") or ".")
    if not repo_root.is_absolute():
        repo_root = (Path.cwd() / repo_root).resolve()
    else:
        repo_root = repo_root.resolve()
    target_root = config.get("TARGET_ROOT", ".")
    target_dir = (repo_root / target_root).resolve()
    rules_file = config.get("ARCH_IMPORT_RULES_FILE", "")

    if not rules_file:
        print("architecture checks skipped: ARCH_IMPORT_RULES_FILE not configured")
        return 0

    rules_path = (repo_root / rules_file).resolve()
    if not rules_path.exists():
        print(f"ARCH CHECK FAILED: rules file not found: {rules_file}", file=sys.stderr)
        return 1

    try:
        proc = subprocess.run(
            ["go", "list", "-json", "./..."],
            cwd=target_dir,
            text=True,
            capture_output=True,
            check=True,
        )
    except FileNotFoundError:
        print("ARCH CHECK FAILED: go is required for go_imports backend", file=sys.stderr)
        return 1
    except subprocess.CalledProcessError as exc:
        print("ARCH CHECK FAILED: failed to load Go package graph", file=sys.stderr)
        print(exc.stderr.strip(), file=sys.stderr)
        return 1

    decoder = json.JSONDecoder()
    payload = proc.stdout.strip()
    index = 0
    packages: list[dict[str, object]] = []
    while index < len(payload):
        while index < len(payload) and payload[index].isspace():
            index += 1
        if index >= len(payload):
            break
        obj, next_index = decoder.raw_decode(payload, index)
        packages.append(obj)
        index = next_index

    import_to_dir: dict[str, str] = {}
    for pkg in packages:
        import_path = str(pkg.get("ImportPath", ""))
        dir_path = str(pkg.get("Dir", ""))
        if not import_path or not dir_path:
            continue
        rel_dir = rel_to_repo(dir_path, repo_root)
        import_to_dir[import_path] = rel_dir

    package_dirs: dict[str, list[str]] = {}
    for pkg in packages:
        dir_path = str(pkg.get("Dir", ""))
        if not dir_path:
            continue
        rel_dir = rel_to_repo(dir_path, repo_root)
        imports: list[str] = []
        for item in pkg.get("Imports", []):
            target = import_to_dir.get(str(item))
            if target is None:
                continue
            imports.append(target)
        package_dirs[rel_dir] = imports

    failures = 0
    for active_rules_path in rule_files(rules_path):
        for raw_line in active_rules_path.read_text(encoding="utf-8").splitlines():
            line = raw_line.strip()
            if not line or line.startswith("#"):
                continue
            try:
                scope, forbidden_scope, message, guidance = parse_rule_line(line)
            except ValueError:
                print(f"ARCH CHECK FAILED: invalid rule line: {line}", file=sys.stderr)
                return 1
            matches: list[str] = []
            for package_dir, imports in package_dirs.items():
                if not (package_dir == scope or package_dir.startswith(f"{scope}/")):
                    continue
                for imported_dir in imports:
                    if imported_dir == forbidden_scope or imported_dir.startswith(f"{forbidden_scope}/"):
                        matches.append(f"{package_dir} imports {imported_dir}")
            if matches:
                print(f"ARCH CHECK FAILED: {message}", file=sys.stderr)
                if guidance:
                    print(f"recommended fix: {guidance}", file=sys.stderr)
                for match in matches:
                    print(match, file=sys.stderr)
                print(file=sys.stderr)
                failures = 1

    if failures:
        return 1

    print("architecture checks passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
