package projector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Projector handles mapping cached assets into the workspace via symlinks.
type Projector struct {
	WorkspaceRoot string
}

func New(workspaceRoot string) *Projector {
	return &Projector{WorkspaceRoot: workspaceRoot}
}

// Project creates a symlink from cachedPath to targetPath.
// targetPath is relative to WorkspaceRoot.
func (p *Projector) Project(cachedPath string, targetPath string, isDir bool) (string, error) {
	absTarget := filepath.Join(p.WorkspaceRoot, targetPath)

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(absTarget), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Remove existing if it's there
	if _, err := os.Lstat(absTarget); err == nil {
		if err := os.RemoveAll(absTarget); err != nil {
			return "", fmt.Errorf("failed to remove existing projection: %w", err)
		}
	}

	// Create symlink
	// On Windows, this may require SeCreateSymbolicLinkPrivilege (Developer Mode)
	// or it may fail. We should consider a fallback to hardlink or copy in the future.
	err := os.Symlink(cachedPath, absTarget)
	if err != nil {
		// Fallback for mobile or systems without symlink support: File Copy
		// (Simplified for now, just trying symlink)
		return "", fmt.Errorf("failed to create symlink: %w. Try enabling Developer Mode (Windows)", err)
	}

	// Ensure gitignored
	if err := p.EnsureGitignored(absTarget); err != nil {
		// Non-fatal, just log?
		fmt.Printf("Warning: failed to update .gitignore: %v\n", err)
	}

	return absTarget, nil
}

// EnsureGitignored adds the projected path to the workspace .gitignore.
func (p *Projector) EnsureGitignored(absPath string) error {
	gitignorePath := filepath.Join(p.WorkspaceRoot, ".gitignore")
	relPath, err := filepath.Rel(p.WorkspaceRoot, absPath)
	if err != nil {
		return err
	}
	relPath = filepath.ToSlash(relPath)

	// Read existing
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == relPath {
			return nil // Already ignored
		}
	}

	// Append under ARCA marker
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	marker := "\n# ARCA managed assets\n"
	if !strings.Contains(string(content), strings.TrimSpace(marker)) {
		if _, err := f.WriteString(marker); err != nil {
			return err
		}
	}

	if _, err := f.WriteString(relPath + "\n"); err != nil {
		return err
	}

	return nil
}

// RemoveProjection deletes the projected symlink.
func (p *Projector) RemoveProjection(targetPath string) error {
	absTarget := filepath.Join(p.WorkspaceRoot, targetPath)
	if _, err := os.Lstat(absTarget); err == nil {
		return os.RemoveAll(absTarget)
	}
	return nil
}
