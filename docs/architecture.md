# 架构说明

## 分层

- `cmd/*`：只负责进程入口和装配
- `internal/transport/httpapi`：只负责 HTTP 协议层和 middleware
- `internal/service`：负责业务规则和领域服务
- `internal/infra`：负责 PostgreSQL、outbound client 等具体适配器
- `internal/runtime`：负责日志、trace、outbound logging 等横切基础设施
- `internal/config`：仓库里唯一允许读取 env 的包
- outbound HTTP client 共享一套 transport profile：timeout、keep-alive 连接池、trace 透传和 outbound logging 都集中配置

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
- request timeout middleware 为每个请求施加 context deadline
- drain middleware 在 graceful shutdown 开始后拒绝新的业务请求
- access log middleware 记录路由、耗时、字节数和状态码
- recover middleware 将 panic 转成统一错误 envelope

## Example 领域示例

`example` 资源用来示范推荐的职责边界：

- transport 负责 DTO 解码和编码
- service 负责输入校验和 ID 生成
- repository 负责 SQL 访问并返回领域模型

这样可以让 handler、service、repository 的职责保持稳定，也便于 Codex 在这个结构上继续扩展。
