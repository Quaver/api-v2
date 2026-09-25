package files

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Quaver/api2/config"
)

const cacheRetention = 90 * 24 * time.Hour

type CachePruneResult struct {
	Files int
	Bytes int64
}

// PruneCache finds cached maps, mapsets, and replays that are unused for 90 days.
func PruneCache(deleteFiles bool) (CachePruneResult, error) {
	var result CachePruneResult
	root := config.Instance.Cache.DataDirectory
	cutoff := time.Now().Add(-cacheRetention)

	for _, name := range []string{"maps", "mapsets", "replays"} {
		directory := filepath.Join(root, name)
		entries, err := os.ReadDir(directory)
		if err != nil {
			return result, fmt.Errorf("read cache directory %s: %w", directory, err)
		}

		for _, entry := range entries {
			path := filepath.Join(directory, entry.Name())
			info, err := entry.Info()
			if err != nil {
				return result, fmt.Errorf("inspect cache file %s: %w", path, err)
			}
			if !info.Mode().IsRegular() || !info.ModTime().Before(cutoff) {
				continue
			}

			if deleteFiles {
				if err := os.Remove(path); err != nil {
					return result, fmt.Errorf("remove cache file %s: %w", path, err)
				}
			}

			result.Files++
			result.Bytes += info.Size()
		}
	}

	return result, nil
}
