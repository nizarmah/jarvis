// Package server provides a TCP server.
package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
)

// OnMessageFunc is the callback for processing messages.
type OnMessageFunc func(ctx context.Context, msg string) error

// TCPServerConfig is the configuration for the TCP server.
type TCPServerConfig struct {
	// Address is the address to listen on.
	Address string
	// Debug enables logging while watching and combining chunks.
	Debug bool
	// OnMessage is the callback for processing messages.
	OnMessage OnMessageFunc
}

// TCPServer is a TCP server.
type TCPServer struct {
	address string
	debug   bool

	onMessage OnMessageFunc
	listener  net.Listener
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(cfg TCPServerConfig) *TCPServer {
	return &TCPServer{
		address:   cfg.Address,
		debug:     cfg.Debug,
		onMessage: cfg.OnMessage,
	}
}

// Start starts the TCP server.
func (s *TCPServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on address %q: %w", s.address, err)
	}

	// Store the listener for future use.
	s.listener = listener

	go s.runListener(ctx, listener)

	log.Println(fmt.Sprintf("server started on address %q", s.address))

	return nil
}

// Stop stops the TCP server.
func (s *TCPServer) stop(reason string) error {
	if s.listener == nil {
		return fmt.Errorf("server not started")
	}

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}

	s.listener = nil

	log.Println(fmt.Sprintf("server stopped: %s", reason))

	return nil
}

// RunListener accepts connections and handles them.
func (s *TCPServer) runListener(ctx context.Context, listener net.Listener) {
	for {
		select {
		// Clean up the goroutine when necessary.
		case <-ctx.Done():
			s.stop("context cancelled")
			return

		// Accept and handle the connection.
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.stop(fmt.Sprintf("accept error: %s", err))
				return
			}

			go s.handleConnection(ctx, conn)
		}
	}
}

// HandleConnection retrieves the messages from the connection and processes them.
func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())

		if err := s.onMessage(ctx, msg); err != nil {
			s.stop(fmt.Sprintf("message error: %s", err))
			return
		}
	}
}
