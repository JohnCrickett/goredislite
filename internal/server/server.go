package server

import (
	"net"
	"time"

	"redis-like-server/internal/connection"
	"redis-like-server/internal/handler"
	"redis-like-server/internal/resp2"
	"redis-like-server/internal/store"
)

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port         int
	MaxClients   int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Server represents the main Redis-like server
type Server struct {
	listener    net.Listener
	store       store.KeyValueStore
	parser      resp2.RESP2Parser
	handler     handler.CommandHandler
	connManager connection.ConnectionManager
	config      *ServerConfig
}

// NewServer creates a new server instance
func NewServer(config *ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

// Start initializes and starts the server
func (s *Server) Start() error {
	// Implementation will be added in later tasks
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	// Implementation will be added in later tasks
	return nil
}

// handleConnection handles individual client connections
func (s *Server) handleConnection(conn net.Conn) {
	// Implementation will be added in later tasks
}