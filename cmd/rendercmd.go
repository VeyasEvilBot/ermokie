package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/rendering"
	overlaythemes "codeberg.org/veya/ermokie/pkg/rendering/overlayThemes"
	"codeberg.org/veya/ermokie/pkg/types"
	"codeberg.org/veya/ermokie/pkg/utils"
	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render a run as a HTML page",
	Long:  `Render a run as a HTML page.`,
}

var renderRunCmd = &cobra.Command{
	Use:   "run [id/name]",
	Short: "Render a run as a HTML page",
	Long:  `Render a run as a HTML page.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateCatFlag, _ := cmd.Flags().GetString("template-cat")
		templateNameFlag, _ := cmd.Flags().GetString("template-name")
		runArg := args[0]
		store, getDbErr := db.Init()
		ctx := context.Background()
		if getDbErr != nil {
			fmt.Printf("Failed to initialize the database %s", getDbErr)
			os.Exit(1)
		}

		var run types.Run
		var splits []types.Split

		id, parseErr := strconv.Atoi(runArg)
		if parseErr == nil {
			runRaw, getRunErr := store.GetRunByID(ctx, int64(id))
			if getRunErr != nil {
				fmt.Printf("Error getting run info: %v\n", getRunErr)
				os.Exit(1)
			}
			run = utils.ToRunType(runRaw)
			splitsRaw, getSplitsErr := store.GetSplitsByRunID(ctx, int64(id))
			if getSplitsErr != nil {
				fmt.Printf("Error getting splits: %v\n", getSplitsErr)
				os.Exit(1)
			}
			for _, split := range splitsRaw {
				splits = append(splits, utils.ToSplitType(split))
			}
		} else {
			runRaw, getRunErr := store.GetRunByName(ctx, runArg)
			if getRunErr != nil {
				fmt.Printf("Error getting run info: %v\n", getRunErr)
				os.Exit(1)
			}
			run = utils.ToRunTypeFromName(runRaw)
			splitsRaw, getSplitsErr := store.GetSplitsByRunID(ctx, runRaw.ID)
			if getSplitsErr != nil {
				fmt.Printf("Error getting splits: %v\n", getSplitsErr)
				os.Exit(1)
			}
			for _, split := range splitsRaw {
				splits = append(splits, utils.ToSplitType(split))
			}
		}

		cfg, getCfgErr := config.LoadConfig()
		if getCfgErr != nil {
			fmt.Printf("Error loading config: %v\n", getCfgErr)
			os.Exit(1)
		}
		if !cmd.Flags().Changed("template-cat") {
			templateCatFlag = cfg.Overlay.TemplateCategory
		}
		if !cmd.Flags().Changed("template-name") {
			templateNameFlag = cfg.Overlay.TemplateName
		}
		templateCat := rendering.TemplateCategory(templateCatFlag)
		templateName := rendering.TemplateName(templateNameFlag)

		overlaythemes := overlaythemes.NewOverlayThemes()
		theme := overlaythemes.GetThemeByName(cfg.Overlay.Theme.Name)

		var params rendering.RenderingParams
		params.Run = run
		params.Splits = splits
		params.Tc = templateCat
		params.Tn = templateName
		params.CssFileName = "output.css"
		params.CounterSettings.DisplayNext.IsLimited = cfg.Overlay.LimitSplitsAbove
		params.CounterSettings.DisplayNext.Count = cfg.Overlay.ShowSplitsAbove
		params.CounterSettings.DisplayPrev.IsLimited = cfg.Overlay.LimitSplitsBellow
		params.CounterSettings.DisplayPrev.Count = cfg.Overlay.ShowSplitsBellow
		params.Theme = theme
		params.TemplateDir = filepath.Join(cfg.General.DataDir, "templates")
		htmlBytes, cssBytes, err := rendering.RenderHTML(params)
		if err != nil {
			fmt.Printf("Error rendering HTML: %v\n", err)
			os.Exit(1)
		}
		outDir := filepath.Join(cfg.General.DataDir, "render")
		if err := rendering.WriteOutput(outDir, htmlBytes, cssBytes); err != nil {
			fmt.Printf("Error writing render output: %v\n", err)
			os.Exit(1)
		}
	},
}

// add show templates and cateogory command and support in config etc

func init() {
	renderRunCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")
	renderRunCmd.Flags().StringP("template-cat", "c", "base", "Template category to use")
	renderRunCmd.Flags().StringP("template-name", "n", "base", "Template name to use")

	renderCmd.AddCommand(renderRunCmd)
	rootCmd.AddCommand(renderCmd)
}
