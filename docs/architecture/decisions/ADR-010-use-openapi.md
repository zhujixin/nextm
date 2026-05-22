# ADR-010: 使用 OpenAPI 3.0 作为 API 文档标准

**日期**: 2026-05-22

**状态**: 已接受

---

## 背景

NextM 需要对外提供标准化的 API 文档，支持第三方开发者和 AI 工具接入。

## 选项

| 方案 | 优点 | 缺点 |
|------|------|------|
| **OpenAPI 3.0** | 行业标准、工具链成熟、代码生成 | 冗长、需要维护 schema |
| **手动文档** | 灵活、无工具约束 | 易过时、不一致、无自动校验 |
| **gRPC Protobuf 直接暴露** | 单一真实来源 | 外部客户端不友好 |

## 决策

使用 **OpenAPI 3.0** 定义和管理 API 规范。

## 理由

1. **行业标准** — Kubernetes、GitHub、Stripe 等均使用 OpenAPI
2. **工具链成熟** — 支持代码生成（oapi-codegen）、Mock 服务、自动化测试
3. **客户端 SDK 生成** — 可为多种语言生成客户端
4. **可读性** — 人类可读的 YAML/JSON，非技术背景也能理解
5. **MCP 兼容** — OpenAPI schema 可直接映射到 MCP 工具定义

## 影响

- API 规范文件独立版本管理（与代码仓库一起）
- 使用 Swagger UI / Stoplight Elements 渲染文档
- Go 服务端可通过 `oapi-codegen` 生成 handler 接口和 model 类型
- API 变更需同步更新 OpenAPI 文件（Code Review 强制要求）
