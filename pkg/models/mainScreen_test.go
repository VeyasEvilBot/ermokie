package models

import (
	"context"
	"database/sql"
	"testing"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	dbsqlc "codeberg.org/veya/ermokie/pkg/db/sqlc"
	"codeberg.org/veya/ermokie/pkg/ipc"
	"codeberg.org/veya/ermokie/pkg/types"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadInitDataRepresentsNoActiveRunAsNil(t *testing.T) {
	config.SetDataDir(t.TempDir())
	t.Cleanup(func() { config.SetDataDir("") })
	store, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })

	msg, ok := loadInitData(store)().(initDataMsg)
	if !ok {
		t.Fatalf("message type = %T", loadInitData(store)())
	}
	if msg.err != nil {
		t.Fatalf("loadInitData error = %v", msg.err)
	}
	if msg.activeRun != nil {
		t.Fatalf("active run = %#v, want nil", msg.activeRun)
	}
}

func TestIPCProgressUpdateWithoutActiveRunDoesNotPanic(t *testing.T) {
	m := model{}
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Update panicked: %v", recovered)
		}
	}()
	updated, _ := m.Update(IPCUpdate{Event: &ipc.AdvanceSplitResult{ActiveSplitIdx: 2}})
	updatedModel, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model type = %T", updated)
	}
	if updatedModel.activeRun != nil {
		t.Fatal("IPC update fabricated an active run")
	}
}

func TestNavigationDoesNotChangeRunProgress(t *testing.T) {
	config.SetDataDir(t.TempDir())
	t.Cleanup(func() { config.SetDataDir("") })
	store, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	ctx := context.Background()
	runID, err := store.InsertRun(ctx, dbsqlc.InsertRunParams{
		Name:        "route",
		Attempts:    sql.NullInt64{Int64: 0, Valid: true},
		ActiveSplit: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	for idx, name := range []string{"one", "two"} {
		insertErr := store.InsertSplit(ctx, dbsqlc.InsertSplitParams{RunID: runID, Name: name, Idx: int64(idx)})
		if insertErr != nil {
			t.Fatal(insertErr)
		}
	}
	updateErr := store.UpdateActiveRunByID(ctx, sql.NullInt64{Int64: runID, Valid: true})
	if updateErr != nil {
		t.Fatal(updateErr)
	}
	LoadKeymapFromConfig("basic")
	m := model{
		db:        store,
		activeRun: &types.Run{ID: int(runID), ActiveSplit: 0},
		runID:     int(runID),
		rows: []rowData{
			{Name: "one", Type: "split", ID: 1},
			{Name: "two", Type: "split", ID: 2},
		},
		vp: viewport.New(80, 20),
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model type = %T", updated)
	}
	if updatedModel.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", updatedModel.cursor)
	}
	run, err := store.GetRunByID(ctx, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.ActiveSplit.Int64 != 0 {
		t.Fatalf("navigation changed active split to %d", run.ActiveSplit.Int64)
	}
}
