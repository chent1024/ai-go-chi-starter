#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

PROTECTED_LOCAL_KEYS = {
    "OPENSPEC_REQUIRED_MODE",
    "OPENSPEC_REQUIRED_PATHS",
}


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
        for key, value in parse_env_file(local_path).items():
            if key in PROTECTED_LOCAL_KEYS:
                continue
            values[key] = value
    return values


def split_paths(value: str) -> list[str]:
    return [item.strip() for item in value.split() if item.strip()]


def path_in_prefixes(path: str, prefixes: list[str]) -> bool:
    return any(path == prefix or path.startswith(f"{prefix}/") for prefix in prefixes)


def git_changed_files(repo_root: Path, base_ref: str) -> list[str]:
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
    return sorted(files)


def ensure_git_repository(repo_root: Path) -> bool:
    result = subprocess.run(
        ["git", "rev-parse", "--is-inside-work-tree"],
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=False,
    )
    return result.returncode == 0 and result.stdout.strip() == "true"


def active_change_dirs(change_root: Path) -> list[Path]:
    if not change_root.exists():
        return []
    candidates: list[Path] = []
    for child in sorted(change_root.iterdir()):
        if not child.is_dir():
            continue
        if child.name.startswith(".") or child.name == "archive":
            continue
        candidates.append(child.resolve())
    return candidates


def is_exempt(path: str) -> bool:
    return path_in_prefixes(path, ["docs", ".github"]) or path.endswith(".md") or path.endswith(".txt")


def main() -> int:
    env_file = Path(sys.argv[1] if len(sys.argv) > 1 else ".orch/rules/global.env")
    if not env_file.exists():
        print(f"OPENSPEC REQUIRED CHECK FAILED: missing env file: {env_file}", file=sys.stderr)
        return 1

    config = load_config(env_file)
    mode = (config.get("OPENSPEC_REQUIRED_MODE", "smart") or "smart").strip()
    if mode == "off":
        print("openspec required checks skipped: OPENSPEC_REQUIRED_MODE=off")
        return 0

    required_paths = split_paths(config.get("OPENSPEC_REQUIRED_PATHS", ""))
    if not required_paths:
        print("openspec required checks skipped: OPENSPEC_REQUIRED_PATHS not configured")
        return 0

    repo_root = Path(config.get("REPO_ROOT") or ".")
    if not repo_root.is_absolute():
        repo_root = (Path.cwd() / repo_root).resolve()
    else:
        repo_root = repo_root.resolve()

    if not ensure_git_repository(repo_root):
        print(
            "OPENSPEC REQUIRED CHECK FAILED: git repository required to evaluate changed files",
            file=sys.stderr,
        )
        return 1

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
        print(f"OPENSPEC REQUIRED CHECK FAILED: unable to inspect changed files: {detail}", file=sys.stderr)
        return 1
    if not changed_files:
        print("openspec required checks passed")
        return 0

    candidate_files = [path for path in changed_files if path_in_prefixes(path, required_paths)]
    if not candidate_files or all(is_exempt(path) for path in candidate_files):
        print("openspec required checks passed")
        return 0

    change_root = (repo_root / config.get("CHANGE_SCOPE_ROOT", "openspec/changes")).resolve()
    active_changes = active_change_dirs(change_root)
    explicit_change = config.get("CHANGE_SCOPE_ACTIVE_CHANGE", "").strip()

    if not active_changes:
        print(
            "OPENSPEC REQUIRED CHECK FAILED: detected non-trivial code changes under OPENSPEC_REQUIRED_PATHS; create an OpenSpec change before implementation",
            file=sys.stderr,
        )
        return 1

    if len(active_changes) == 1:
        print("openspec required checks passed")
        return 0

    if explicit_change:
        explicit_dir = (change_root / explicit_change).resolve()
        if explicit_dir in active_changes:
            print("openspec required checks passed")
            return 0
        print(
            f"OPENSPEC REQUIRED CHECK FAILED: CHANGE_SCOPE_ACTIVE_CHANGE does not match an active change: {explicit_change}",
            file=sys.stderr,
        )
        return 1

    print(
        "OPENSPEC REQUIRED CHECK FAILED: multiple active OpenSpec changes exist; set CHANGE_SCOPE_ACTIVE_CHANGE to disambiguate",
        file=sys.stderr,
    )
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
