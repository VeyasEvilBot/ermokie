package db

import (
	"database/sql"
	_ "embed"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schemas/schema.sql
var Schema string

func Init() *sql.DB {
	db, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("DB connection failed:", err)
	}

	for stmt := range strings.SplitSeq(Schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatalf("Error executing statement %q: %s", stmt, err)
		}
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

func GetAllRuns(db *sql.DB) ([]Run, error) {
	rows, err := db.Query("SELECT id, name FROM runs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}

	return runs, rows.Err()
}
