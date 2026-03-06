package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/adryledo/arca-cli/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	instructionsFolder string
	skillsFolder       string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an arca-manifest.yaml file by scanning folders",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		manifestPath := filepath.Join(cwd, "arca-manifest.yaml")

		manifest := models.Manifest{
			Schema: "1.0",
			Assets: make(map[string]models.ManifestAsset),
		}

		// Try to read existing manifest
		data, err := os.ReadFile(manifestPath)
		if err == nil {
			if err := yaml.Unmarshal(data, &manifest); err != nil {
				return fmt.Errorf("failed to parse existing manifest: %w", err)
			}
			if manifest.Assets == nil {
				manifest.Assets = make(map[string]models.ManifestAsset)
			}
		}

		addAssets := func(folder string, kind models.AssetKind, isDir bool) error {
			if folder == "" {
				return nil
			}
			fullPath := filepath.Join(cwd, folder)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				fmt.Printf("⚠️  Folder %s not found, skipping.\n", folder)
				return nil
			}
			fmt.Printf("🔍 Scanning %s for %s...\n", folder, kind)
			entries, err := os.ReadDir(fullPath)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.IsDir() != isDir {
					continue
				}
				name := entry.Name()
				if strings.HasPrefix(name, ".") {
					continue
				}

				var assetID string
				if isDir {
					assetID = name
				} else {
					assetID = strings.TrimSuffix(strings.TrimSuffix(name, filepath.Ext(name)), ".instructions")
				}

				relPath := filepath.ToSlash(filepath.Join(folder, name))

				desc := fmt.Sprintf("Auto-discovered %s", kind)
				if extDesc := extractDescription(filepath.Join(fullPath, entry.Name()), kind); extDesc != "" {
					desc = extDesc
				}

				asset, ok := manifest.Assets[assetID]
				if !ok {
					asset = models.ManifestAsset{
						Kind:        kind,
						Description: desc,
						Versions:    make(map[string]models.ManifestVersion),
					}
				}

				if asset.Versions == nil {
					asset.Versions = make(map[string]models.ManifestVersion)
				}

				if len(asset.Versions) == 0 {
					asset.Versions["0.0.1"] = models.ManifestVersion{
						Path: relPath,
					}
				} else {
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
						v := asset.Versions[highestVStr]
						v.Path = relPath
						asset.Versions[highestVStr] = v
					}
				}

				manifest.Assets[assetID] = asset
				fmt.Printf("   ✅ Added %s (%s)\n", assetID, kind)
			}
			return nil
		}

		if err := addAssets(instructionsFolder, models.KindInstruction, false); err != nil {
			return err
		}
		if err := addAssets(skillsFolder, models.KindSkill, true); err != nil {
			return err
		}

		newData, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}

		if err := os.WriteFile(manifestPath, newData, 0644); err != nil {
			return err
		}

		fmt.Println("✨ arca-manifest.yaml generated successfully.")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&instructionsFolder, "instructions", "instructions", "Folder containing instructions")
	initCmd.Flags().StringVar(&skillsFolder, "skills", "skills", "Folder containing skills")
	rootCmd.AddCommand(initCmd)
}
