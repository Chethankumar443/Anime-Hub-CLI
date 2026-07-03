package player

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type PlayerManager struct {
	PlayerName string // mpv or vlc
	BinaryPath string
}

func NewPlayerManager(playerName string) *PlayerManager {
	// Locate path or default
	binary := playerName
	if binary == "" {
		binary = "mpv"
	}

	path, err := exec.LookPath(binary)
	if err != nil {
		path = binary // fallback to command name
	}

	return &PlayerManager{
		PlayerName: binary,
		BinaryPath: path,
	}
}

// StartPlayer launches the player in a separate thread, binding it to the specified socket
func (m *PlayerManager) StartPlayer(streamURL string, ipcPipePath string, headers map[string]string) (*exec.Cmd, error) {
	var args []string

	if strings.ToLower(m.PlayerName) == "vlc" {
		// Launch VLC with RC socket control
		// Windows: vlc.exe --extraintf rc --rc-host localhost:4212 streamURL
		// Unix: vlc --extraintf rc --rc-unix ipcPipePath streamURL
		args = append(args, streamURL, "--play-and-exit", "--extraintf", "rc")
		if runtime.GOOS == "windows" {
			args = append(args, "--rc-host", ipcPipePath)
		} else {
			args = append(args, "--rc-unix", ipcPipePath)
		}
	} else {
		// Launch MPV with standard JSON-IPC
		args = append(args,
			streamURL,
			"--no-terminal",
			"--force-window",
			"--title=anime-cli Playback Window",
			fmt.Sprintf("--input-ipc-server=%s", ipcPipePath),
		)

		// Parse spoofer headers for request bypass
		if len(headers) > 0 {
			var fields []string
			for k, v := range headers {
				fields = append(fields, fmt.Sprintf("%s: %s", k, v))
			}
			args = append(args, fmt.Sprintf("--http-header-fields=%s", strings.Join(fields, ",")))
		}
	}

	cmd := exec.Command(m.BinaryPath, args...)
	
	// Separate stdout/stderr so they don't break TUI drawing
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Platform specific adjustments
	setPlatformSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start player %s: %w", m.PlayerName, err)
	}

	return cmd, nil
}

// CheckPlayerInstalled returns true if the player binary is present on the path.
func (m *PlayerManager) CheckPlayerInstalled() bool {
	_, err := exec.LookPath(m.PlayerName)
	return err == nil
}
