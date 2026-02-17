package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adryledo/arca-cli/internal/models"
	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName   = ".arca-assets.yaml"
	LockFileName     = ".arca-assets.lock"
	DefaultSchemaVer = "1.0"
)

type Manager struct {
	WorkspaceRoot string
}

func NewManager(workspaceRoot string) *Manager {
	return &Manager{WorkspaceRoot: workspaceRoot}
}

// LoadConfig loads the .arca-assets.yaml file.
func (m *Manager) LoadConfig() (*models.Config, error) {
	path := filepath.Join(m.WorkspaceRoot, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Config{
				Schema:  DefaultSchemaVer,
				Sources: make(map[string]models.SourceConfig),
				Assets:  []models.AssetEntry{},
			}, nil
		}
		return nil, err
	}

	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig saves the configuration to .arca-assets.yaml.
func (m *Manager) SaveConfig(cfg *models.Config) error {
	path := filepath.Join(m.WorkspaceRoot, ConfigFileName)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// EnsureSource registers a source if it doesn't exist and returns its alias.
func (m *Manager) EnsureSource(cfg *models.Config, url string, stype models.SourceType) string {
	// Check if URL already exists
	for alias, src := range cfg.Sources {
		if (stype == models.SourceGit && src.URL == url) || (stype == models.SourceLocal && src.Path == url) {
			return alias
		}
	}

	// Derive alias from URL/Path
	alias := m.deriveAlias(url)

	// Ensure uniqueness
	baseAlias := alias
	counter := 1
	for {
		if _, exists := cfg.Sources[alias]; !exists {
			break
		}
		alias = fmt.Sprintf("%s-%d", baseAlias, counter)
		counter++
	}

	srcCfg := models.SourceConfig{Type: stype}
	if stype == models.SourceGit {
		srcCfg.URL = url
		// Simplified: infer provider
		if strings.Contains(url, "github.com") {
			srcCfg.Provider = "github"
		} else if strings.Contains(url, "azure.com") {
			srcCfg.Provider = "azure"
		}
	} else {
		srcCfg.Path = url
	}

	if cfg.Sources == nil {
		cfg.Sources = make(map[string]models.SourceConfig)
	}
	cfg.Sources[alias] = srcCfg
	return alias
}

// AddAsset adds or updates an asset entry in the config.
func (m *Manager) AddAsset(cfg *models.Config, entry models.AssetEntry) {
	for i, a := range cfg.Assets {
		if a.ID == entry.ID && a.Source == entry.Source {
			cfg.Assets[i] = entry
			return
		}
	}
	cfg.Assets = append(cfg.Assets, entry)
}

// LoadLockfile loads the .arca-assets.lock file.
func (m *Manager) LoadLockfile() (*models.Lockfile, error) {
	path := filepath.Join(m.WorkspaceRoot, LockFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Lockfile{Assets: []models.LockedAsset{}}, nil
		}
		return nil, err
	}

	var lock models.Lockfile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lockfile: %w", err)
	}
	return &lock, nil
}

// SaveLockfile saves the lockfile to .arca-assets.lock.
func (m *Manager) SaveLockfile(lock *models.Lockfile) error {
	path := filepath.Join(m.WorkspaceRoot, LockFileName)
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *Manager) deriveAlias(input string) string {
	// Simple alias derivation
	clean := strings.TrimRight(input, "/")
	parts := strings.Split(clean, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		return strings.TrimSuffix(name, ".git")
	}
	return "source"
}
