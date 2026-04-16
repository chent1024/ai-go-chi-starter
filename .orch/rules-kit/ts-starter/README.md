# TypeScript Starter Profile

这个 profile 基于通用 `starter`，但预填了更适合 TypeScript 服务仓的默认值。

## 额外内容

- 面向 Node.js/TypeScript 项目的 `Makefile.template`
- 预填的 `rules.env`
- 一份最小 `package.json` / `tsconfig.json` 骨架的初始化入口

## 默认假设

- 仓库使用 `npm` 作为默认包管理器
- 源码位于 `src/`
- 配置入口位于 `src/config/env.ts`
- 通过 `make verify` / `make verify-strict` 或等价命令接入约束

## 当前定位

这个目录现在主要作为 `install.sh` 的 TypeScript 默认值来源，不再推荐用户直接拿 profile 名称安装。

公开用法改为：

```bash
bash rules-kit/install.sh /path/to/repo --mode new --layout single --targets "app:ts:."
cd /path/to/repo
bash rules-kit/doctor.sh . --strict
```
