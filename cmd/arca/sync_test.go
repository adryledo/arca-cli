package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSync_RepopulatesLocalDirAsset(t *testing.T) {
	srcDir, _ := localSkillFixture(t)
	ws, home := workspaceEnv(t)

	if err := installCmd.RunE(installCmd, []string{srcDir, testSkillID, testSkillVer}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Simulate drift: a fresh clone has the config but neither cache nor projection.
	cacheDir := filepath.Join(home, ".arca-cache", testSkillAlias, testSkillID, testSkillVer)
	projection := filepath.Join(ws, ".arca", "assets", testSkillAlias, testSkillID)
	if err := os.RemoveAll(cacheDir); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(projection); err != nil {
		t.Fatal(err)
	}

	if err := syncCmd.RunE(syncCmd, []string{}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(cacheDir, "SKILL.md")); err != nil {
		t.Fatalf("sync did not repopulate the cache: %v; entries=%v", err, entriesOf(t, cacheDir))
	}
	got, err := os.ReadFile(filepath.Join(projection, "SKILL.md"))
	if err != nil {
		t.Fatalf("sync did not restore a usable projection: %v", err)
	}
	if len(got) == 0 {
		t.Error("restored projection is empty")
	}
}
