# Migration Checklist

这份清单默认面向 Go、TypeScript 或 Go+TypeScript 混合仓库。

## 5 步迁移

1. 准备好 `rules` CLI。
   - 本地开发可直接：
     ```bash
     pip install -e .
     ```
   - 如果需要离线分发一份独立目录，再导出：
     ```bash
     rules export /tmp/constraint-package
     ```

2. 用单入口安装器写入仓库约束。
   - 已有仓库：
     ```bash
     rules install /path/to/repo --mode existing --targets "backend:go:app,frontend:ts:web"
     ```
   - 新仓库：
     ```bash
     rules install /path/to/repo --mode new --layout monorepo --targets "backend:go:app,frontend:ts:web"
     ```
   - 如果上层系统先生成仓库清单，就用：
     ```bash
     rules install /path/to/repo --manifest /path/to/manifest.json
     ```

3. 立刻运行体检，先看结构是否接好。
   ```bash
   cd /path/to/repo
   rules doctor .
   rules doctor . --strict
   ```
   `--strict` 失败时，优先修：
   - target root 配置错误
   - 配置入口路径错误
   - `.env.example` 缺失
   - `ARCH_IMPORT_RULES_FILE` 缺失

4. 把验证入口接到目标仓库的 CI / 分支保护，并按需启用本地 hooks。
   - 确认根级 `Makefile` 已经 include `Makefile.rules`
   - 确认你的 GitLab / Gitee CI 会执行 `make verify` 和 `make verify-strict`
   - 如需本地提交前校验，执行 `rules install --hooks .`
   - 确认根级 `AGENTS.md` 已存在并按仓库语义调整
   - 如果你的 AI 平台支持仓库级软规则，可在平台侧直接配置，不再依赖 toolkit 额外导出的文本文件

5. 按目标仓库语义补最少的项目特定规则。
   - 优先修改 `.orch/rules/<target>/local.env`
   - 只有在明确调整受管默认值时才直接修改 `.orch/rules/<target>/rules.env`
   - 补 `.orch/rules/<target>/arch.rules`
   - 调整前后端各自的 contract 路径、config 路径、migration 路径
   - 再跑一遍：
     ```bash
     rules doctor . --strict
     ```

## 迁移后建议立刻确认的 6 项

- `.orch/rules/manifest.json` 是否反映了真实 target 列表
- `Makefile.rules` 是否已被根 `Makefile` 引入
- `.githooks/pre-commit` 与 `.githooks/pre-push` 是否已落盘
- `.orch/rules/<target>/rules.env` 中的路径是否真实存在
- `.orch/rules/<target>/local.env` 中的仓库私有覆盖是否与当前仓库一致
- `.orch/rules/<target>/arch.rules` 是否已经不再是占位示例

## 什么时候不要直接复用

先不要直接复用当前模板，如果目标仓库：

- 不是 Go / TypeScript 仓库
- 没有 `Makefile` 或不希望用 `make` 作为统一验证入口
- 配置加载方式完全不是文件路径扫描可识别的模式
- 不接受 root 级 `Makefile.rules` 之外的旧接线方式

这几类仓库应该先做一轮模板裁剪，再迁移。
