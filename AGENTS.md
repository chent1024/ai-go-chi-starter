# 仓库约束

这个仓库不依赖 prompt 自觉执行规则。所有改动默认必须通过 `make verify`；涉及并发、性能、存储、队列、runtime wiring、契约或 migration 的改动，必须通过 `make verify-strict`。

## 1. 强制验证

- 普通改动完成后必须执行 `make verify`
- 涉及并发、性能、存储、队列、runtime wiring、契约或 migration 的改动，必须执行 `make verify-strict`
- 不允许把“代码已改但未验证”视为完成态
- 用户请求“提交”或“推送”时，必须先完成对应验证，再执行 `git add`、`git commit`、`git push`
- `make verify` 或 `make verify-strict` 未通过前，不得将提交、推送、已完成或“可稍后补验证”视为可接受状态
- 除非用户明确要求，不得建议或使用 `--no-verify`、跳过 hooks 或其他绕过仓库校验的做法

## 2. 约束入口

- `.orch/rules/` 是当前仓库的运行态约束目录；`.githooks/`、`Makefile.rules` 和根级 `AGENTS.md` 一起构成正式约束入口
- 调整校验目标、路径或 env 约束时，优先修改 `.orch/rules/manifest.json` 或 `.orch/rules/<target>/rules.env`
- 不要手工改 `Makefile.rules`；需要刷新生成结果时，执行 `rules install . --manifest .orch/rules/manifest.json --force --yes`
- 除纯文档和小型说明性改动外，开始实现前默认先执行 `rules preflight`
- 如果改动命中仓库配置的 OpenSpec 受控路径，且 `rules preflight` 要求先创建 change，则必须先创建 OpenSpec change 再继续实现
- `CHANGE_SCOPE_ACTIVE_CHANGE` 仅用于多个 active OpenSpec changes 并存时的人工歧义消解，不是常规必填项
- 运行态 target 配置位于 `.orch/rules/<target>/`；不要假设默认 target 一定叫 `service`
- 新增 env key 时，优先复用仓库现有配置读取入口与调用形态；避免引入仓库配置扫描规则无法识别的新 helper 或分散读取方式
- 新增 env key 时，默认同步 `.env.example` 以及仓库内已有的本地开发 / 部署示例 env 文件

## 3. 变更边界

- 修改前先阅读现有实现、测试和调用链
- 改动保持最小且聚焦当前任务
- 除非任务明确要求，否则不要做无关清理、大重构、批量重命名、目录搬迁、全仓格式化或依赖升级
- 不要覆盖或回退与当前任务无关的用户改动
- 优先沿用仓库现有模式、结构、技术选型、装配方式和测试风格
- 不要因为个人偏好替换框架、重做分层或引入新的抽象
- 新增或修改代码时，默认保持函数和文件短小可拆分；避免把复杂逻辑堆进超长函数
- 单个生产函数原则上不超过 80 行；单行原则上不超过 120 列，除非 URL、生成内容或无法拆分的字面量

## 4. 契约与配置同步

- 如果 API 行为、路由、请求/响应字段、鉴权要求、错误语义、配置加载、持久化行为、状态流转或并发行为发生变化，必须同步更新相关契约文件、文档和测试
- 修改 API surface 目录下的文件时，先评估是否需要同步契约文件；不要等严格校验失败后再补 contract
- 即使本次只改测试，也要预判仓库的 contract-sync / doc-sync 是否会把该目录视为 API surface 变化
- 新增或调整配置项时，除了代码，还要同步 `.env.example`、相关配置文档，并在交付说明里列出新增 key、默认值和影响到的 runtime
- 不要把“只改了代码”视为完成，前提是契约或配置文档已经过期
- 是否完成以仓库内可执行的检查结果为准，而不是以主观判断为准

## 5. 生成文件

- 不要手工修改生成文件；应修改源输入并重新执行生成命令

## 6. 测试与安全

- 测试默认不应依赖真实外部服务、本机私有状态或手工环境准备
- 涉及 API、鉴权、存储、配置、持久化、并发或状态流转的改动，通常应补充测试
- 不要在代码、fixture、日志、文档或示例中提交真实 secret、token、key 或其他敏感载荷
- 提交前检查 `git status --short`，确认 `tmp/`、`.runtime/`、本地对象存储目录等运行产物没有进入暂存区
- 如果同步了本地示例文件但该文件未受 git 跟踪，最终说明里要显式写明“已更新本地文件，但未进入版本控制”

## 7. 交付要求

- 如果新增、删除或升级依赖，必须明确说明为什么当前技术栈不足
- 如果验证失败，先说明失败来自代码、契约/文档未同步，还是仓库规则触发；在未定位失败来源前，不要声称任务已完成
- 最终交付时必须说明：
  1. 改了什么
  2. 同步了哪些契约 / 配置 / 文档
  3. 实际执行了哪些验证
  4. 剩余风险、假设或缺口

## 8. Starter 模板放置约定

- `cmd/api`、`cmd/worker`、`cmd/migrate` 只负责进程装配，不写业务规则
- `internal/transport/httpapi/v1` 只放 handler 和 HTTP DTO，不直接操作 SQL 或业务状态机
- `internal/service/<domain>` 放 domain model、service、repository interface
- `internal/infra/store/postgres` 放 PostgreSQL 具体实现，不返回 HTTP DTO
- `internal/runtime` 放日志、trace、outbound logging 等横切基础设施
- 新增 env key 时，只能在 `internal/config/config.go` 读取，并同步 `.env.example`、`deploy/.env.runtime.example`、`deploy/.env.dev.example`、`docs/config.md`
- 新增 API 路由或响应字段时，必须同步 `docs/api.md` 和 `openapi/openapi.yaml`
- 新增或调整保留错误码时，必须同步 `docs/errors.md`
