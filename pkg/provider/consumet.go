package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

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

func (cm *ConsumetManager) Start() error {
	// Dynamically scan for open ports if default is bound
	if cm.isPortInUse(cm.port) {
		return nil
	}

	if _, err := os.Stat(cm.pathToConsumetBinary); os.IsNotExist(err) {
		// As per documentation, "First Run Downloader" would go here.
		// For now, we return a mock error detailing this requirement.
		return fmt.Errorf("consumet binary not found at %s. Please download and place it there", cm.pathToConsumetBinary)
	}

	cm.cmd = exec.Command(cm.pathToConsumetBinary)
	cm.cmd.Env = append(os.Environ(), "PORT="+cm.port)

	// Detach child processes and mask consoles under Windows platforms
	if runtime.GOOS == "windows" {
		cm.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	if err := cm.cmd.Start(); err != nil {
		return fmt.Errorf("initialization failure on provider binary: %w", err)
	}

	return cm.waitForReady()
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
	// Consumet root or health
	ln, err := http.Get("http://localhost:" + port + "/")
	if err == nil {
		ln.Body.Close()
		return true
	}
	return false
}

func (cm *ConsumetManager) Stop() {
	if cm.cmd != nil && cm.cmd.Process != nil {
		_ = cm.cmd.Process.Kill()
	}
}

// ConsumetProvider implements the AnimeProvider interface for Consumet API
type ConsumetProvider struct {
	client *resty.Client
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
