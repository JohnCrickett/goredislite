package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	shutdown    chan struct{}
}

// NewServer creates a new server instance
func NewServer(config *ServerConfig) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		shutdown: make(chan struct{}),
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
	
	// Set up signal handling for graceful shutdown
	s.setupSignalHandling()
	
	// Start connection acceptance loop
	s.wg.Add(1)
	go s.acceptConnections()
	
	return nil
}

// setupSignalHandling sets up signal handlers for graceful shutdown
func (s *Server) setupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, initiating graceful shutdown...")
		s.Stop()
	}()
}

// acceptConnections handles the connection acceptance loop
func (s *Server) acceptConnections() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.ctx.Done():
			// Server is shutting down
			return
		default:
			// Accept incoming TCP connections
			conn, err := s.listener.Accept()
			if err != nil {
				// Check if this is due to listener being closed during shutdown
				select {
				case <-s.ctx.Done():
					return
				default:
					// Log other errors but continue accepting connections
					fmt.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}
			
			// Spawn goroutine for each client connection
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	// Signal shutdown to all goroutines
	s.cancel()
	
	// Close the listener to stop accepting new connections
	if s.listener != nil {
		s.listener.Close()
	}
	
	// Close all existing client connections
	if s.connManager != nil {
		fmt.Println("Closing all client connections...")
		s.connManager.CloseAllConnections()
	}
	
	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		fmt.Println("All connections closed gracefully")
	case <-time.After(10 * time.Second):
		fmt.Println("Timeout waiting for connections to close, forcing shutdown")
	}
	
	// Signal that shutdown is complete
	close(s.shutdown)
	
	return nil
}

// WaitForShutdown blocks until the server has shut down
func (s *Server) WaitForShutdown() {
	<-s.shutdown
}

// GetListener returns the server's listener (for testing purposes)
func (s *Server) GetListener() net.Listener {
	return s.listener
}

// handleConnection handles individual client connections
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	
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
		select {
		case <-s.ctx.Done():
			// Server is shutting down, close connection gracefully
			return
		default:
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
}