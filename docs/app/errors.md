# 错误码说明

这个仓库为横切层保留了一组最小错误码基线。

## 保留错误码

| Code | HTTP Status | Retryable | 含义 |
| --- | --- | --- | --- |
| `INTERNAL` | `500` | 通常为 `false` | 服务端发生了预期外错误。 |
| `INVALID_ARGUMENT` | `400`、`413` 或 `415` | `false` | 请求输入不合法、协议不匹配，或请求体超过允许上限。 |
| `NOT_READY` | `503` | `true` | 依赖未就绪，或服务正处于 draining。 |
| `REQUEST_TIMEOUT` | `504` | `true` | 请求超过了服务端配置的超时。 |

## 使用规则

- 从这个 starter 派生出的服务应保持这些错误码稳定
- 新的领域错误码应建立在这组基线之上，而不是直接改动它
- 新增横切层错误码前，先在这里补文档，再在 transport 或 middleware 中使用
- 业务失败优先使用领域错误码，基础设施或协议层行为优先使用保留错误码
- `INVALID_ARGUMENT` 可以同时覆盖 JSON 解码失败、空请求体、unknown fields、错误的 `Content-Type`、多个 JSON document 以及 body limit 超限等 transport 层输入错误
- service/shared 只保留领域错误语义，例如 `Code`、`Kind`、`Retryable` 和 `details`
- HTTP status 由 `internal/transport/httpapi/httpx` 基于 `shared.Kind` 统一映射，不在领域错误对象里保存 transport 语义
- 错误日志字段名统一归 `internal/runtime/logging`，不要回流到 `service/shared`

## 领域错误码规则

- 领域错误码应在对应 domain 包内集中声明为常量，不要在 handler、repository 或测试里散写字符串
- 命名建议使用大写蛇形命名，并带上资源前缀，例如 `EXAMPLE_NOT_FOUND`
- 常见的 HTTP 映射应通过共享错误构造器固定下来，再由 transport 统一写回 envelope

## Domain Error Checklist

每个新增 domain 默认按下面的清单落地：

1. 在对应 domain 包内新增或维护 `error_codes.go`
2. 把领域错误码声明为常量，不在调用点散写字符串
3. 为常见领域错误提供 helper，例如 `ErrNotFound()`
4. 通过 `shared.ErrInvalidArgument`、`shared.ErrNotFound` 这类共享构造器固定 HTTP status
5. 对外 `message` 使用稳定短句，不拼接底层驱动或第三方原始错误
6. 参数校验类错误优先附带 `details.field_errors`
7. 在 service、repository、handler tests 中至少覆盖一个典型领域错误
8. 如果新增的是跨多个接口都会出现的领域错误，补充本文件或对应 domain 文档说明

## Message 与 Details

- `message` 是返回给客户端的稳定文案，不应直接拼接底层驱动、SQL 或第三方 SDK 的原始错误
- 底层原始错误原因应进入日志，不应直接作为响应文案透出
- `details` 是可选的机器可读扩展字段，用于承载结构化错误信息
- 参数校验类错误优先通过 `details.field_errors` 返回字段级信息

### Details V1 允许结构

第一版只允许下面这些形态：

- `details.field_errors`
  - 类型：数组
  - 元素结构：`{ "field": string, "message": string }`
  - 用途：字段级参数校验失败

当前不建议在 starter 里直接扩展这些 key：

- `reason`
- `hint`
- `resource`
- 任意自定义深层嵌套对象

如果后续确实需要新增新的 `details` 结构，应先更新本文档，再更新 OpenAPI 和相关测试。

## 共享构造器

- `shared.ErrInternal`
- `shared.ErrInvalidArgument`
- `shared.ErrNotReady`
- `shared.ErrRequestTimeout`
- `shared.ErrNotFound`
- `shared.WithDetails`
- `shared.WithFieldErrors`
