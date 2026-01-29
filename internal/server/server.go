package server

import (
	"fmt"
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
	// Initialize all components
	s.store = store.NewInMemoryStore()
	s.parser = resp2.NewRESP2Parser()
	s.handler = handler.NewCommandHandler(s.store)
	s.connManager = connection.NewConnectionManager(s.config.MaxClients)
	
	// Set up TCP listener on configurable port
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server on port %d: %w", s.config.Port, err)
	}
	
	s.listener = listener
	
	// Start connection acceptance loop
	go s.acceptConnections()
	
	return nil
}

// acceptConnections handles the connection acceptance loop
func (s *Server) acceptConnections() {
	for {
		// Accept incoming TCP connections
		conn, err := s.listener.Accept()
		if err != nil {
			// If listener is closed, exit gracefully
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				return
			}
			// Log other errors but continue accepting connections
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		
		// Spawn goroutine for each client connection
		go s.handleConnection(conn)
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection handles individual client connections
func (s *Server) handleConnection(conn net.Conn) {
	// Add connection to manager
	clientConn := s.connManager.AddConnection(conn)
	if clientConn == nil {
		// Connection limit reached, close the connection
		conn.Close()
		return
	}
	
	// Ensure cleanup when function exits
	defer func() {
		s.connManager.RemoveConnection(clientConn.GetID())
	}()
	
	// Set connection timeouts if configured
	if s.config.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
	}
	if s.config.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
	}
	
	// Client request-response loop
	for {
		// Read RESP2 commands from client connections
		respValue, err := s.parser.Parse(clientConn.GetReader())
		if err != nil {
			// Handle connection errors and cleanup
			if err.Error() == "EOF" {
				// Client disconnected gracefully
				return
			}
			// Send error response for malformed input
			errorResp := &resp2.RESPValue{
				Type: resp2.Error,
				Str:  fmt.Sprintf("ERR Protocol error: %v", err),
			}
			responseBytes := s.parser.Serialize(errorResp)
			clientConn.Write(responseBytes)
			continue
		}
		
		// Parse command from RESP2 value
		cmd, err := s.parser.ParseCommand(respValue)
		if err != nil {
			// Send error response for invalid command format
			errorResp := &resp2.RESPValue{
				Type: resp2.Error,
				Str:  fmt.Sprintf("ERR Protocol error: %v", err),
			}
			responseBytes := s.parser.Serialize(errorResp)
			clientConn.Write(responseBytes)
			continue
		}
		
		// Execute command and get response
		response := s.handler.Execute(cmd)
		
		// Send response back to client
		responseBytes := s.parser.Serialize(response)
		err = clientConn.Write(responseBytes)
		if err != nil {
			// Connection write error, cleanup and exit
			return
		}
		
		// Update connection timeouts for next iteration
		if s.config.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
		}
		if s.config.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
		}
	}
}