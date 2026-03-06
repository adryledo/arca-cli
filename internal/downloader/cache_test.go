package downloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheProvider(t *testing.T) {
	tmpDir := t.TempDir()

	cp := NewCacheProvider(tmpDir)

	if cp.CacheRoot != tmpDir {
		t.Errorf("Expected cache root %s, got %s", tmpDir, cp.CacheRoot)
	}

	alias := "test-repo"
	id := "test-skill"
	version := "1.0.0"

	expectedDir := filepath.Join(tmpDir, alias, id, version)

	if got := cp.GetAssetDir(alias, id, version); got != expectedDir {
		t.Errorf("GetAssetDir() = %v, want %v", got, expectedDir)
	}

	if got := cp.GetAssetPath(alias, id, version, true); got != expectedDir {
		t.Errorf("GetAssetPath(isDir=true) = %v, want %v", got, expectedDir)
	}

	expectedFile := filepath.Join(expectedDir, id+".md")
	if got := cp.GetAssetPath(alias, id, version, false); got != expectedFile {
		t.Errorf("GetAssetPath(isDir=false) = %v, want %v", got, expectedFile)
	}

	// EnsureDir
	dir, err := cp.EnsureDir(alias, id, version)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	if dir != expectedDir {
		t.Errorf("EnsureDir() = %v, want %v", dir, expectedDir)
	}

	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		t.Errorf("EnsureDir did not create directory")
	}

	// Clear
	if err := cp.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	if _, err := os.Stat(cp.CacheRoot); !os.IsNotExist(err) {
		t.Errorf("Clear did not remove cache root")
	}
}
