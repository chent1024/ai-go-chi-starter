# API

## Envelope

Every JSON response uses the same envelope:

```json
{
  "code": "OK",
  "message": "",
  "request_id": "req_123",
  "data": {}
}
```

Errors add `retryable`.

Reserved cross-cutting codes are documented in [docs/errors.md](/Users/xihe0000/workspace/code/ai-go-chi-starter/docs/errors.md).

## Routes

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness probe. |
| `GET` | `/readyz` | Readiness probe backed by database ping. |
| `POST` | `/v1/examples` | Create an example. |
| `GET` | `/v1/examples/{id}` | Fetch a single example. |
| `GET` | `/v1/examples` | List examples. |

## Cross-Cutting Behavior

- every request gets `X-Request-Id` and `Traceparent`
- every request is subject to a server-side request timeout
- timeout responses return `504` with envelope code `REQUEST_TIMEOUT`
- when shutdown draining starts, new application requests return `503` with envelope code `NOT_READY`

## Example Resource

### Create Request

```json
{
  "name": "demo"
}
```

### Example Object

```json
{
  "id": "exm_0123456789abcdef",
  "name": "demo",
  "created_at": "2026-04-16T12:00:00Z",
  "updated_at": "2026-04-16T12:00:00Z"
}
```
