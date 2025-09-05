package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stefanistkuhl/ermokie/pkg/config"
)

//go:embed schemas/schema.sql
var Schema string

func Init() *sql.DB {
	dataDir := config.GetDataDir()
	dbPath := filepath.Join(dataDir, "app.db")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("DB connection failed:", err)
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal("begin tx:", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatal("enable foreign_keys:", err)
	}

	if _, err := tx.ExecContext(ctx, Schema); err != nil {
		log.Fatalf("apply schema failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("commit schema:", err)
	}

	return db
}

func CheckEmpty(db *sql.DB) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

func QueryRows[T any](
	db *sql.DB,
	query string,
	scanFn func(*sql.Rows) (T, error),
	args ...any,
) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}

func UpdateRows(
	db *sql.DB,
	query string,
	args ...any,
) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
