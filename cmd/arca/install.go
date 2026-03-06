package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adryledo/arca-cli/internal/config"
	"github.com/adryledo/arca-cli/internal/downloader"
	"github.com/adryledo/arca-cli/internal/hasher"
	"github.com/adryledo/arca-cli/internal/models"
	"github.com/adryledo/arca-cli/internal/projector"
	"github.com/adryledo/arca-cli/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	targetPath string
	projName   string
)

var installCmd = &cobra.Command{
	Use:   "install [url|path] [id] [version]",
	Short: "Install an asset from a source",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceStr := args[0]
		assetID := args[1]
		versionConstraint := "latest"
		if len(args) > 2 {
			versionConstraint = args[2]
		}

		cwd, _ := os.Getwd()
		res := resolver.New(cwd)
		proj := projector.New(cwd)
		cfgMgr := config.NewManager(cwd)
		cache := downloader.NewCacheProvider("")

		// 1. Load existing config
		cfg, err := cfgMgr.LoadConfig()
		if err != nil {
			return err
		}

		// 2. Identify source
		stype := models.SourceGit
		if info, err := os.Stat(sourceStr); err == nil && info.IsDir() {
			stype = models.SourceLocal
		}
		sourceAlias := cfgMgr.EnsureSource(cfg, sourceStr, stype)

		fmt.Printf("🔍 Resolving asset %s from %s (%s)...\n", assetID, sourceStr, sourceAlias)

		// 3. Load Manifest
		manifest, err := res.LoadManifest(cfg.Sources[sourceAlias], "main")
		if err != nil {
			return err
		}

		// 4. Resolve full graph
		assets, err := res.ResolveGraph(manifest, assetID, versionConstraint)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Resolved %d asset(s) including dependencies\n", len(assets))

		lock, err := cfgMgr.LoadLockfile()
		if err != nil {
			return err
		}

		for _, item := range assets {
			fmt.Printf("📦 Installing %s@%s...\n", item.ID, item.Version)

			isDir := item.Kind == models.KindSkill
			commitSHA := ""
			assetPath := cache.GetAssetPath(sourceAlias, item.ID, item.Version, isDir)
			cacheDir, _ := cache.EnsureDir(sourceAlias, item.ID, item.Version)

			if stype == models.SourceLocal {
				absPath := filepath.Join(sourceStr, item.Meta.Path)
				if isDir {
					commitSHA = "local"
				} else {
					data, err := os.ReadFile(absPath)
					if err != nil {
						return err
					}
					err = os.WriteFile(assetPath, data, 0644)
					if err != nil {
						return err
					}
					commitSHA = "local"
				}
			} else {
				gitDownloader := downloader.NewGitDownloader()
				ref := item.Meta.Ref
				if ref == "" {
					ref = "main"
				}
				if isDir {
					sha, err := gitDownloader.FetchDirectory(sourceStr, item.Meta.Path, ref, cacheDir)
					if err != nil {
						return err
					}
					commitSHA = sha
				} else {
					data, sha, err := gitDownloader.FetchFile(sourceStr, item.Meta.Path, ref)
					if err != nil {
						return err
					}
					err = os.WriteFile(assetPath, []byte(data), 0644)
					if err != nil {
						return err
					}
					commitSHA = sha
				}
			}

			// Projection
			actualTarget := targetPath
			actualProjName := projName

			if item.ID != assetID {
				// Dependencies go to default location
				ext := ".md"
				if isDir {
					ext = ""
				}
				actualTarget = fmt.Sprintf(".arca/assets/%s/%s%s", sourceAlias, item.ID, ext)
				actualProjName = "default"
			} else if actualTarget == "" {
				ext := ".md"
				if isDir {
					ext = ""
				}
				actualTarget = fmt.Sprintf(".arca/assets/%s/%s%s", sourceAlias, item.ID, ext)
			}

			_, err = proj.Project(assetPath, actualTarget, isDir)
			if err != nil {
				return err
			}
			fmt.Printf("   🔗 Projected %s to %s\n", item.ID, actualTarget)

			// Update Config Entry (Only for the root asset)
			if item.ID == assetID {
				entry := models.AssetEntry{
					ID:      item.ID,
					Kind:    item.Kind,
					Source:  sourceAlias,
					Version: item.Version,
					Projections: map[string]string{
						actualProjName: actualTarget,
					},
				}
				cfgMgr.AddAsset(cfg, entry)
			}

			// Update Lockfile Entry
			var contentHash string
			if isDir {
				contentHash, err = hasher.HashDir(assetPath)
			} else {
				contentHash, err = hasher.HashFile(assetPath)
			}
			if err != nil {
				return fmt.Errorf("failed to hash asset: %w", err)
			}
			locked := models.LockedAsset{
				ID:         item.ID,
				Version:    item.Version,
				Source:     sourceAlias,
				Commit:     commitSHA,
				SHA256:     contentHash,
				ResolvedAt: time.Now(),
			}

			found := false
			for i, la := range lock.Assets {
				if la.ID == item.ID && la.Source == sourceAlias {
					lock.Assets[i] = locked
					found = true
					break
				}
			}
			if !found {
				lock.Assets = append(lock.Assets, locked)
			}
		}

		if err := cfgMgr.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if err := cfgMgr.SaveLockfile(lock); err != nil {
			return fmt.Errorf("failed to save lockfile: %w", err)
		}

		fmt.Println("✨ Installation complete and persisted.")
		return nil
	},
}

func init() {
	installCmd.Flags().StringVarP(&targetPath, "target", "t", "", "Projection target path")
	installCmd.Flags().StringVarP(&projName, "name", "n", "default", "Projection name")
	rootCmd.AddCommand(installCmd)
}
