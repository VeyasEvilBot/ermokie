package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml/v2"
)

const (
	EnvDataDir = "ERMOKIE_DATA_DIR"
)

var ErrNoConfigFile = errors.New("no config file")

func LoadConfig() (Config, error) {
	var cfg Config
	path, getPathErr := GetConfigPath()
	if getPathErr != nil {
		return cfg, getPathErr
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, ErrNoConfigFile
	}

	cfg, getConfErr := readConfigFile(path)
	if getConfErr != nil {
		return cfg, getConfErr
	}
	SetDataDir(cfg.General.DataDir)

	return cfg, nil
}

func CreateDefaultConfigFile(cfg Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	conf, confErr := toml.Marshal(cfg)
	if confErr != nil {
		return fmt.Errorf("failed to marshal default config values: %w", confErr)
	}
	if writeErr := os.WriteFile(path, conf, 0o644); writeErr != nil {
		return fmt.Errorf("failed to write default config to %s: %w", path, writeErr)
	}
	return nil
}

func GetConfigPath() (string, error) {
	configRoot, err := os.UserConfigDir()

	if err != nil {
		home, herr := homedir.Dir()
		if herr != nil {
			return "", fmt.Errorf("could not resolve the home directory: %w", herr)
		}
		configRoot = filepath.Join(home, ".config")
	}

	configDir := filepath.Join(configRoot, "ermokie")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	return configPath, nil
}

func readConfigFile(path string) (Config, error) {
	cfg := NewConfig()
	f, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("could not open config file: %w", err)
	}
	confErr := toml.Unmarshal(f, &cfg)
	if confErr != nil {
		return cfg, fmt.Errorf("failed to load the config data: %w", confErr)
	}

	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := GetConfigPath()
	if err != nil && !errors.Is(err, ErrNoConfigFile) {
		return err
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if werr := os.WriteFile(path, b, 0o644); werr != nil {
		return fmt.Errorf("write config: %w", werr)
	}
	return nil
}

var globalDataDir string

func EnsureUserDataDir(appName string) (string, error) {
	var dir string

	if v := os.Getenv(EnvDataDir); v != "" {
		dir = v
	} else {
		switch runtime.GOOS {
		case "windows":
			if d := os.Getenv("APPDATA"); d != "" {
				dir = filepath.Join(d, appName)
			} else if d := os.Getenv("LOCALAPPDATA"); d != "" {
				dir = filepath.Join(d, appName)
			} else {
				home, herr := os.UserHomeDir()
				if herr != nil {
					return "", herr
				}
				dir = filepath.Join(home, "AppData", "Roaming", appName)
			}
		case "darwin":
			home, herr := os.UserHomeDir()
			if herr != nil {
				return "", herr
			}
			dir = filepath.Join(home, "Library", "Application Support", appName)
		default:
			if d := os.Getenv("XDG_DATA_HOME"); d != "" {
				dir = filepath.Join(d, appName)
			} else {
				home, herr := os.UserHomeDir()
				if herr != nil {
					return "", herr
				}
				dir = filepath.Join(home, ".local", "share", appName)
			}
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func GetDataDirWithFallback(appName string) (string, error) {
	if globalDataDir != "" {
		return globalDataDir, nil
	}

	if cfg, err := LoadConfig(); err == nil && cfg.General.DataDir != "" {
		dir := cfg.General.DataDir
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		return dir, nil
	}

	return EnsureUserDataDir(appName)
}

func SetDataDir(dir string) {
	globalDataDir = dir
}

func GetDataDir() string {
	return globalDataDir
}
