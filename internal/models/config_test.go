package models

import (
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestConfigUnmarshal(t *testing.T) {
	yamlData := `
schema: "1.0"
sources:
  github:
    type: "git"
    provider: "github"
    url: "https://github.com/adryledo/arca-assets.git"
  local-skills:
    type: "local"
    path: "../my-skills"
assets:
  - id: "my-skill"
    kind: "skill"
    source: "github"
    version: "1.0.0"
    projections:
      default: ".arca/my-skill"
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(yamlData), &cfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if cfg.Schema != "1.0" {
		t.Errorf("Expected schema 1.0, got %s", cfg.Schema)
	}

	if len(cfg.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(cfg.Sources))
	}

	if cfg.Sources["github"].Type != SourceGit {
		t.Errorf("Expected git source type, got %s", cfg.Sources["github"].Type)
	}

	if len(cfg.Assets) != 1 {
		t.Fatalf("Expected 1 asset, got %d", len(cfg.Assets))
	}

	asset := cfg.Assets[0]
	if asset.ID != "my-skill" || asset.Kind != KindSkill || asset.Source != "github" || asset.Version != "1.0.0" {
		t.Errorf("Asset fields did not unmarshal correctly: %+v", asset)
	}
	if asset.Projections["default"] != ".arca/my-skill" {
		t.Errorf("Projections did not unmarshal correctly")
	}
}

func TestManifestUnmarshal(t *testing.T) {
	yamlData := `
schema: "1.0"
assets:
  test-agent:
    kind: "skill"
    description: "A test agent"
    versions:
      "1.0.0":
        path: "skills/test-agent"
`
	var manifest Manifest
	if err := yaml.Unmarshal([]byte(yamlData), &manifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if manifest.Schema != "1.0" {
		t.Errorf("Expected schema 1.0, got %s", manifest.Schema)
	}

	asset, ok := manifest.Assets["test-agent"]
	if !ok {
		t.Fatalf("Asset 'test-agent' not found")
	}

	if asset.Kind != KindSkill {
		t.Errorf("Expected kind skill, got %s", asset.Kind)
	}

	version, ok := asset.Versions["1.0.0"]
	if !ok {
		t.Fatalf("Version 1.0.0 not found")
	}

	if version.Path != "skills/test-agent" {
		t.Errorf("Expected path 'skills/test-agent', got '%s'", version.Path)
	}
}

func TestLockedAsset(t *testing.T) {
	// Simple struct test to ensure constraints
	now := time.Now()
	locked := LockedAsset{
		ID:         "test",
		Version:    "1.0.0",
		Source:     "local",
		Commit:     "local",
		SHA256:     "abc",
		ResolvedAt: now,
	}

	if locked.ID != "test" || !reflect.DeepEqual(locked.ResolvedAt, now) {
		t.Errorf("LockedAsset fields malformed")
	}
}
