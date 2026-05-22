# NextM

**便捷 · 智能 · 可视化** — 下一代个人知识管理工具（PKM），你的第二大脑。

NextM 是一款本地优先、AI 驱动的知识管理工具，帮助知识工作者高效采集、组织、检索和复用碎片化信息。支持多端同步、多智能体协作、视频截屏即笔记、CRDT 冲突-free 同步等特性。

---

## 项目状态

> **当前阶段**: 设计/规划阶段 — 核心设计文档已完成，即将进入 MVP 编码阶段。

## 文档目录

| 文档 | 版本 | 说明 |
|------|------|------|
| [产品需求说明书](NextM-PRD.md) | v1.2 | 产品定位、功能清单、用户画像、路线图 |
| [产品设计说明书](NextM-PDD.md) | v1.0 | UI/UX 交互设计、信息架构、原型 |
| [架构设计说明书](NextM-ADD.md) | v1.0 | 后端模块划分、技术选型、部署拓扑 |
| [OpenAPI 接口规范](NextM-API.md) | v1.0 | REST API 定义 |
| [数据库设计说明书](NextM-DB.md) | v1.0 | 47 张表、ER 图、索引策略 |
| [开发环境搭建指南](docs/guides/setup.md) | — | 从零搭建开发环境 |
| [编码规范](docs/guides/coding-standards.md) | — | Go / TypeScript / Dart 编码约定 |
| [Git 工作流规范](docs/guides/git-workflow.md) | — | 分支策略、Commit 规范 |
| [测试策略](docs/guides/testing.md) | — | 单元/集成/E2E 测试规范 |
| [贡献指南](CONTRIBUTING.md) | — | PR 流程、开发指南 |

## 技术栈概览

| 领域 | 选型 |
|------|------|
| 后端语言 | Go 1.24+ |
| 前端 Web | React 19 + Vite + Tailwind CSS 4 |
| 桌面端 | Tauri 2 (Rust) |
| 移动端 | Flutter 3 (Dart) |
| 浏览器扩展 | Chrome / Firefox Extension |
| 本地数据库 | SQLite (modernc.org/sqlite) |
| 云端数据库 | PostgreSQL + pgvector |
| 搜索引擎 | SQLite FTS5 / Meilisearch |
| 向量搜索 | LanceDB / pgvector |
| 消息总线 | NATS JetStream |
| AI 网关 | LiteLLM |
| 同步引擎 | CRDT (Yrs) |

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/your-org/nextm
cd nextm

# 安装依赖（详见 docs/guides/setup.md）
make deps

# 启动开发服务
make dev

# 运行测试
make test
```

## 路线图

| 阶段 | 周期 | 核心交付 |
|------|------|---------|
| **MVP** | 第 1-3 月 | Web 端 + 快速剪藏 + 基础笔记 + 全文搜索 + 单账号 |
| **Phase 2** | 第 4-6 月 | 桌面端 + 移动端 + 多端同步 + 图片提取 + 多账号 |
| **Phase 3** | 第 7-9 月 | AI 问答 + 知识图谱 + 智能标签 + 智能体架构 |
| **Phase 4** | 第 10-12 月 | 智能体通信 + 团队空间 + API 开放 + 模板市场 |

## 许可

[License TBD]
