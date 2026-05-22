# ADR-001: 使用 Chi Router 作为 HTTP 路由框架

**日期**: 2026-05-22

**状态**: 已接受

---

## 背景

NextM 后端需要 HTTP 路由框架来处理 REST API 请求。Go 标准库的 `net/http` 在 Go 1.22+ 中已支持路径参数，但对于中间件链、路由分组等功能仍需额外实现。

## 选项

| 方案 | 优点 | 缺点 |
|------|------|------|
| **Chi router** | stdlib-compatible、轻量、中间件链式组合、社区活跃 | 额外的依赖 |
| **Gin** | 性能高、广泛使用、内置验证器 | 非 stdlib 接口、自定义 Context、框架较重 |
| **标准库 net/http (Go 1.22+)** | 零依赖、Go 官方支持 | 缺少中间件编排、路由分组等 DX 提升 |
| **HttpRouter** | 极高性能 | 不支持中间件链（需自己封装）、功能过于精简 |

## 决策

选择 **Chi router** (`github.com/go-chi/chi/v5`)。

## 理由

1. **stdlib-compatible** — Chi 的 Handler 签名与 `net/http` 完全兼容，不会锁定生态
2. **中间件链** — Chi 的中间件模型简洁，按路由分组应用中间件（如 `/api/v1` 组用 auth，`/health` 不用）
3. **轻量无魔法** — 没有自定义 Context、没有反射，代码可读性强
4. **社区成熟** — 广泛用于生产环境（包括 Kubernetes 相关项目）
5. **Go 1.22+ 兼容** — 即使标准库增加了模式匹配，Chi 的 DX（中间件、路由分组）仍有价值

## 影响

- 新增依赖 `github.com/go-chi/chi/v5`
- 所有 handler 签名保持 `func(w http.ResponseWriter, r *http.Request)`
- 中间件编写采用 `func(next http.Handler) http.Handler` 模式
- 不会引入 Gin 式的自定义 Context，避免 handler 测试时需要 mock 框架 Context

## 相关

- 无 Gin 式自定义 Context，意味着 handler 测试可以直接使用 `httptest.NewRecorder()` + `httptest.NewRequest()`
