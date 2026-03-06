package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/adryledo/arca-cli/internal/config"
	"github.com/adryledo/arca-cli/internal/models"
	"github.com/spf13/cobra"
)

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

			fmt.Printf("%s %s (%s) from %s\n", status, asset.ID, asset.Kind, asset.Source)
			fmt.Printf("   Version: %s\n", versionInfo)
			for name, path := range asset.Projections {
				fmt.Printf("   🔗 %s -> %s\n", name, path)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
