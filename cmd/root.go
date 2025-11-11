package cmd

import (
	"fmt"
	"log"
	"os"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/ermokie/pkg/config"
	"github.com/stefanistkuhl/ermokie/pkg/db"
	hcmmigration "github.com/stefanistkuhl/ermokie/pkg/hcmMigration"
	"github.com/stefanistkuhl/ermokie/pkg/models"
	"github.com/stefanistkuhl/ermokie/pkg/models/setup"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
	"github.com/stefanistkuhl/ermokie/pkg/presets"
	"github.com/stefanistkuhl/ermokie/pkg/types"
)

var launchSetup bool = false

var rootCmd = &cobra.Command{
	Use:   "ermokie",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil || launchSetup {
			themeManager := setup.NewThemeManager()
			availableThemes := themeManager.GetAvailableThemes()
			cfg = config.NewConfig()
			cfg.General.DataDir = config.GetDataDir()

			steps := []setup.StepSpec{
				{
					ID:          "choose-theme",
					Kind:        setup.StepList,
					Title:       "Theme Selector",
					Description: "Pick your preferred color theme (use arrow keys to preview)",
					Options:     availableThemes,
					OnDone: func(v any) tea.Cmd {
						cfg.Theme.Overlay.Name = fmt.Sprintf("%v", v)
						cfg.Theme.App.Name = fmt.Sprintf("%v", v)

						return tea.Batch(
							tea.ClearScreen,
							tea.EnterAltScreen,
						)
					},
				},
				{
					ID:          "choose-keymap",
					Kind:        setup.StepList,
					Title:       "Keymap Selector",
					Description: "Choose your preferred keybinding style",
					Options:     models.GetAvailableKeymaps(),
					OnDoneWithStyles: func(v any, currentStyles styles.Styles) tea.Cmd {
						cfg.KeyBinds.Keymap = fmt.Sprintf("%v", v)
						models.LoadKeymapFromConfig(cfg.KeyBinds.Keymap)

						return tea.Batch(
							tea.ClearScreen,
							tea.EnterAltScreen,
						)
					},
				},
				{
					ID:          "show-help",
					Kind:        setup.StepHelp,
					Title:       "Help - Keybindings",
					Description: "Review the keybindings for your selected keymap",
					OnDone: func(v any) tea.Cmd {
						return tea.Batch(
							tea.ClearScreen,
							tea.EnterAltScreen,
						)
					},
				},
				{
					ID:          "import-splits",
					Kind:        setup.StepAsk,
					Title:       "Import Splits",
					Description: "Import your splits from Hit Counter Manager?",
					Question:    "Import existing splits?",
					DefaultY:    false,
					OnDoneWithStyles: func(v any, currentStyles styles.Styles) tea.Cmd {
						if v == true {
							xmlFile := models.NewFilePickerWithTheme(currentStyles)
							if xmlFile == "" {
							} else {
								store, err := db.Init()
								if err != nil {
									log.Fatalf("Failed to initialize the database %s", err)
								}
								profiles, err := hcmmigration.LoadProfiles(xmlFile)
								if err != nil {
									log.Fatalf("Failed to convert the XML file to Hit Counter Manager Profiles %s", err)
								}
								insertErr := hcmmigration.ImportProfiles(store, profiles)
								if insertErr != nil {
									log.Fatalf("Failed to insert the Hit Counter Manager Profiles into the DB %s", insertErr)
								}
							}
						}
						return tea.Batch(
							tea.ClearScreen,
							tea.EnterAltScreen,
						)
					},
				},
				{
					ID:          "preset-runs",
					Kind:        setup.StepAsk,
					Title:       "Preset Runs",
					Description: "Add preset runs to your app?",
					Question:    "Add preset runs?",
					DefaultY:    true,
					OnDoneWithResults: func(v any, currentStyles styles.Styles, results map[string]any) tea.Cmd {
						var toImport []presets.Preset
						if v == true {
							gameNames := presets.GetPresetGameNames()
							selectedGames := models.NewFuzzyFinderWithTheme(gameNames, true, currentStyles, "Select Game(s) to import splits from")
							for _, game := range selectedGames {
								runNames := presets.GetPresetRunNamesByGame(game)
								selectedRuns := models.NewFuzzyFinderWithTheme(runNames, true, currentStyles, fmt.Sprintf("Select Runs(s) to import splits from %s", game))
								for _, run := range selectedRuns {
									toImport = append(toImport, presets.GetPresetByName(run))
								}

							}
						}

						store, getDbErr := db.Init()
						if getDbErr != nil {
							log.Fatalf("Failed to initialize the database %s", err)
						}
						cat := presets.NewPresets()
						var runs []types.RunCreate
						for _, preset := range toImport {
							run, ok := presets.PresetToRunType(preset, cat)
							if !ok {

							}
							runs = append(runs, run)
						}

						err := presets.ImportPresets(store, runs)
						if err != nil {
							log.Fatal(err)
						}

						return tea.Batch(
							tea.ClearScreen,
							tea.EnterAltScreen,
						)
					},
				},
				{
					ID:          "rename-behavior",
					Kind:        setup.StepAsk,
					Title:       "Rename Behavior",
					Description: "When renaming, place cursor at end or delete existing text?",
					Question:    "Cursor at end? (No = delete existing)",
					DefaultY:    true,
					OnDone: func(v any) tea.Cmd {
						if reflect.TypeOf(v).Kind() == reflect.Bool {
							cfg.General.ClearOnRename = reflect.ValueOf(v).Bool()
						} else {
							cfg.General.ClearOnRename = true
						}
						return tea.Batch(
							tea.ClearScreen,
							tea.EnterAltScreen,
						)
					},
				},
			}
			p := tea.NewProgram(
				setup.NewSetupFlow(func() error { return loadDefaultsAndSave(cfg) }, steps, &cfg),
				tea.WithAltScreen(),
			)

			if _, err := p.Run(); err != nil {
				log.Println("setup error:", err)
				os.Exit(1)
			}
		} else {
			models.LoadKeymapFromConfig(cfg.KeyBinds.Keymap)
			theme := "default"
			themeManager := setup.NewThemeManager()
			if themeManager.ValidateTheme(cfg.Theme.App.Name) {
				theme = cfg.Theme.App.Name
			}
			styles := themeManager.ApplyTheme(theme, styles.DefaultStyles())
			store, getDbErr := db.Init()
			if getDbErr != nil {
				log.Fatalf("Failed to initialize the database %s", err)
			}
			models.NewMainScreen(styles, store)
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVar(&launchSetup, "setup", false, "Run the setup")
}

func loadDefaultsAndSave(cfg config.Config) error {
	return config.SaveConfig(cfg)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
