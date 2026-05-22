# ADR-005: 使用 NATS 作为消息总线

**日期**: 2026-05-22

**状态**: 已接受

---

## 背景

NextM 需要消息总线支持多智能体间的通信，包括发布/订阅、请求/回复、持久化队列等模式。

## 选项

| 方案 | 优点 | 缺点 |
|------|------|------|
| **NATS + JetStream** | 轻量、高性能、持久化、通配符订阅 | 功能较 RabbitMQ 少 |
| **RabbitMQ** | 功能丰富、AMQP 协议标准 | 较重、Erlang 运行时、运维复杂 |
| **Redis Stream** | 轻量、无额外依赖 | 持久化有限、消费组功能不如专业消息队列 |

## 决策

选择 **NATS**，使用 **JetStream** 提供持久化能力。

## 理由

1. **轻量** — 单二进制部署，内存占用低（~10MB）
2. **JetStream 持久化** — 消息持久化到磁盘，支持 At-Least-Once 投递
3. **通配符订阅** — `agents.>` 匹配所有智能体消息，天然支持智能体主题路由
4. **请求-回复模式** — 内置 `INBOX` 机制，适合智能体间的 RPC 调用
5. **Go 原生客户端** — `github.com/nats-io/nats.go`，与 Go 后端深度集成

## 影响

- 部署依赖中新增 NATS 服务
- 智能体间通信统一走 NATS 主题（`agents.<agent_id>.<action>`）
- JetStream 的 Consumer 配置可能需要调优
- 本地开发通过 Docker Compose 启动
