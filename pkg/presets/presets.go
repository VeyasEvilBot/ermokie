package presets

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stefanistkuhl/ermokie/pkg/db"
	"github.com/stefanistkuhl/ermokie/pkg/db/sqlc"
	"github.com/stefanistkuhl/ermokie/pkg/types"
)

type Preset struct {
	Game     string
	Category string
	Name     string
	Splits   []string
}

func NewPresets() []Preset {
	p := []Preset{
		// Elden Ring prests
		{Game: "Elden Ring", Category: "Any%", Name: "Iron Balls", Splits: []string{"Torrenting", "DTS", "Patches", "Bernahl/Manor", "Margit", "Godrick", "Radahn", "Goldfrey", "Morgott", "Fire Giant", "Godskin Dou", "Maliketh", "Gideon", "Horah Loux", "Radabeast"}},
		{Game: "Elden Ring", Category: "All Great Runes", Name: "Star Fist", Splits: []string{"Setup", "Noble", "Rykard", "Margit", "Godrick", "DTS", "Goldfrey", "Guardian Golem", "Morgott", "Fire Giant", "Godskin Dou", "Maliketh", "Gideon", "Horah Loux", "Niall", "Loretta", "Melanie", "Mohg", "Red Wolf", "Rennala", "Radabeast"}},
		// Trash souls 1 presets
		{Game: "Dark Souls 1", Category: "Any%", Name: "Crystal Halberd", Splits: []string{"Asylum Demon", "Gargoyles", "Quelaag", "Ceaseless", "Iron Golem", "Ornstein & Smough", "Pinwheel", "Sif", "Seath", "Nito", "Bed of Chaos", "Four Kings", "Gwyn"}},
		{Game: "Dark Souls 1", Category: "All Main Game Bosses", Name: "Crystal Halberd/Blacksmith Giant Hammer", Splits: []string{"Asylum Demon", "Gargoyles", "Quelaag", "Ceaseless", "Iron Golem", "Ornstein & Smough", "Pinwheel", "Stray Demon", "Butterfly", "Sif", "Taurus Demon", "Capra Demon", "Gaping Dragon", "Seath", "Demon Firesage", "Centipede Demon", "Bed of Chaos", "Four Kings", "Nito", "Gwyndolin", "Priscilla", "Gwyn"}},
		{Game: "Dark Souls 1", Category: "All Bosses", Name: "Hybrid", Splits: []string{"Asylum Demon", "Gargoyles", "Quelaag", "Iron Golem", "Ornstein & Smough", "Pinwheel", "Stray Demon", "Seath", "Sanctuary Guardian", "Artorias", "Manus", "Gwyndolin", "Priscilla", "Ceaseless", "Demon Firesage", "Centipede Demon", "Bed of Chaos", "Sif", "Butterfly", "Taurus Demon", "Capra Demon", "Gaping Dragon", "Four Kings", "Nito", "Kalameet", "Gwyn"}},
	}
	return p
}

func GetPresetGameNames() []string {
	var games []string
	presets := NewPresets()
	for _, preset := range presets {
		games = append(games, preset.Game)
	}
	return deduplicate(games)
}
func GetPresetRunNamesByGame(gameName string) []string {
	var runs []string
	presets := NewPresets()
	for _, preset := range presets {
		if preset.Game == gameName {
			runs = append(runs, fmt.Sprintf("%s %s %s", preset.Game, preset.Category, preset.Name))
		}
	}

	return runs
}

func deduplicate(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func GetPresetByName(runName string) Preset {
	presets := NewPresets()
	for _, preset := range presets {
		if fmt.Sprintf("%s %s %s", preset.Game, preset.Category, preset.Name) == runName {
			return preset
		}
	}
	var p Preset
	return p
}

func PresetToRunType(toImport Preset, presets []Preset) (types.RunCreate, bool) {
	var run types.RunCreate
	for _, preset := range presets {
		if preset.Game == toImport.Game &&
			preset.Category == toImport.Category &&
			preset.Name == toImport.Name {
			run.Name = preset.Game + " " + preset.Category + " " + preset.Name
			run.Game = sql.NullString{String: preset.Game, Valid: true}
			run.Category = sql.NullString{String: preset.Category, Valid: true}
			run.Attempts = 0
			run.ActiveSplit = 0
			for i, split := range preset.Splits {
				var splitCreate types.SplitCreate
				splitCreate.Name = split
				splitCreate.Hits = 0
				splitCreate.PBHits = 0
				splitCreate.Idx = i
				splitCreate.SaveFile = sql.NullString{String: "", Valid: false}
				run.Splits = append(run.Splits, splitCreate)

			}
			return run, true
		}
	}
	return run, false
}

func ImportPresets(s *db.Store, runs []types.RunCreate) error {
	tx, err := s.DB.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, run := range runs {
		insRun := sqlc.InsertRunParams{
			Name:        run.Name,
			Attempts:    sql.NullInt64{Int64: int64(run.Attempts), Valid: true},
			ActiveSplit: sql.NullInt64{Int64: int64(run.ActiveSplit), Valid: true},
			Game:        run.Game,
			Category:    run.Category,
		}
		runID, err := s.Queries.InsertRun(context.Background(), insRun)
		if err != nil {
			return err
		}

		for idx, split := range run.Splits {
			splitIns := sqlc.InsertSplitParams{
				RunID:      runID,
				Name:       split.Name,
				HitCount:   sql.NullInt64{Int64: int64(split.Hits), Valid: true},
				PbHitCount: sql.NullInt64{Int64: int64(split.PBHits), Valid: true},
				Idx:        int64(idx + 1),
				SaveFile:   split.SaveFile,
			}
			err := s.Queries.InsertSplit(context.Background(), splitIns)
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
