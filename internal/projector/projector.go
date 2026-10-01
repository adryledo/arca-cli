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

// Project creates a projection at targetPath pointing at cachedPath.
// targetPath is relative to WorkspaceRoot.
func (p *Projector) Project(cachedPath string, targetPath string, isDir bool) (string, error) {
	// A projection must always point at real cached content. Without this
	// guard a missing or empty cache entry silently becomes a broken
	// projected skill (issue #9).
	absCached, err := filepath.Abs(cachedPath)
	if err != nil {
		return "", fmt.Errorf("resolve cache path %s: %w", cachedPath, err)
	}
	cachedPath = absCached

	info, err := os.Stat(cachedPath)
	if err != nil {
		return "", fmt.Errorf("cannot project %s: cache entry unavailable: %w", cachedPath, err)
	}
	if isDir {
		if !info.IsDir() {
			return "", fmt.Errorf("cannot project %s: expected a directory cache entry", cachedPath)
		}
		entries, err := os.ReadDir(cachedPath)
		if err != nil {
			return "", fmt.Errorf("cannot read cache dir %s: %w", cachedPath, err)
		}
		if len(entries) == 0 {
			return "", fmt.Errorf("cannot project %s: cache directory is empty", cachedPath)
		}
	}

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

	// Platform-specific projection: symlink on POSIX, junction/hard link on
	// Windows (see link_unix.go / link_windows.go).
	if err := createLink(cachedPath, absTarget, isDir); err != nil {
		return "", fmt.Errorf("failed to project %s -> %s: %w", absTarget, cachedPath, err)
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
