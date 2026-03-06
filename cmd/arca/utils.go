package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adryledo/arca-cli/internal/models"
	"gopkg.in/yaml.v3"
)

func extractDescription(filePath string, kind models.AssetKind) string {
	if kind == models.KindSkill {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return ""
		}
		found := false
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(entry.Name(), "SKILL.md") {
				filePath = filepath.Join(filePath, entry.Name())
				found = true
				break
			}
		}
		if !found {
			return ""
		}
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	content := string(data)
	if strings.HasPrefix(content, "---\n") || strings.HasPrefix(content, "---\r\n") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			var fm struct {
				Description string `yaml:"description"`
			}
			if err := yaml.Unmarshal([]byte(parts[1]), &fm); err == nil {
				return strings.TrimSpace(fm.Description)
			}
		}
	}
	return ""
}
