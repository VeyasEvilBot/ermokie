package db

import "database/sql"

type Run struct {
	ID       int
	Name     string
	Game     sql.NullString
	Category sql.NullString
	Attempts int
}
