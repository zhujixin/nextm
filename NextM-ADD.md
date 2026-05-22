# NextM — 架构设计说明书

**版本**: V1.0
**状态**: 初稿
**最后更新**: 2026-05-22

---

## 1. 文档概述

### 1.1 目的

本文档描述 NextM 系统的软件架构设计，包括 Go 后端的模块划分、组件依赖关系、关键交互流程、部署拓扑以及横切关注点的设计决策。本文档面向开发团队，作为编码实现的技术蓝图。

### 1.2 范围

本文档覆盖 NextM 后端服务的架构设计，包括：
- Go 模块结构与包依赖规则
- Handler → Service → Repository 分层架构
- 多智能体编排引擎设计
- CRDT 同步引擎设计
- 本地优先 + 云端同步的部署架构
- 安全、可观测、配置管理等横切关注点

前端客户端（React Web、Tauri Desktop、Flutter Mobile、浏览器插件）的架构不在本文档范围内。

### 1.3 参考文档

| 文档 | 说明 |
|------|------|
| NextM-PRD.md | 产品需求说明书 v1.2 |
| NextM-PDD.md | 产品设计说明书 v1.0 |
| NextM-DB.md | 数据库设计说明书 v1.0 |

---

## 2. 架构原则与约束

### 2.1 架构原则

| 原则 | 说明 |
|------|------|
| **Clean Architecture** | 依赖方向由外向内：Handler → Service → Repository，内部层不依赖外部层 |
| **接口驱动** | 所有 Service 间依赖通过 Go 接口表达，实现位于 `internal/service` |
| **SQL 优先** | 数据访问使用 SQLc 生成类型安全代码，不使用 ORM |
| **本地优先** | 核心功能离线可用，云端为同步和协作增强 |
| **优雅降级** | 外部依赖不可用时降级而非崩溃，降级路径有定义 |
| **可观测性内置** | 每个请求/任务携带 Trace Context，日志结构化输出 |

### 2.2 约束

| 约束 | 说明 |
|------|------|
| **语言** | 后端主语言 Go 1.24+，FFI 调用 CRDT(CGo) 和 AI(Python) |
| **无环形依赖** | `go tool mod` 不允许模块循环依赖，包间通过接口解耦 |
| **数据库双轨** | 本地 SQLite + 云端 PostgreSQL，SQLc 生成双方言查询 |
| **部署形态** | 单二进制部署（可水平扩展），不支持热插拔插件 |

---

## 3. Go 模块架构

### 3.1 模块依赖图

```
nextm (root module: github.com/nextm/nextm)
│
├── cmd/                 # main 入口，依赖所有 internal 包
│   ├── server           → internal/api, internal/service/...
│   ├── sync             → internal/service/sync, internal/crdt
│   ├── worker           → internal/eventbus, internal/service/agent
│   ├── migrator         → internal/repository/db
│   └── cli              → internal/config, internal/pkg/...
│
├── internal/
│   ├── api              → internal/service/*, internal/config, internal/pkg/logger
│   │   ├── handler/     → internal/service/* (按模块分文件)
│   │   ├── middleware/  → internal/pkg/logger, internal/config
│   │   ├── router/      → internal/api/handler/*
│   │   └── dto/         → 无依赖
│   │
│   ├── service/         → internal/repository/*, internal/eventbus, internal/crdt
│   │   ├── object/      → internal/repository/db
│   │   ├── search/      → internal/repository/db, internal/vector
│   │   ├── vision/      → internal/repository/db (image_queue)
│   │   ├── agent/       → internal/eventbus, internal/repository/db
│   │   │   ├── orchestrator/ → internal/eventbus
│   │   │   └── messagebus/   → internal/eventbus
│   │   ├── sync/        → internal/crdt, internal/repository/db
│   │   ├── mcp/         → internal/service/object, search
│   │   └── auth/        → internal/repository/db, internal/pkg/crypto
│   │
│   ├── eventbus/        → 无依赖 (NATS 客户端封装)
│   ├── crdt/            → 无依赖 (Go-Yrs CGo FFI)
│   ├── vector/          → internal/repository/db
│   │
│   ├── repository/      → internal/model, internal/config
│   │   ├── db/           → SQLc 生成代码
│   │   └── cache/       → go-redis
│   │
│   ├── model/           → 无依赖 (纯 struct 定义)
│   ├── config/          → 无依赖 (Viper)
│   ├── telemetry/       → internal/config
│   │
│   └── pkg/             → 无依赖 (工具包)
│       ├── logger/
│       ├── idgen/
│       ├── crypto/
│       └── httputil/
```

### 3.2 依赖规则

```
handler → service → repository → db (SQLc)
            ↘
          eventbus (NATS)
          crdt (Go-Yrs)
          vector (LanceDB)
```

- **Handler 层**：只负责 HTTP/gRPC 协议转换，调用 Service 接口，不包含业务逻辑
- **Service 层**：纯业务逻辑，通过接口依赖 Repository 和其他 Service
- **Repository 层**：SQLc 生成的类型安全查询代码，仅负责数据存取
- **禁止**：Handler 直接调用 Repository；Service 之间直接实例化具体实现
- **依赖注入**：所有依赖在 `main.go` 中通过构造函数注入（wire 或手动）

---

## 4. 组件架构

### 4.1 分层职责

```
┌─────────────────────────────────────────────────────────────┐
│  Transport Layer (Chi Router)                               │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐ │
│  │  REST Handler │  │  gRPC Server │  │  WebSocket Hub    │ │
│  │  (JSON)       │  │  (internal)  │  │  (goroutine/conn) │ │
│  └──────┬───────┘  └──────┬───────┘  └────────┬──────────┘ │
├─────────┼─────────────────┼───────────────────┼─────────────┤
│         ▼                 ▼                   ▼              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Middleware Pipeline                       │   │
│  │  Recovery ─ Logger ─ CORS ─ Auth(JWT) ─ RateLimit ─  │   │
│  │  Trace ─ RequestID ─ Timeout ─ Compress               │   │
│  └──────────────────────┬───────────────────────────────┘   │
├─────────────────────────┼───────────────────────────────────┤
│                         ▼                                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Service Layer (Business Logic)            │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────┐  │   │
│  │  │ Object │ │ Search │ │ Vision │ │ Agent  │ │ ...│  │   │
│  │  │Service │ │Service │ │Service │ │Service │ │    │  │   │
│  │  └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘ └────┘  │   │
│  │      │           │          │          │               │   │
│  │  ┌───┴───────────┴──────────┴──────────┴────────────┐  │   │
│  │  │          Sync Service (跨模块编排)                │  │   │
│  │  │    CRDT Merge → Conflict Resolve → Broadcast     │  │   │
│  │  └──────────────────────┬───────────────────────────┘  │   │
│  └─────────────────────────┼──────────────────────────────┘   │
├───────────────────────────┼──────────────────────────────────┤
│                           ▼                                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Repository Layer (Data Access)           │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │   │
│  │  │  SQLite Repo │  │  PostgreSQL  │  │  Redis     │  │   │
│  │  │  (SQLc)      │  │  Repo (SQLc) │  │  Cache     │  │   │
│  │  └──────────────┘  └──────────────┘  └────────────┘  │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              External Integrations                     │   │
│  │  ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐  │   │
│  │  │  NATS    │ │  LanceDB │ │LiteLLM │ │ Meili-   │  │   │
│  │  │  Client  │ │  Go SDK  │ │HTTP/Cli│ │ search   │  │   │
│  │  └──────────┘ └──────────┘ └────────┘ └──────────┘  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 核心组件说明

#### 4.2.1 Transport Layer

| 组件 | 技术 | 职责 |
|------|------|------|
| REST Handler | Chi Router + net/http | 外部 API 入口，JSON 请求/响应 |
| gRPC Server | google.golang.org/grpc | 服务间内部通信，推送通知 |
| WebSocket Hub | gorilla/websocket / nhooyr.io/websocket | 实时同步连接管理 |
| Middleware | Chi 中间件链 | 认证、日志、限流、追踪 |

**WS Hub 设计要点**：
```go
// WebSocket Hub — goroutine-per-connection 模型
type WSHub struct {
    mu      sync.RWMutex
    conns   map[string]*WSConn   // connID → connection
    spaceConns map[string]map[string]*WSConn  // spaceID → {connID}
}

type WSConn struct {
    ID        string
    SpaceID   string
    UserID    string
    Conn      *websocket.Conn
    Send      chan []byte
    Done      chan struct{}
}
```

#### 4.2.2 Service Layer

每个核心模块：Object, Collection, Search, Vision, Agent, Relation, Sync, MCP, Auth, AI

**接口模式示例**：

```go
// ObjectService 接口定义 (位于 internal/service/object)
type ObjectService interface {
    Create(ctx context.Context, req *dto.CreateObjectReq) (*model.Object, error)
    Get(ctx context.Context, id string) (*model.Object, error)
    List(ctx context.Context, req *dto.ListObjectReq) (*dto.ListObjectResp, error)
    Update(ctx context.Context, req *dto.UpdateObjectReq) (*model.Object, error)
    Delete(ctx context.Context, id string) error
}

// objectServiceImpl 具体实现
type objectServiceImpl struct {
    repo   db.ObjectRepository    // SQLc 生成接口
    types  db.TypeRepository
    blocks db.BlockRepository
    search SearchIndexer         // 搜索索引回调
    event  EventPublisher        // 事件发布 (NATS)
}

func NewObjectService(repo db.ObjectRepository, types db.TypeRepository, blocks db.BlockRepository, search SearchIndexer, event EventPublisher) ObjectService {
    return &objectServiceImpl{repo: repo, types: types, blocks: blocks, search: search, event: event}
}
```

#### 4.2.3 Repository Layer

SQLc 生成模式：

```
internal/repository/db/
├── sqlc.yaml              # SQLc 配置 (SQLite + PG 双引擎)
├── migrations/            # golang-migrate SQL
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── queries/               # *.sql → SQLc 生成 *.sql.go
│   ├── objects.sql
│   ├── blocks.sql
│   ├── tags.sql
│   ├── agents.sql
│   └── ...
├── sqlite/                # SQLite 方言覆盖 (使用 sqlite dialect)
│   └── queries/
└── postgres/              # PG 方言覆盖
    └── queries/
```

SQLc 生成产物示例：
```go
// Code generated by sqlc. DO NOT EDIT.
// queries/objects.sql.go

type ObjectRepository interface {
    CreateObject(ctx context.Context, arg CreateObjectParams) (model.Object, error)
    GetObject(ctx context.Context, id string) (model.Object, error)
    ListObjects(ctx context.Context, arg ListObjectsParams) ([]model.Object, error)
    UpdateObject(ctx context.Context, arg UpdateObjectParams) (model.Object, error)
    DeleteObject(ctx context.Context, id string) error
}
```

### 4.3 依赖注入

使用构造函数手动注入（权衡后不引入 wire，保持显式可追踪）：

```go
// cmd/server/main.go
func main() {
    cfg := config.Load()
    logger := logger.New(cfg.LogLevel)
    db := openDB(cfg)
    
    // Repository
    objectRepo := db.NewObjectRepository(db)
    blockRepo  := db.NewBlockRepository(db)
    
    // Infrastructure
    eventBus   := eventbus.New(cfg.NATS)
    searchIdx  := search.NewIndexer(cfg.Meilisearch)
    
    // Service
    objectSvc := object.NewService(objectRepo, blockRepo, searchIdx, eventBus)
    authSvc   := auth.NewService(db, cfg.JWT)
    
    // Handler
    objectHandler := handler.NewObjectHandler(objectSvc, logger)
    authHandler   := handler.NewAuthHandler(authSvc, logger)
    
    // Router
    r := router.New(cfg, authHandler, objectHandler, ...)
    
    log.Fatal(http.ListenAndServe(cfg.ListenAddr, r))
}
```

---

## 5. 进程架构

### 5.1 进程模型

```
┌─────────────────────────────────────────────────────────┐
│  cmd/server             cmd/sync              cmd/worker  │
│  ┌─────────────────┐  ┌───────────────┐  ┌────────────┐ │
│  │ HTTP Server      │  │ WS Sync Hub   │  │ Asynq      │ │
│  │ (Chi Router)     │  │ (goroutine    │  │ Worker     │ │
│  │ :8080            │  │  per conn)    │  │ (并发消费)  │ │
│  │                  │  │ :8081          │  │            │ │
│  │ gRPC Server      │  │               │  │ OCR Worker │ │
│  │ :9090            │  │ Sync Logic    │  │ AI Worker  │ │
│  └────────┬─────────┘  └───────┬───────┘  │ Export     │ │
│           │                     │          │ Worker     │ │
│           └────────┬────────────┘          └────────────┘ │
│                    │                                        │
│              ┌─────▼──────┐                                │
│              │ PostgreSQL  │                                │
│              └────────────┘                                │
└─────────────────────────────────────────────────────────────┘
```

| 进程 | 入口 | 职责 | 扩缩容 |
|------|------|------|--------|
| `cmd/server` | `main.go` | REST + gRPC API, 同步编排 | 水平扩展 (stateless) |
| `cmd/sync` | `main.go` | WebSocket 长连接, CRDT 变更推送 | 按连接数 (stateful via NATS) |
| `cmd/worker` | `main.go` | 后台任务: AI 推理, OCR, 导出, 智能体 | 水平扩展 |

### 5.2 关键流程

#### 5.2.1 请求生命周期

```
Client                     API Server                        Repository              External
  │                          │                                  │                      │
  │  POST /api/v1/objects    │                                  │                      │
  │ ────────────────────────▶│                                  │                      │
  │                          │                                  │                      │
  │                     ┌────┴────┐                             │                      │
  │                     │ Middleware                             │                      │
  │                     │ ① Recovery                            │                      │
  │                     │ ② Logger (traceID)                   │                      │
  │                     │ ③ Auth (JWT parse)                   │                      │
  │                     │ ④ Rate Limit                         │                      │
  │                     │ ⑤ RequestID                          │                      │
  │                     └────┬────┘                             │                      │
  │                          │                                  │                      │
  │                     ┌────┴────┐                             │                      │
  │                     │ Handler                               │                      │
  │                     │ ① Bind JSON → dto                     │                      │
  │                     │ ② Validate dto                       │                      │
  │                     │ ③ Call Service                       │                      │
  │                     └────┬────┘                             │                      │
  │                          │                                  │                      │
  │                     ┌────┴────┐                             │                      │
  │                     │ Service                                │                      │
  │                     │ ① Repo: Insert object                │────▶ SQLite/PG ──────▶│
  │                     │ ② Repo: Insert blocks               │◀──── response ────────│
  │                     │ ③ Publish: object.created (NATS)    │────▶ NATS ────────────▶│
  │                     │ ④ Search: index object              │────▶ Meilisearch ─────▶│
  │                     └────┬────┘                             │                      │
  │                          │                                  │                      │
  │  201 Created + object    │                                  │                      │
  │ ◀────────────────────────│                                  │                      │
```

#### 5.2.2 实时同步流程

```
Device A                    cmd/sync                     cmd/sync                    Device B
  │                          │                             │                          │
  │  WS Connect              │                             │                          │
  │ (space_id, user_id)      │                             │                          │
  │ ────────────────────────▶│                             │                          │
  │                     ┌────┴────┐                        │                          │
  │                     │ Hub.Register                     │                          │
  │                     │ spaceConns[spaceA][connA] = conn  │                          │
  │                     └────┬────┘                        │                          │
  │                          │                             │                          │
  │  编辑对象 X (离线)       │                             │                          │
  │  ┌─────────────────┐     │                             │                          │
  │  │ SQLite:          │     │                             │                          │
  │  │ sync_log insert  │     │                             │                          │
  │  │ CRDT binary      │     │                             │                          │
  │  └────────┬────────┘     │                             │                          │
  │           │ 上线         │                             │                          │
  │  ┌────────┴────────┐     │                             │                          │
  │  │ FlushOfflineQueue│     │                             │                          │
  │  │ WS Send SyncMsg  │────▶│                             │                          │
  │  └─────────────────┘     │                             │                          │
  │                          │  Receive SyncMsg             │                          │
  │                          │  ① Write to PG              │                          │
  │                          │  ② Publish to NATS JetStream│                          │
  │                          │     (spaceA.objectX.delta)   │                          │
  │                          └────────┬────────────────────┘                          │
  │                                   │                                               │
  │                                   │  NATS Subscribe: spaceA.*.delta               │
  │                                   │                                               │
  │                          ┌────────┴────────────────────┐                          │
  │                          │ Hub.Broadcast(spaceA, msg)   │                          │
  │                          │ connB.Send <- msg            │────▶ Device B ──────────▶│
  │                          │                              │    ① CRDT Merge         │
  │                          │                              │    ② Update SQLite      │
  │                          └─────────────────────────────┘    ③ UI Update           │
```

#### 5.2.3 智能体编排流程

```
User Request                          Worker
  │                                     │
  │  POST /api/v1/agents/:id/trigger    │
  │ ───────────────────────────────────▶│
  │                                     │
  │                              ┌──────┴──────┐
  │                              │ Orchestrator  │
  │                              │  ① 解析 DAG   │
  │                              │  ② 分解任务    │
  │                              └──────┬──────┘
  │                                     │
  │              ┌──────────────────────┼──────────────────────┐
  │              │                      │                      │
  │         ┌────┴────┐          ┌─────┴─────┐         ┌─────┴─────┐
  │         │ 采集智能体 │          │ 提取智能体  │         │ 总结智能体  │
  │         │ (goroutine)│         │ (goroutine) │         │ (goroutine) │
  │         └────┬────┘          └─────┬─────┘         └─────┬─────┘
  │              │                      │                      │
  │              │  并行执行 (DAG)       │                      │
  │              ├──────────────────────┼──────────────────────┤
  │              │  1. 搜索/采集         │                      │
  │              │  2. OCR/提取          │                      │
  │              │  3. AI 总结           │                      │
  │              └──────────────────────┴──────────────────────┘
  │                                     │
  │                              ┌──────┴──────┐
  │                              │ Orchestrator  │
  │                              │  汇总结果      │
  │                              │  保存对象      │
  │                              │  发送通知      │
  │                              └──────┬──────┘
  │                                     │
  │  Response + notification           │
  │ ◀───────────────────────────────────│
```

---

## 6. 部署架构

### 6.1 本地优先架构

```
┌──────────────────────────────────────────────────────────────┐
│                    客户端设备                                  │
│                                                               │
│  ┌─────────────┐    ┌─────────────────────────────────────┐  │
│  │ 前端 UI     │    │  Go 后端进程 (嵌入式)                  │  │
│  │ (React/     │    │  ┌─────────┐ ┌───────────┐          │  │
│  │  Flutter)   │◀───│  │ 本地 API │ │ Sync      │          │  │
│  │             │    │  │ (net/http)  │ │ Engine    │          │  │
│  └─────────────┘    │  └─────────┘ └───────────┘          │  │
│                     │  ┌─────────┐ ┌───────────┐          │  │
│                     │  │ SQLite  │ │ LanceDB   │          │  │
│                     │  │ (WAL)   │ │ (vectors) │          │  │
│                     │  └─────────┘ └───────────┘          │  │
│                     └─────────────────────────────────────┘  │
│                                                               │
│  离线模式: 全部本地读写，变更记入 sync_log                      │
│  在线模式: 本地 + 云端双向同步 (CRDT)                          │
└──────────────────────────────────────────────────────────────┘
```

**嵌入式 Go 后端**: 桌面端 Go 二进制通过 Tauri sidecar 启动，移动端通过 gomobile 编译。

### 6.2 云端部署

```
                          ┌──────────────┐
                          │  Cloudflare   │
                          │  CDN + DNS    │
                          └──────┬───────┘
                                 │
                          ┌──────▼───────┐
                          │  LB / ALB    │
                          └──────┬───────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
       ┌──────▼──────┐   ┌──────▼──────┐   ┌──────▼──────┐
       │ cmd/server   │   │ cmd/server   │   │ cmd/server   │
       │ (:8080 REST) │   │ (:8080)      │   │ (:8080)      │
       │ (:9090 gRPC) │   │              │   │              │
       └──────┬───────┘   └──────┬───────┘   └──────┬───────┘
              │                  │                  │
              └──────────────────┼──────────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
       ┌──────▼──────┐   ┌──────▼──────┐   ┌──────▼──────┐
       │ PostgreSQL   │   │   Redis     │   │   NATS      │
       │ + pgvector   │   │   Cache     │   │   JetStream │
       └──────────────┘   └─────────────┘   └──────┬──────┘
                                                    │
              ┌─────────────────────────────────────┼──────────┐
              │                 │                            │
       ┌──────▼──────┐   ┌──────▼──────┐   ┌──────────────┐  │
       │ cmd/sync     │   │ cmd/worker  │   │  Meilisearch │  │
       │ (:8081 WS)   │   │ (Asynq)     │   │              │  │
       └──────┬───────┘   └──────┬───────┘   └──────────────┘  │
              │                  │                              │
       ┌──────▼──────┐   ┌──────▼──────┐                       │
       │  MinIO S3   │   │  LiteLLM    │                       │
       │  (附件)      │   │  (AI 网关)  │                       │
       └─────────────┘   └─────────────┘                       │
```

### 6.3 Go 二进制构建

```dockerfile
# 多阶段构建
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o /bin/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/server /app/server
EXPOSE 8080 9090
ENTRYPOINT ["/app/server"]
```

---

## 7. 数据架构

### 7.1 数据流拓扑

```
┌─────────────────────────────────────────────────────────────────┐
│                         写入路径                                 │
│                                                                  │
│  Handler → Service → Repository → DB (本地: SQLite / 云端: PG)  │
│                           │                                     │
│                           ▼                                     │
│                     EventBus (NATS)                              │
│                           │                                     │
│              ┌────────────┼────────────┐                        │
│              ▼            ▼            ▼                        │
│         Search Index   Sync Engine  Agent Workers               │
│         (Meilisearch)  (推送变更)     (触发响应)                │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│                         搜索路径                                 │
│                                                                  │
│  Handler → SearchService → HybridSearch                         │
│                ├── FulltextSearch (Meilisearch/FTS5)             │
│                ├── SemanticSearch (pgvector/LanceDB)            │
│                └── RRF Fusion → Ranked Results                   │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│                        同步路径                                  │
│                                                                  │
│  本地变更 → SyncLog (CRDT) → WebSocket → SyncService            │
│                                        → PG (权威副本)          │
│                                        → NATS (广播)            │
│                                        → 其他在线设备            │
└─────────────────────────────────────────────────────────────────┘
```

### 7.2 双数据库读写策略

| 场景 | 本地 (SQLite) | 云端 (PostgreSQL) |
|------|--------------|-------------------|
| 读取 (在线) | 可选 (本地优先) | 权威数据 |
| 读取 (离线) | 唯一来源 | — |
| 写入 (在线) | 即写 + sync_log | 同步后写入 |
| 写入 (离线) | 即写 + sync_log | 上线后推送 |
| 全文搜索 | FTS5 | Meilisearch |
| 向量搜索 | LanceDB | pgvector |

**抽象层设计**：

```go
// DBSelector 根据连接状态选择数据库
type DBSelector struct {
    online bool
}

func (s *DBSelector) ObjectRepo() ObjectRepository {
    if s.online {
        return pgRepo    // PostgreSQL
    }
    return sqliteRepo   // SQLite
}

// 但 CRUD 操作需双写以保证数据一致性（在线时）
// WriteStrategy: 在线时先写本地再推送；离线时写本地+日志
```

---

## 8. 安全架构

### 8.1 认证与授权

```
Client                     API Server
  │                          │
  │  POST /api/v1/auth/login │
  │  {email, password}       │
  │ ────────────────────────▶│
  │                          │  ① Verify credentials (bcrypt)
  │                          │  ② Generate Access Token (JWT, 15min)
  │                          │  ③ Generate Refresh Token (random, 7d)
  │                          │  ④ Store refresh_token hash in PG
  │  200 {access_token,      │
  │       refresh_token}     │
  │ ◀────────────────────────│
  │                          │
  │  GET /api/v1/objects     │
  │  Authorization: Bearer <access_token>
  │ ────────────────────────▶│
  │                          │  ① Auth Middleware: validate JWT
  │                          │  ② Extract user_id, space_id from claims
  │                          │  ③ Pass context with claims
  │                          │  ④ Handler checks space membership
  │  200 {...}               │
  │ ◀────────────────────────│
  │                          │
  │  POST /api/v1/auth/refresh
  │  {refresh_token}         │
  │ ────────────────────────▶│
  │                          │  ① Verify refresh_token hash
  │                          │  ② Rotate: invalidate old, issue new
  │                          │  ③ Generate new access_token
  │  200 {access_token,      │
  │       refresh_token}     │
  │ ◀────────────────────────│
```

### 8.2 多账号隔离

```go
// context key
type ctxKey string
const (
    CtxUserID  ctxKey = "user_id"
    CtxSpaceID ctxKey = "space_id"
)

// Space-level isolation — 所有 Repository 查询带 space_id
func (r *objectRepo) ListObjects(ctx context.Context, arg ListObjectsParams) ([]model.Object, error) {
    spaceID := ctx.Value(CtxSpaceID).(string)
    arg.SpaceID = spaceID
    return r.queries.ListObjects(ctx, arg)
}
```

---

## 9. 横切关注点

### 9.1 错误处理规范

```go
// 错误分级
type ErrorSeverity int

const (
    SevDebug    ErrorSeverity = iota // 调试信息
    SevInfo                          // 预期错误（验证失败等）
    SevWarn                          // 外部错误（依赖不可用）
    SevError                         // 内部错误（bug/panic）
)

// 统一错误响应格式
type APIError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details any       `json:"details,omitempty"`
    // 内部字段（不序列化）
    Severity ErrorSeverity
    Err      error
}

func (e *APIError) Error() string { return e.Message }
func (e *APIError) Unwrap() error { return e.Err }

// Handler 层统一错误处理
func (h *ObjectHandler) Create(w http.ResponseWriter, r *http.Request) {
    obj, err := h.svc.Create(r.Context(), req)
    if err != nil {
        h.writeError(w, err)  // 根据 type assertion 转 APIError
        return
    }
    writeJSON(w, http.StatusCreated, obj)
}
```

### 9.2 可观测性

```go
// OpenTelemetry 集成
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/nextm/nextm")

// Service 层自动追踪
func (s *objectServiceImpl) Create(ctx context.Context, req *dto.CreateObjectReq) (*model.Object, error) {
    ctx, span := tracer.Start(ctx, "ObjectService.Create")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("object.type", req.TypeID),
        attribute.Int("block.count", len(req.Blocks)),
    )
    
    // ...
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    return obj, nil
}
```

### 9.3 配置管理

```go
// config.go — Viper 配置结构
type Config struct {
    ListenAddr  string       `mapstructure:"listen_addr"`
    LogLevel    string       `mapstructure:"log_level"`

    JWT struct {
        Secret          string `mapstructure:"secret"`
        AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
        RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
    } `mapstructure:"jwt"`

    Database struct {
        SQLite struct {
            Path string `mapstructure:"path"`
        } `mapstructure:"sqlite"`
        Postgres struct {
            DSN string `mapstructure:"dsn"`
        } `mapstructure:"postgres"`
    } `mapstructure:"database"`

    NATS struct {
        URL     string `mapstructure:"url"`
        Cluster string `mapstructure:"cluster"`
    } `mapstructure:"nats"`

    Meilisearch struct {
        Host   string `mapstructure:"host"`
        APIKey string `mapstructure:"api_key"`
    } `mapstructure:"meilisearch"`

    LiteLLM struct {
        BaseURL string `mapstructure:"base_url"`
        APIKey  string `mapstructure:"api_key"`
    } `mapstructure:"litellm"`

    Redis struct {
        URL string `mapstructure:"url"`
    } `mapstructure:"redis"`

    Asynq struct {
        Concurrency int `mapstructure:"concurrency"`
    } `mapstructure:"asynq"`
}

// 加载顺序: 默认值 → 配置文件 → 环境变量
func Load() *Config {
    v := viper.New()
    v.SetConfigName("nextm")
    v.SetConfigPaths([]string{".", "$HOME/.nextm", "/etc/nextm/"})
    v.AutomaticEnv()
    v.SetEnvPrefix("NEXTM")
    _ = v.ReadInConfig()
    
    var cfg Config
    v.Unmarshal(&cfg)
    return &cfg
}
```

### 9.4 日志规范

```go
// 结构化日志 — 使用 slog (Go 1.21+)
import "log/slog"

// Middleware 注入 logger
func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            traceID := r.Header.Get("X-Trace-ID")
            if traceID == "" {
                traceID = uuid.New().String()
            }
            
            ctx := context.WithValue(r.Context(), "trace_id", traceID)
            r = r.WithContext(ctx)
            
            logger.Info("request started",
                "method", r.Method,
                "path", r.URL.Path,
                "trace_id", traceID,
            )
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 9.5 优雅关闭

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: cfg.ListenAddr, Handler: r}

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s", err)
        }
    }()

    <-ctx.Done()
    log.Println("shutting down gracefully...")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    srv.Shutdown(shutdownCtx)
    db.Close()
    eventBus.Close()
}
```

---

## 10. 关键设计决策记录 (ADR)

| ID | 决策 | 选项 | 选择 | 理由 |
|----|------|------|------|------|
| ADR-001 | Go Web 框架 | Chi / Gin / Echo / Fiber | **Chi** | stdlib net/http 兼容、中间件接口标准化、社区活跃 |
| ADR-002 | SQL 访问层 | SQLc / GORM / Ent / sqlx | **SQLc** | 类型安全、无运行时开销、SQL 即代码 |
| ADR-003 | 依赖注入 | Wire / 手动 / Uber Dig | **手动** | 显式可追踪、无代码生成、小团队可控 |
| ADR-004 | 本地数据库 | SQLite / BoltDB / Badger | **SQLite** | 关系型查询能力、FTS5 全文搜索、生态成熟 |
| ADR-005 | 消息队列 | NATS / RabbitMQ / Redis Stream | **NATS** | 轻量、JetStream 持久化、智能体通信模式匹配 |
| ADR-006 | 任务队列 | Asynq / Machinery / 自建 | **Asynq** | Redis 后端无额外依赖、延迟任务、Go 原生 |
| ADR-007 | CRDT 实现 | Yrs (CGo) / 自建 Go CRDT / Yjs (WASM) | **Yrs CGo** | 经过生产验证、Rust 核心性能好、CGo 封装可行 |
| ADR-008 | 向量搜索 | LanceDB / Qdrant / Milvus | **LanceDB** | 嵌入式无服务端、Go SDK 原生、适合本地优先 |
| ADR-009 | API 协议 | REST / GraphQL / tRPC | **REST + gRPC** | REST 对外标准化、gRPC 内部高效 |
| ADR-010 | API 文档 | Swagger/OpenAPI / 手动 | **OpenAPI 3.0** | 行业标准、代码生成工具链成熟 |

---

## 11. 演进路线

| 阶段 | 架构状态 | 关键里程碑 |
|------|---------|-----------|
| **MVP** (第 1-3 月) | 单体单进程 `cmd/server` | SQLite-only, REST API, 基础 CRUD, 纯文本搜索 |
| **Phase 2** (第 4-6 月) | 进程分离 `server` + `sync` | PostgreSQL 支持, WebSocket 同步, 多账号 |
| **Phase 3** (第 7-9 月) | 引入 `worker` 进程 | Asynq 后台任务, AI 服务集成, 知识图谱 |
| **Phase 4** (第 10-12 月) | 多进程 + NATS 消息总线 | 智能体 DAG 编排, MCP 协议, 团队协作 |
