# Rules Toolkit Templates

这是一套可移植的工程约束模板，目标是把“软提示”与“硬校验”拆开，方便复用到其他仓库。

当前源码仓通过 `rules` CLI 分发这套 toolkit。安装到目标仓库后，
正式运行态是根级 `.orch/rules/`、`.githooks`、`Makefile.rules`，以及由 `install` 自动刷新的
`.orch/rules-kit/` standalone toolkit。旧的根级 `rules-kit/` 不再是必需目录。下面的命令示例默认以“在目标仓库中通过已安装 CLI 执行”为主。

如果你正在当前源码仓中开发或调试 toolkit，优先使用统一入口：

```bash
rules install /path/to/repo --mode existing --targets "backend:go:app"
rules doctor /path/to/repo --strict
rules install --local /path/to/repo --yes
rules export /tmp/rules-kit
rules install --hooks /path/to/repo
rules init --lang go /tmp/new-service --name github.com/acme/new-service
rules init --lang py /tmp/new-app --name acme_demo
rules install /path/to/repo --legacy starter
```

也可以把当前目录当作本地 Python CLI 包安装：

```bash
pip install -e .
rules --help
```

更适合 CLI 隔离安装的方式：

```bash
pipx install .
```

这层 `pip/pipx` 安装只负责分发 CLI 和模板资产；安装到目标仓库后，正式运行态仍然会写入
`.orch/rules/`、`.githooks/` 和 `Makefile.rules`。

在目标仓库根目录里，直接执行一次：

```bash
rules
```

默认等价于：

```bash
rules install . --local
```

## 公开入口

日常只需要记两个命令：

```bash
rules install /path/to/repo --mode existing --targets "backend:go:app"
rules doctor /path/to/repo --strict
```

现在 `install` 在未显式传入 `--mode`、`--layout`、`--targets` 时，会优先自动探测：

- `repo_mode`: `new` 或 `existing`
- `layout`: `single` 或 `monorepo`
- `targets`: 常见 Go / TypeScript / Python 单仓与 monorepo 目标

高置信度场景会直接采用探测结果；如果探测到的是模糊 monorepo，则会拒绝猜测并要求显式传入 `--targets`。
这些 detect 与参数归并逻辑现在优先在 `rules` 的 Python 层完成，再把显式参数下发给底层安装器。
同一层里还包含 target spec 解析和核心生成文件渲染模型，用于稳定 `manifest.json`、`global.env`、`<target>/rules.env`、`<target>/arch.rules`、`<target>/docsync.rules` 与 `Makefile.rules` 的输出。
当前 installed CLI 下，`rules install`、`rules init`、`rules doctor`、`rules install --hooks`、`rules install --local` 和 `rules install --legacy ...` 的主路径都优先走 Python；明确保留 shell helper 语义的主要只剩 exported standalone toolkit 本身。
也就是说：

- 已安装的 `rules` 命令优先走 Python runtime
- `rules export` 导出的 `rules-kit/` 目录则有意保留为自包含 shell toolkit，便于在没有安装 Python 包的环境里分发和使用

`rules install` 现在还会在目标仓库下自动刷新 `.orch/rules-kit/`。这为 `preflight` 提供了仓库内标准 fallback：

```bash
rules preflight .
bash .orch/rules-kit/preflight.sh .
```

推荐顺序是先尝试 `rules preflight .`；如果当前环境的 `PATH` 看不到 `rules`，再退回到仓库内的 `.orch/rules-kit/preflight.sh`。

如需启用本地 Git hooks，可以额外执行：

```bash
rules install --hooks /path/to/repo
```

如果你想先把它打包成一个可独立拷贝的目录，再发到别的仓库使用，可以执行：

```bash
rules export /tmp/rules-kit
rules export /tmp/rules-kit --force
```

默认不会覆盖一个已有且非 toolkit 导出的目录；需要显式传 `--force`。

如果你想直接初始化一个新仓库，可以执行：

```bash
rules init --lang go /tmp/new-service --name github.com/acme/new-service
rules init --lang ts /tmp/new-web --name @acme/new-web
rules init --lang py /tmp/new-app --name acme_demo
```

如果你需要直接落 legacy profile，可以执行：

```bash
rules install /path/to/repo --legacy starter
rules install /path/to/repo --legacy minimal --force
```

迁移步骤清单见 [MIGRATION_CHECKLIST.md](MIGRATION_CHECKLIST.md)。

如果你希望让上层 AI 或平台先产出仓库清单，再交给安装器执行，可以改用：

```bash
rules install /path/to/repo --manifest /path/to/manifest.json
rules doctor /path/to/repo --strict
```

预演但不落盘：

```bash
rules install /path/to/repo --dry-run --mode new --layout monorepo --targets "backend:go:app,frontend:ts:web"
```

`--dry-run` 会先输出一份 Python 侧的解析计划摘要，再输出底层安装器的详细 dry-run 动作。

已安装 CLI 的公开安装入口是 `rules install`。如果你只是想看 standalone toolkit 中对应的 helper 命令，可以执行：

```bash
rules --print install /path/to/repo --mode existing --targets "backend:go:app"
```

导出的 standalone toolkit 中，对应入口仍然是 `install.sh`。它负责：

- 识别或创建 `existing` / `new` 仓库
- 选择 `single` / `monorepo` 布局
- 配置 `go` / `ts` / `py` / 混合 target
- 生成仓库级 `.orch/rules/manifest.json`
- 安装仓库运行态约束文件、`.githooks/` 和根级 `Makefile.rules`
- 生成根级 `AGENTS.md`
- 提供一个可选的本地 Git hooks 启用动作，但不会自动修改开发机 Git 配置

`rules install` 当前公开参数是显式的，不再依赖任意参数透传。主参数包括：

- `target_repo`
- `--manifest`
- `--mode`
- `--layout`
- `--targets`
- `--yes`
- `--force`
- `--dry-run`
- `--legacy`
- `--hooks`
- `--local`
- `--target`

如果传入 `--manifest`，则优先从 manifest 读取 `repo_mode`、`layout` 和 `targets`，命令行显式传入的同名参数会覆盖 manifest。
已安装 CLI 的 `rules install` 会先对输入 manifest 和即将写出的 manifest 做显式校验；`--dry-run` 会保留同样的校验，但只输出计划动作，不落盘。

已安装 CLI 的公开检查入口是 `rules doctor`。如需查看 standalone toolkit 中对应的 helper 命令，可以执行：

```bash
rules --print doctor /path/to/repo
```

导出的 standalone toolkit 中，对应检查入口仍然是 `doctor.sh`。它优先读取 `.orch/rules/manifest.json`，先做 manifest 显式校验，再逐个 target 检查 Go、TS、Python 或混合 monorepo 的接线状态，并确认 `.githooks/` 已落盘。

如需机器可读输出，可以执行：

```bash
rules doctor /path/to/repo --json
```

这套主路径默认会生成并检查根级 `AGENTS.md`。新版本会把 `AGENTS.md` 写成“受管块 + 用户块”：

- `rules install` 会自动刷新受管块
- 仓库自己补充的说明应写在用户块中
- 对没有 marker 的旧版 `AGENTS.md`，默认保留原文件；传 `--force` 时会迁移到用户块并写入最新受管块
- 刷新仓库生成结果时，默认先执行 `rules install ... --yes`；只有在迁移旧版 `AGENTS.md` 或需要覆盖受管 base 文件时才使用 `--force --yes`

## 常见场景

已有仓库：

```bash
rules install /path/to/repo --mode existing --targets "backend:go:app,frontend:ts:web"
rules doctor /path/to/repo --strict
```

新建单体 Go 仓库：

```bash
rules install /path/to/repo --mode new --layout single --targets "app:go:."
rules doctor /path/to/repo --strict
```

新建单体 Python 仓库：

```bash
rules init --lang py /tmp/new-app --name acme_demo
rules doctor /tmp/new-app --strict
```

新建混合 monorepo：

```bash
rules install /path/to/repo --mode new --layout monorepo --targets "backend:go:app,frontend:ts:web"
rules doctor /path/to/repo --strict
```

## 设计原则

- `AGENTS.md` 是仓库内默认软约束入口；现在由 `rules install` 维护受管块，并保留仓库自定义用户块。
- `AGENTS.base.md` 作为兼容参考保留在 toolkit 资产中，可按需写入目标仓库。
- `scripts/verify-*.sh` 放可执行、可阻塞的校验。
- `verify-layout.sh` 负责可参数化的目录/布局硬约束。
- `verify-doc-sync.sh` 负责代码路径与文档路径的一致性硬约束。
- `verify-contract-parity.sh` 提供插件化的路由/契约一致性框架，toolkit 负责比较 manifest，具体 route/contract 导出由仓库实现。
- `verify-change-scope.sh` 提供插件化的 change gating 框架；当前内置 OpenSpec adapter，读取 change metadata 并校验改动范围。
- 所有 verify 脚本都会先读取 `.orch/rules/<target>/rules.env`，再按约定自动叠加 `.orch/rules/<target>/local.env`；适合把仓库私有覆盖项放在 local 文件里，避免 `upgrade` 被覆盖。
- 日常调整仓库私有路径或 env 约束时，优先修改 `.orch/rules/<target>/local.env`；只有在明确调整受管默认值时才直接修改 `.orch/rules/<target>/rules.env`。
- `verify-arch.sh` 会同时读取 `.orch/rules/<target>/arch.rules` 和 `.orch/rules/<target>/local.arch.rules`。
- `verify-doc-sync.sh` 会同时读取 `.orch/rules/<target>/docsync.rules` 和 `.orch/rules/<target>/local.docsync.rules`。
- `rules install --local /repo --yes` 会为每个 target 生成缺失的 `local.*` 文件，并把当前 base env 里的常见覆盖项复制到 `.orch/rules/<target>/local.env`，便于把仓库私有约束迁到 local 层。
- Go target 现在内置 `service_layered` 骨架 profile：新建仓库会生成 `cmd/`、`internal/{config,runtime,transport,service,infra}`、`openapi/`、`db/migrations/` 等目录，并为 Go import graph 生成分层规则起点。
- `lint/` 放运行期需要的静态分析配置。
- `.githooks/` 放本地仓库 hook 资产，是否启用由开发者显式决定。
- `manifest.json` 是新的仓库级配置入口；正式运行态使用 `.orch/rules/global.env` 与 `.orch/rules/<target>/rules.env`。
- `manifest.json` 负责 target 结构、target 根目录和受管 env 文件定位；仓库私有覆盖应落在 `.orch/rules/<target>/local.env`。

## 标准资产目录

为了让包结构更稳定，`rules_kit` 现在额外提供一层标准资产入口：

- `templates/`: 共享基础模板入口，例如 `AGENTS.base.md`、`README.base.md`、`Makefile.base`
- `examples/`: manifest 示例
- `policies/`: review policy 与协作模板
- `scripts/`: repo-side verify 与 helper 脚本

当前 export 的 standalone toolkit 仍然会保留原有兼容目录；这些标准目录主要用于 Python 包内读取、后续模板收敛和测试对齐。

## Manifest 结构

`rules install` 会生成：

- `.orch/rules/manifest.json`
- `.orch/rules/global.env`
- `.orch/rules/<target>/rules.env`
- `.githooks/pre-commit`
- `.githooks/pre-push`
- `Makefile.rules`
- `AGENTS.md`

manifest 核心字段：

- `repo_mode`
- `layout`
- `targets[].name`
- `targets[].language`
- `targets[].root`
- `targets[].env_file`

模板示例见 [repo/manifest.example.json](repo/manifest.example.json)。

## 兼容与内部脚本

- `install-into-repo.sh`
- `fixup.sh`
- `init-go-service.sh`
- `init-ts-service.sh`
- `minimal` / `starter` / `go-starter` / `ts-starter` legacy profile 导出

这些仍保留，但现在属于兼容层或内部辅助，不再是推荐主路径。

## 注意事项

- 这些模板默认依赖 `bash`、`git`、`rg`，个别脚本还会使用 `python3`。
- `verify-contract-sync.sh` 提供的是目录级强提醒模板，不试图自动推断“哪一次接口代码修改一定意味着契约变化”；最终仍建议配合 review 或 API diff 工具。
- `verify-arch.sh` 是最保守的占位模板，建议目标仓库自行补分层规则。
- `verify-arch.sh` 现在支持后端选择：`text` 继续使用保守的文本匹配；Go target 可配置 `ARCH_BACKEND=go_imports`，基于真实 import graph 校验目录分层依赖。
- `verify-arch.sh` / `verify-arch-go.py` 的规则格式兼容三段式，也支持可选第 4 段 `recommended fix`；命中时会在失败输出里提示推荐迁移位置。
- `GO_LAYOUT_PROFILE=service_layered` 只针对 Go target；新建 Go 仓库默认启用这套 required dirs，已有仓库安装 rules 时不会强制打开，避免误伤现有结构。
- `service_layered` 还会为新建 Go 仓默认填充一组保守的 `LAYOUT_FORBIDDEN_DIRS`，用于阻止常见漂移目录，例如 `internal/handlers`、`internal/repository`、`internal/storage`。
- `verify-contract-parity.sh` 当前内置 `manifest` backend，适合先让仓库自行导出 `route manifest` 与 `contract manifest` 再做一致性对比；后续可在此之上扩展 Go+OpenAPI 等专用适配器。
- `verify-change-scope.sh` 当前内置 `openspec` backend，要求 change 带 `.openspec.yaml` 且声明 `scope_paths`；toolkit 只负责 active change 选择和 diff 路径校验。
- `verify-change-scope.sh` 还支持：
  - `CHANGE_SCOPE_MODE=strict|allow_docs|off`
  - `CHANGE_SCOPE_CODE_PATHS="<path> <path>"`：只对这些代码路径启用 gating
  - `CHANGE_SCOPE_ALLOW_PATHS="<path> <path>"`：允许少量共享路径在 change scope 之外修改
- `verify-layout.sh` 默认未配置即跳过；目标仓库可通过 `LAYOUT_REQUIRED_DIRS` / `LAYOUT_FORBIDDEN_DIRS` 打开目录级硬约束。
- `verify-doc-sync.sh` 默认未配置即跳过；目标仓库可通过 `DOC_SYNC_RULES_FILE` 把实现路径和文档路径绑定起来。
- 安装器现在会为每个 target 生成一个带注释模板的 `*.docsync.rules` 文件，并把 `DOC_SYNC_RULES_FILE` 指向它；是否填入实际规则由目标仓库决定。
- `rules install --force --yes` 会刷新 `.orch/rules/<target>/rules.env`、`.orch/rules/<target>/arch.rules`、`.orch/rules/<target>/docsync.rules` 这些 base 文件；如果仓库需要保留手写定制，请写到 `.orch/rules/<target>/local.env`、`.orch/rules/<target>/local.arch.rules`、`.orch/rules/<target>/local.docsync.rules`。
