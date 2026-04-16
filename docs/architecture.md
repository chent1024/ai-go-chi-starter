# 架构说明

## 分层

- `cmd/*`：只负责进程入口和装配
- `internal/transport/httpapi`：只负责 HTTP 协议层和 middleware
- `internal/service`：负责业务规则和领域服务
- `internal/infra`：负责 PostgreSQL、outbound client 等具体适配器
- `internal/runtime`：负责日志、trace、span、outbound logging 等横切基础设施
- `internal/config`：仓库里唯一允许读取 env 的包
- outbound HTTP client 共享一套 transport profile：timeout、keep-alive 连接池、trace 透传、child span 和 outbound logging 都集中配置
- API 还内置 body limit、安全响应头、build info 和基础 metrics

## 进程装配

### API

`cmd/api` 负责加载配置、创建 logger、按显式连接池配置打开 PostgreSQL、装配 example service 和 handler，然后启动带 graceful shutdown 的 chi 服务。HTTP server timeout 和 draining 状态也在这里完成 wiring，shutdown 期间会拒绝新的业务请求。

### Worker

`cmd/worker` 负责加载配置、创建 logger，并通过 `JobHandler` 接口运行最小 ticker loop。shutdown 日志会记录 drain 开始、当前 inflight job 数和 drain 完成。

### Migrate

`cmd/migrate` 负责加载配置、打开 PostgreSQL、获取 PostgreSQL advisory lock，并执行 `db/migrations` 下的 forward-only SQL 文件。

## HTTP 横切层

- request ID middleware 保证请求里始终有 `X-Request-Id`
- trace middleware 保证请求里始终有 `Traceparent`
- request ID middleware 除了写入 header，也会把 request id 放进 context，便于 outbound client 和内部 span 统一打点
- request timeout middleware 为每个请求施加 context deadline
- request timeout middleware 在 deadline 之后会丢弃后续写入，避免 late write 覆盖超时响应
- body limit middleware 为每个请求施加统一 body size 上限
- security headers middleware 为响应写入基础安全头
- drain middleware 在 graceful shutdown 开始后拒绝新的业务请求
- access log middleware 记录路由、耗时、字节数和状态码
- recover middleware 将 panic 转成统一错误 envelope

## 日志与 Trace 基线

- request logger 默认带 `service`、`request_id`、`trace_id`、`span_id`
- outbound request log 会复用当前 context 中的 `request_id` 和 trace/span 字段
- `APP_LOG_OUTBOUND_ENABLED` 只控制 outbound 成功日志；失败日志始终保留，级别为 `warn/error`
- API、worker、migrate 在 logger 完成装配前，使用 bootstrap logger 输出结构化 fatal 日志，避免启动早期只剩裸 stderr
- runtime 提供最小 span API：`runtime.StartSpan(...).End(...)`
- span 默认只在 `debug` 级别输出，用于追踪链路，不替代顶层错误日志
- `/metrics` 除请求总数和延迟总和外，还会暴露 in-flight request、process uptime 和按路由的延迟最大值
- timeout middleware 会记录 `http_request_timeout_late_write_total{route=...}`，用于发现超时后仍继续写响应的链路

## Child Span 约定

- HTTP 入站请求是整条链路的父 trace
- outbound HTTP 调用创建 child span：`outbound.http.roundtrip`
- PostgreSQL repository 调用创建 child span：`postgres.<resource>.<operation>`
- worker job 执行创建 child span：`worker.job.handle`
- 后续新增 domain/service 的内部 span，应沿用 `<layer>.<resource>.<operation>` 命名模式
- 只有跨边界或明显长耗时步骤才应创建 child span；普通 getter、DTO 转换、轻量校验不要滥起 span
- 新 span 必须基于当前 request/job 的 context 继续派生，不允许脱离当前链路使用 `context.Background()`
- 不要在逐行扫描、逐条列表项、紧密循环里创建 per-item span，避免把 debug trace 噪音放大
- handler 不应重复创建“http.*”类型 span；HTTP 父链路已由 transport middleware 建立，业务层只补真正有价值的子步骤

## Context 取消约束

- service、repository、outbound client、worker handler 都必须接受 `context.Context`
- 所有长耗时操作必须主动尊重 `ctx.Done()`，而不是只依赖顶层 timeout middleware
- PostgreSQL 访问应使用 `QueryContext`、`ExecContext`、`BeginTx`
- 出站 HTTP 应使用 `http.NewRequestWithContext`
- timeout middleware 只能阻止超时后的 late write，不能强制杀掉不响应取消的 goroutine；真正的退出责任仍在下游实现

## Example 领域示例

`example` 资源用来示范推荐的职责边界：

- transport 负责 DTO 解码和编码
- service 负责输入校验和 ID 生成
- repository 负责 SQL 访问并返回领域模型

这样可以让 handler、service、repository 的职责保持稳定，也便于 Codex 在这个结构上继续扩展。
