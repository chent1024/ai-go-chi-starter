# Go Starter Profile

这个 profile 基于通用 `starter`，但预填了更适合 Go 服务仓的默认值。

## 额外内容

- 面向 Go 项目的 `Makefile.template`
- 预填的 `rules.env`
- 一份可直接作为起点的 `.golangci` 基线
- `service_layered` 骨架 profile 默认值与对应的 Go import graph 分层规则起点

## 默认假设

- 仓库使用 `make verify` 和 `make verify-strict`
- 存在 `internal/config/config.go` 或等价配置入口
- 普通 `make verify` 使用 `go test`
- `make verify-strict` 再叠加 `go test -race` 和 `go test -bench`
- 希望通过 `.orch/rules/make/verify.mk` 接入可移植约束
- 新建 Go 服务更接近这种目录结构：
  - `cmd/{api,worker}`
  - `internal/{config,runtime,transport/httpapi,service,infra}`
  - `openapi/`
  - `db/migrations/`
- 默认还会给出一组保守的禁止目录：
  - `internal/handlers`
  - `internal/repository`
  - `internal/storage`
  - `internal/upload`
  - `internal/misc`
  - `internal/tmp`

## 当前定位

这个目录现在主要作为 `install.sh` 的 Go 默认值来源，不再推荐用户直接拿 profile 名称安装。

公开用法改为：

```bash
bash rules-kit/install.sh /path/to/repo --mode new --layout single --targets "app:go:."
cd /path/to/repo
bash rules-kit/doctor.sh . --strict
```
