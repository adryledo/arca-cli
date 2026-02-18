package resolver

import (
	"testing"

	"github.com/adryledo/arca-cli/internal/models"
)

func TestResolveVersion(t *testing.T) {
	manifest := &models.Manifest{
		Schema: "1.0",
		VersionStrategy: &models.VersionStrategy{
			Template: "v{{version}}",
		},
		Assets: map[string]models.ManifestAsset{
			"test-asset": {
				Kind: models.KindPrompt,
				Versions: map[string]models.ManifestVersion{
					"1.0.0": {Path: "p1.md"},
					"1.1.0": {Path: "p2.md", Ref: "special-tag"},
					"2.0.0": {Path: "p3.md"},
				},
			},
		},
	}

	r := New("/tmp")

	tests := []struct {
		constraint      string
		expectedVersion string
		expectedRef     string
	}{
		{"latest", "2.0.0", "v2.0.0"},
		{"^1.0.0", "1.1.0", "special-tag"},
		{"1.0.0", "1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		v, meta, err := r.ResolveVersion(manifest, "test-asset", tt.constraint)
		if err != nil {
			t.Errorf("Constraint %s failed: %v", tt.constraint, err)
			continue
		}
		if v != tt.expectedVersion {
			t.Errorf("Constraint %s: expected version %s, got %s", tt.constraint, tt.expectedVersion, v)
		}
		if meta.Ref != tt.expectedRef {
			t.Errorf("Constraint %s: expected ref %s, got %s", tt.constraint, tt.expectedRef, meta.Ref)
		}
	}
}

func TestResolveVersionFallbacks(t *testing.T) {
	manifest := &models.Manifest{
		Assets: map[string]models.ManifestAsset{
			"non-semver": {
				Versions: map[string]models.ManifestVersion{
					"alpha": {Path: "a.md"},
					"beta":  {Path: "b.md"},
				},
			},
		},
	}
	r := New("/tmp")

	// Test exact match for non-semver
	v, _, err := r.ResolveVersion(manifest, "non-semver", "alpha")
	if err != nil {
		t.Fatal(err)
	}
	if v != "alpha" {
		t.Errorf("Expected alpha, got %s", v)
	}

	// Test latest for non-semver (alphabetical fallback)
	v, _, err = r.ResolveVersion(manifest, "non-semver", "latest")
	if err != nil {
		t.Fatal(err)
	}
	if v != "beta" {
		t.Errorf("Expected beta, got %s", v)
	}
}
