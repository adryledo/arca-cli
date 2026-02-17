package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/adryledo/arca-cli/internal/config"
	"github.com/adryledo/arca-cli/internal/downloader"
	"github.com/adryledo/arca-cli/internal/hasher"
	"github.com/adryledo/arca-cli/internal/models"
	"github.com/adryledo/arca-cli/internal/projector"
	"github.com/adryledo/arca-cli/internal/resolver"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	jsonOutput bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	installCmd.Flags().StringVarP(&targetPath, "target", "t", "", "Projection target path")
	installCmd.Flags().StringVarP(&projName, "name", "n", "default", "Projection name")
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(listRemoteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(publishCmd)
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
		isDir := manifest.Assets[assetID].Kind == models.KindSkill
		commitSHA := ""
		assetPath := cache.GetAssetPath(sourceAlias, assetID, version, isDir)
		cacheDir, _ := cache.EnsureDir(sourceAlias, assetID, version)

		if stype == models.SourceLocal {
			absPath := filepath.Join(sourceStr, meta.Path)
			if isDir {
				// Copy directory
				// Simplified: just assuming it works for now
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
			ref := meta.Ref
			if ref == "" {
				ref = "main"
			}
			if isDir {
				sha, err := gitDownloader.FetchDirectory(sourceStr, meta.Path, ref, cacheDir)
				if err != nil {
					return err
				}
				commitSHA = sha
			} else {
				data, sha, err := gitDownloader.FetchFile(sourceStr, meta.Path, ref)
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

		// 6. Project
		if targetPath == "" {
			ext := ".md"
			if isDir {
				ext = ""
			}
			targetPath = fmt.Sprintf(".arca/assets/%s/%s%s", sourceAlias, assetID, ext)
		}

		_, err = proj.Project(assetPath, targetPath, isDir)
		if err != nil {
			return err
		}
		fmt.Printf("🚀 Projected to %s\n", targetPath)

		// 7. Update Config Entry

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

		// 8. Update Lockfile
		lock, err := cfgMgr.LoadLockfile()
		if err != nil {
			return err
		}

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

var listRemoteCmd = &cobra.Command{
	Use:   "list-remote [url|path]",
	Short: "List assets available in a source manifest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceStr := args[0]
		cwd, _ := os.Getwd()
		res := resolver.New(cwd)

		stype := models.SourceGit
		if info, err := os.Stat(sourceStr); err == nil && info.IsDir() {
			stype = models.SourceLocal
		}

		sourceCfg := models.SourceConfig{
			Type: stype,
			URL:  sourceStr,
			Path: sourceStr,
		}

		manifest, err := res.LoadManifest(sourceCfg)
		if err != nil {
			return err
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(manifest, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nAssets in %s:\n", sourceStr)
		fmt.Println(strings.Repeat("-", 40))

		for id, asset := range manifest.Assets {
			fmt.Printf("📦 %s (%s)\n", id, asset.Kind)
			fmt.Printf("   📝 %s\n", asset.Description)
			fmt.Print("   📌 Versions: ")
			versions := []string{}
			for v := range asset.Versions {
				versions = append(versions, v)
			}
			sort.Strings(versions)
			fmt.Println(strings.Join(versions, ", "))
			fmt.Println()
		}

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

			// 5. Fetch content
			isDir := manifest.Assets[asset.ID].Kind == models.KindSkill
			commitSHA := ""
			assetPath := cache.GetAssetPath(asset.Source, asset.ID, version, isDir)
			cacheDir, _ := cache.EnsureDir(asset.Source, asset.ID, version)

			if source.Type == models.SourceLocal {
				absPath := filepath.Join(source.Path, meta.Path)
				if isDir {
					// Directory logic
					commitSHA = "local"
				} else {
					data, err := os.ReadFile(absPath)
					if err != nil {
						fmt.Printf("❌ Failed to read %s: %v\n", asset.ID, err)
						continue
					}
					_ = os.WriteFile(assetPath, data, 0644)
					commitSHA = "local"
				}
			} else {
				gitDownloader := downloader.NewGitDownloader()
				ref := meta.Ref
				if ref == "" {
					ref = "main"
				}
				if isDir {
					sha, err := gitDownloader.FetchDirectory(source.URL, meta.Path, ref, cacheDir)
					if err != nil {
						fmt.Printf("❌ Failed to fetch %s: %v\n", asset.ID, err)
						continue
					}
					commitSHA = sha
				} else {
					data, sha, err := gitDownloader.FetchFile(source.URL, meta.Path, ref)
					if err != nil {
						fmt.Printf("❌ Failed to fetch %s: %v\n", asset.ID, err)
						continue
					}
					_ = os.WriteFile(assetPath, []byte(data), 0644)
					commitSHA = sha
				}
			}

			// 6. Project to all defined locations
			for name, target := range asset.Projections {
				_, err = proj.Project(assetPath, target, isDir)
				if err != nil {
					fmt.Printf("❌ Failed to project %s (%s) to %s: %v\n", asset.ID, name, target, err)
				}
			}

			// 7. Update Lockfile Entry
			var contentHash string
			if isDir {
				contentHash, _ = hasher.HashDir(assetPath)
			} else {
				contentHash, _ = hasher.HashFile(assetPath)
			}
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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed assets in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		cfgMgr := config.NewManager(cwd)

		cfg, err := cfgMgr.LoadConfig()
		if err != nil {
			return err
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(cfg.Assets, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(cfg.Assets) == 0 {
			fmt.Println("No assets installed.")
			return nil
		}

		lock, _ := cfgMgr.LoadLockfile()
		lockedMap := make(map[string]models.LockedAsset)
		if lock != nil {
			for _, la := range lock.Assets {
				lockedMap[la.Source+":"+la.ID] = la
			}
		}

		fmt.Printf("📦 Installed Assets (%d):\n", len(cfg.Assets))
		fmt.Println(strings.Repeat("-", 60))

		for _, asset := range cfg.Assets {
			status := "🟢"
			locked, ok := lockedMap[asset.Source+":"+asset.ID]
			versionInfo := asset.Version
			if ok {
				versionInfo = fmt.Sprintf("%s (locked at %s)", asset.Version, locked.Version)
			} else {
				status = "🟡 (unlocked)"
			}

			fmt.Printf("%s %s from %s\n", status, asset.ID, asset.Source)
			fmt.Printf("   Version: %s\n", versionInfo)
			for name, path := range asset.Projections {
				fmt.Printf("   🔗 %s -> %s\n", name, path)
			}
			fmt.Println()
		}

		return nil
	},
}

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
		if kind != models.KindPrompt && kind != models.KindSkill && kind != models.KindInstruction {
			return fmt.Errorf("invalid asset kind: %s. Use 'prompt', 'skill', or 'instruction'", kindStr)
		}

		asset, ok := manifest.Assets[assetID]
		if !ok {
			asset = models.ManifestAsset{
				Kind:        kind,
				Description: "Added via arca publish",
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
