//go:build !windows

package projector

import "os"

// createLink projects oldname at newname using a symlink, the only mechanism
// ARCA needs outside Windows.
func createLink(oldname, newname string, isDir bool) error {
	return os.Symlink(oldname, newname)
}
