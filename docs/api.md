# API 文档

## Envelope

所有 JSON 响应都使用同一层 envelope：

```json
{
  "code": "OK",
  "message": "",
  "request_id": "req_123",
  "data": {}
}
```

错误响应会额外携带 `retryable`。

横切层保留错误码见 [docs/errors.md](/Users/xihe0000/workspace/code/ai-go-chi-starter/docs/errors.md)。

## 路由

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/healthz` | 存活探针。 |
| `GET` | `/readyz` | 基于数据库探活的就绪探针。 |
| `POST` | `/v1/examples` | 创建 example。 |
| `GET` | `/v1/examples/{id}` | 获取单个 example。 |
| `GET` | `/v1/examples` | 获取 example 列表。 |

## 横切行为

- 每个请求都会获得 `X-Request-Id` 和 `Traceparent`
- 每个请求都会受到服务端 request timeout 约束
- 请求超时时返回 `504`，envelope code 为 `REQUEST_TIMEOUT`
- graceful shutdown 进入 draining 后，新业务请求返回 `503`，envelope code 为 `NOT_READY`

## Example 资源

### 创建请求

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
