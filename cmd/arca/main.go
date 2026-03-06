package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arca",
	Short: "ARCA - Asset Resolution for AI Assistants",
	Long: `ARCA is a high-performance CLI for managing versioned agentic assets 
(skills, instructions) from Git-based or local manifests.`,
}

var (
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
}
