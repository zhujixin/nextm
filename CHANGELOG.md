# Changelog

## [Unreleased]

### Added
- 项目初始化，创建设计文档体系
  - PRD (v1.2) — 产品需求说明书
  - PDD (v1.0) — 产品设计说明书
  - ADD (v1.0) — 架构设计说明书
  - API (v1.0) — OpenAPI 接口规范
  - DB (v1.0) — 数据库设计说明书
- 开发基础设施文档
  - 开发环境搭建指南
  - 编码规范（Go / TypeScript / Dart）
  - Git 工作流规范（Trunk-Based + Conventional Commits）
  - 测试策略文档
- 安全设计说明书
- 前端架构设计说明书
- 架构决策记录 (ADR-001 至 ADR-010)

### Changed
- 交叉验证并修复文档间不一致
  - PRD v1.3 — 数据库视图新增"列表"类型
  - PDD v1.1 — 补充 ViewTypeList 枚举
  - API v1.1 — 新增 Export 导出模块路由
  - 新增 ADR-004 至 ADR-010（SQLite、NATS、Asynq、CRDT Yrs CGo、LanceDB、REST+gRPC、OpenAPI）
