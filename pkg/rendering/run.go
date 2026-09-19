package rendering

import (
	"context"
	"fmt"
	"path/filepath"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	overlaythemes "codeberg.org/veya/ermokie/pkg/rendering/overlayThemes"
	"codeberg.org/veya/ermokie/pkg/types"
	"codeberg.org/veya/ermokie/pkg/utils"
)

// RenderRun writes the configured overlay for a persisted run. Both the CLI
// and the TUI use this path so templates behave identically in live and manual
// rendering flows.
func RenderRun(ctx context.Context, store *db.Store, cfg *config.Config, runID int64) error {
	runRaw, err := store.GetRunByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("get run: %w", err)
	}
	splitsRaw, err := store.GetSplitsByRunID(ctx, runID)
	if err != nil {
		return fmt.Errorf("get run splits: %w", err)
	}
	splits := make([]types.Split, 0, len(splitsRaw))
	for i := range splitsRaw {
		splits = append(splits, utils.ToSplitType(splitsRaw[i]))
	}

	themes := overlaythemes.NewOverlayThemes()
	theme := themes.GetThemeByName(cfg.Overlay.Theme.Name)
	if theme.Name == "" {
		return fmt.Errorf("unknown overlay theme %q", cfg.Overlay.Theme.Name)
	}
	params := RenderingParams{
		Run:         utils.ToRunType(runRaw),
		Splits:      splits,
		Tc:          TemplateCategory(cfg.Overlay.TemplateCategory),
		Tn:          TemplateName(cfg.Overlay.TemplateName),
		CssFileName: "output.css",
		Theme:       theme,
		TemplateDir: filepath.Join(cfg.General.DataDir, "templates"),
	}
	params.CounterSettings.DisplayNext.IsLimited = cfg.Overlay.LimitSplitsAbove
	params.CounterSettings.DisplayNext.Count = cfg.Overlay.ShowSplitsAbove
	params.CounterSettings.DisplayPrev.IsLimited = cfg.Overlay.LimitSplitsBellow
	params.CounterSettings.DisplayPrev.Count = cfg.Overlay.ShowSplitsBellow

	htmlBytes, cssBytes, err := RenderHTML(params)
	if err != nil {
		return err
	}
	return WriteOutput(filepath.Join(cfg.General.DataDir, "render"), htmlBytes, cssBytes)
}
