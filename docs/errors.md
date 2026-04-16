# Error Codes

This repository keeps a small reserved error code baseline for cross-cutting behavior.

## Reserved Codes

| Code | HTTP Status | Retryable | Meaning |
| --- | --- | --- | --- |
| `INTERNAL` | `500` | usually `false` | Unexpected server-side failure. |
| `INVALID_ARGUMENT` | `400` | `false` | Request input is invalid. |
| `NOT_READY` | `503` | `true` | Dependency not ready or service is draining. |
| `REQUEST_TIMEOUT` | `504` | `true` | Request exceeded the configured server-side timeout. |

## Usage Rules

- keep these codes stable across services derived from this starter
- add new domain-specific codes on top of this baseline instead of changing it
- document new cross-cutting codes here before using them in transport or middleware
- prefer domain-specific codes for business failures and reserved codes for infrastructure or protocol behavior
