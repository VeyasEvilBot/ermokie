package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/veya/ermokie/pkg/config"
	dbsqlc "codeberg.org/veya/ermokie/pkg/db/sqlc"
)

func TestInitUsesConfiguredDataDirectory(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv(config.EnvDataDir, dataDir)
	config.SetDataDir("")
	t.Cleanup(func() { config.SetDataDir("") })

	store, err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })

	if _, err := os.Stat(filepath.Join(dataDir, "app.db")); err != nil {
		t.Fatalf("configured database was not created: %v", err)
	}
}

func TestAdvanceAndBackOnlyMutateActiveRun(t *testing.T) {
	config.SetDataDir(t.TempDir())
	t.Cleanup(func() { config.SetDataDir("") })
	store, err := Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	ctx := context.Background()

	insert := func(name string) int64 {
		id, insertErr := store.InsertRun(ctx, dbsqlc.InsertRunParams{
			Name:        name,
			Attempts:    sql.NullInt64{Int64: 0, Valid: true},
			ActiveSplit: sql.NullInt64{Int64: 0, Valid: true},
		})
		if insertErr != nil {
			t.Fatal(insertErr)
		}
		return id
	}
	activeID := insert("active")
	otherID := insert("other")
	insertSplitErr := store.InsertSplit(ctx, dbsqlc.InsertSplitParams{RunID: activeID, Name: "second", Idx: 1})
	if insertSplitErr != nil {
		t.Fatal(insertSplitErr)
	}
	updateErr := store.UpdateActiveRunByID(ctx, sql.NullInt64{Int64: activeID, Valid: true})
	if updateErr != nil {
		t.Fatal(updateErr)
	}
	if advanceErr := store.AdvanceSplitInActiveRun(ctx); advanceErr != nil {
		t.Fatal(advanceErr)
	}

	active, err := store.GetRunByID(ctx, activeID)
	if err != nil {
		t.Fatal(err)
	}
	other, err := store.GetRunByID(ctx, otherID)
	if err != nil {
		t.Fatal(err)
	}
	if active.ActiveSplit.Int64 != 1 || other.ActiveSplit.Int64 != 0 {
		t.Fatalf("active split values = (%d, %d), want (1, 0)", active.ActiveSplit.Int64, other.ActiveSplit.Int64)
	}

	if backErr := store.GoBackSplitInActiveRun(ctx); backErr != nil {
		t.Fatal(backErr)
	}
	if backErr := store.GoBackSplitInActiveRun(ctx); backErr != nil {
		t.Fatal(backErr)
	}
	active, err = store.GetRunByID(ctx, activeID)
	if err != nil {
		t.Fatal(err)
	}
	if active.ActiveSplit.Int64 != 0 {
		t.Fatalf("back moved active split below zero: %d", active.ActiveSplit.Int64)
	}
}
