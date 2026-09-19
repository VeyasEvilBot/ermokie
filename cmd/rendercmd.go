package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/rendering"
	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render an overlay",
}

var renderRunCmd = &cobra.Command{
	Use:   "run [id/name]",
	Short: "Render a run as an HTML overlay",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		store, err := db.Init()
		if err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
		defer func() { _ = store.DB.Close() }()

		ctx := context.Background()
		runID, parseErr := strconv.ParseInt(args[0], 10, 64)
		if parseErr != nil {
			run, getErr := store.GetRunByName(ctx, args[0])
			if getErr != nil {
				return fmt.Errorf("get run %q: %w", args[0], getErr)
			}
			runID = run.ID
		}

		if cmd.Flags().Changed("template-cat") {
			cfg.Overlay.TemplateCategory, _ = cmd.Flags().GetString("template-cat")
		}
		if cmd.Flags().Changed("template-name") {
			cfg.Overlay.TemplateName, _ = cmd.Flags().GetString("template-name")
		}
		renderErr := rendering.RenderRun(ctx, store, &cfg, runID)
		if renderErr != nil {
			return fmt.Errorf("render overlay: %w", renderErr)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Rendered %s\n", filepath.Join(cfg.General.DataDir, "render", "output.html"))
		return err
	},
}

func init() {
	renderRunCmd.Flags().StringP("template-cat", "c", "base", "Template category to use")
	renderRunCmd.Flags().StringP("template-name", "n", "base", "Template name to use")
	renderCmd.AddCommand(renderRunCmd)
	rootCmd.AddCommand(renderCmd)
}
