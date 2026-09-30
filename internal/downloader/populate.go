package downloader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adryledo/arca-cli/internal/fsutil"
)

// PopulateDir replaces the cache directory of an asset version with an exact
// copy of a local source directory. It is the local-source counterpart of
// GitDownloader.FetchDirectory: after it returns, hashing, the lockfile and
// projections all see real content instead of an empty directory.
//
// Existing cached content is removed first so files deleted upstream do not
// survive a re-install.
func (c *CacheProvider) PopulateDir(sourceAlias, assetID, version, srcDir string) error {
	dst := c.GetAssetDir(sourceAlias, assetID, version)

	srcAbs, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("resolve source dir %s: %w", srcDir, err)
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve cache dir %s: %w", dst, err)
	}

	// Refusing to clear the source is a destructive-operation guard, not
	// theoretical: the cache and the source are both plain directories.
	if srcAbs == dstAbs {
		return fmt.Errorf("cache dir equals source dir %s; refusing to overwrite it", srcAbs)
	}

	if err := os.RemoveAll(dstAbs); err != nil {
		return fmt.Errorf("clear cache dir %s: %w", dstAbs, err)
	}
	if err := fsutil.CopyDir(srcAbs, dstAbs); err != nil {
		return fmt.Errorf("copy %s into cache: %w", srcAbs, err)
	}
	return nil
}
