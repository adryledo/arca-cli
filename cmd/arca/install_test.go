package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adryledo/arca-cli/internal/config"
	"github.com/adryledo/arca-cli/internal/hasher"
)

const (
	testSkillID    = "my-skill"
	testSkillAlias = "mysource"
	testSkillVer   = "1.0.0"
)

// localSkillFixture builds a local maintainer source at <tmp>/mysource with one
// directory (skill) asset. A fixed basename makes the derived source alias
// deterministic: deriveAlias(".../mysource") == "mysource".
func localSkillFixture(t *testing.T) (srcDir, skillDir string) {
	t.Helper()
	srcDir = filepath.Join(t.TempDir(), testSkillAlias)
	skillDir = filepath.Join(srcDir, "skills", "my-skill")

	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}

	files := map[string]string{
		"arca-manifest.yaml":                "schema: \"1.0\"\nassets:\n    my-skill:\n        kind: skill\n        versions:\n            1.0.0:\n                path: skills/my-skill\n",
		"skills/my-skill/SKILL.md":          "---\ndescription: A real skill\n---\n# My Skill\nbody\n",
		"skills/my-skill/scripts/helper.py": "print('helper')\n",
	}
	for rel, body := range files {
		if err := os.WriteFile(filepath.Join(srcDir, filepath.FromSlash(rel)), []byte(body), 0644); err != nil {
			t.Fatalf("write fixture %s: %v", rel, err)
		}
	}
	return srcDir, skillDir
}

// workspaceEnv points HOME (and USERPROFILE so the test also works on Windows)
// at a throwaway dir so ~/.arca-cache is isolated, and chdirs into an empty
// workspace. Cleanup restores both.
func workspaceEnv(t *testing.T) (ws, home string) {
	t.Helper()
	ws = filepath.Join(t.TempDir(), "ws")
	home = t.TempDir()
	if err := os.MkdirAll(ws, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(ws); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	// reset package-level flag state shared by the cobra commands
	targetPath, projName = "", "default"
	return ws, home
}

func entriesOf(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range ents {
		names = append(names, e.Name())
	}
	return names
}

func TestInstall_LocalDirAsset_PopulatesCacheAndProjects(t *testing.T) {
	srcDir, skillDir := localSkillFixture(t)
	ws, home := workspaceEnv(t)

	if err := installCmd.RunE(installCmd, []string{srcDir, testSkillID, testSkillVer}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	cacheDir := filepath.Join(home, ".arca-cache", testSkillAlias, testSkillID, testSkillVer)
	if _, err := os.Stat(filepath.Join(cacheDir, "SKILL.md")); err != nil {
		t.Fatalf("cache not populated with SKILL.md: %v; cache entries=%v", err, entriesOf(t, cacheDir))
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "scripts", "helper.py")); err != nil {
		t.Fatalf("cache not populated with scripts/helper.py: %v; cache entries=%v", err, entriesOf(t, cacheDir))
	}

	projection := filepath.Join(ws, ".arca", "assets", testSkillAlias, testSkillID)
	got, err := os.ReadFile(filepath.Join(projection, "SKILL.md"))
	if err != nil {
		t.Fatalf("projected skill unreadable: %v (is the projection %s?)", err, projection)
	}
	if string(got) != "---\ndescription: A real skill\n---\n# My Skill\nbody\n" {
		t.Errorf("projected SKILL.md has wrong content:\n%s", string(got))
	}

	// Acceptance criterion: lockfile + hashing operate against cached content.
	wantHash, err := hasher.HashDir(skillDir)
	if err != nil {
		t.Fatalf("hash source skill: %v", err)
	}
	lock, err := config.NewManager(ws).LoadLockfile()
	if err != nil {
		t.Fatalf("load lockfile: %v", err)
	}
	if len(lock.Assets) != 1 {
		t.Fatalf("expected 1 lock entry, got %d", len(lock.Assets))
	}
	if lock.Assets[0].SHA256 != wantHash {
		t.Errorf("lockfile SHA256 mismatch:\n got %s\nwant %s", lock.Assets[0].SHA256, wantHash)
	}
	if lock.Assets[0].Commit != "local" {
		t.Errorf("expected commit %q, got %q", "local", lock.Assets[0].Commit)
	}
}
