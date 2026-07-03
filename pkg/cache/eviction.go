package cache

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

type CacheEntry struct {
	Path     string
	Size     int64
	Accessed time.Time
}

// EvictCache scans the cache folder and deletes the oldest accessed files until total size is under maxBytes.
func EvictCache(dir string, maxBytes int64) error {
	var entries []CacheEntry
	var totalSize int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		entries = append(entries, CacheEntry{
			Path:     path,
			Size:     info.Size(),
			Accessed: info.ModTime(),
		})
		totalSize += info.Size()
		return nil
	})
	if err != nil {
		return err
	}

	if totalSize <= maxBytes {
		return nil
	}

	// Sort oldest accessed first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Accessed.Before(entries[j].Accessed)
	})

	for _, entry := range entries {
		if totalSize <= maxBytes {
			break
		}
		if err := os.Remove(entry.Path); err == nil {
			totalSize -= entry.Size
		}
	}
	return nil
}
