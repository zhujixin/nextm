package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nextm/nextm/internal/api/router"
	"github.com/nextm/nextm/internal/config"
	"github.com/nextm/nextm/internal/pkg/logger"
	"github.com/nextm/nextm/internal/repository/postgres"
	"github.com/nextm/nextm/internal/repository/sqlite"
)

func main() {
	// 加载配置
	cfg, err := config.Load(".")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 初始化日志
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	log.Info("starting nextm server", "version", "0.1.0")

	// 连接本地 SQLite
	sqliteDB, err := sqlite.Open(cfg.Database.SQLite)
	if err != nil {
		log.Error("failed to open sqlite", "error", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()
	log.Info("sqlite connected", "path", cfg.Database.SQLite.Path)

	// 可选连接 PostgreSQL
	var postgresDB *sql.DB
	if cfg.Database.Postgres.DSN != "" {
		pgDB, err := postgres.Open(cfg.Database.Postgres)
		if err != nil {
			log.Warn("postgres not available, running in local-only mode", "error", err)
		} else {
			postgresDB = pgDB
			defer pgDB.Close()
			log.Info("postgres connected")
		}
	}

	// 创建路由
	r := router.New(cfg, log, sqliteDB, postgresDB)

	// HTTP 服务器
	srv := &http.Server{
		Addr:         cfg.Server.ListenAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("server listening", "addr", cfg.Server.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}
