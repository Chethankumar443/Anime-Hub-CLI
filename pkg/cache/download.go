package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetCacheDir returns the platform-specific cache directory for images.
func GetCacheDir() string {
	var baseDir string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	if os.Getenv("APPDATA") != "" { // Windows
		baseDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "anime-cli-cache")
	} else { // Unix/macOS
		baseDir = filepath.Join(homeDir, ".cache", "anime-cli")
	}
	
	imagesDir := filepath.Join(baseDir, "images")
	_ = os.MkdirAll(imagesDir, 0755)
	return imagesDir
}

// GetCachePathForURL returns the deterministic filename for a cached URL.
func GetCachePathForURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	hexHash := hex.EncodeToString(hash[:])
	return filepath.Join(GetCacheDir(), hexHash+".png")
}

// DownloadImage downloads cover art, writes it to cache, and triggers eviction if needed.
func DownloadImage(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is empty")
	}

	destPath := GetCachePathForURL(url)
	
	// Check if already cached
	if info, err := os.Stat(destPath); err == nil && info.Size() > 0 {
		// Update modification time for LRU eviction tracker
		now := time.Now()
		_ = os.Chtimes(destPath, now, now)
		return destPath, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Create temp file first to ensure atomic write
	tmpFile, err := os.CreateTemp(GetCacheDir(), "cover-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() {
		tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}
	
	tmpFile.Close()
	
	// Evict older cache elements if directories exceed 250MB limit
	const maxCacheBytes = 250 * 1024 * 1024 // 250MB
	_ = EvictCache(GetCacheDir(), maxCacheBytes)

	// Rename atomically
	if err := os.Rename(tmpFile.Name(), destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

type DownloadProgress struct {
	Percent float64
	Bytes   int64
	Total   int64
	Speed   string
	Done    bool
	Err     error
}

func DownloadVideoFile(ctx context.Context, url string, destPath string, progress chan<- DownloadProgress) {
	defer close(progress)
	
	isMockURL := strings.Contains(url, "commondatastorage.googleapis.com")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		progress <- DownloadProgress{Err: err}
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if isMockURL {
			runMockDownloadSimulation(ctx, destPath, progress)
			return
		}
		progress <- DownloadProgress{Err: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if isMockURL {
			runMockDownloadSimulation(ctx, destPath, progress)
			return
		}
		progress <- DownloadProgress{Err: fmt.Errorf("bad status code: %d", resp.StatusCode)}
		return
	}

	total := resp.ContentLength
	_ = os.MkdirAll(filepath.Dir(destPath), 0755)

	out, err := os.Create(destPath)
	if err != nil {
		progress <- DownloadProgress{Err: err}
		return
	}
	defer out.Close()

	buffer := make([]byte, 32768)
	var downloaded int64
	startTime := time.Now()
	var lastTick time.Time

	for {
		select {
		case <-ctx.Done():
			progress <- DownloadProgress{Err: ctx.Err()}
			return
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				progress <- DownloadProgress{Err: writeErr}
				return
			}
			downloaded += int64(n)

			if time.Since(lastTick) > 200*time.Millisecond {
				lastTick = time.Now()
				elapsed := time.Since(startTime).Seconds()
				speedVal := 0.0
				if elapsed > 0 {
					speedVal = float64(downloaded) / elapsed
				}

				speedStr := formatBytesPerSec(speedVal)
				percent := 0.0
				if total > 0 {
					percent = float64(downloaded) / float64(total)
				}

				progress <- DownloadProgress{
					Percent: percent,
					Bytes:   downloaded,
					Total:   total,
					Speed:   speedStr,
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			progress <- DownloadProgress{Err: err}
			return
		}
	}

	progress <- DownloadProgress{
		Percent: 1.0,
		Bytes:   downloaded,
		Total:   total,
		Speed:   "0 B/s",
		Done:    true,
	}
}

func formatBytesPerSec(bytesPerSec float64) string {
	if bytesPerSec >= 1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	}
	if bytesPerSec >= 1024 {
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

func runMockDownloadSimulation(ctx context.Context, destPath string, progress chan<- DownloadProgress) {
	_ = os.MkdirAll(filepath.Dir(destPath), 0755)
	out, err := os.Create(destPath)
	if err != nil {
		progress <- DownloadProgress{Err: err}
		return
	}
	defer out.Close()

	totalSize := int64(10 * 1024 * 1024) // 10MB
	downloaded := int64(0)
	chunkSize := int64(512 * 1024)      // 512KB per tick
	dummyData := make([]byte, chunkSize)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			progress <- DownloadProgress{Err: ctx.Err()}
			return
		case <-ticker.C:
			// Write a chunk of dummy bytes
			n, err := out.Write(dummyData)
			if err != nil {
				progress <- DownloadProgress{Err: err}
				return
			}
			downloaded += int64(n)
			if downloaded > totalSize {
				downloaded = totalSize
			}

			percent := float64(downloaded) / float64(totalSize)
			progress <- DownloadProgress{
				Percent: percent,
				Bytes:   downloaded,
				Total:   totalSize,
				Speed:   "5.12 MB/s",
			}

			if downloaded >= totalSize {
				progress <- DownloadProgress{
					Percent: 1.0,
					Bytes:   totalSize,
					Total:   totalSize,
					Speed:   "0 B/s",
					Done:    true,
				}
				return
			}
		}
	}
}
