package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nextm/nextm/internal/config"
)

func Open(cfg config.PostgresConfig) (*sql.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("postgres dsn is empty")
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
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

const defaultTimeout = 30 * time.Second

func PingWithTimeout(db *sql.DB) error {
	done := make(chan error, 1)
	go func() {
		done <- db.Ping()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(defaultTimeout):
		return fmt.Errorf("postgres ping timeout after %s", defaultTimeout)
	}
}
