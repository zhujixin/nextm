# ADR-004: 使用 SQLite 作为本地数据库

**日期**: 2026-05-22

**状态**: 已接受

---

## 背景

NextM 需要本地优先的嵌入式数据库，用于离线存储、全文搜索和本地 CRDT 变更日志。

## 选项

| 方案 | 优点 | 缺点 |
|------|------|------|
| **SQLite** | 零配置、关系查询、FTS5 全文搜索、生态成熟、BLOB 支持 | 写入并发受限、无原生向量支持 |
| **BoltDB** | 纯 Go、零依赖、读写性能好 | 无 SQL、无索引、不适用复杂查询 |
| **Badger** | 高性能 KV 存储、事务支持 | 无关系查询、生态不如 SQLite |

## 决策

选择 **SQLite**（WAL 模式），驱动使用 `modernc.org/sqlite`（纯 Go）为主，`go-sqlite3`（CGo）为备选。

## 理由

1. **关系查询能力** — 支持 JOIN、WHERE 过滤、聚合，与云端 PostgreSQL 共享 SQL schema
2. **FTS5 全文搜索** — 内建全文索引，无需额外搜索引擎即可满足本地搜索需求
3. **SQLc 兼容** — SQLite dialect 与 PostgreSQL dialect 可通过 SQLc 双目标编译
4. **零配置** — 嵌入式无服务端，应用启动即打开
5. **WAL 模式** — 读写不互斥，适合本地优先的并发场景

## 影响

- 使用 WAL 模式提升并发性能
- 纯 Go 驱动 `modernc.org/sqlite` 避免 CGo 交叉编译问题
- CGo 驱动 `go-sqlite3` 作为性能备选（特定平台优化）
- 与 PostgreSQL 的 SQL 差异需在 SQLc 层面处理（`*.sqlite.sql` vs `*.postgres.sql`）
