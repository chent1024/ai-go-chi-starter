#!/usr/bin/env python3
from __future__ import annotations

import json
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


def load_routes(path: Path) -> set[str]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    if isinstance(payload, dict):
        items = payload.get("routes", payload)
    else:
        items = payload
    routes: set[str] = set()
    if not isinstance(items, list):
        raise ValueError(f"manifest must contain a list or routes field: {path}")
    for item in items:
        if isinstance(item, str):
            routes.add(item.strip())
            continue
        if not isinstance(item, dict):
            raise ValueError(f"invalid route entry in {path}: {item!r}")
        method = str(item.get("method", "")).upper().strip()
        route = str(item.get("path", "")).strip()
        if not method or not route:
            raise ValueError(f"route entry must include method and path: {item!r}")
        routes.add(f"{method} {route}")
    return routes


def main() -> int:
    env_file = Path(sys.argv[1] if len(sys.argv) > 1 else ".orch/rules/global.env")
    if not env_file.exists():
        print(f"CONTRACT PARITY CHECK FAILED: missing env file: {env_file}", file=sys.stderr)
        return 1

    config = load_config(env_file)
    repo_root = Path(config.get("REPO_ROOT") or ".")
    if not repo_root.is_absolute():
        repo_root = (Path.cwd() / repo_root).resolve()
    else:
        repo_root = repo_root.resolve()

    route_manifest = config.get("CONTRACT_PARITY_ROUTE_MANIFEST_FILE", "")
    contract_manifest = config.get("CONTRACT_PARITY_CONTRACT_MANIFEST_FILE", "")
    if not route_manifest or not contract_manifest:
        print(
            "CONTRACT PARITY CHECK FAILED: manifest backend requires CONTRACT_PARITY_ROUTE_MANIFEST_FILE and CONTRACT_PARITY_CONTRACT_MANIFEST_FILE",
            file=sys.stderr,
        )
        return 1

    route_path = (repo_root / route_manifest).resolve()
    contract_path = (repo_root / contract_manifest).resolve()
    if not route_path.exists():
        print(f"CONTRACT PARITY CHECK FAILED: route manifest not found: {route_manifest}", file=sys.stderr)
        return 1
    if not contract_path.exists():
        print(f"CONTRACT PARITY CHECK FAILED: contract manifest not found: {contract_manifest}", file=sys.stderr)
        return 1

    try:
        route_routes = load_routes(route_path)
        contract_routes = load_routes(contract_path)
    except ValueError as exc:
        print(f"CONTRACT PARITY CHECK FAILED: {exc}", file=sys.stderr)
        return 1

    only_in_route = sorted(route_routes - contract_routes)
    only_in_contract = sorted(contract_routes - route_routes)
    if only_in_route or only_in_contract:
        print("CONTRACT PARITY CHECK FAILED: route and contract manifests differ", file=sys.stderr)
        if only_in_route:
            print("Routes missing from contract:", file=sys.stderr)
            for item in only_in_route:
                print(item, file=sys.stderr)
        if only_in_contract:
            print("Contract entries missing from route manifest:", file=sys.stderr)
            for item in only_in_contract:
                print(item, file=sys.stderr)
        return 1

    print("contract parity checks passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
