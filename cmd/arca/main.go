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

var rootCmd = &cobra.Command{
	Use:   "arca",
	Short: "ARCA - Asset Resolution for AI Assistants",
	Long: `ARCA is a high-performance CLI for managing versioned agentic assets 
(prompts, skills, instructions) from Git-based or local manifests.`,
}

var (
	targetPath string
	projName   string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	installCmd.Flags().StringVarP(&targetPath, "target", "t", "", "Projection target path")
	installCmd.Flags().StringVarP(&projName, "name", "n", "default", "Projection name")
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(syncCmd)
}

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
		manifest, err := res.LoadManifest(cfg.Sources[sourceAlias])
		if err != nil {
			return err
		}

		// 4. Resolve version
		version, meta, err := res.ResolveVersion(manifest, assetID, versionConstraint)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Resolved %s at %s\n", assetID, version)

		// 5. Download/Reference content
		var content string
		var commitSHA string
		if stype == models.SourceLocal {
			absPath := filepath.Join(sourceStr, meta.Path)
			data, err := os.ReadFile(absPath)
			if err != nil {
				return err
			}
			content = string(data)
			commitSHA = "local"
		} else {
			gitDownloader := downloader.NewGitDownloader()
			ref := meta.Ref
			if ref == "" {
				ref = "main"
			}
			data, sha, err := gitDownloader.FetchFile(sourceStr, meta.Path, ref)
			if err != nil {
				return err
			}
			content = data
			commitSHA = sha
		}

		// 6. Cache
		_, err = cache.EnsureDir(sourceAlias, assetID, version)
		if err != nil {
			return err
		}

		isDir := false // Asset resolution for single file or dir based on Kind
		assetPath := cache.GetAssetPath(sourceAlias, assetID, version, isDir)

		err = os.WriteFile(assetPath, []byte(content), 0644)
		if err != nil {
			return err
		}

		// 7. Project
		if targetPath == "" {
			// Fallback to default mapping if not provided
			targetPath = fmt.Sprintf(".arca/assets/%s/%s.md", sourceAlias, assetID)
		}

		_, err = proj.Project(assetPath, targetPath, false)
		if err != nil {
			return err
		}
		fmt.Printf("🚀 Projected to %s\n", targetPath)

		// 8. Update Config Entry
		entry := models.AssetEntry{
			ID:      assetID,
			Source:  sourceAlias,
			Version: version,
			Projections: map[string]string{
				projName: targetPath,
			},
		}
		cfgMgr.AddAsset(cfg, entry)

		if err := cfgMgr.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// 9. Update Lockfile
		lock, err := cfgMgr.LoadLockfile()
		if err != nil {
			return err
		}

		contentHash := hasher.HashString(content)
		locked := models.LockedAsset{
			ID:         assetID,
			Version:    version,
			Source:     sourceAlias,
			Commit:     commitSHA,
			SHA256:     contentHash,
			ResolvedAt: time.Now(),
		}

		// Update or Add to lock
		found := false
		for i, la := range lock.Assets {
			if la.ID == assetID && la.Source == sourceAlias {
				lock.Assets[i] = locked
				found = true
				break
			}
		}
		if !found {
			lock.Assets = append(lock.Assets, locked)
		}

		if err := cfgMgr.SaveLockfile(lock); err != nil {
			return fmt.Errorf("failed to save lockfile: %w", err)
		}

		fmt.Println("✨ Installation complete and persisted.")
		return nil
	},
}

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

		fmt.Println("🔄 Syncing assets...")

		for _, asset := range cfg.Assets {
			source, ok := cfg.Sources[asset.Source]
			if !ok {
				fmt.Printf("⚠️  Source %s not found for asset %s, skipping.\n", asset.Source, asset.ID)
				continue
			}

			// Load manifest for the source
			manifest, err := res.LoadManifest(source)
			if err != nil {
				fmt.Printf("❌ Failed to load manifest for %s: %v\n", asset.Source, err)
				continue
			}

			// Resolve version (honors constraint in config)
			version, meta, err := res.ResolveVersion(manifest, asset.ID, asset.Version)
			if err != nil {
				fmt.Printf("❌ Failed to resolve %s: %v\n", asset.ID, err)
				continue
			}

			// Fetch content
			var content string
			var commitSHA string
			if source.Type == models.SourceLocal {
				absPath := filepath.Join(source.Path, meta.Path)
				data, err := os.ReadFile(absPath)
				if err != nil {
					fmt.Printf("❌ Failed to read %s: %v\n", asset.ID, err)
					continue
				}
				content = string(data)
				commitSHA = "local"
			} else {
				gitDownloader := downloader.NewGitDownloader()
				ref := meta.Ref
				if ref == "" {
					ref = "main"
				}
				data, sha, err := gitDownloader.FetchFile(source.URL, meta.Path, ref)
				if err != nil {
					fmt.Printf("❌ Failed to fetch %s: %v\n", asset.ID, err)
					continue
				}
				content = data
				commitSHA = sha
			}

			// Cache
			_, err = cache.EnsureDir(asset.Source, asset.ID, version)
			if err != nil {
				return err
			}
			isDir := false
			assetPath := cache.GetAssetPath(asset.Source, asset.ID, version, isDir)
			err = os.WriteFile(assetPath, []byte(content), 0644)
			if err != nil {
				return err
			}

			// Project to all defined locations
			for name, target := range asset.Projections {
				_, err = proj.Project(assetPath, target, false)
				if err != nil {
					fmt.Printf("❌ Failed to project %s (%s) to %s: %v\n", asset.ID, name, target, err)
				}
			}

			// Update Lockfile Entry
			contentHash := hasher.HashString(content)
			locked := models.LockedAsset{
				ID:         asset.ID,
				Version:    version,
				Source:     asset.Source,
				Commit:     commitSHA,
				SHA256:     contentHash,
				ResolvedAt: time.Now(),
			}

			// Upsert into lock
			found := false
			for i, la := range lock.Assets {
				if la.ID == asset.ID && la.Source == asset.Source {
					lock.Assets[i] = locked
					found = true
					break
				}
			}
			if !found {
				lock.Assets = append(lock.Assets, locked)
			}

			fmt.Printf("✅ Synced %s@%s\n", asset.ID, version)
		}

		if err := cfgMgr.SaveLockfile(lock); err != nil {
			return err
		}

		fmt.Println("✨ Sync complete.")
		return nil
	},
}
