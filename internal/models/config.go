package models

import "time"

// AssetKind defines the type of asset
type AssetKind string

const (
	KindPrompt      AssetKind = "prompt"
	KindSkill       AssetKind = "skill"
	KindInstruction AssetKind = "instruction"
)

// --- Consumer Config (.arca-assets.yaml) ---

type Config struct {
	Schema  string                  `yaml:"schema"`
	Sources map[string]SourceConfig `yaml:"sources"`
	Assets  []AssetEntry            `yaml:"assets"`
}

type SourceType string

const (
	SourceGit   SourceType = "git"
	SourceLocal SourceType = "local"
)

type SourceConfig struct {
	Type     SourceType `yaml:"type"`
	Provider string     `yaml:"provider,omitempty"` // github, azure, etc.
	URL      string     `yaml:"url,omitempty"`      // for git
	Path     string     `yaml:"path,omitempty"`     // for local
}

type AssetEntry struct {
	ID          string            `yaml:"id"`
	Source      string            `yaml:"source"`
	Version     string            `yaml:"version"`
	Projections map[string]string `yaml:"projections"` // name -> path (e.g. "default" -> ".github/prompts/...")
}

// --- Source Manifest (arca-manifest.yaml) ---

type Manifest struct {
	Schema          string                    `yaml:"schema"`
	VersionStrategy *VersionStrategy          `yaml:"version-strategy,omitempty"`
	Assets          map[string]ManifestAsset `yaml:"assets"`
}

type VersionStrategy struct {
	Template string `yaml:"template"`
}

type ManifestAsset struct {
	Kind        AssetKind                  `yaml:"kind"`
	Description string                     `yaml:"description,omitempty"`
	Versions    map[string]ManifestVersion `yaml:"versions"`
}

type ManifestVersion struct {
	Ref     string           `yaml:"ref,omitempty"`
	Path    string           `yaml:"path"`
	Runtime *AssetRuntime    `yaml:"runtime,omitempty"`
}

type AssetRuntime struct {
	LLM              []LLMTarget `yaml:"llm,omitempty"`
	MinContextTokens int         `yaml:"min_context_tokens,omitempty"`
	RequiresTools    bool        `yaml:"requires_tools,omitempty"`
}

type LLMTarget struct {
	Provider string   `yaml:"provider"`
	Models   []string `yaml:"models"`
}

// --- Lockfile (.arca-assets.lock) ---

type Lockfile struct {
	Assets []LockedAsset `json:"assets"`
}

type LockedAsset struct {
	ID           string    `json:"id"`
	Version      string    `json:"version"`
	Source       string    `json:"source"`
	Commit       string    `json:"commit"`
	SHA256       string    `json:"sha256"`
	ManifestHash string    `json:"manifestHash"`
	ResolvedAt   time.Time `json:"resolvedAt"`
}
