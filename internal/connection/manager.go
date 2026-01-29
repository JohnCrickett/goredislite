package connection

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"time"
)

// ConnectionManager manages all active client connections and their lifecycle
type ConnectionManager interface {
	AddConnection(conn net.Conn) *ClientConnection
	RemoveConnection(id string)
	GetConnection(id string) *ClientConnection
	GetActiveCount() int
	CleanupStaleConnections()
	CloseAllConnections()
}

// ClientConnection wraps a network connection with additional metadata
type ClientConnection struct {
	conn       net.Conn
	id         string
	lastActive time.Time
	reader     *bufio.Reader
	writer     *bufio.Writer
}

// GetID returns the connection ID
func (cc *ClientConnection) GetID() string {
	return cc.id
}

// GetConn returns the underlying network connection
func (cc *ClientConnection) GetConn() net.Conn {
	return cc.conn
}

// GetReader returns the buffered reader
func (cc *ClientConnection) GetReader() *bufio.Reader {
	return cc.reader
}

// GetWriter returns the buffered writer
func (cc *ClientConnection) GetWriter() *bufio.Writer {
	return cc.writer
}

// UpdateLastActive updates the last activity timestamp
func (cc *ClientConnection) UpdateLastActive() {
	cc.lastActive = time.Now()
}

// GetLastActive returns the last activity timestamp
func (cc *ClientConnection) GetLastActive() time.Time {
	return cc.lastActive
}

// Close closes the connection and flushes any pending writes
func (cc *ClientConnection) Close() error {
	// Flush any pending writes
	if cc.writer != nil {
		cc.writer.Flush()
	}
	// Close the underlying connection
	return cc.conn.Close()
}

// Write writes data to the connection and flushes
func (cc *ClientConnection) Write(data []byte) error {
	cc.UpdateLastActive()
	_, err := cc.writer.Write(data)
	if err != nil {
		return err
	}
	return cc.writer.Flush()
}

// IsStale checks if the connection is stale based on timeout
func (cc *ClientConnection) IsStale(timeout time.Duration) bool {
	return time.Since(cc.lastActive) > timeout
}

// DefaultConnectionManager is the default implementation of ConnectionManager
type DefaultConnectionManager struct {
	connections map[string]*ClientConnection
	mutex       sync.RWMutex
	maxClients  int
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(maxClients int) ConnectionManager {
	return &DefaultConnectionManager{
		connections: make(map[string]*ClientConnection),
		maxClients:  maxClients,
	}
}

// AddConnection adds a new client connection
func (cm *DefaultConnectionManager) AddConnection(conn net.Conn) *ClientConnection {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Check if we've reached the maximum number of clients
	if cm.maxClients > 0 && len(cm.connections) >= cm.maxClients {
		return nil // Connection limit reached
	}
	
	// Generate a unique connection ID
	id := generateConnectionID()
	
	// Create the client connection wrapper
	clientConn := &ClientConnection{
		conn:       conn,
		id:         id,
		lastActive: time.Now(),
		reader:     bufio.NewReader(conn),
		writer:     bufio.NewWriter(conn),
	}
	
	// Add to connections map
	cm.connections[id] = clientConn
	
	return clientConn
}

// RemoveConnection removes a client connection
func (cm *DefaultConnectionManager) RemoveConnection(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if clientConn, exists := cm.connections[id]; exists {
		// Close the client connection (which handles flushing and closing)
		clientConn.Close()
		// Remove from connections map
		delete(cm.connections, id)
	}
}

// GetConnection retrieves a connection by ID
func (cm *DefaultConnectionManager) GetConnection(id string) *ClientConnection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if clientConn, exists := cm.connections[id]; exists {
		return clientConn
	}
	return nil
}

// GetActiveCount returns the number of active connections
func (cm *DefaultConnectionManager) GetActiveCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return len(cm.connections)
}

// CleanupStaleConnections removes stale connections
func (cm *DefaultConnectionManager) CleanupStaleConnections() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	staleTimeout := 5 * time.Minute // Connections inactive for 5 minutes are considered stale
	
	var staleIDs []string
	
	// Find stale connections
	for id, clientConn := range cm.connections {
		if clientConn.IsStale(staleTimeout) {
			staleIDs = append(staleIDs, id)
		}
	}
	
	// Remove stale connections
	for _, id := range staleIDs {
		if clientConn, exists := cm.connections[id]; exists {
			clientConn.Close()
			delete(cm.connections, id)
		}
	}
}

// CloseAllConnections closes all active connections
func (cm *DefaultConnectionManager) CloseAllConnections() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Close all connections
	for id, clientConn := range cm.connections {
		clientConn.Close()
		delete(cm.connections, id)
	}
}

// generateConnectionID generates a unique connection identifier
func generateConnectionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("conn_%x", bytes)
}