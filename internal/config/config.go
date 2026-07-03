package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type UserPreferences struct {
	DefaultPlayer string `json:"default_player"`
	MaxCacheBytes int64  `json:"max_cache_bytes"`
}

type ConfigManager struct {
	mu       sync.RWMutex
	filePath string
	prefs    UserPreferences
}

func NewConfigManager(filePath string) (*ConfigManager, error) {
	cm := &ConfigManager{
		filePath: filePath,
	}
	if err := cm.Load(); err != nil {
		if os.IsNotExist(err) {
			cm.prefs = UserPreferences{
				DefaultPlayer: "mpv",
				MaxCacheBytes: 250 * 1024 * 1024, // 250MB
			}
			if err := cm.Save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return cm, nil
}

func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	file, err := os.Open(cm.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &cm.prefs)
}

func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := json.MarshalIndent(cm.prefs, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(cm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	tmpFile.Close()

	return os.Rename(tmpFile.Name(), cm.filePath)
}

func (cm *ConfigManager) GetPreferences() UserPreferences {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.prefs
}

func (cm *ConfigManager) SetPreferences(prefs UserPreferences) error {
	cm.mu.Lock()
	cm.prefs = prefs
	cm.mu.Unlock()
	return cm.Save()
}
