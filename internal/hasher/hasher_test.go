package hasher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeLF(t *testing.T) {
	input := "line1\r\nline2\r\nline3"
	expected := "line1\nline2\nline3"
	actual := NormalizeLF(input)
	if actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

func TestHashString(t *testing.T) {
	c1 := "content\r\nwith crlf"
	c2 := "content\nwith crlf"

	h1 := HashString(c1)
	h2 := HashString(c2)

	if h1 != h2 {
		t.Errorf("Hashes should be identical regardless of line endings")
	}
}

func TestHashDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "arca-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a nested file structure
	files := map[string]string{
		"file1.txt":       "hello\r\nworld",
		"subdir/file2.md": "nested content\n",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	h1, err := HashDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Change line endings in one file
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello\nworld"), 0644)

	h2, err := HashDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if h1 != h2 {
		t.Errorf("Directory hashes should be deterministic and LF-normalized")
	}
}
