package downloader

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestGitRepo(t *testing.T) string {
	t.Helper()

	repoDir := t.TempDir()

	// Initialize git
	cmdInit := exec.Command("git", "init")
	cmdInit.Dir = repoDir
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create test file
	testFile := filepath.Join(repoDir, "test.md")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create test directory containing a file
	testDir := filepath.Join(repoDir, "test-skill")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	testDirFile := filepath.Join(testDir, "SKILL.md")
	if err := os.WriteFile(testDirFile, []byte("skill contents"), 0644); err != nil {
		t.Fatalf("failed to write skill file : %v", err)
	}

	// Commit
	cmdAdd := exec.Command("git", "add", ".")
	cmdAdd.Dir = repoDir
	if err := cmdAdd.Run(); err != nil {
		t.Fatalf("failed to 'git add': %v", err)
	}

	// Provide author name/email so commit succeeds in clean CI environments
	cmdConfigName := exec.Command("git", "config", "user.name", "Test User")
	cmdConfigName.Dir = repoDir
	_ = cmdConfigName.Run()
	cmdConfigEmail := exec.Command("git", "config", "user.email", "test@example.com")
	cmdConfigEmail.Dir = repoDir
	_ = cmdConfigEmail.Run()

	cmdCommit := exec.Command("git", "commit", "-m", "Initial commit")
	cmdCommit.Dir = repoDir
	if err := cmdCommit.Run(); err != nil {
		t.Fatalf("failed to 'git commit': %v", err)
	}

	return repoDir
}

func TestGitDownloader_FetchFile(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	// file:// protocol requires abs path
	repoURL := "file://" + filepath.ToSlash(repoDir)

	dl := NewGitDownloader()

	content, sha, err := dl.FetchFile(repoURL, "test.md", "main")
	// If standard branch is 'master' (git older versions), handle fallback
	if err != nil {
		content, sha, err = dl.FetchFile(repoURL, "test.md", "master")
	}

	if err != nil {
		t.Fatalf("FetchFile failed: %v", err)
	}

	if content != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", content)
	}
	if sha == "" {
		t.Errorf("Expected commit SHA, got empty string")
	}
}

func TestGitDownloader_FetchDirectory(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	repoURL := "file://" + filepath.ToSlash(repoDir)

	destDir := t.TempDir()
	dl := NewGitDownloader()

	sha, err := dl.FetchDirectory(repoURL, "test-skill", "main", destDir)
	// fallback for branch name master
	if err != nil {
		sha, err = dl.FetchDirectory(repoURL, "test-skill", "master", destDir)
	}

	if err != nil {
		t.Fatalf("FetchDirectory failed: %v", err)
	}

	if sha == "" {
		t.Errorf("Expected commit SHA, got empty string")
	}

	// verify file exists
	content, err := os.ReadFile(filepath.Join(destDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed to read fetched file: %v", err)
	}
	if string(content) != "skill contents" {
		t.Errorf("Expected 'skill contents', got '%s'", string(content))
	}
}
