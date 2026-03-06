package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adryledo/arca-cli/internal/models"
)

func TestExtractDescription(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Instructions - standard markdown frontmatter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "instruction.md")
		content := `---
description: This is a test instruction
---
# Header
Test content`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %v", err)
		}

		desc := extractDescription(file, models.KindInstruction)
		if desc != "This is a test instruction" {
			t.Errorf("Expected 'This is a test instruction', got '%s'", desc)
		}
	})

	t.Run("Instructions - no frontmatter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "instruction2.md")
		content := `# Header
Test content`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %v", err)
		}

		desc := extractDescription(file, models.KindInstruction)
		if desc != "" {
			t.Errorf("Expected empty string, got '%s'", desc)
		}
	})

	t.Run("Skill - SKILL.md in dir case insensitive", func(t *testing.T) {
		skillDir := filepath.Join(tmpDir, "my-skill")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		file := filepath.Join(skillDir, "sKiLl.Md")
		content := `---
description: A powerful skill
---
# Header`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %v", err)
		}

		desc := extractDescription(skillDir, models.KindSkill)
		if desc != "A powerful skill" {
			t.Errorf("Expected 'A powerful skill', got '%s'", desc)
		}
	})

	t.Run("Skill - no SKILL.md", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty-skill")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		desc := extractDescription(emptyDir, models.KindSkill)
		if desc != "" {
			t.Errorf("Expected empty string, got '%s'", desc)
		}
	})
}
