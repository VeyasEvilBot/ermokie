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
		  s.diff,
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
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Diff, &s.Idx, &s.SaveFile, &s.IsActive)
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
		  s.diff,
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
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Diff, &s.Idx, &s.SaveFile, &s.IsActive)
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
		`SELECT id, run_id, name, hit_count, pb_hit_count, diff, idx, save_file 
		 FROM splits 
		 WHERE run_id = ? AND idx = ?`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Diff, &s.Idx, &s.SaveFile)
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
		`SELECT s.id, s.run_id, s.name, s.hit_count, s.pb_hit_count, s.diff, s.idx, s.save_file
		 FROM splits s
		 JOIN runs r ON s.run_id = r.id
		 WHERE r.name = ? AND s.idx = ?`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Diff, &s.Idx, &s.SaveFile)
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

func GetActiveRun(data *sql.DB) (*Run, error) {
	results, err := db.QueryRows(data,
		`SELECT id, name, game, category, attempts, active_split
		FROM runs r JOIN misc_info m ON m.active_run = r.id
		`,
		func(row *sql.Rows) (Run, error) {
			var r Run
			err := row.Scan(&r.ID, &r.Name, &r.Game, &r.Category, &r.Attempts, &r.ActiveSplit)
			return r, err
		},
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}

func GetRunAttemptsByName(data *sql.DB, runName string) (int, error) {
	results, err := db.QueryRows(data,
		"SELECT attempts FROM runs WHERE name = ?",
		func(rows *sql.Rows) (Run, error) {
			var r Run
			err := rows.Scan(&r.Attempts)
			return r, err
		},
		runName,
	)
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, sql.ErrNoRows
	}
	return results[0].Attempts, nil
}

func GetRunIDByName(data *sql.DB, runName string) (int, error) {
	results, err := db.QueryRows(data,
		"SELECT id FROM runs WHERE name = ?",
		func(rows *sql.Rows) (Run, error) {
			var r Run
			err := rows.Scan(&r.ID)
			return r, err
		},
		runName,
	)
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, sql.ErrNoRows
	}
	return results[0].ID, nil
}

func GetSplitByID(data *sql.DB, splitID int) (*Split, error) {
	results, err := db.QueryRows(data,
		`SELECT id, run_id, name, hit_count, pb_hit_count, diff, idx, save_file 
		 FROM splits 
		 WHERE id = ?`,
		func(rows *sql.Rows) (Split, error) {
			var s Split
			err := rows.Scan(&s.ID, &s.RunID, &s.Name, &s.Hits, &s.PBHits, &s.Diff, &s.Idx, &s.SaveFile)
			return s, err
		},
		splitID,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}

type SplitTotals struct {
	TotalHits int
	TotalPB   int
	TotalDiff int
}

func GetSplitTotalsByRunID(data *sql.DB, runID int) (*SplitTotals, error) {
	results, err := db.QueryRows(data,
		`SELECT 
			SUM(hit_count) as total_hits,
			SUM(pb_hit_count) as total_pb,
			SUM(diff) as total_diff
		FROM splits 
		WHERE run_id = ?`,
		func(rows *sql.Rows) (SplitTotals, error) {
			var t SplitTotals
			err := rows.Scan(&t.TotalHits, &t.TotalPB, &t.TotalDiff)
			return t, err
		},
		runID,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &SplitTotals{0, 0, 0}, nil
	}
	return &results[0], nil
}
