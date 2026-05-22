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
	collectionService "github.com/nextm/nextm/internal/service/collection"
	objectService "github.com/nextm/nextm/internal/service/object"
	relationService "github.com/nextm/nextm/internal/service/relation"
	tagService "github.com/nextm/nextm/internal/service/tag"
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
	tagRepo := db.NewTagRepository(sqliteDB)
	relRepo := db.NewRelationRepository(sqliteDB)
	colRepo := db.NewCollectionRepository(sqliteDB)

	// Auth 模块
	authSvc := authService.NewService(authRepo, jwtManager, authService.Config{
		BcryptCost: cfg.Auth.BcryptCost,
	})
	authHandler := handler.NewAuthHandler(authSvc)

	// Object 模块
	objectSvc := objectService.NewService(objectRepo, blockRepo)
	objectHandler := handler.NewObjectHandler(objectSvc)

	// Tag 模块
	tagSvc := tagService.NewService(tagRepo)
	tagHandler := handler.NewTagHandler(tagSvc)

	// Relation 模块
	relSvc := relationService.NewService(relRepo)
	relHandler := handler.NewRelationHandler(relSvc)

	// Collection 模块
	colSvc := collectionService.NewService(colRepo)
	colHandler := handler.NewCollectionHandler(colSvc)

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

			// 标签
			r.Get("/tags", tagHandler.List)
			r.Post("/tags", tagHandler.Create)
			r.Get("/tags/{id}", tagHandler.Get)
			r.Put("/tags/{id}", tagHandler.Update)
			r.Delete("/tags/{id}", tagHandler.Delete)
			r.Get("/objects/{id}/tags", tagHandler.GetObjectTags)
			r.Post("/objects/{id}/tags", tagHandler.AssignTags)
			r.Delete("/objects/{id}/tags/{tagId}", tagHandler.UnassignTag)

			// 关系
			r.Get("/objects/{id}/relations", relHandler.ListByObject)
			r.Post("/relations", relHandler.Create)
			r.Put("/relations/{id}", relHandler.Update)
			r.Delete("/relations/{id}", relHandler.Delete)

			// 集合
			r.Get("/collections", colHandler.List)
			r.Post("/collections", colHandler.Create)
			r.Get("/collections/{id}", colHandler.Get)
			r.Put("/collections/{id}", colHandler.Update)
			r.Delete("/collections/{id}", colHandler.Delete)
			r.Get("/collections/{id}/views", colHandler.ListViews)
			r.Post("/collections/{id}/views", colHandler.CreateView)
			r.Put("/collections/views/{id}", colHandler.UpdateView)
			r.Delete("/collections/views/{id}", colHandler.DeleteView)
			r.Get("/collections/{id}/items", colHandler.ListItems)
			r.Post("/collections/{id}/items", colHandler.AddItem)
			r.Delete("/collections/items/{itemId}", colHandler.RemoveItem)
		})
	})

	return r
}
