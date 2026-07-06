package provider

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)


func init() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		port := os.Getenv("PORT")
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "Welcome to Consumet API!"}`))
		})
		err := http.ListenAndServe(":"+port, mux)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func TestConsumetManager_ZombieProcessReuse(t *testing.T) {
	// Start a mock Consumet server on port 13000
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Consumet API!"}`))
	})
	server := &http.Server{Addr: ":13000", Handler: mux}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create manager on port 13000 with a dummy non-existent binary path
	cm := NewConsumetManager("13000")
	cm.pathToConsumetBinary = "non_existent_binary_file"

	// Start should succeed because it detects Consumet is already running and reuses it
	err := cm.Start()
	if err != nil {
		t.Fatalf("expected Start() to succeed by reusing port, got error: %v", err)
	}
	if cm.Port() != "13000" {
		t.Errorf("expected port to be 13000, got: %s", cm.Port())
	}
}

func TestConsumetManager_PortConflictFallback(t *testing.T) {
	// Start a dummy non-Consumet server on port 14000
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("unrelated dummy service"))
	})
	server := &http.Server{Addr: ":14000", Handler: mux}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Set env var to tell the spawned process to act as the mock Consumet server
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	defer os.Unsetenv("GO_WANT_HELPER_PROCESS")

	// Create manager starting at port 14000, pointing to the test binary itself
	cm := NewConsumetManager("14000")
	cm.pathToConsumetBinary = os.Args[0]

	// Start should skip 14000 due to conflict, and start helper process on 14001
	err := cm.Start()
	if err != nil {
		t.Fatalf("expected Start() to succeed with fallback, got error: %v", err)
	}
	defer cm.Stop()

	if cm.Port() != "14001" {
		t.Errorf("expected port fallback to 14001, got: %s", cm.Port())
	}
}

func TestConsumetManager_EnsureConsumetBinary(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "anime-cli-downloader-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Start mock HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock binary content"))
	})
	server := &http.Server{Addr: ":15000", Handler: mux}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Point downloadURLBase to our mock server
	oldDownloadURLBase := downloadURLBase
	downloadURLBase = "http://localhost:15000"
	defer func() { downloadURLBase = oldDownloadURLBase }()

	// Create manager with path inside temp dir
	targetPath := filepath.Join(tmpDir, "consumet-test-bin")
	cm := &ConsumetManager{
		port:                 "15001",
		pathToConsumetBinary: targetPath,
	}

	err = cm.ensureConsumetBinary()
	if err != nil {
		t.Fatalf("expected ensureConsumetBinary to succeed, got: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(data) != "mock binary content" {
		t.Errorf("expected 'mock binary content', got: %s", string(data))
	}

	// Verify permissions
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("failed to stat downloaded file: %v", err)
	}
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		if mode != 0700 {
			t.Errorf("expected file mode 0700, got %o", mode)
		}
	}
}
