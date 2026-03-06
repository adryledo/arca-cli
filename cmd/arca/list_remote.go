package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/adryledo/arca-cli/internal/models"
	"github.com/adryledo/arca-cli/internal/resolver"
	"github.com/spf13/cobra"
)

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

		manifest, err := res.LoadManifest(sourceCfg, "main")
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

func init() {
	rootCmd.AddCommand(listRemoteCmd)
}
