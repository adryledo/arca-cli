package downloader

import (
	"os"
	"path/filepath"
)

type CacheProvider struct {
	CacheRoot string
}

func NewCacheProvider(customPath string) *CacheProvider {
	if customPath == "" {
		home, _ := os.UserHomeDir()
		customPath = filepath.Join(home, ".arca-cache")
	}
	return &CacheProvider{CacheRoot: customPath}
}

// GetAssetDir returns the path to a specific asset version in the cache.
func (c *CacheProvider) GetAssetDir(sourceAlias, assetID, version string) string {
	return filepath.Join(c.CacheRoot, sourceAlias, assetID, version)
}

// GetAssetPath returns the path to the actual asset file/directory.
func (c *CacheProvider) GetAssetPath(sourceAlias, assetID, version string, isDir bool) string {
	dir := c.GetAssetDir(sourceAlias, assetID, version)
	if isDir {
		return dir
	}
	return filepath.Join(dir, assetID+".md")
}

// EnsureDir makes sure the cache directory for an asset exists.
func (c *CacheProvider) EnsureDir(sourceAlias, assetID, version string) (string, error) {
	dir := c.GetAssetDir(sourceAlias, assetID, version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// Clear removes everything from the cache.
func (c *CacheProvider) Clear() error {
	return os.RemoveAll(c.CacheRoot)
}
