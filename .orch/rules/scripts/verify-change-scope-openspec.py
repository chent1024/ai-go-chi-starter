#!/usr/bin/env python3
from __future__ import annotations

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


def parse_scope_paths(metadata_path: Path) -> list[str]:
    scope_paths: list[str] = []
    in_scope = False
    for raw_line in metadata_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.rstrip()
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if not in_scope:
            if stripped == "scope_paths:":
                in_scope = True
            continue
        if raw_line.startswith((" ", "\t")) and stripped.startswith("- "):
            value = stripped[2:].strip().strip("'\"")
            if value:
                scope_paths.append(value)
            continue
        break
    return scope_paths


def git_changed_files(repo_root: Path, base_ref: str) -> set[str]:
    files: set[str] = set()

    def collect(*args: str) -> None:
        result = subprocess.run(
            ["git", *args],
            cwd=repo_root,
            text=True,
            capture_output=True,
            check=True,
        )
        for line in result.stdout.splitlines():
            value = line.strip()
            if value:
                files.add(value)

    collect("diff", "--name-only", base_ref, "--")
    collect("diff", "--name-only", "--cached", "--")
    collect("ls-files", "--others", "--exclude-standard")
    return files


def ensure_git_repository(repo_root: Path) -> bool:
    result = subprocess.run(
        ["git", "rev-parse", "--is-inside-work-tree"],
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=False,
    )
    return result.returncode == 0 and result.stdout.strip() == "true"


def find_active_change_dir(change_root: Path, explicit_change: str) -> Path | None:
    if explicit_change:
        explicit_dir = (change_root / explicit_change).resolve()
        if explicit_dir.is_dir() and explicit_dir.name != "archive" and not explicit_dir.name.startswith("."):
            return explicit_dir
        raise ValueError(f"CHANGE_SCOPE_ACTIVE_CHANGE does not match an active OpenSpec change: {explicit_change}")

    candidates = []
    if not change_root.exists():
        return None
    for child in sorted(change_root.iterdir()):
        if not child.is_dir():
            continue
        if child.name.startswith(".") or child.name == "archive":
            continue
        candidates.append(child)
    if len(candidates) == 1:
        return candidates[0].resolve()
    if not candidates:
        return None
    raise ValueError("multiple active OpenSpec changes found; set CHANGE_SCOPE_ACTIVE_CHANGE explicitly")


def allowed_path(path: str, allowed_prefixes: list[str]) -> bool:
    return any(path == prefix or path.startswith(f"{prefix}/") for prefix in allowed_prefixes)


def split_paths(value: str) -> list[str]:
    return [item.strip() for item in value.split() if item.strip()]


def main() -> int:
    env_file = Path(sys.argv[1] if len(sys.argv) > 1 else ".orch/rules/global.env")
    if not env_file.exists():
        print(f"CHANGE SCOPE CHECK FAILED: missing env file: {env_file}", file=sys.stderr)
        return 1

    config = load_config(env_file)
    repo_root = Path(config.get("REPO_ROOT") or ".")
    if not repo_root.is_absolute():
        repo_root = (Path.cwd() / repo_root).resolve()
    else:
        repo_root = repo_root.resolve()

    if not ensure_git_repository(repo_root):
        print("CHANGE SCOPE CHECK FAILED: git repository required to evaluate changed files", file=sys.stderr)
        return 1

    change_root = (repo_root / config.get("CHANGE_SCOPE_ROOT", "openspec/changes")).resolve()

    default_base_ref = "HEAD"
    head_parent = subprocess.run(
        ["git", "rev-parse", "--verify", "HEAD~1"],
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=False,
    )
    if head_parent.returncode == 0:
        default_base_ref = "HEAD~1"
    base_ref = config.get("CHANGE_SCOPE_BASE_REF", default_base_ref)
    verify_base = subprocess.run(
        ["git", "rev-parse", "--verify", base_ref],
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=False,
    )
    if verify_base.returncode != 0:
        base_ref = default_base_ref

    try:
        changed_files = git_changed_files(repo_root, base_ref)
    except subprocess.CalledProcessError as exc:
        detail = exc.stderr.strip() or exc.stdout.strip() or "git command failed"
        print(f"CHANGE SCOPE CHECK FAILED: unable to inspect changed files: {detail}", file=sys.stderr)
        return 1
    if not changed_files:
        print("change scope checks passed")
        return 0

    mode = config.get("CHANGE_SCOPE_MODE", "strict").strip() or "strict"
    code_paths = split_paths(config.get("CHANGE_SCOPE_CODE_PATHS", ""))
    allow_paths = split_paths(config.get("CHANGE_SCOPE_ALLOW_PATHS", ""))

    if mode == "off":
        print("change scope checks skipped: CHANGE_SCOPE_MODE=off")
        return 0
    if mode == "allow_docs":
        allow_paths = [*allow_paths, "docs"]

    candidate_files = sorted(changed_files)
    if code_paths:
        candidate_files = [path for path in candidate_files if allowed_path(path, code_paths)]
        if not candidate_files:
            print("change scope checks passed")
            return 0

    try:
        change_dir = find_active_change_dir(
            change_root,
            config.get("CHANGE_SCOPE_ACTIVE_CHANGE", "").strip(),
        )
    except ValueError as exc:
        print(f"CHANGE SCOPE CHECK FAILED: {exc}", file=sys.stderr)
        return 1
    if change_dir is None:
        print("change scope checks skipped: no active OpenSpec change")
        return 0

    metadata_name = config.get("CHANGE_SCOPE_METADATA_FILE", ".openspec.yaml")
    metadata_path = change_dir / metadata_name
    if not metadata_path.exists():
        print(
            f"CHANGE SCOPE CHECK FAILED: metadata file not found for change {change_dir.name}: {metadata_name}",
            file=sys.stderr,
        )
        return 1

    scope_paths = parse_scope_paths(metadata_path)
    if not scope_paths:
        print(
            f"CHANGE SCOPE CHECK FAILED: no scope_paths declared in {metadata_path.relative_to(repo_root)}",
            file=sys.stderr,
        )
        return 1

    allowed_prefixes = [str(change_dir.relative_to(repo_root)).replace("\\", "/"), *scope_paths, *allow_paths]

    out_of_scope = sorted(path for path in candidate_files if not allowed_path(path, allowed_prefixes))
    if out_of_scope:
        print(
            f"CHANGE SCOPE CHECK FAILED: files changed outside active change scope: {change_dir.name}",
            file=sys.stderr,
        )
        print("Allowed prefixes:", file=sys.stderr)
        for prefix in allowed_prefixes:
            print(prefix, file=sys.stderr)
        print("Out-of-scope files:", file=sys.stderr)
        for path in out_of_scope:
            print(path, file=sys.stderr)
        return 1

    print("change scope checks passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
