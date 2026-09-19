package tracker

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	dbsqlc "codeberg.org/veya/ermokie/pkg/db/sqlc"
)

func testStore(t *testing.T) (store *db.Store, runID int64) {
	t.Helper()
	config.SetDataDir(t.TempDir())
	t.Cleanup(func() { config.SetDataDir("") })
	var err error
	store, err = db.Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	ctx := context.Background()
	runID, err = store.InsertRun(ctx, dbsqlc.InsertRunParams{
		Name:        "route",
		Attempts:    sql.NullInt64{Int64: 2, Valid: true},
		ActiveSplit: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	for idx, name := range []string{"one", "two"} {
		insertErr := store.InsertSplit(ctx, dbsqlc.InsertSplitParams{
			RunID: runID, Name: name, Idx: int64(idx),
			HitCount: sql.NullInt64{Int64: int64(idx), Valid: true},
		})
		if insertErr != nil {
			t.Fatal(insertErr)
		}
	}
	if err := store.UpdateActiveRunByID(ctx, sql.NullInt64{Int64: runID, Valid: true}); err != nil {
		t.Fatal(err)
	}
	return store, runID
}

func TestServiceTracksHitsAndBoundsProgress(t *testing.T) {
	store, runID := testStore(t)
	service := New(store)
	ctx := context.Background()

	if err := service.AddHit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.RemoveHit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.RemoveHit(ctx); err != nil {
		t.Fatal(err)
	}
	first, err := store.GetSplitByRunIDAndIdx(ctx, dbsqlc.GetSplitByRunIDAndIdxParams{RunID: runID, Idx: 0})
	if err != nil {
		t.Fatal(err)
	}
	if first.HitCount.Int64 != 0 {
		t.Fatalf("first split hits = %d, want 0", first.HitCount.Int64)
	}

	for range 3 {
		advanceErr := service.Advance(ctx)
		if advanceErr != nil {
			t.Fatal(advanceErr)
		}
	}
	run, err := store.GetRunByID(ctx, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.ActiveSplit.Int64 != 1 {
		t.Fatalf("active split = %d, want bounded value 1", run.ActiveSplit.Int64)
	}
}

func TestServiceResetIsAtomicUserAction(t *testing.T) {
	store, runID := testStore(t)
	service := New(store)
	ctx := context.Background()
	if err := service.Advance(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.AddHit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.Reset(ctx); err != nil {
		t.Fatal(err)
	}

	run, err := store.GetRunByID(ctx, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.ActiveSplit.Int64 != 0 || run.Attempts.Int64 != 3 {
		t.Fatalf("run after reset = split %d attempts %d", run.ActiveSplit.Int64, run.Attempts.Int64)
	}
	splits, err := store.GetSplitsByRunID(ctx, runID)
	if err != nil {
		t.Fatal(err)
	}
	for _, split := range splits {
		if split.HitCount.Int64 != 0 {
			t.Fatalf("split %q retained %d hits", split.Name, split.HitCount.Int64)
		}
	}
}

func TestServiceRequiresActiveRun(t *testing.T) {
	config.SetDataDir(t.TempDir())
	t.Cleanup(func() { config.SetDataDir("") })
	store, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	if err := New(store).AddHit(context.Background()); !errors.Is(err, ErrNoActiveRun) {
		t.Fatalf("AddHit error = %v", err)
	}
}
