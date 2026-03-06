package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/adryledo/arca-cli/internal/models"
)

func TestManager_Config(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// 1. Load empty config (should not error, just return defaults)
	cfg, err := mgr.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load generic config: %v", err)
	}
	if cfg.Schema != DefaultSchemaVer {
		t.Errorf("Expected schema %s, got %s", DefaultSchemaVer, cfg.Schema)
	}

	// 2. Add source
	alias := mgr.EnsureSource(cfg, "https://github.com/test/repo.git", models.SourceGit)
	if alias != "repo" {
		t.Errorf("Expected alias 'repo', got '%s'", alias)
	}

	// ensure unique alias
	alias2 := mgr.EnsureSource(cfg, "https://github.com/other/repo.git", models.SourceGit)
	if alias2 != "repo-1" {
		t.Errorf("Expected alias 'repo-1', got '%s'", alias2)
	}

	// 3. Add asset
	mgr.AddAsset(cfg, models.AssetEntry{
		ID:      "test-skill",
		Kind:    models.KindSkill,
		Source:  alias,
		Version: "1.0.0",
		Projections: map[string]string{
			"default": ".arca/assets/test-skill",
		},
	})

	// 4. Save and reload
	if err := mgr.SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	cfgLoaded, err := mgr.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if len(cfgLoaded.Assets) != 1 || cfgLoaded.Assets[0].ID != "test-skill" {
		t.Errorf("Loaded config does not match saved config")
	}
}

func TestManager_Lockfile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// 1. Load empty
	lock, err := mgr.LoadLockfile()
	if err != nil {
		t.Fatalf("Failed to load generic lockfile: %v", err)
	}
	if len(lock.Assets) != 0 {
		t.Errorf("Expected empty lockfile")
	}

	now := time.Now().Truncate(time.Second) // JSON marshaling might lose sub-second precision
	// 2. Add lock and save
	lock.Assets = append(lock.Assets, models.LockedAsset{
		ID:         "test",
		Version:    "1.0.0",
		Source:     "provider",
		Commit:     "abcdef",
		SHA256:     "hash",
		ResolvedAt: now,
	})

	if err := mgr.SaveLockfile(lock); err != nil {
		t.Fatalf("Failed to save lockfile: %v", err)
	}

	// 3. Reload
	lockLoaded, err := mgr.LoadLockfile()
	if err != nil {
		t.Fatalf("Failed to load saved lockfile: %v", err)
	}

	if len(lockLoaded.Assets) != 1 {
		t.Fatalf("Expected 1 asset in lockfile, got %d", len(lockLoaded.Assets))
	}

	if !reflect.DeepEqual(lockLoaded.Assets[0].ResolvedAt.UTC(), lock.Assets[0].ResolvedAt.UTC()) {
		t.Errorf("Lockfile loading did not preserve time. Expected %v, got %v", lock.Assets[0].ResolvedAt, lockLoaded.Assets[0].ResolvedAt)
	}
}

func TestDeriveAlias(t *testing.T) {
	mgr := NewManager("/tmp")

	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/foo/bar.git", "bar"},
		{"https://gitlab.com/repo", "repo"},
		{"/local/path/to/my-skill", "my-skill"},
		{"invalid/chars!@#", "chars---"},
		{"", "source"},
		{"/", "source"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mgr.deriveAlias(tt.input)
			if got != tt.expected {
				t.Errorf("deriveAlias(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
