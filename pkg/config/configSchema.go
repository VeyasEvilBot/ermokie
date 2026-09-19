package config

import (
	"fmt"
	"log"
)

type Config struct {
	General  GeneralSettings `toml:"general"`
	Theme    ThemeConfig     `toml:"theme"`
	KeyBinds Binds           `toml:"keybinds"`
	Overlay  OverlayConfig   `toml:"overlay"`
}

type GeneralSettings struct {
	ClearOnRename      bool   `toml:"clear_on_rename"`
	DataDir            string `toml:"data_dir"`
	ShowAttemptCounter bool   `toml:"show_attempt_counter"`
	ShowDiff           bool   `toml:"show_hit_diff"`
}

type ThemeConfig struct {
	// Name    string             `toml:"name"`
	App AppThemeConfig `toml:"app"`
}

type Binds struct {
	Keymap string              `toml:"keymap"`
	Custom map[string][]string `toml:"custom"`
}

type AppThemeConfig struct {
	// add exposing colors
	Name string `toml:"name"`
}

type OverlayThemeConfig struct {
	// add exposing colors
	Name string `toml:"name"`
}

type OverlayConfig struct {
	Port              int                `toml:"port"`
	Theme             OverlayThemeConfig `toml:"theme"`
	TemplateCategory  string             `toml:"template_category"`
	TemplateName      string             `toml:"template_name"`
	ShowSplitsAbove   int                `toml:"show_splits_above"`
	LimitSplitsAbove  bool               `toml:"limit_splits_above"`
	ShowSplitsBellow  int                `toml:"show_splits_bellow"`
	LimitSplitsBellow bool               `toml:"limit_splits_bellow"`
}

func NewConfig() Config {
	var c Config
	dataDir, err := EnsureUserDataDir("ermokie")
	if err != nil {
		log.Fatalf("Failed to create or get the dir to store the data of the app, %v", err)
	}
	SetDataDir(dataDir)
	c.General.DataDir = GetDataDir()
	c.General.ClearOnRename = true

	c.Theme.App.Name = "default"
	c.Overlay.Theme.Name = "default"
	c.Overlay.TemplateCategory = "base"
	c.Overlay.TemplateName = "base"
	c.KeyBinds.Keymap = "arrow-vim"
	c.General.ShowAttemptCounter = true
	c.General.ShowDiff = false
	c.Overlay.Port = 5539
	c.Overlay.ShowSplitsAbove = 5
	c.Overlay.ShowSplitsBellow = 5
	c.Overlay.LimitSplitsAbove = false
	c.Overlay.LimitSplitsBellow = false
	return c
}

func (c Config) PrintConfig() {
	fmt.Println("General Settings")
	fmt.Println("Clear on rename:", c.General.ClearOnRename)
	fmt.Println("Data Dir:", c.General.DataDir)
	fmt.Println("Show Attempt Counter:", c.General.ShowAttemptCounter)
	fmt.Println("Show Diff:", c.General.ShowDiff)
	fmt.Println()
	fmt.Println("Theme Settings")
	fmt.Println("App Theme:", c.Theme.App.Name)
	fmt.Println()
	fmt.Println("Keybinds")
	fmt.Println("Keymap:", c.KeyBinds.Keymap)
	fmt.Println()
	fmt.Println("Overlay Settings")
	fmt.Println("Port:", c.Overlay.Port)
	fmt.Println("Theme:", c.Overlay.Theme.Name)
	fmt.Println("Show Splits Above:", c.Overlay.ShowSplitsAbove)
	fmt.Println("Limit Splits Above:", c.Overlay.LimitSplitsAbove)
	fmt.Println("Show Splits Below:", c.Overlay.ShowSplitsBellow)
	fmt.Println("Limit Splits Below:", c.Overlay.LimitSplitsBellow)
}
