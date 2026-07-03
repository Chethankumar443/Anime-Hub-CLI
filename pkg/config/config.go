package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type ProviderSettings struct {
	BaseURL    string `json:"base_url"`
	TimeoutSec int    `json:"timeout_sec"`
}

type Config struct {
	Theme             string                      `json:"theme"`
	DefaultPlayer     string                      `json:"default_player"`
	PreferredQuality  string                      `json:"preferred_quality"`
	BufferSizeMB      int                         `json:"buffer_size_mb"`
	SyncplayServerURL string                      `json:"syncplay_server_url"`
	DefaultProvider   string                      `json:"default_provider"`
	ProvidersConfig   map[string]ProviderSettings `json:"providers_config"`
}

var (
	configLock sync.RWMutex
	appConfig  *Config
)

// GetAppDir returns the platform-specific application directory.
func GetAppDir() string {
	var baseDir string
	if home := os.Getenv("APPDATA"); home != "" { // Windows
		baseDir = filepath.Join(home, "anime-cli")
	} else { // Unix/macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		baseDir = filepath.Join(homeDir, ".config", "anime-cli")
	}
	_ = os.MkdirAll(baseDir, 0755)
	return baseDir
}

// GetConfigPath returns the path to config.json
func GetConfigPath() string {
	return filepath.Join(GetAppDir(), "config.json")
}

// LoadConfig reads the configuration file from disk.
func LoadConfig() (*Config, error) {
	configLock.Lock()
	defer configLock.Unlock()

	if appConfig != nil {
		return appConfig, nil
	}

	path := GetConfigPath()
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			appConfig = DefaultConfig()
			_ = saveConfigLocked(appConfig)
			return appConfig, nil
		}
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		if err == io.EOF {
			appConfig = DefaultConfig()
			return appConfig, nil
		}
		return nil, err
	}

	appConfig = &cfg
	return appConfig, nil
}

// SaveConfig commits current settings to storage.
func SaveConfig(cfg *Config) error {
	configLock.Lock()
	defer configLock.Unlock()
	appConfig = cfg
	return saveConfigLocked(cfg)
}

func saveConfigLocked(cfg *Config) error {
	path := GetConfigPath()
	tmpPath := path + ".tmp"

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	file.Close() // Close before rename
	return os.Rename(tmpPath, path)
}

// DefaultConfig builds a default configurations profile.
func DefaultConfig() *Config {
	return &Config{
		Theme:             "dark-emerald",
		DefaultPlayer:     "mpv",
		PreferredQuality:  "1080p",
		BufferSizeMB:      64,
		SyncplayServerURL: "wss://sync.animecli.dev",
		DefaultProvider:   "gogoanime",
		ProvidersConfig: map[string]ProviderSettings{
			"gogoanime": {
				BaseURL:    "https://anitaku.pe",
				TimeoutSec: 10,
			},
		},
	}
}
