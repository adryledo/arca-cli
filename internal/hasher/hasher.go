package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HashString computes a SHA-256 hash of a string after LF normalization.
func HashString(content string) string {
	normalized := NormalizeLF(content)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// NormalizeLF converts all CRLF to LF.
func NormalizeLF(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}

// HashFile computes a SHA-256 hash of a file's content after LF normalization.
func HashFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return HashString(string(content)), nil
}

// HashDir computes a deterministic SHA-256 hash of a directory.
// It includes relative paths and LF-normalized file contents.
func HashDir(dirPath string) (string, error) {
	hash := sha256.New()
	var paths []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(paths)

	for _, path := range paths {
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return "", err
		}
		// Use forward slashes for relative paths in hash to ensure consistency across OS
		normalizedRelPath := filepath.ToSlash(relPath)

		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		normalizedContent := NormalizeLF(string(content))

		hash.Write([]byte(normalizedRelPath))
		hash.Write([]byte(normalizedContent))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// NewHashReader returns a reader that computes SHA-256 while reading.
// Note: This does NOT do LF normalization automatically because it's a stream.
// Use for binary validation if needed, but primary ARCA assets use HashString/HashFile.
func NewHashReader(r io.Reader) (io.Reader, func() string) {
	h := sha256.New()
	tee := io.TeeReader(r, h)
	return tee, func() string {
		return hex.EncodeToString(h.Sum(nil))
	}
}
