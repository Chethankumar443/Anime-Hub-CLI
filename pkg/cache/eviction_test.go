package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEvictCache(t *testing.T) {
	// Create a temp cache directory
	tmpDir, err := os.MkdirTemp("", "anime-cli-cache-test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create three dummy files with different sizes and mod times
	// File 1: 100 bytes, oldest
	file1 := filepath.Join(tmpDir, "file1.png")
	if err := os.WriteFile(file1, make([]byte, 100), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(file1, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set mod time for file1: %v", err)
	}

	// File 2: 150 bytes, medium age
	file2 := filepath.Join(tmpDir, "file2.png")
	if err := os.WriteFile(file2, make([]byte, 150), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	medTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(file2, medTime, medTime); err != nil {
		t.Fatalf("failed to set mod time for file2: %v", err)
	}

	// File 3: 200 bytes, newest
	file3 := filepath.Join(tmpDir, "file3.png")
	if err := os.WriteFile(file3, make([]byte, 200), 0644); err != nil {
		t.Fatalf("failed to write file3: %v", err)
	}
	newTime := time.Now()
	if err := os.Chtimes(file3, newTime, newTime); err != nil {
		t.Fatalf("failed to set mod time for file3: %v", err)
	}

	// Total size of cache is 450 bytes.
	// We run eviction with a limit of 350 bytes.
	// It should evict file1 (100 bytes, oldest), leaving file2 and file3 (total 350 bytes).
	err = EvictCache(tmpDir, 350)
	if err != nil {
		t.Fatalf("eviction failed: %v", err)
	}

	// Check file1 has been evicted
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Errorf("expected file1 to be evicted (deleted)")
	}

	// Check file2 and file3 still exist
	if _, err := os.Stat(file2); err != nil {
		t.Errorf("expected file2 to exist, got error: %v", err)
	}
	if _, err := os.Stat(file3); err != nil {
		t.Errorf("expected file3 to exist, got error: %v", err)
	}

	// Now we run eviction with a limit of 220 bytes.
	// It should evict file2 (150 bytes, oldest now), leaving only file3 (200 bytes).
	err = EvictCache(tmpDir, 220)
	if err != nil {
		t.Fatalf("second eviction failed: %v", err)
	}

	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Errorf("expected file2 to be evicted in second pass")
	}
	if _, err := os.Stat(file3); err != nil {
		t.Errorf("expected file3 to exist in second pass, got error: %v", err)
	}
}
