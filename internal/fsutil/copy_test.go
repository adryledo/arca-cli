package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir_CopiesNestedContents(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "cache")

	if err := os.MkdirAll(filepath.Join(src, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "scripts", "helper.py"), []byte("print(1)"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir returned error: %v", err)
	}

	for _, rel := range []string{"SKILL.md", "scripts/helper.py"} {
		got, err := os.ReadFile(filepath.Join(dst, rel))
		if err != nil {
			t.Fatalf("copied file %s unreadable: %v", rel, err)
		}
		if len(got) == 0 {
			t.Errorf("copied file %s is empty", rel)
		}
	}

	// permission bits survive for executables
	info, err := os.Stat(filepath.Join(dst, "scripts", "helper.py"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Errorf("expected executable bit preserved, got %v", info.Mode().Perm())
	}
}

func TestCopyDir_MissingSource(t *testing.T) {
	err := CopyDir(filepath.Join(t.TempDir(), "nope"), t.TempDir()+"/dst")
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestCopyDir_FileSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "a.md")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CopyDir(src, filepath.Join(t.TempDir(), "dst")); err == nil {
		t.Fatal("expected error when source is a file, got nil")
	}
}

func TestCopyFile_Overwrites(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.md")
	dst := filepath.Join(dir, "dst.md")
	if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old-old-old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("expected %q, got %q", "new", string(got))
	}
}
