//go:build windows

package projector

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// createLink projects oldname at newname without os.Symlink, which on Windows
// requires Developer Mode, administrator rights or SeCreateSymbolicLinkPrivilege.
//
//   - directories: junction via `mklink /J` — no elevation required, works for
//     local paths (the ARCA cache is always local)
//   - files: hard link via os.Link (the Go equivalent of `mklink /H`); if the
//     link cannot be created (different volume, unsupported filesystem) the
//     file is copied instead
func createLink(oldname, newname string, isDir bool) error {
	if isDir {
		out, err := exec.Command("cmd", "/c", "mklink", "/J", newname, oldname).CombinedOutput()
		if err != nil {
			return fmt.Errorf("mklink /J failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}

	if err := os.Link(oldname, newname); err == nil {
		return nil
	}
	return copyForProjection(oldname, newname)
}

// copyForProjection is the hard-link fallback for file assets.
func copyForProjection(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
