package updating

import (
	"database/sql"

	"github.com/stefanistkuhl/ermokie/pkg/db"
)

func UpdateActiveRun(data *sql.DB, runName string) (int64, error) {
	return db.UpdateRows(data,
		`UPDATE misc_info
		 SET active_run = r.id
		 FROM runs r
		 WHERE r.name = ?`,
		runName)
}

func UpdateActiveRunByID(data *sql.DB, runID int) (int64, error) {
	return db.UpdateRows(data,
		`UPDATE misc_info
		 SET active_run = ?`,
		runID)
}

func UpdateActiveSplitByID(data *sql.DB, runID int, splitID int) (int64, error) {
	return db.UpdateRows(data,
		`UPDATE runs
		 SET active_split = (
		   SELECT idx FROM splits WHERE id = ?
		 )
		 WHERE id = ?`,
		splitID, runID)
}
