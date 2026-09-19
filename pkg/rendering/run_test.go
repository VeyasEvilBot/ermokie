package rendering

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	dbsqlc "codeberg.org/veya/ermokie/pkg/db/sqlc"
)

func TestRenderRunWritesConfiguredOverlay(t *testing.T) {
	dataDir := t.TempDir()
	config.SetDataDir(dataDir)
	t.Cleanup(func() { config.SetDataDir("") })
	store, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	ctx := context.Background()
	runID, err := store.InsertRun(ctx, dbsqlc.InsertRunParams{
		Name:        "Rendered route",
		Attempts:    sql.NullInt64{Int64: 4, Valid: true},
		ActiveSplit: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	insertSplitErr := store.InsertSplit(ctx, dbsqlc.InsertSplitParams{RunID: runID, Name: "Boss", Idx: 0})
	if insertSplitErr != nil {
		t.Fatal(insertSplitErr)
	}

	cfg := config.Config{
		General: config.GeneralSettings{DataDir: dataDir},
		Overlay: config.OverlayConfig{
			Theme:            config.OverlayThemeConfig{Name: "default"},
			TemplateCategory: "base",
			TemplateName:     "base",
		},
	}
	renderErr := RenderRun(ctx, store, &cfg, runID)
	if renderErr != nil {
		t.Fatal(renderErr)
	}
	htmlBytes, err := os.ReadFile(filepath.Join(dataDir, "render", "output.html")) // #nosec G304 -- temp test fixture.
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(htmlBytes), "Rendered route") || !strings.Contains(string(htmlBytes), "Boss") {
		t.Fatalf("generated overlay missing run data: %s", htmlBytes)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "render", "output.css")); err != nil {
		t.Fatal(err)
	}
}
