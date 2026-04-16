#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

compose() {
  docker compose -f deploy/docker-compose.dev.yaml --env-file deploy/.env.dev.example "$@"
}

cleanup() {
  compose down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

compose down -v >/dev/null 2>&1 || true
compose up -d

for _ in $(seq 1 90); do
  health="$(docker inspect --format '{{.State.Health.Status}}' ai-go-chi-starter-postgres 2>/dev/null || true)"
  if [[ "$health" == "healthy" ]]; then
    break
  fi
  sleep 1
done

if [[ "${health:-}" != "healthy" ]]; then
  echo "postgres did not become healthy" >&2
  exit 1
fi

set -a
source deploy/.env.dev.example
set +a

go test -count=1 -tags=integration ./internal/infra/store/postgres -run TestExampleRepositoryIntegrationCRUD
