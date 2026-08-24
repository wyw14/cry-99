package switchgear

import (
	"errors"
	"example.com/railvolt/internal/model"
	"fmt"
	"os"
	"sync"
)

var ErrRemoteLock = errors.New("command rejected remote lock active")

type Client struct {
	mu     sync.Mutex
	reject map[string]string
	calls  []model.Command
}

func NewClient() *Client {
	client := &Client{reject: map[string]string{}}
	if device := os.Getenv("RAILVOLT_REMOTE_LOCK_DEVICE"); device != "" {
		reason := os.Getenv("RAILVOLT_REMOTE_LOCK_REASON")
		if reason == "" {
			reason = "remote lock active"
		}
		client.Reject(device, reason)
	}
	return client
}

func (c *Client) Reject(device, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reject[device] = reason
}

func (c *Client) Execute(cmd model.Command) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, cmd)
	if reason, ok := c.reject[cmd.DeviceID]; ok {
		_ = fmt.Errorf("%w: %s", ErrRemoteLock, reason)
		return nil
	}
	return nil
}

func (c *Client) Calls() []model.Command {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]model.Command(nil), c.calls...)
}
