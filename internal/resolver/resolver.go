package resolver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/adryledo/arca-cli/internal/models"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/yaml.v3"
)

type Resolver struct {
	WorkspaceRoot string
}

func New(workspaceRoot string) *Resolver {
	return &Resolver{WorkspaceRoot: workspaceRoot}
}

// LoadManifest fetches and parses the arca-manifest.yaml from a source.
func (r *Resolver) LoadManifest(source models.SourceConfig) (*models.Manifest, error) {
	var data []byte
	var err error

	switch source.Type {
	case models.SourceLocal:
		manifestPath := filepath.Join(source.Path, "arca-manifest.yaml")
		if !filepath.IsAbs(manifestPath) {
			manifestPath = filepath.Join(r.WorkspaceRoot, manifestPath)
		}
		data, err = os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read local manifest: %w", err)
		}
	case models.SourceGit:
		data, err = r.fetchManifestFromGit(source.URL)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}

	var manifest models.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

func (r *Resolver) fetchManifestFromGit(url string) ([]byte, error) {
	// Clone manifest from 'main' (simplified ref resolution for now)
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:   url,
		Depth: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone manifest repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	f, err := wt.Filesystem.Open("arca-manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("arca-manifest.yaml not found in repo: %w", err)
	}
	defer f.Close()

	return io.ReadAll(f)
}

// ResolveVersion finds the best version matching a constraint for an asset.
func (r *Resolver) ResolveVersion(manifest *models.Manifest, assetID string, constraint string) (string, models.ManifestVersion, error) {
	asset, ok := manifest.Assets[assetID]
	if !ok {
		return "", models.ManifestVersion{}, fmt.Errorf("asset %s not found in manifest", assetID)
	}

	var availableVersions []string
	for v := range asset.Versions {
		availableVersions = append(availableVersions, v)
	}

	// Simple sort if not using semver, but we should use it
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		// Fallback to exact match or "latest" logic
		if constraint == "latest" {
			// Find highest version string (simplified)
			highest := ""
			for v := range asset.Versions {
				if v > highest {
					highest = v
				}
			}
			return highest, asset.Versions[highest], nil
		}
		version, ok := asset.Versions[constraint]
		if !ok {
			return "", models.ManifestVersion{}, fmt.Errorf("version %s not found", constraint)
		}
		return constraint, version, nil
	}

	var bestVersion *semver.Version
	for vStr := range asset.Versions {
		v, err := semver.NewVersion(vStr)
		if err != nil {
			continue
		}
		if c.Check(v) {
			if bestVersion == nil || v.GreaterThan(bestVersion) {
				bestVersion = v
			}
		}
	}

	if bestVersion == nil {
		return "", models.ManifestVersion{}, fmt.Errorf("no version matching %s found", constraint)
	}

	bestStr := bestVersion.Original()
	return bestStr, asset.Versions[bestStr], nil
}
