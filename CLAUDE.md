# CLAUDE.md

使用中文回答所有问题。

## Project Status

**设计/规划阶段** — 尚无实现代码。详见各设计文档：

| 文档 | 说明 |
|------|------|
| [NextM-PRD.md](NextM-PRD.md) | 产品需求（功能清单、用户画像、路线图） |
| [NextM-PDD.md](NextM-PDD.md) | 产品设计（UI/UX、交互、信息架构） |
| [NextM-ADD.md](NextM-ADD.md) | 系统架构（模块划分、技术选型、部署拓扑） |
| [NextM-API.md](NextM-API.md) | OpenAPI 接口规范 |
| [NextM-DB.md](NextM-DB.md) | 数据库设计（47 表、ER 图、索引） |
| [docs/architecture/frontend.md](docs/architecture/frontend.md) | 前端架构（Web/Desktop/Mobile/Extension） |
| [docs/architecture/decisions/](docs/architecture/decisions/) | 架构决策记录 (ADR) |
| [docs/security/security-design.md](docs/security/security-design.md) | 安全设计（加密、认证、隐私合规） |
| [docs/guides/coding-standards.md](docs/guides/coding-standards.md) | 编码规范（Go/TS/Dart） |
| [docs/guides/git-workflow.md](docs/guides/git-workflow.md) | Git 工作流 |
| [docs/guides/setup.md](docs/guides/setup.md) | 环境搭建 |
| [docs/guides/testing.md](docs/guides/testing.md) | 测试策略 |

## 项目概要

**NextM** — "便捷、智能、可视化"的 PKM 工具（第二大脑）。核心差异化：视频截屏即笔记、多智能体 DAG 编排、本地优先 + CRDT 同步、多账号、MCP 协议兼容。目标用户：知识工作者、研究人员/学生、创作者。

## 目录结构

```
nextm/
├── cmd/{server,sync,worker,migrator,cli}    # 入口
├── internal/
│   ├── api/{handler,middleware,router,dto}   # HTTP 层
│   ├── service/{object,collection,search,vision,capture,
│   │            agent,relation,sync,mcp,auth,ai,export}  # 业务层
│   ├── repository/{db,sqlite,postgres}       # 数据层（SQLc）
│   ├── model/                               # 领域模型
│   ├── eventbus/                            # NATS 消息总线
│   ├── crdt/                                # CRDT 合并引擎
│   ├── vector/                              # 向量搜索
│   ├── search/                              # 搜索引擎客户端
│   ├── ai/                                  # AI 网关客户端
│   ├── config/                              # 配置管理
│   ├── telemetry/                           # 可观测性
│   └── pkg/{logger,idgen,crypto,validator,httputil}
├── frontend/{web,desktop,mobile,extension}  # 多端客户端
├── infra/{k8s,docker,terraform}             # 部署
└── docs/{api,architecture,guides}           # 文档
```

## 关键架构约束

代码生成时必须遵守以下规则：

### 分层依赖

`Handler → Service → Repository` 单向依赖，禁止循环。

- **Handler**: 解析请求、调用 Service、返回响应，不含业务逻辑
- **Service**: 业务逻辑，定义 Go interface 供 handler 依赖
- **Repository**: SQLc 生成的类型安全查询，无 ORM

### 接口定义

**Consumer-side interfaces** — 在调用方侧定义接口：

```go
// service/object/service.go（消费者侧定义接口）
type ObjectRepository interface {
    FindByID(ctx context.Context, id string) (*model.Object, error)
}

type Service struct {
    repo ObjectRepository
}
```

### 关键选型

| 领域 | 选型 | ADR |
|------|------|-----|
| HTTP 路由 | **Chi router**（stdlib 兼容） | [ADR-001](docs/architecture/decisions/ADR-001-use-chi-router.md) |
| 同步引擎 | **CRDT (Yrs)** 冲突-free 合并 | [ADR-002](docs/architecture/decisions/ADR-002-use-crdt-yrs-for-sync.md) |
| 数据访问 | **SQLc** 类型安全代码生成，无 ORM | [ADR-003](docs/architecture/decisions/ADR-003-use-sqlc-over-orm.md) |
| 消息总线 | NATS JetStream（智能体间通信） | |
| API 协议 | REST (JSON) 外部 + gRPC 内部 | |
| 本地 DB | SQLite (modernc.org/sqlite) | |
| 云端 DB | PostgreSQL + pgvector | |
| 搜索 | SQLite FTS5(本地) / Meilisearch(云端) | |
| 向量搜索 | LanceDB(本地) / pgvector(云端) | |
| AI 网关 | LiteLLM 侧车 | |
| 认证 | JWT + Refresh Token 轮转 | |
| 加密 | AES-256-GCM | |
| 配置 | Viper + envconfig | |

### DB 约定

- UUID v4 主键、Unix ms 时间戳 (UTC)、软删除
- Space 级数据隔离（`space_id`）
- SQL 文件命名：`*.sqlite.sql` / `*.postgres.sql`

### 代码生成命令

```bash
make sqlc-gen    # SQLc Go 代码生成
make mock-gen    # Mock 代码生成
make migrate-up  # 数据库迁移
```

## 测试规范

| 层 | 工具 | 覆盖率 |
|----|------|--------|
| 单元 | `go test` + testify | 核心 > 90% |
| 集成 | testcontainers-go (需 Docker) | Repository + Service |
| API | httptest | Handler |
| E2E | Playwright | 关键路径 |

```bash
make test        # 全部
make test-unit   # 仅单元（-short）
make test-int    # 含集成（-tags=integration）
```

- 测试文件与源码同目录：`service/object/object_test.go`
- Service 层 mock repository interface，不 mock 外部 API（用 test server）
- 集成测试加 `//go:build integration` 标签
- `go test -race` 默认开启

## 设计原则

1. **极简采集** — 3 步内从灵感到记录
2. **渐进呈现** — 高级功能不干扰基础流程
3. **直觉交互** — 拖拽、右键、快捷键、点选
4. **即时反馈** — 同步状态、AI 进度、操作结果
5. **暗色模式** — 默认支持
