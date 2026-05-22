# ADR-009: 使用 REST + gRPC 作为 API 协议

**日期**: 2026-05-22

**状态**: 已接受

---

## 背景

NextM 需要确定外部客户端和服务间通信使用的 API 协议。

## 选项

| 方案 | 优点 | 缺点 |
|------|------|------|
| **REST + gRPC** | REST 对外标准化、gRPC 内部高效 | 维护两套协议增加了复杂度 |
| **GraphQL** | 灵活查询、减少 Over-fetching | 服务端复杂度高、缓存困难、Go 生态支持弱 |
| **tRPC** | 类型安全端到端 | 仅 TypeScript 生态、Go 端无成熟支持 |

## 决策

外部 API 使用 **REST (JSON)**，内部服务间通信使用 **gRPC**。

## 理由

1. **REST 标准化** — 外部客户端（Web、移动端、第三方开发者）REST 是最通用的选择
2. **gRPC 内部高效** — Protobuf 序列化、流式传输、强类型接口定义
3. **HTTP/2 兼容** — REST 和 gRPC 可共享同一端口（通过 gorilla/mux 或代理层）
4. **双向流** — gRPC Stream 适合同步通知、智能体消息推送等场景
5. **Protocol Buffers** — 接口即文档，生成多语言客户端

## 影响

- REST 使用 `net/http` + Chi router
- gRPC 使用 `google.golang.org/grpc`
- 外部客户端仅通过 REST 通信（WebSocket 除外）
- 内部 `cmd/sync` 和 `cmd/worker` 通过 gRPC 与 `cmd/server` 通信
- 部署时可能需要 gRPC-Web 代理层（envoy/grpc-gateway）
