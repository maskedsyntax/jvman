package config

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/maskedsyntax/jvman/internal/paths"
)

type InstalledJVM struct {
	Path   string `json:"path"`
	Vendor string `json:"vendor"`
}

type Config struct {
	Global         string                  `json:"global"`
	LocalOverrides map[string]string       `json:"local_overrides"`
	Installed      map[string]InstalledJVM `json:"installed"`
}

var (
	instance *Config
	mu       sync.RWMutex
)

func defaultConfig() *Config {
	return &Config{
		Global:         "",
		LocalOverrides: make(map[string]string),
		Installed:      make(map[string]InstalledJVM),
	}
}

func Load() (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	configPath, err := paths.ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			instance = defaultConfig()
			return instance, nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.LocalOverrides == nil {
		cfg.LocalOverrides = make(map[string]string)
	}
	if cfg.Installed == nil {
		cfg.Installed = make(map[string]InstalledJVM)
	}

	instance = cfg
	return instance, nil
}

func Save(cfg *Config) error {
	mu.Lock()
	defer mu.Unlock()

	if err := paths.EnsureDirectories(); err != nil {
		return err
	}

	configPath, err := paths.ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	instance = cfg
	return nil
}

func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if instance == nil {
		return defaultConfig()
	}
	return instance
}
