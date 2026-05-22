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

	// 执行数据库迁移
	if err := runMigrations(sqliteDB, "internal/repository/db/migrations/000001_initial.up.sql"); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations applied")

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

// runMigrations 执行 SQL 迁移文件
func runMigrations(db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 按 `;` 分割 SQL 语句，逐条执行
	statements := splitSQL(string(data))
	for _, stmt := range statements {
		if len(stmt) < 5 {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// splitSQL 将 SQL 文本按 `;` 分割为独立语句，保留注释和 DDL
func splitSQL(text string) []string {
	var stmts []string
	var current []byte
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '-' && i+1 < len(text) && text[i+1] == '-' {
			// 跳过单行注释
			for i < len(text) && text[i] != '\n' {
				i++
			}
			continue
		}
		if c == ';' {
			stmts = append(stmts, trimSQL(string(current)))
			current = current[:0]
			continue
		}
		current = append(current, c)
	}
	if len(current) > 0 {
		stmts = append(stmts, trimSQL(string(current)))
	}
	return stmts
}

func trimSQL(s string) string {
	// 去除前导和尾随空白
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\r' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\r' || s[end-1] == '\t') {
		end--
	}
	if start >= end {
		return ""
	}
	return s[start:end]
}
