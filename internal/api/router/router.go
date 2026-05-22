package router

import (
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/nextm/nextm/internal/api/handler"
	"github.com/nextm/nextm/internal/api/middleware"
	"github.com/nextm/nextm/internal/config"
)

func New(cfg *config.Config, log *slog.Logger, sqliteDB *sql.DB, postgresDB *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	// 全局中间件
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins, cfg.CORS.AllowedHeaders))
	r.Use(chimw.Timeout(cfg.Server.GracefulTimeout))

	// 健康检查（无需认证）
	healthHandler := handler.NewHealthHandler(sqliteDB, postgresDB)

	r.Get("/health", healthHandler.Health)
	r.Get("/health/ready", healthHandler.Ready)
	r.Get("/health/live", healthHandler.Live)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// 认证模块
		// TODO: authHandler.RegisterRoutes(r)

		// 知识对象
		// TODO: objectHandler.RegisterRoutes(r)

		// 内容块
		// TODO: blockHandler.RegisterRoutes(r)

		// 对象类型
		// TODO: typeHandler.RegisterRoutes(r)

		// 集合/视图
		// TODO: collectionHandler.RegisterRoutes(r)

		// 搜索
		// TODO: searchHandler.RegisterRoutes(r)

		// 视觉
		// TODO: visionHandler.RegisterRoutes(r)

		// 采集
		// TODO: captureHandler.RegisterRoutes(r)

		// 智能体
		// TODO: agentHandler.RegisterRoutes(r)

		// 关系/图谱
		// TODO: relationHandler.RegisterRoutes(r)

		// MCP
		// TODO: mcpHandler.RegisterRoutes(r)

		// 同步
		// TODO: syncHandler.RegisterRoutes(r)

		// 导出
		// TODO: exportHandler.RegisterRoutes(r)
	})

	return r
}
