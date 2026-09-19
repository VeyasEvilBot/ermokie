package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"path/filepath"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db/sqlc"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var Schema string

type Store struct {
	*sqlc.Queries
	DB *sql.DB
}

func openDB(ctx context.Context, dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}
	if _, err := database.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	return database, nil
}

func Init() (*Store, error) {
	ctx := context.Background()
	dataDir := config.GetDataDir()
	if dataDir == "" {
		var err error
		dataDir, err = config.GetDataDirWithFallback("ermokie")
		if err != nil {
			return nil, fmt.Errorf("resolve data directory: %w", err)
		}
		config.SetDataDir(dataDir)
	}

	database, err := openDB(ctx, filepath.Join(dataDir, "app.db"))
	if err != nil {
		return nil, err
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("begin schema transaction: %w", err)
	}
	if _, err := tx.ExecContext(ctx, Schema); err != nil {
		_ = tx.Rollback()
		_ = database.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := tx.Commit(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("commit schema: %w", err)
	}

	return &Store{Queries: sqlc.New(database), DB: database}, nil
}

// CheckEmpty fails closed: an unreadable database is treated as non-empty so
// import commands never overwrite or append data after an inspection error.
func CheckEmpty(database *sql.DB) bool {
	var count int
	if err := database.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
		return false
	}
	return count == 0
}
