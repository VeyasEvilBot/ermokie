package fetching

import (
	"database/sql"

	"github.com/stefanistkuhl/ermokie/pkg/db"
)

func GetAllRunNames(data *sql.DB) ([]RunName, error) {
	return db.QueryRows(data,
		"SELECT id, name FROM runs",
		func(rows *sql.Rows) (RunName, error) {
			var r RunName
			err := rows.Scan(&r.ID, &r.Name)
			return r, err
		},
	)
}

func GetAllRuns(data *sql.DB) ([]Run, error) {
	return db.QueryRows(data,
		"SELECT id, name, game, category, attempts, active_split FROM runs",
		func(rows *sql.Rows) (Run, error) {
			var r Run
			err := rows.Scan(&r.ID, &r.Name, &r.Game, &r.Category, &r.Attempts, &r.ActiveSplit)
			return r, err
		},
	)
}

func GetSplitsByRunName(data *sql.DB, runName string) ([]Split, error) {
	return db.QueryRows(data,
		`SELECT
		  s.id,
		  s.run_id,
		  s.name,
		  s.hit_count,
		  s.pb_hit_count,
		  s.idx,
		  s.save_file,
		  (s.idx = r.active_split) AS is_active
		FROM
		  splits AS s
		JOIN
		  runs AS r
		  ON r.id = s.run_id
		WHERE
		  r.name = ?
		ORDER BY
		  s.idx;
		`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Idx, &s.SaveFile, &s.IsActive)
			return s, err
		},
		runName,
	)
}

func GetSplitsByRunID(data *sql.DB, runID int) ([]Split, error) {
	return db.QueryRows(data,
		`SELECT
		  s.id,
		  s.run_id,
		  s.name,
		  s.hit_count,
		  s.pb_hit_count,
		  s.idx,
		  s.save_file,
		  (s.idx = r.active_split) AS is_active
		FROM
		  splits AS s
		JOIN
		  runs AS r
		  ON r.id = s.run_id
		WHERE
		  s.run_id = ?
		ORDER BY
		  s.idx;
		`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Idx, &s.SaveFile, &s.IsActive)
			return s, err
		},
		runID,
	)
}

func GetRunByName(data *sql.DB, runName string) (*Run, error) {
	results, err := db.QueryRows(data,
		"SELECT id, name, game, category, attempts, active_split FROM runs WHERE name = ?",
		func(rows *sql.Rows) (Run, error) {
			var r Run
			err := rows.Scan(&r.ID, &r.Name, &r.Game, &r.Category, &r.Attempts, &r.ActiveSplit)
			return r, err
		},
		runName,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}

func GetRunByID(data *sql.DB, runID int) (*Run, error) {
	results, err := db.QueryRows(data,
		"SELECT id, name, game, category, attempts, active_split FROM runs WHERE id = ?",
		func(rows *sql.Rows) (Run, error) {
			var r Run
			err := rows.Scan(&r.ID, &r.Name, &r.Game, &r.Category, &r.Attempts, &r.ActiveSplit)
			return r, err
		},
		runID,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}

func GetSplitByRunIDAndIdx(data *sql.DB, runID int, splitIdx int) (*Split, error) {
	results, err := db.QueryRows(data,
		`SELECT id, run_id, name, hit_count, pb_hit_count, idx, save_file 
		 FROM splits 
		 WHERE run_id = ? AND idx = ?`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Idx, &s.SaveFile)
			return s, err
		},
		runID, splitIdx,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}

func GetSplitByRunNameAndIdx(data *sql.DB, runName string, splitIdx int) (*Split, error) {
	results, err := db.QueryRows(data,
		`SELECT s.id, s.run_id, s.name, s.hit_count, s.pb_hit_count, s.idx, s.save_file 
		 FROM splits s 
		 JOIN runs r ON s.run_id = r.id 
		 WHERE r.name = ? AND s.idx = ?`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Idx, &s.SaveFile)
			return s, err
		},
		runName, splitIdx,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}
