package router

import (
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/nextm/nextm/internal/api/handler"
	"github.com/nextm/nextm/internal/api/middleware"
	"github.com/nextm/nextm/internal/config"
	"github.com/nextm/nextm/internal/pkg/crypto"
	db "github.com/nextm/nextm/internal/repository/db/sqlite"
	authService "github.com/nextm/nextm/internal/service/auth"
	objectService "github.com/nextm/nextm/internal/service/object"
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

	// ─── 初始化 Service 层 ─────────────────────────────
	jwtManager := crypto.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// 使用 SQLite 作为主仓库
	authRepo := db.NewAuthRepository(sqliteDB)
	objectRepo := db.NewObjectRepository(sqliteDB)
	blockRepo := db.NewBlockRepository(sqliteDB)

	// Auth 模块
	authSvc := authService.NewService(authRepo, jwtManager, authService.Config{
		BcryptCost: cfg.Auth.BcryptCost,
	})
	authHandler := handler.NewAuthHandler(authSvc)

	// Object 模块
	objectSvc := objectService.NewService(objectRepo, blockRepo)
	objectHandler := handler.NewObjectHandler(objectSvc)

	// ─── API v1 ────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// 公开路由（无需认证）
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)

		// 需认证路由
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtManager))

			// 认证
			r.Get("/auth/accounts", authHandler.GetAccounts)
			r.Post("/auth/switch", authHandler.SwitchAccount)
			r.Delete("/auth/accounts/{id}", authHandler.Logout)

			// 对象
			r.Get("/objects", objectHandler.List)
			r.Post("/objects", objectHandler.Create)
			r.Get("/objects/search", objectHandler.Search)
			r.Get("/objects/{id}", objectHandler.Get)
			r.Put("/objects/{id}", objectHandler.Update)
			r.Delete("/objects/{id}", objectHandler.Delete)
			r.Patch("/objects/{id}/archive", objectHandler.Archive)

			// 块
			r.Get("/objects/{id}/blocks", objectHandler.ListBlocks)
			r.Post("/objects/{id}/blocks", objectHandler.CreateBlock)
			r.Put("/blocks/{id}", objectHandler.UpdateBlock)
			r.Delete("/blocks/{id}", objectHandler.DeleteBlock)
		})
	})

	return r
}
