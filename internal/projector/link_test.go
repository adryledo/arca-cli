package projector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProject_RejectsMissingCacheEntry(t *testing.T) {
	ws := t.TempDir()
	p := New(ws)
	missing := filepath.Join(t.TempDir(), "gone.md")

	if _, err := p.Project(missing, ".arca/instructions/gone.md", false); err == nil {
		t.Fatal("expected an error when the cache entry does not exist, got nil")
	}
	if _, err := os.Lstat(filepath.Join(ws, ".arca", "instructions", "gone.md")); !os.IsNotExist(err) {
		t.Errorf("no projection should have been created, lstat err=%v", err)
	}
}

func TestProject_RejectsEmptyCacheDir(t *testing.T) {
	ws := t.TempDir()
	p := New(ws)
	emptyCache := filepath.Join(t.TempDir(), "empty-skill")
	if err := os.MkdirAll(emptyCache, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := p.Project(emptyCache, ".arca/skills/empty-skill", true)
	if err == nil {
		t.Fatal("expected an error when the cache directory is empty, got nil")
	}
	if _, lerr := os.Lstat(filepath.Join(ws, ".arca", "skills", "empty-skill")); !os.IsNotExist(lerr) {
		t.Errorf("no projection should have been created, lstat err=%v", lerr)
	}
}

func TestProject_DirectoryProjectionIsUsable(t *testing.T) {
	ws := t.TempDir()
	cache := t.TempDir()
	p := New(ws)

	if err := os.WriteFile(filepath.Join(cache, "SKILL.md"), []byte("skill body"), 0644); err != nil {
		t.Fatal(err)
	}

	absTarget, err := p.Project(cache, ".arca/skills/my-skill", true)
	if err != nil {
		t.Fatalf("Project failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(absTarget, "SKILL.md"))
	if err != nil {
		t.Fatalf("cannot read through projection: %v", err)
	}
	if string(got) != "skill body" {
		t.Errorf("expected %q, got %q", "skill body", string(got))
	}

	if err := p.RemoveProjection(".arca/skills/my-skill"); err != nil {
		t.Fatalf("RemoveProjection: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cache, "SKILL.md")); err != nil {
		t.Errorf("RemoveProjection deleted the cache content instead of only the link: %v", err)
	}
}

func TestProject_FileProjectionIsUsable(t *testing.T) {
	ws := t.TempDir()
	cache := t.TempDir()
	p := New(ws)

	cachedFile := filepath.Join(cache, "test.md")
	if err := os.WriteFile(cachedFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	absTarget, err := p.Project(cachedFile, ".arca/instructions/test.md", false)
	if err != nil {
		t.Fatalf("Project failed: %v", err)
	}
	got, err := os.ReadFile(absTarget)
	if err != nil {
		t.Fatalf("cannot read projected file: %v", err)
	}
	if string(got) != "test content" {
		t.Errorf("expected %q, got %q", "test content", string(got))
	}
	if err := p.RemoveProjection(".arca/instructions/test.md"); err != nil {
		t.Fatalf("RemoveProjection: %v", err)
	}
}
