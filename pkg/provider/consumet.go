package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

var downloadURLBase = "https://github.com/cheth/anime-cli/releases/download"

type ConsumetManager struct {
	cmd                  *exec.Cmd
	port                 string
	pathToConsumetBinary string
}

func NewConsumetManager(port string) *ConsumetManager {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	// E.g., ~/.config/animehub/consumet.exe
	binaryName := "consumet"
	if runtime.GOOS == "windows" {
		binaryName = "consumet.exe"
	}
	binPath := filepath.Join(home, ".config", "animehub", binaryName)

	return &ConsumetManager{
		port:                 port,
		pathToConsumetBinary: binPath,
	}
}

func (cm *ConsumetManager) Port() string {
	return cm.port
}

func (cm *ConsumetManager) ensureConsumetBinary() error {
	binDir := filepath.Dir(cm.pathToConsumetBinary)
	if err := os.MkdirAll(binDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if already exists
	if _, err := os.Stat(cm.pathToConsumetBinary); err == nil {
		return nil
	}

	// Determine asset name based on OS
	var assetName string
	switch runtime.GOOS {
	case "windows":
		assetName = "consumet-win.exe"
	case "darwin":
		assetName = "consumet-mac"
	case "linux":
		assetName = "consumet-linux"
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	tag := "v1.0.0"
	downloadURL := fmt.Sprintf("%s/%s/%s", downloadURLBase, tag, assetName)

	// Create file atomically by writing to temporary file first and renaming it
	tmpFilePattern := filepath.Base(cm.pathToConsumetBinary) + "-tmp-*"
	tmpFile, err := os.CreateTemp(binDir, tmpFilePattern)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	// Download the binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download binary from %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download server returned status %s for url %s", resp.Status, downloadURL)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save binary: %w", err)
	}

	// Close the file so we can rename/chmod it
	tmpFile.Close()

	// Set executable permissions
	if err := os.Chmod(tmpPath, 0700); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Atomic rename to final path
	if err := os.Rename(tmpPath, cm.pathToConsumetBinary); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func (cm *ConsumetManager) Start() error {
	originalPort := cm.port
	portInt, err := strconv.Atoi(originalPort)
	if err != nil {
		portInt = 3000
	}

	// Try scanning up to 10 ports starting from the configured port
	for i := 0; i < 10; i++ {
		currentPort := strconv.Itoa(portInt + i)

		if cm.isPortInUse(currentPort) {
			if cm.isConsumetRunning(currentPort) {
				// Reusing the already running/orphaned Consumet process
				cm.port = currentPort
				return nil
			}
			// Port is in use by another application. Continue scanning.
			continue
		}

		// Port is free, start the binary here
		cm.port = currentPort

		if _, err := os.Stat(cm.pathToConsumetBinary); os.IsNotExist(err) {
			if downloadErr := cm.ensureConsumetBinary(); downloadErr != nil {
				return fmt.Errorf("consumet binary not found and download failed: %w", downloadErr)
			}
		}

		cm.cmd = exec.Command(cm.pathToConsumetBinary)
		cm.cmd.Env = append(os.Environ(), "PORT="+cm.port)

		if runtime.GOOS == "windows" {
			cm.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		}

		if err := cm.cmd.Start(); err != nil {
			return fmt.Errorf("initialization failure on provider binary: %w", err)
		}

		return cm.waitForReady()
	}

	return fmt.Errorf("failed to find an available port to start Consumet (tried ports %d to %d)", portInt, portInt+9)
}

func (cm *ConsumetManager) waitForReady() error {
	url := fmt.Sprintf("http://localhost:%s/", cm.port)
	for i := 0; i < 30; i++ {
		resp, err := http.Get(url)
		if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 404) {
			resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("timeout limit reached waiting for child service initialization")
}

func (cm *ConsumetManager) isPortInUse(port string) bool {
	addr := ":" + port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return true // Port is in use
	}
	listener.Close()
	return false // Port is free
}

func (cm *ConsumetManager) isConsumetRunning(port string) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get("http://localhost:" + port + "/")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(body)), "consumet")
}

func (cm *ConsumetManager) Stop() {
	if cm.cmd != nil && cm.cmd.Process != nil {
		_ = cm.cmd.Process.Kill()
	}
}

// ConsumetProvider implements the AnimeProvider interface for Consumet API
type ConsumetProvider struct {
	client  *resty.Client
	baseURL string
}

func NewConsumetProvider(port string) *ConsumetProvider {
	return &ConsumetProvider{
		client:  resty.New(),
		baseURL: fmt.Sprintf("http://localhost:%s", port),
	}
}

func (p *ConsumetProvider) Search(query string) ([]Anime, error) {
	// e.g. /anime/gogoanime/{query}
	resp, err := p.client.R().
		SetQueryParam("page", "1").
		Get(p.baseURL + "/anime/gogoanime/" + query)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.Status())
	}

	var result struct {
		Results []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Image string `json:"image"`
		} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	var animes []Anime
	for _, r := range result.Results {
		animes = append(animes, Anime{
			ID:    r.ID,
			Title: r.Title,
			Image: r.Image,
		})
	}

	return animes, nil
}

func (p *ConsumetProvider) GetEpisodes(animeID string) ([]Episode, error) {
	resp, err := p.client.R().
		Get(p.baseURL + "/anime/gogoanime/info/" + animeID)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.Status())
	}

	var result struct {
		Episodes []struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
		} `json:"episodes"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	var episodes []Episode
	for _, e := range result.Episodes {
		episodes = append(episodes, Episode{
			ID:     e.ID,
			Number: e.Number,
		})
	}

	return episodes, nil
}

func (p *ConsumetProvider) GetStreamURL(episodeID, lang string) (string, error) {
	resp, err := p.client.R().
		SetQueryParam("server", "vidstreaming").
		Get(p.baseURL + "/anime/gogoanime/watch/" + episodeID)

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("API error: %s", resp.Status())
	}

	var result struct {
		Sources []struct {
			URL     string `json:"url"`
			Quality string `json:"quality"`
		} `json:"sources"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}

	for _, src := range result.Sources {
		if src.Quality == "1080p" || src.Quality == "auto" || src.Quality == "default" {
			return src.URL, nil
		}
	}

	if len(result.Sources) > 0 {
		return result.Sources[0].URL, nil
	}

	return "", errors.New("no stream sources found")
}
