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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all assets defined in .arca-assets.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		cfgMgr := config.NewManager(cwd)
		res := resolver.New(cwd)
		proj := projector.New(cwd)
		cache := downloader.NewCacheProvider("")

		// 1. Load config and lockfile
		cfg, err := cfgMgr.LoadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Assets) == 0 {
			fmt.Println("No assets defined in .arca-assets.yaml")
			return nil
		}

		lock, err := cfgMgr.LoadLockfile()
		if err != nil {
			return err
		}

		// Use a map to collect all assets to sync (avoids duplicates)
		type syncItem struct {
			ID          string
			Source      models.SourceConfig
			SourceAlias string
			Version     string
			Meta        models.ManifestVersion
			Kind        models.AssetKind
			Projections map[string]string
		}
		toSync := make(map[string]syncItem)

		for _, asset := range cfg.Assets {
			source, ok := cfg.Sources[asset.Source]
			if !ok {
				fmt.Printf("⚠️  Source %s not found for asset %s, skipping.\n", asset.Source, asset.ID)
				continue
			}

			// Determine manifest revision (pin to locked commit if available)
			manifestRef := "main"
			if lock != nil {
				for _, la := range lock.Assets {
					if la.ID == asset.ID && la.Source == asset.Source {
						manifestRef = la.Commit
						break
					}
				}
			}

			manifest, err := res.LoadManifest(source, manifestRef)
			if err != nil {
				manifest, err = res.LoadManifest(source, "main")
				if err != nil {
					fmt.Printf("❌ Failed to load manifest for %s: %v\n", asset.Source, err)
					continue
				}
			}

			// Resolve full graph for this top-level asset
			graph, err := res.ResolveGraph(manifest, asset.ID, asset.Version)
			if err != nil {
				fmt.Printf("❌ Failed to resolve graph for %s: %v\n", asset.ID, err)
				continue
			}

			for id, item := range graph {
				key := asset.Source + ":" + id
				projections := make(map[string]string)
				if id == asset.ID {
					projections = asset.Projections
				} else {
					ext := ".md"
					if item.Kind == models.KindSkill {
						ext = ""
					}
					projections["default"] = fmt.Sprintf(".arca/assets/%s/%s%s", asset.Source, id, ext)
				}

				toSync[key] = syncItem{
					ID:          id,
					Source:      source,
					SourceAlias: asset.Source,
					Version:     item.Version,
					Meta:        item.Meta,
					Kind:        item.Kind,
					Projections: projections,
				}
			}
		}

		for _, item := range toSync {
			version := item.Version
			isDir := item.Kind == models.KindSkill
			commitSHA := ""
			assetPath := cache.GetAssetPath(item.SourceAlias, item.ID, version, isDir)
			cacheDir, _ := cache.EnsureDir(item.SourceAlias, item.ID, version)

			if item.Source.Type == models.SourceLocal {
				absPath := filepath.Join(item.Source.Path, item.Meta.Path)
				if isDir {
					commitSHA = "local"
				} else {
					data, err := os.ReadFile(absPath)
					if err != nil {
						fmt.Printf("❌ Failed to read %s: %v\n", item.ID, err)
						continue
					}
					_ = os.WriteFile(assetPath, data, 0644)
					commitSHA = "local"
				}
			} else {
				gitDownloader := downloader.NewGitDownloader()
				ref := item.Meta.Ref
				if ref == "" {
					ref = "main"
				}
				if isDir {
					sha, err := gitDownloader.FetchDirectory(item.Source.URL, item.Meta.Path, ref, cacheDir)
					if err != nil {
						fmt.Printf("❌ Failed to fetch %s: %v\n", item.ID, err)
						continue
					}
					commitSHA = sha
				} else {
					data, sha, err := gitDownloader.FetchFile(item.Source.URL, item.Meta.Path, ref)
					if err != nil {
						fmt.Printf("❌ Failed to fetch %s: %v\n", item.ID, err)
						continue
					}
					_ = os.WriteFile(assetPath, []byte(data), 0644)
					commitSHA = sha
				}
			}

			// Project to all defined locations
			for _, target := range item.Projections {
				_, err = proj.Project(assetPath, target, isDir)
				if err != nil {
					fmt.Printf("❌ Failed to project %s to %s: %v\n", item.ID, target, err)
				}
			}

			// Update Lockfile Entry
			var contentHash string
			if isDir {
				contentHash, _ = hasher.HashDir(assetPath)
			} else {
				contentHash, _ = hasher.HashFile(assetPath)
			}
			locked := models.LockedAsset{
				ID:         item.ID,
				Version:    version,
				Source:     item.SourceAlias,
				Commit:     commitSHA,
				SHA256:     contentHash,
				ResolvedAt: time.Now(),
			}

			found := false
			for i, la := range lock.Assets {
				if la.ID == item.ID && la.Source == item.SourceAlias {
					lock.Assets[i] = locked
					found = true
					break
				}
			}
			if !found {
				lock.Assets = append(lock.Assets, locked)
			}

			fmt.Printf("✅ Synced %s@%s\n", item.ID, version)
		}

		if err := cfgMgr.SaveLockfile(lock); err != nil {
			return err
		}

		fmt.Println("✨ Sync complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
