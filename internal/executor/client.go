// Package executor provides a client for the executor server.
package executor

import (
	"context"
	"fmt"
	"net"
)

// ClientConfig is the configuration for the client.
type ClientConfig struct {
	Address string
	Debug   bool
}

// Client is the client for the executor server.
type Client struct {
	address string
	debug   bool
}

// NewClient creates a new client.
func NewClient(cfg ClientConfig) (*Client, error) {
	client := &Client{
		address: cfg.Address,
		debug:   cfg.Debug,
	}

	if err := client.Healthcheck(context.Background()); err != nil {
		return nil, fmt.Errorf("executor is not running: %w", err)
	}

	return client, nil
}

// Healthcheck checks if the executor server is running.
func (c *Client) Healthcheck(ctx context.Context) error {
	// Connect to the executor.
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to connect to executor: %w", err)
	}
	defer conn.Close()

	return nil
}

// SendCommand sends a command to the executor server.
func (c *Client) SendCommand(ctx context.Context, command string) error {
	// Connect to the executor.
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to connect to executor: %w", err)
	}
	defer conn.Close()

	// Send the command to the executor.
	_, err = conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("failed to send command to executor: %w", err)
	}

	return nil
}
