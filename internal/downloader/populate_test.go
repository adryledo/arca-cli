package downloader

import (
	"os"
	"path/filepath"
	"testing"
)

func skillFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	src := filepath.Join(root, "mysource")
	if err := os.MkdirAll(filepath.Join(src, "skills", "my-skill", "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"skills/my-skill/SKILL.md":          "---\ndescription: A real skill\n---\n# My Skill\nbody\n",
		"skills/my-skill/scripts/helper.py": "print('helper')\n",
	}
	for rel, body := range files {
		if err := os.WriteFile(filepath.Join(src, filepath.FromSlash(rel)), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return src
}

func TestPopulateDir_CopiesLocalSkillIntoCache(t *testing.T) {
	cache := NewCacheProvider(filepath.Join(t.TempDir(), "cache"))
	src := skillFixture(t)

	if err := cache.PopulateDir("mysource", "my-skill", "1.0.0", filepath.Join(src, "skills", "my-skill")); err != nil {
		t.Fatalf("PopulateDir: %v", err)
	}

	cacheDir := cache.GetAssetDir("mysource", "my-skill", "1.0.0")
	if _, err := os.Stat(filepath.Join(cacheDir, "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md missing from cache: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "scripts", "helper.py")); err != nil {
		t.Fatalf("scripts/helper.py missing from cache: %v", err)
	}
}

func TestPopulateDir_ReplacesStaleContent(t *testing.T) {
	cache := NewCacheProvider(filepath.Join(t.TempDir(), "cache"))
	src := skillFixture(t)

	dst := cache.GetAssetDir("mysource", "my-skill", "1.0.0")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(dst, "STALE.md")
	if err := os.WriteFile(stale, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := cache.PopulateDir("mysource", "my-skill", "1.0.0", filepath.Join(src, "skills", "my-skill")); err != nil {
		t.Fatalf("PopulateDir: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale file survived the refresh: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "SKILL.md")); err != nil {
		t.Errorf("fresh content missing: %v", err)
	}
}

func TestPopulateDir_MissingSource(t *testing.T) {
	cache := NewCacheProvider(filepath.Join(t.TempDir(), "cache"))
	if err := cache.PopulateDir("mysource", "my-skill", "1.0.0", filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing source dir, got nil")
	}
}

func TestPopulateDir_RefusesWhenSourceEqualsCache(t *testing.T) {
	cache := NewCacheProvider(filepath.Join(t.TempDir(), "cache"))
	dst := cache.GetAssetDir("mysource", "my-skill", "1.0.0")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "SKILL.md"), []byte("keep me"), 0644); err != nil {
		t.Fatal(err)
	}

	err := cache.PopulateDir("mysource", "my-skill", "1.0.0", dst)
	if err == nil {
		t.Fatal("expected an error when source == cache dir, got nil")
	}
	// the guard must fire *before* the destructive RemoveAll
	if _, statErr := os.Stat(filepath.Join(dst, "SKILL.md")); statErr != nil {
		t.Fatalf("source content was deleted by the guard path: %v", statErr)
	}
}
