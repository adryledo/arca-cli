package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/adryledo/arca-cli/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var publishCmd = &cobra.Command{
	Use:   "publish [id] [version] [kind] [file-path]",
	Short: "Add or update an asset version in the local arca-manifest.yaml",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		assetID := args[0]
		version := args[1]
		kindStr := args[2]
		assetFile := args[3]

		cwd, _ := os.Getwd()
		manifestPath := filepath.Join(cwd, "arca-manifest.yaml")

		// 1. Load or init manifest
		var manifest models.Manifest
		data, err := os.ReadFile(manifestPath)
		if err == nil {
			if err := yaml.Unmarshal(data, &manifest); err != nil {
				return fmt.Errorf("failed to parse existing manifest: %w", err)
			}
		} else if os.IsNotExist(err) {
			manifest = models.Manifest{
				Schema: "1.0",
				Assets: make(map[string]models.ManifestAsset),
			}
		} else {
			return err
		}

		// 2. Validate asset file
		if _, err := os.Stat(assetFile); err != nil {
			return fmt.Errorf("asset file not found: %w", err)
		}

		// 3. Update manifest
		kind := models.AssetKind(kindStr)
		if kind != models.KindSkill && kind != models.KindInstruction {
			return fmt.Errorf("invalid asset kind: %s. Use 'skill', or 'instruction'", kindStr)
		}

		desc := "Added via arca publish"
		if extDesc := extractDescription(assetFile, kind); extDesc != "" {
			desc = extDesc
		}

		asset, ok := manifest.Assets[assetID]
		if !ok {
			asset = models.ManifestAsset{
				Kind:        kind,
				Description: desc,
				Versions:    make(map[string]models.ManifestVersion),
			}
		} else {
			// Update kind if it changed? Or keep existing?
			// The original TS implementation uses the passed kind or existing.
			// Let's allow updating the kind.
			asset.Kind = kind
		}

		// 4. Checkpointing: Pin the previous version to the current HEAD
		var highestV *semver.Version
		var highestVStr string
		for vStr := range asset.Versions {
			v, err := semver.NewVersion(vStr)
			if err != nil {
				continue
			}
			if highestV == nil || v.GreaterThan(highestV) {
				highestV = v
				highestVStr = vStr
			}
		}

		if highestVStr != "" {
			prevMeta := asset.Versions[highestVStr]
			if prevMeta.Ref == "" {
				// Try to get current commit
				gitCmd := exec.Command("git", "rev-parse", "HEAD")
				gitCmd.Dir = cwd
				out, err := gitCmd.Output()
				if err == nil {
					commit := strings.TrimSpace(string(out))
					prevMeta.Ref = commit
					asset.Versions[highestVStr] = prevMeta
				}
			}
		}

		// 5. Add new version
		if asset.Versions == nil {
			asset.Versions = make(map[string]models.ManifestVersion)
		}

		asset.Versions[version] = models.ManifestVersion{
			Path: assetFile,
		}
		if manifest.Assets == nil {
			manifest.Assets = make(map[string]models.ManifestAsset)
		}
		manifest.Assets[assetID] = asset

		// 6. Save
		newData, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}

		if err := os.WriteFile(manifestPath, newData, 0644); err != nil {
			return err
		}

		fmt.Printf("🚀 Published %s@%s to arca-manifest.yaml\n", assetID, version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
