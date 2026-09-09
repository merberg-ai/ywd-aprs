package kiss

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type ClientConfig struct {
	Address        string
	KISSPort       int
	ReconnectDelay time.Duration
}

type Status struct {
	Connected     bool      `json:"connected"`
	ConnectedAt   time.Time `json:"connected_at,omitempty"`
	LastRX        time.Time `json:"last_rx,omitempty"`
	RXFrames      uint64    `json:"rx_frames"`
	Reconnects    uint64    `json:"reconnects"`
	IgnoredFrames uint64    `json:"ignored_frames"`
	LastError     string    `json:"last_error,omitempty"`
}

type Client struct {
	cfg      ClientConfig
	onData   func(port int, payload []byte)
	onStatus func(Status)
	mu       sync.Mutex
	status   Status
}

func NewClient(cfg ClientConfig, onData func(port int, payload []byte), onStatus func(Status)) *Client {
	if cfg.ReconnectDelay <= 0 {
		cfg.ReconnectDelay = 3 * time.Second
	}
	return &Client{cfg: cfg, onData: onData, onStatus: onStatus}
}

func (c *Client) Run(ctx context.Context) {
	first := true
	for {
		if ctx.Err() != nil {
			return
		}
		if !first {
			c.bumpReconnect()
		}
		first = false

		dialer := net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", c.cfg.Address)
		if err != nil {
			c.setDisconnected(fmt.Sprintf("connect: %v", err))
			if !sleepContext(ctx, c.cfg.ReconnectDelay) {
				return
			}
			continue
		}
		c.setConnected()
		err = c.readLoop(ctx, conn)
		_ = conn.Close()
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			err = io.EOF
		}
		c.setDisconnected(err.Error())
		if !sleepContext(ctx, c.cfg.ReconnectDelay) {
			return
		}
	}
}

func (c *Client) readLoop(ctx context.Context, conn net.Conn) error {
	decoder := &Decoder{}
	buf := make([]byte, 4096)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(buf)
		if n > 0 {
			for _, frame := range decoder.Feed(buf[:n]) {
				if frame.Command != 0 || frame.Port != c.cfg.KISSPort {
					c.bumpIgnored()
					continue
				}
				c.bumpRX()
				if c.onData != nil {
					c.onData(frame.Port, frame.Payload)
				}
			}
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			}
			return err
		}
	}
}

func (c *Client) snapshot() Status { c.mu.Lock(); defer c.mu.Unlock(); return c.status }
func (c *Client) publish() {
	if c.onStatus != nil {
		c.onStatus(c.snapshot())
	}
}
func (c *Client) setConnected() {
	c.mu.Lock()
	c.status.Connected = true
	c.status.ConnectedAt = time.Now()
	c.status.LastError = ""
	c.mu.Unlock()
	c.publish()
}
func (c *Client) setDisconnected(err string) {
	c.mu.Lock()
	c.status.Connected = false
	c.status.LastError = err
	c.mu.Unlock()
	c.publish()
}
func (c *Client) bumpReconnect() { c.mu.Lock(); c.status.Reconnects++; c.mu.Unlock(); c.publish() }
func (c *Client) bumpIgnored()   { c.mu.Lock(); c.status.IgnoredFrames++; c.mu.Unlock(); c.publish() }
func (c *Client) bumpRX() {
	c.mu.Lock()
	c.status.RXFrames++
	c.status.LastRX = time.Now()
	c.mu.Unlock()
	c.publish()
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
