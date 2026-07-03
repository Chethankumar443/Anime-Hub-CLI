package player

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PlaybackUpdate struct {
	ElapsedSec  int
	DurationSec int
	Paused      bool
	Closed      bool
}

type IPCClient struct {
	conn       net.Conn
	mu         sync.Mutex
	requestID  int
	pending    map[int]chan string
	pipePath   string
	isTCP      bool
	updates    chan PlaybackUpdate
	cancelFunc context.CancelFunc
	PlayerName string // "mpv" or "vlc"
}

func NewIPCClient(pipePath string) *IPCClient {
	isTCP := strings.Contains(pipePath, ":")
	return &IPCClient{
		pipePath:   pipePath,
		isTCP:      isTCP,
		pending:    make(map[int]chan string),
		updates:    make(chan PlaybackUpdate, 10),
		PlayerName: "mpv", // default for backwards compatibility
	}
}

// Connect establishes connection with the player IPC server.
func (c *IPCClient) Connect(ctx context.Context) error {
	var conn net.Conn
	var err error

	// Retry connection for up to 3 seconds as the player process starts
	for i := 0; i < 15; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if c.isTCP {
			conn, err = net.DialTimeout("tcp", c.pipePath, 500*time.Millisecond)
		} else {
			conn, err = net.DialTimeout("unix", c.pipePath, 500*time.Millisecond)
		}

		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to player IPC at %s: %w", c.pipePath, err)
	}

	c.conn = conn
	return nil
}

func (c *IPCClient) Updates() <-chan PlaybackUpdate {
	return c.updates
}

// Close closes the connection.
func (c *IPCClient) Close() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// StartMonitoring spawns background goroutines to read ticks from the connection.
func (c *IPCClient) StartMonitoring(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	if c.PlayerName == "vlc" {
		go c.vlcPollLoop(ctx)
	} else {
		// Read loop
		go c.readLoop(ctx)
		// Poll loop (get position and pause status every 1 second)
		go c.pollLoop(ctx)
	}
}

func (c *IPCClient) sendCommand(ctx context.Context, cmd []interface{}) (string, error) {
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	ch := make(chan string, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	payload := map[string]interface{}{
		"command":    cmd,
		"request_id": id,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return "", fmt.Errorf("connection is nil")
	}
	_, err = c.conn.Write(append(data, '\n'))
	c.mu.Unlock()

	if err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case resp := <-ch:
		return resp, nil
	}
}

func (c *IPCClient) readLoop(ctx context.Context) {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Bytes()
		var msg map[string]json.RawMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		// Handle events (pause, unpause, end-file)
		if ev, ok := msg["event"]; ok {
			var eventName string
			_ = json.Unmarshal(ev, &eventName)
			switch eventName {
			case "pause":
				c.updates <- PlaybackUpdate{Paused: true}
			case "unpause":
				c.updates <- PlaybackUpdate{Paused: false}
			case "end-file":
				c.updates <- PlaybackUpdate{Closed: true}
				return
			}
			continue
		}

		// Handle responses
		if reqIDVal, ok := msg["request_id"]; ok {
			var reqID int
			_ = json.Unmarshal(reqIDVal, &reqID)

			c.mu.Lock()
			ch, exists := c.pending[reqID]
			c.mu.Unlock()

			if exists {
				var dataVal json.RawMessage
				if data, ok := msg["data"]; ok {
					dataVal = data
				}
				ch <- string(dataVal)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		// Log/handle error if necessary
	}

	// Scanner exited -> socket closed
	c.updates <- PlaybackUpdate{Closed: true}
}

func (c *IPCClient) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Fetch position
			posStr, err := c.sendCommand(ctx, []interface{}{"get_property", "time-pos"})
			if err != nil {
				continue
			}

			// Fetch duration
			durStr, err := c.sendCommand(ctx, []interface{}{"get_property", "duration"})
			if err != nil {
				continue
			}

			// Fetch pause
			pauseStr, err := c.sendCommand(ctx, []interface{}{"get_property", "pause"})
			if err != nil {
				continue
			}

			var elapsed float64
			var duration float64
			var paused bool

			_ = json.Unmarshal([]byte(posStr), &elapsed)
			_ = json.Unmarshal([]byte(durStr), &duration)
			_ = json.Unmarshal([]byte(pauseStr), &paused)

			c.updates <- PlaybackUpdate{
				ElapsedSec:  int(elapsed),
				DurationSec: int(duration),
				Paused:      paused,
			}
		}
	}
}

func (c *IPCClient) sendVLCCommand(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return "", fmt.Errorf("connection is nil")
	}

	// Flush any pending data in read buffer first by setting a short read deadline
	_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Millisecond))
	tmp := make([]byte, 1024)
	for {
		n, err := c.conn.Read(tmp)
		if n == 0 || err != nil {
			break
		}
	}

	_ = c.conn.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
	_, err := c.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}

	_ = c.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	reader := bufio.NewReader(c.conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip status messages or banners
		if strings.Contains(trimmed, "VLC") || strings.Contains(trimmed, "status change") || strings.Contains(trimmed, "Type 'help'") {
			continue
		}
		return trimmed, nil
	}
}

func (c *IPCClient) vlcPollLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Wait a moment for player startup
	time.Sleep(500 * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Fetch position
			timeStr, err := c.sendVLCCommand("get_time")
			if err != nil {
				c.updates <- PlaybackUpdate{Closed: true}
				return
			}
			elapsed, _ := strconv.Atoi(timeStr)

			// Fetch duration
			lengthStr, err := c.sendVLCCommand("get_length")
			if err != nil {
				c.updates <- PlaybackUpdate{Closed: true}
				return
			}
			duration, _ := strconv.Atoi(lengthStr)

			// Fetch playing status
			playingStr, err := c.sendVLCCommand("is_playing")
			if err != nil {
				c.updates <- PlaybackUpdate{Closed: true}
				return
			}
			paused := playingStr == "0"

			c.updates <- PlaybackUpdate{
				ElapsedSec:  elapsed,
				DurationSec: duration,
				Paused:      paused,
			}
		}
	}
}

// Seek sets the playback offset in seconds
func (c *IPCClient) Seek(ctx context.Context, seconds int) error {
	if c.PlayerName == "vlc" {
		_, err := c.sendVLCCommand(fmt.Sprintf("seek %d", seconds))
		return err
	}
	_, err := c.sendCommand(ctx, []interface{}{"seek", seconds, "absolute"})
	return err
}

// Pause toggles play/pause state
func (c *IPCClient) SetPause(ctx context.Context, pause bool) error {
	if c.PlayerName == "vlc" {
		_, err := c.sendVLCCommand("pause")
		return err
	}
	_, err := c.sendCommand(ctx, []interface{}{"set_property", "pause", pause})
	return err
}
