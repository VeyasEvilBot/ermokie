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
		// rawFlag, _ := cmd.Flags().GetBool("raw")
		var templateCat rendering.TemplateCategory
		var templateName rendering.TemplateName
		templateCatFlag, getTemplateCatErr := cmd.Flags().GetString("template-cat")
		if getTemplateCatErr != nil {
			templateCat = rendering.TemplateCategoryBase
		}
		templateNameFlag, getTemplateNameErr := cmd.Flags().GetString("template-name")
		if getTemplateNameErr != nil {
			templateName = rendering.TemplateNameBase
		}
		templateCat = rendering.TemplateCategory(templateCatFlag)
		templateName = rendering.TemplateName(templateNameFlag)
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
		htmlBytes, cssBytes, err := rendering.RenderHtml(params)
		if err != nil {
			fmt.Printf("Error rendering HTML: %v\n", err)
			os.Exit(1)
		}
		outDir := filepath.Join(cfg.General.DataDir, "render")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
		fHtml, createHtmlErr := os.Create(filepath.Join(outDir, "output.html"))
		if createHtmlErr != nil {
			fmt.Printf("Error creating file: %v\n", createHtmlErr)
			os.Exit(1)
		}
		fCss, createCssErr := os.Create(filepath.Join(outDir, "output.css"))
		if createCssErr != nil {
			fmt.Printf("Error creating file: %v\n", createCssErr)
			os.Exit(1)
		}
		defer fHtml.Close()
		defer fCss.Close()
		fHtml.Write(htmlBytes)
		fCss.Write(cssBytes)
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
