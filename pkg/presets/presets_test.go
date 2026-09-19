package presets

import (
	"context"
	"database/sql"
	"testing"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/types"
)

func TestImportPresetsRollsBackWholeBatch(t *testing.T) {
	config.SetDataDir(t.TempDir())
	t.Cleanup(func() { config.SetDataDir("") })
	store, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })

	runs := []types.RunCreate{
		{
			Name:     "valid",
			Game:     sql.NullString{String: "Elden Ring", Valid: true},
			Category: sql.NullString{String: "Any%", Valid: true},
			Splits:   []types.SplitCreate{{Name: "boss"}},
		},
		{
			Name:     "invalid",
			Game:     sql.NullString{String: "Elden Ring", Valid: true},
			Category: sql.NullString{String: "not allowed by legacy schema", Valid: true},
		},
	}
	if importErr := ImportPresets(store, runs); importErr == nil {
		t.Fatal("ImportPresets() accepted invalid batch")
	}
	all, err := store.GetAllRuns(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Fatalf("partial import survived rollback: %#v", all)
	}
}
