package downloader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adryledo/arca-cli/internal/auth"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

// GitDownloader handles fetching assets from Git repositories using memory storage
// to avoid local git dependency and ensure portability.
type GitDownloader struct{}

func NewGitDownloader() *GitDownloader {
	return &GitDownloader{}
}

// FetchFile fetches a single file from a Git URL at a specific ref.
func (g *GitDownloader) FetchFile(url, path, ref string) (string, string, error) {
	opts := &git.CloneOptions{
		URL:           url,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + ref),
	}
	if a := auth.GetGitAuth(); a != nil {
		opts.Auth = a
	}

	repo, err := git.Clone(memory.NewStorage(), memfs.New(), opts)
	if err != nil {
		return "", "", fmt.Errorf("failed to clone repo: %w", err)
	}

	head, _ := repo.Head()
	commitSHA := head.Hash().String()

	wt, err := repo.Worktree()
	if err != nil {
		return "", "", err
	}

	f, err := wt.Filesystem.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("file not found in repo: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return "", "", err
	}

	return string(content), commitSHA, nil
}

// FetchDirectory fetches a directory and saves it to a local destination.
func (g *GitDownloader) FetchDirectory(url, repoPath, ref, destDir string) (string, error) {
	opts := &git.CloneOptions{
		URL:   url,
		Depth: 1,
	}
	if a := auth.GetGitAuth(); a != nil {
		opts.Auth = a
	}

	repo, err := git.Clone(memory.NewStorage(), memfs.New(), opts)
	if err != nil {
		return "", err
	}

	head, _ := repo.Head()
	commitSHA := head.Hash().String()

	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	// Read recursive directory
	fis, err := wt.Filesystem.ReadDir(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read repo directory: %w", err)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			// Basic recursion for demo
			_, err = g.FetchDirectory(url, filepath.Join(repoPath, fi.Name()), ref, filepath.Join(destDir, fi.Name()))
			if err != nil {
				return "", err
			}
			continue
		}

		f, err := wt.Filesystem.Open(filepath.Join(repoPath, fi.Name()))
		if err != nil {
			return "", err
		}

		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return "", err
		}

		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(filepath.Join(destDir, fi.Name()), data, 0644)
		if err != nil {
			return "", err
		}
	}

	return commitSHA, nil
}
