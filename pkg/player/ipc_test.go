package player

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestIPCClientConnectionAndCommands(t *testing.T) {
	// Create mock socket server
	var listener net.Listener
	var err error
	var socketPath string

	if runtime.GOOS == "windows" {
		// Use a local TCP listener to mock Windows Named Pipe behavior for the test
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to listen on TCP: %v", err)
		}
		socketPath = listener.Addr().String()
	} else {
		// Unix socket
		tmpDir, err := os.MkdirTemp("", "anime-cli-ipc-test")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		socketPath = filepath.Join(tmpDir, "test-mpv.sock")
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			t.Fatalf("failed to listen on Unix socket: %v", err)
		}
	}
	defer listener.Close()

	// Spawn mock player socket handler
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read and reply loop
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Bytes()
			var req map[string]interface{}
			if err := json.Unmarshal(line, &req); err != nil {
				continue
			}

			reqID := req["request_id"].(float64)
			cmd := req["command"].([]interface{})

			var resp []byte
			if cmd[0] == "get_property" && cmd[1] == "time-pos" {
				// Return elapsed time: 100.5 seconds
				resp, _ = json.Marshal(map[string]interface{}{
					"data":       100.5,
					"error":      "success",
					"request_id": reqID,
				})
			} else {
				resp, _ = json.Marshal(map[string]interface{}{
					"error":      "success",
					"request_id": reqID,
				})
			}

			_, _ = conn.Write(append(resp, '\n'))
		}
		_ = scanner.Err()
	}()

	// Initialize TUI IPC client
	client := NewIPCClient(socketPath)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("client connection failed: %v", err)
	}
	defer client.Close()

	client.StartMonitoring(ctx)

	// Test elapsed query command
	posStr, err := client.sendCommand(ctx, []interface{}{"get_property", "time-pos"})
	if err != nil {
		t.Fatalf("failed to send time-pos command: %v", err)
	}

	var elapsed float64
	err = json.Unmarshal([]byte(posStr), &elapsed)
	if err != nil {
		t.Fatalf("failed to parse mock response: %v", err)
	}

	if elapsed != 100.5 {
		t.Errorf("expected mock playback coordinate 100.5, got %.1f", elapsed)
	}
}
