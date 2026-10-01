package projector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjector_Project(t *testing.T) {
	wsDir := t.TempDir()
	cacheDir := t.TempDir()

	p := New(wsDir)

	// Create dummy cached file
	cachedFile := filepath.Join(cacheDir, "test.md")
	if err := os.WriteFile(cachedFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create cached file: %v", err)
	}

	targetRelPath := ".arca/instructions/test.md"
	absTarget, err := p.Project(cachedFile, targetRelPath, false)
	if err != nil {
		t.Fatalf("Project failed: %v", err)
	}

	content, err := os.ReadFile(absTarget)
	if err != nil {
		t.Fatalf("failed to read projected file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}

	// Verify EnsureGitignored worked during Project()
	gitignorePath := filepath.Join(wsDir, ".gitignore")
	gitignoreContent, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if !strings.Contains(string(gitignoreContent), ".arca/instructions/test.md") {
		t.Errorf("Expected .gitignore to contain projected file path. got: %s", string(gitignoreContent))
	}

	// Test RemoveProjection
	if err := p.RemoveProjection(targetRelPath); err != nil {
		t.Fatalf("Failed to remove projection: %v", err)
	}

	if _, err := os.Stat(absTarget); !os.IsNotExist(err) {
		t.Errorf("Expected file to be deleted")
	}
}

func TestEnsureGitignored_ExistingContent(t *testing.T) {
	wsDir := t.TempDir()
	p := New(wsDir)

	gitignorePath := filepath.Join(wsDir, ".gitignore")
	initialContent := "node_modules/\n.env\n"
	if err := os.WriteFile(gitignorePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to init gitignore: %v", err)
	}

	err := p.EnsureGitignored(filepath.Join(wsDir, "my-path", "test.txt"))
	if err != nil {
		t.Fatalf("EnsureGitignored failed: %v", err)
	}

	content, _ := os.ReadFile(gitignorePath)
	expectedSub := "\n# ARCA managed assets\nmy-path/test.txt\n"

	if !strings.Contains(string(content), expectedSub) {
		t.Errorf("Expected .gitignore to contain updated section. Got: %s", string(content))
	}
}
