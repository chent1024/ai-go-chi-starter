# API 文档

## Envelope

除 `/metrics` 外，所有 JSON 响应都使用同一层 envelope：

```json
{
  "code": "OK",
  "message": "",
  "request_id": "req_123",
  "data": {}
}
```

错误响应会额外携带 `retryable`，并可选携带机器可读的 `details`。

```json
{
  "code": "INVALID_ARGUMENT",
  "message": "name is required",
  "request_id": "req_123",
  "data": null,
  "details": {
    "field_errors": [
      {
        "field": "name",
        "message": "is required"
      }
    ]
  },
  "retryable": false
}
```

横切层保留错误码见 `docs/app/errors.md`。

## 路由

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/healthz` | 存活探针。 |
| `GET` | `/readyz` | 基于数据库探活的就绪探针。 |
| `GET` | `/version` | 返回当前 API build 信息。 |
| `GET` | `/metrics` | 返回 Prometheus 文本指标。 |
| `POST` | `/v1/examples` | 创建 example。 |
| `GET` | `/v1/examples/{id}` | 获取单个 example。 |
| `GET` | `/v1/examples` | 获取 example 列表。 |

## 横切行为

- 每个请求都会获得 `X-Request-Id` 和 `Traceparent`
- 每个请求都会带上基础安全响应头
- 每个请求体都会受到统一 body size limit 约束
- 所有写入型 JSON 接口都要求 `Content-Type: application/json`
- 所有写入型 JSON 接口都只接受单个 JSON document，不接受多个 JSON 对象拼接
- 每个请求都会受到服务端 request timeout 约束
- 如果请求已经超时，后续 handler 写入会被丢弃，最终统一返回 `504`
- 请求超时时返回 `504`，envelope code 为 `REQUEST_TIMEOUT`
- graceful shutdown 进入 draining 后，新业务请求返回 `503`，envelope code 为 `NOT_READY`

## Example 资源

### 创建请求

创建 `example` 时，请求必须满足：

- `Content-Type: application/json`
- body 中只能包含单个 JSON 对象

```json
{
  "name": "demo"
}
```

### Example 对象

```json
{
  "id": "exm_0123456789abcdef",
  "name": "demo",
  "created_at": "2026-04-16T12:00:00Z",
  "updated_at": "2026-04-16T12:00:00Z"
}
```
