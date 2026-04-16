#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any


FRONTEND_NAMES = {"frontend", "web", "client", "ui", "site", "console", "admin"}
BACKEND_NAMES = {"backend", "server", "api", "service", "svc"}
GO_BACKEND_NAMES = BACKEND_NAMES | {"app"}
UTILITY_NAMES = {
    "shared",
    "common",
    "utils",
    "util",
    "lib",
    "libs",
    "config",
    "types",
    "schemas",
    "eslint-config",
}
CONTAINER_NAMES = {"apps", "services", "packages"}
MONOREPO_MARKER_FILES = {"pnpm-workspace.yaml", "turbo.json", "nx.json", "lerna.json"}
PYTHON_MARKER_FILES = ("pyproject.toml", "requirements.txt", "requirements-dev.txt", "setup.py", "setup.cfg")


@dataclass(frozen=True)
class Candidate:
    root: str
    language: str
    name_hint: str
    role: str
    confidence: str
    reason: str


def normalize_root(path: Path) -> str:
    return "." if str(path) in {"", "."} else path.as_posix().lstrip("./")


def load_json(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    return payload if isinstance(payload, dict) else {}


def package_has_workspaces(path: Path) -> bool:
    payload = load_json(path / "package.json")
    workspaces = payload.get("workspaces")
    if isinstance(workspaces, list):
        return True
    return isinstance(workspaces, dict) and bool(workspaces.get("packages"))


def has_ts_marker(path: Path) -> bool:
    return (path / "package.json").exists() or (path / "tsconfig.json").exists()


def has_py_marker(path: Path) -> bool:
    return any((path / marker).exists() for marker in PYTHON_MARKER_FILES)


def has_project_marker(path: Path) -> bool:
    return (path / "go.mod").exists() or has_ts_marker(path) or has_py_marker(path)


def detect_language(path: Path) -> str:
    if (path / "go.mod").exists():
        return "go"
    if has_ts_marker(path):
        return "ts"
    if has_py_marker(path):
        return "py"
    return ""


def classify_candidate(relative_root: str, language: str) -> tuple[str, str, str]:
    if relative_root == ".":
        return ("app", "root", "high")

    parts = relative_root.split("/")
    leaf = parts[-1].lower()
    parent = parts[-2].lower() if len(parts) > 1 else ""

    if leaf in UTILITY_NAMES:
        return (leaf, "utility", "low")
    if parent == "apps":
        if leaf in UTILITY_NAMES:
            return (leaf, "utility", "low")
        return ("frontend" if leaf in FRONTEND_NAMES else leaf, "frontend", "high")
    if parent == "services":
        if leaf in UTILITY_NAMES:
            return (leaf, "utility", "low")
        return ("backend" if leaf in BACKEND_NAMES else leaf, "backend", "high")
    if parent == "packages":
        if leaf in UTILITY_NAMES:
            return (leaf, "utility", "low")
        if leaf in FRONTEND_NAMES:
            return ("frontend", "frontend", "medium")
        if leaf in BACKEND_NAMES:
            return ("backend", "backend", "medium")
        return (leaf, "generic", "low")
    if language == "go" and leaf in GO_BACKEND_NAMES:
        return ("backend", "backend", "high")
    if language == "ts" and leaf in FRONTEND_NAMES:
        return ("frontend", "frontend", "high")
    if language == "ts" and leaf in BACKEND_NAMES:
        return ("backend", "backend", "medium")
    return (leaf, "generic", "medium")


def discover_candidate_paths(repo: Path) -> list[Path]:
    candidates: list[Path] = []
    if has_project_marker(repo):
        candidates.append(Path("."))

    for child in sorted(repo.iterdir(), key=lambda item: item.name):
        if not child.is_dir() or child.name.startswith("."):
            continue
        if child.name in CONTAINER_NAMES:
            for grandchild in sorted(child.iterdir(), key=lambda item: item.name):
                if not grandchild.is_dir() or grandchild.name.startswith("."):
                    continue
                if has_project_marker(grandchild):
                    candidates.append(grandchild.relative_to(repo))
            continue
        if has_project_marker(child):
            candidates.append(child.relative_to(repo))
    return candidates


def assign_target_names(candidates: list[Candidate]) -> list[dict[str, str]]:
    name_counts: dict[str, int] = {}
    for candidate in candidates:
        name_counts[candidate.name_hint] = name_counts.get(candidate.name_hint, 0) + 1

    assigned: list[dict[str, str]] = []
    used_names: set[str] = set()
    for candidate in candidates:
        base_name = candidate.name_hint if name_counts[candidate.name_hint] == 1 else candidate.root.split("/")[-1]
        name = base_name
        suffix = 2
        while name in used_names:
            name = f"{base_name}{suffix}"
            suffix += 1
        used_names.add(name)
        assigned.append(
            {
                "name": name,
                "language": candidate.language,
                "root": candidate.root,
            }
        )
    return assigned


def detect_repo(repo: Path) -> dict[str, Any]:
    repo_exists = repo.exists()
    if not repo_exists:
        return {
            "repo_mode": "new",
            "layout": "single",
            "targets": [],
            "targets_spec": "",
            "confidence": "medium",
            "ambiguous": False,
            "reasons": [f"target repository does not exist yet: {repo}"],
            "warnings": [],
        }

    entries = [item.name for item in repo.iterdir()]
    repo_mode = "existing" if entries else "new"

    marker_files = [name for name in sorted(MONOREPO_MARKER_FILES) if (repo / name).exists()]
    workspace_root = package_has_workspaces(repo)
    raw_candidate_paths = discover_candidate_paths(repo)
    if (workspace_root or marker_files) and any(path != Path(".") for path in raw_candidate_paths):
        raw_candidate_paths = [path for path in raw_candidate_paths if path != Path(".")]

    candidates: list[Candidate] = []
    warnings: list[str] = []
    reasons: list[str] = []
    has_generic = False
    has_nested_projects = False

    for relative_path in raw_candidate_paths:
        root = normalize_root(relative_path)
        if root != ".":
            has_nested_projects = True
        language = detect_language(repo / relative_path)
        if not language:
            continue
        name_hint, role, confidence = classify_candidate(root, language)
        if role == "utility":
            warnings.append(f"ignored utility package at {root}")
            continue
        if role == "generic":
            has_generic = True
        reasons.append(f"found {language} project marker under {root}")
        candidates.append(
            Candidate(
                root=root,
                language=language,
                name_hint=name_hint,
                role=role,
                confidence=confidence,
                reason=f"{root}:{language}:{role}",
            )
        )

    layout = "single"
    if marker_files or workspace_root or len(candidates) > 1 or (has_nested_projects and candidates):
        layout = "monorepo"

    ambiguous = False
    confidence = "high"

    if marker_files:
        for marker in marker_files:
            reasons.append(f"found monorepo marker file {marker}")
    if workspace_root:
        reasons.append("root package.json declares workspaces")

    if layout == "single" and candidates:
        selected_candidates = [candidates[0]]
    else:
        selected_candidates = candidates

    if layout == "monorepo" and not selected_candidates:
        ambiguous = True
        confidence = "low"
        reasons.append("did not find clear app/service targets for monorepo layout")
    elif layout == "monorepo" and any(candidate.confidence == "low" for candidate in selected_candidates):
        ambiguous = True
        confidence = "low"
        reasons.append("found only low-confidence monorepo targets")
    elif has_generic and layout == "monorepo":
        ambiguous = True
        confidence = "low"
        reasons.append("found generic package targets without clear frontend/backend naming")
    elif any(candidate.confidence == "medium" for candidate in selected_candidates):
        confidence = "medium"

    targets = assign_target_names(selected_candidates)
    targets_spec = ",".join(f"{item['name']}:{item['language']}:{item['root']}" for item in targets)

    return {
        "repo_mode": repo_mode,
        "layout": layout,
        "targets": targets,
        "targets_spec": targets_spec,
        "confidence": confidence,
        "ambiguous": ambiguous,
        "reasons": reasons,
        "warnings": warnings,
    }


def field_value(payload: dict[str, Any], field: str) -> str:
    if field == "reasons_text":
        return "\n".join(str(item) for item in payload.get("reasons") or [])
    if field == "warnings_text":
        return "\n".join(str(item) for item in payload.get("warnings") or [])
    value = payload.get(field, "")
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, (list, dict)):
        return json.dumps(value, ensure_ascii=False)
    return str(value)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Detect repository mode/layout/targets for rules install")
    parser.add_argument("repo", help="target repository path")
    parser.add_argument("--field", default="", help="print only one field")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    repo = Path(args.repo).expanduser().resolve()
    payload = detect_repo(repo)
    if args.field:
        print(field_value(payload, args.field))
        return 0
    print(json.dumps(payload, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
