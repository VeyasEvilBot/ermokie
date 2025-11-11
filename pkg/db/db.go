package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/stefanistkuhl/ermokie/pkg/config"
	"github.com/stefanistkuhl/ermokie/pkg/db/sqlc"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var Schema string

type Store struct {
	*sqlc.Queries
	DB *sql.DB
}

func openDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func Init() (*Store, error) {
	dataDir := config.GetDataDir()
	dbPath := filepath.Join(dataDir, "app.db")

	_, statErr := os.Stat(dbPath)
	if os.IsNotExist(statErr) {
		db, err := openDB(dbPath)
		if err != nil {
			return nil, err
		}
		tx, err := db.Begin()
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		if _, err := tx.Exec(Schema); err != nil {
			_ = tx.Rollback()
			_ = db.Close()
			return nil, fmt.Errorf("apply schema: %w", err)
		}
		if err := tx.Commit(); err != nil {
			_ = db.Close()
			return nil, err
		}
		return &Store{
			Queries: sqlc.New(db),
			DB:      db,
		}, nil
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
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

	return &Store{
		Queries: sqlc.New(db),
		DB:      db,
	}, nil
}

func CheckEmpty(db *sql.DB) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}
