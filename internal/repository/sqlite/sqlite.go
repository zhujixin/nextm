package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/nextm/nextm/internal/config"
)

func Open(cfg config.SQLiteConfig) (*sql.DB, error) {
	dsn := cfg.Path
	if cfg.WAL {
		dsn += "?_journal_mode=WAL"
	}
	if cfg.ForeignKeys {
		dsn += "&_foreign_keys=on"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite WAL 模式下写并发受限
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	// 启用 WAL 模式和 foreign keys
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000", // 64MB
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("pragma %s: %w", p, err)
		}
	}

	return db, nil
}

// WithTx 在事务中执行 fn
func WithTx(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Retry 用于处理 SQLite 锁冲突的重试
func Retry(fn func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if isLockError(err) {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			continue
		}
		return err
	}
	return err
}

func isLockError(err error) bool {
	// SQLite 锁错误代码 5 (SQLITE_BUSY) 和 6 (SQLITE_LOCKED)
	errStr := err.Error()
	return contains(errStr, "database is locked") ||
		contains(errStr, "database table is locked") ||
		contains(errStr, "BUSY") ||
		contains(errStr, "SQLITE_BUSY")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
