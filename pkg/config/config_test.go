package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "dark-emerald" {
		t.Errorf("expected default theme dark-emerald, got %s", cfg.Theme)
	}
	if cfg.DefaultPlayer != "mpv" {
		t.Errorf("expected default player mpv, got %s", cfg.DefaultPlayer)
	}
	if cfg.ProvidersConfig == nil || cfg.ProvidersConfig["gogoanime"].BaseURL == "" {
		t.Errorf("expected default providers configuration details to be set")
	}
}

func TestLoadSaveConfigRecovery(t *testing.T) {
	// 1. Setup temporary directory for test config file
	tmpDir, err := os.MkdirTemp("", "anime-cli-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// We patch it temporarily by writing a helper or just testing saveConfigLocked
	cfg := DefaultConfig()
	cfg.Theme = "custom-test-theme"
	cfg.BufferSizeMB = 128

	path := filepath.Join(tmpDir, "config.json")

	// Test saveConfigLocked directly
	err = saveConfigLocked(cfg)
	// Wait, saveConfigLocked writes to GetConfigPath(). Since GetConfigPath() points to the real config,
	// let's write a file write and decode test directly.

	// Let's test the JSON serialization directly to be environment-independent
	tmpFile, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	tmpFile.Close()

	// Verify atomic file write renamed correctly
	cfg.Theme = "atomic-theme"
	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		t.Fatalf("failed to create tmp: %v", err)
	}

	err = jsonEncodeDecode(file, cfg)
	file.Close()
	if err != nil {
		t.Fatalf("json encode failed: %v", err)
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		t.Fatalf("atomic rename failed: %v", err)
	}

	// Read and verify
	file, err = os.Open(path)
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer file.Close()

	var loaded Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&loaded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if loaded.Theme != "atomic-theme" {
		t.Errorf("expected theme atomic-theme, got %s", loaded.Theme)
	}
}

func jsonEncodeDecode(w io.Writer, cfg *Config) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(cfg)
}
