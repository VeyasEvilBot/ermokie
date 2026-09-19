package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfigMergesFileOverDefaults(t *testing.T) {
	configRoot := t.TempDir()
	dataRoot := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv(EnvDataDir, dataRoot)

	path := filepath.Join(configRoot, "ermokie", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	contents := []byte("[keybinds]\nkeymap = 'wasd'\n[keybinds.custom]\nup = ['i']\nquit = ['x', 'ctrl+c']\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.KeyBinds.Keymap != "wasd" {
		t.Fatalf("keymap = %q", cfg.KeyBinds.Keymap)
	}
	if got := cfg.KeyBinds.Custom["up"]; !reflect.DeepEqual(got, []string{"i"}) {
		t.Fatalf("custom up = %#v", got)
	}
	if cfg.Overlay.Port != 5539 || cfg.Overlay.TemplateCategory != "base" || cfg.Overlay.TemplateName != "base" {
		t.Fatalf("overlay defaults were not preserved: %#v", cfg.Overlay)
	}
	if cfg.General.DataDir != dataRoot {
		t.Fatalf("data dir = %q, want %q", cfg.General.DataDir, dataRoot)
	}
}
