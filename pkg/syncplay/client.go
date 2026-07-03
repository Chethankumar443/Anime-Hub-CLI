package syncplay

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SyncEvent struct {
	Event      string `json:"event"` // join, play, pause, seek, status
	Username   string `json:"username"`
	Room       string `json:"room"`
	ElapsedSec int    `json:"elapsed_sec"`
}

type SyncplayClient struct {
	conn       net.Conn
	mu         sync.Mutex
	Room       string
	Username   string
	ServerURL  string
	events     chan SyncEvent
	cancelFunc context.CancelFunc
}

func NewSyncplayClient(serverURL, room, username string) *SyncplayClient {
	return &SyncplayClient{
		ServerURL: serverURL,
		Room:      room,
		Username:  username,
		events:    make(chan SyncEvent, 10),
	}
}

func (c *SyncplayClient) Connect(ctx context.Context) error {
	u, err := url.Parse(c.ServerURL)
	if err != nil {
		return err
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	var conn net.Conn
	dialer := &net.Dialer{Timeout: 5 * time.Second}

	if u.Scheme == "wss" {
		conn, err = tls.DialWithDialer(dialer, "tcp", host, &tls.Config{
			InsecureSkipVerify: true, // Allow dev certs
		})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", host)
	}

	if err != nil {
		return err
	}

	c.conn = conn

	// Sec-WebSocket-Key generation
	randKey := make([]byte, 16)
	_, _ = rand.Read(randKey)
	secKey := base64.StdEncoding.EncodeToString(randKey)

	// Perform WebSocket handshake
	path := u.Path
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}

	handshake := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Key: %s\r\n"+
			"Sec-WebSocket-Version: 13\r\n\r\n",
		path, u.Host, secKey,
	)

	_, err = conn.Write([]byte(handshake))
	if err != nil {
		conn.Close()
		return err
	}

	// Read handshake response headers
	respHeader := make([]byte, 1024)
	n, err := conn.Read(respHeader)
	if err != nil {
		conn.Close()
		return err
	}

	respStr := string(respHeader[:n])
	if !strings.Contains(respStr, "101 Switching Protocols") {
		conn.Close()
		return fmt.Errorf("websocket handshake failed: %s", respStr)
	}

	return nil
}

func (c *SyncplayClient) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	// Send join message
	_ = c.SendEvent(SyncEvent{
		Event:    "join",
		Username: c.Username,
		Room:     c.Room,
	})

	go c.readLoop(ctx)
}

func (c *SyncplayClient) Close() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()
}

func (c *SyncplayClient) Events() <-chan SyncEvent {
	return c.events
}

func (c *SyncplayClient) SendEvent(ev SyncEvent) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	return c.writeFrame(0x01, data) // Text frame
}

func (c *SyncplayClient) writeFrame(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	var header []byte
	header = append(header, 0x80|opcode) // FIN = 1, Opcode

	length := len(payload)
	// Client frames must be masked (so OR length with 0x80)
	if length <= 125 {
		header = append(header, 0x80|byte(length))
	} else if length <= 65535 {
		header = append(header, 0x80|126)
		header = append(header, byte(length>>8), byte(length&0xFF))
	} else {
		// Too large for basic Syncplay event
		return fmt.Errorf("payload too large")
	}

	// Generate masking key
	maskKey := make([]byte, 4)
	_, _ = rand.Read(maskKey)
	header = append(header, maskKey...)

	// Mask the payload
	maskedPayload := make([]byte, length)
	for i := 0; i < length; i++ {
		maskedPayload[i] = payload[i] ^ maskKey[i%4]
	}

	// Write header and payload
	_, err := c.conn.Write(append(header, maskedPayload...))
	return err
}

func (c *SyncplayClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		opcode, payload, err := c.readFrame()
		if err != nil {
			return
		}

		if opcode == 0x08 { // Close frame
			return
		}

		if opcode == 0x01 { // Text frame
			var ev SyncEvent
			if err := json.Unmarshal(payload, &ev); err == nil {
				c.events <- ev
			}
		}
	}
}

func (c *SyncplayClient) readFrame() (byte, []byte, error) {
	header := make([]byte, 2)
	_, err := io.ReadFull(c.conn, header)
	if err != nil {
		return 0, nil, err
	}

	opcode := header[0] & 0x0F
	mask := (header[1] & 0x80) != 0
	length := int(header[1] & 0x7F)

	if length == 126 {
		lenBytes := make([]byte, 2)
		_, err = io.ReadFull(c.conn, lenBytes)
		if err != nil {
			return 0, nil, err
		}
		length = int(lenBytes[0])<<8 | int(lenBytes[1])
	} else if length == 127 {
		lenBytes := make([]byte, 8)
		_, err = io.ReadFull(c.conn, lenBytes)
		if err != nil {
			return 0, nil, err
		}
		// Syncplay events shouldn't exceed 64KB, but we read it to be spec-compliant
		length = 0
		for _, b := range lenBytes {
			length = (length << 8) | int(b)
		}
	}

	var maskKey []byte
	if mask {
		maskKey = make([]byte, 4)
		_, err = io.ReadFull(c.conn, maskKey)
		if err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, length)
	_, err = io.ReadFull(c.conn, payload)
	if err != nil {
		return 0, nil, err
	}

	if mask {
		for i := 0; i < length; i++ {
			payload[i] = payload[i] ^ maskKey[i%4]
		}
	}

	return opcode, payload, nil
}
