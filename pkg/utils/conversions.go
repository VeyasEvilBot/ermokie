package utils

import (
	"database/sql"

	"codeberg.org/veya/ermokie/pkg/db/sqlc"
	"codeberg.org/veya/ermokie/pkg/types"
)

func ToRunType(run sqlc.GetRunByIDRow) types.Run {
	var game sql.NullString
	var category sql.NullString
	if run.Game.Valid {
		game = sql.NullString{String: run.Game.String, Valid: true}
	} else {
		game = sql.NullString{String: "", Valid: false}
	}
	if run.Category.Valid {
		category = sql.NullString{String: run.Category.String, Valid: true}
	} else {
		category = sql.NullString{String: "", Valid: false}
	}
	return types.Run{
		ID:          int(run.ID),
		Name:        run.Name,
		Game:        game,
		Category:    category,
		Attempts:    int(run.Attempts.Int64),
		ActiveSplit: int(run.ActiveSplit.Int64),
	}
}

func ToRunTypeFromName(run sqlc.GetRunByNameRow) types.Run {
	var game sql.NullString
	var category sql.NullString
	if run.Game.Valid {
		game = sql.NullString{String: run.Game.String, Valid: true}
	} else {
		game = sql.NullString{String: "", Valid: false}
	}
	if run.Category.Valid {
		category = sql.NullString{String: run.Category.String, Valid: true}
	} else {
		category = sql.NullString{String: "", Valid: false}
	}
	return types.Run{
		ID:          int(run.ID),
		Name:        run.Name,
		Game:        game,
		Category:    category,
		Attempts:    int(run.Attempts.Int64),
		ActiveSplit: int(run.ActiveSplit.Int64),
	}
}

func ToRunFromActiveRun(run sqlc.GetActiveRunRow) types.Run {
	var game sql.NullString
	var category sql.NullString
	if run.Game.Valid {
		game = sql.NullString{String: run.Game.String, Valid: true}
	} else {
		game = sql.NullString{String: "", Valid: false}
	}
	if run.Category.Valid {
		category = sql.NullString{String: run.Category.String, Valid: true}
	} else {
		category = sql.NullString{String: "", Valid: false}
	}
	return types.Run{
		ID:          int(run.ID),
		Name:        run.Name,
		Game:        game,
		Category:    category,
		Attempts:    int(run.Attempts.Int64),
		ActiveSplit: int(run.ActiveSplit.Int64),
	}
}

func ToSplitType(split sqlc.GetSplitsByRunIDRow) types.Split {
	var saveFile sql.NullString
	if split.SaveFile.Valid {
		saveFile = sql.NullString{String: split.SaveFile.String, Valid: true}
	} else {
		saveFile = sql.NullString{String: "", Valid: false}
	}
	isActive := false
	if val, ok := split.IsActive.(int64); ok {
		isActive = val == 1
	}
	return types.Split{
		ID:       int(split.Idx),
		RunID:    int(split.RunID),
		Name:     split.Name,
		Hits:     int(split.HitCount.Int64),
		PBHits:   int(split.PbHitCount.Int64),
		Diff:     int(split.Diff.Int64),
		Idx:      int(split.Idx),
		SaveFile: saveFile,
		IsActive: isActive,
	}
}
