package config

import "log"

type Config struct {
	General  GeneralSettings `toml:"general"`
	Theme    ThemeConfig     `toml:"theme"`
	KeyBinds Binds           `toml:"keybinds"`
}

type GeneralSettings struct {
	ClearOnRename bool   `toml:"clear_on_rename"`
	DataDir       string `toml:"data_dir"`
}

type ThemeConfig struct {
	// Name    string             `toml:"name"`
	App     AppThemeConfig     `toml:"app"`
	Overlay OverlayThemeConfig `toml:"overlay"`
}

type Binds struct {
	// Quit string `toml:"quit"`
	// Save string `toml:"save"`
}

type AppThemeConfig struct {
	Name string `toml:"name"`
}

type OverlayThemeConfig struct {
	Name string `toml:"name"`
}

func NewConfig() Config {
	var c Config
	dataDir, err := GetDataDirWithFallback("ermokie")
	if err != nil {
		log.Fatalf("Failed to create or get the dir to store the data of the app, %v", err)
	}
	SetDataDir(dataDir)
	c.General.DataDir = GetDataDir()
	c.General.ClearOnRename = true

	c.Theme.App.Name = "default"
	c.Theme.Overlay.Name = "default"
	return c
}
