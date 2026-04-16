#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

compose() {
  docker compose -f deploy/docker-compose.dev.yaml --env-file deploy/.env.dev.example "$@"
}

API_PID=""
WORKER_PID=""

cleanup() {
  if [[ -n "$API_PID" ]]; then
    kill "$API_PID" >/dev/null 2>&1 || true
    wait "$API_PID" >/dev/null 2>&1 || true
  fi
  if [[ -n "$WORKER_PID" ]]; then
    kill "$WORKER_PID" >/dev/null 2>&1 || true
    wait "$WORKER_PID" >/dev/null 2>&1 || true
  fi
  compose down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

mkdir -p .runtime
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

go run ./cmd/migrate -action up

go run ./cmd/api >.runtime/smoke-api.log 2>&1 &
API_PID=$!

APP_WORKER_POLL_INTERVAL=1s go run ./cmd/worker >.runtime/smoke-worker.log 2>&1 &
WORKER_PID=$!

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8080/healthz >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! kill -0 "$WORKER_PID" >/dev/null 2>&1; then
  echo "worker exited before smoke checks completed" >&2
  exit 1
fi

health_body="$(curl -fsS http://127.0.0.1:8080/healthz)"
ready_body="$(curl -fsS http://127.0.0.1:8080/readyz)"
create_body="$(curl -fsS -H 'Content-Type: application/json' -d '{"name":"demo-smoke"}' http://127.0.0.1:8080/v1/examples)"
example_id="$(printf '%s' "$create_body" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
if [[ -z "$example_id" ]]; then
  echo "failed to parse example id from create response" >&2
  exit 1
fi
get_body="$(curl -fsS "http://127.0.0.1:8080/v1/examples/$example_id")"
list_body="$(curl -fsS http://127.0.0.1:8080/v1/examples)"

printf 'HEALTH=%s\n---\nREADY=%s\n---\nCREATE=%s\n---\nGET=%s\n---\nLIST=%s\n' \
  "$health_body" \
  "$ready_body" \
  "$create_body" \
  "$get_body" \
  "$list_body"
