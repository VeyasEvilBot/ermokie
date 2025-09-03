package fetching

import "database/sql"

type Run struct {
	ID          int
	Name        string
	Game        sql.NullString
	Category    sql.NullString
	Attempts    int
	ActiveSplit int
}

type RunName struct {
	ID   int
	Name string
}

type Split struct {
	ID       int
	RunID    int
	Name     string
	Hits     int
	PBHits   int
	Idx      int
	SaveFile sql.NullString
	IsActive bool
}

type RunCreate struct {
	Name        string
	Game        sql.NullString
	Category    sql.NullString
	Attempts    int
	ActiveSplit int
	Splits      []SplitCreate
}

type SplitCreate struct {
	Name     string
	Hits     int
	PBHits   int
	Idx      int
	SaveFile sql.NullString
	IsActive bool
}
