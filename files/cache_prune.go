package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	cacheDirectories := []struct {
		name      string
		extension string
	}{
		{"maps", ".qua"},
		{"mapsets", ".qp"},
		{"replays", ".qr"},
	}

	for _, cacheDirectory := range cacheDirectories {
		directory := filepath.Join(root, cacheDirectory.name)
		entries, err := os.ReadDir(directory)
		if err != nil {
			return result, fmt.Errorf("read cache directory %s: %w", directory, err)
		}

		for _, entry := range entries {
			if !isCacheFile(entry.Name(), cacheDirectory.extension) {
				continue
			}

			path := filepath.Join(directory, entry.Name())
			info, err := entry.Info()
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return result, fmt.Errorf("inspect cache file %s: %w", path, err)
			}
			if !info.Mode().IsRegular() || !info.ModTime().Before(cutoff) {
				continue
			}

			if deleteFiles {
				// Skip files that were used or replaced while scanning.
				currentInfo, err := os.Lstat(path)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return result, fmt.Errorf("inspect cache file %s: %w", path, err)
				}
				if !os.SameFile(info, currentInfo) || !currentInfo.ModTime().Before(cutoff) {
					continue
				}
				if err := os.Remove(path); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					return result, fmt.Errorf("remove cache file %s: %w", path, err)
				}
			}

			result.Files++
			result.Bytes += info.Size()
		}
	}

	return result, nil
}

func isCacheFile(name string, extension string) bool {
	if !strings.HasSuffix(name, extension) {
		return false
	}

	id, err := strconv.Atoi(strings.TrimSuffix(name, extension))
	return err == nil && id > 0
}
