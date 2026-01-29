package server

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestServer(t *testing.T) {
	// Unit tests will be added in later tasks
}

// Property-based test setup for concurrent client processing
func TestConcurrentClientProcessing(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 3: Concurrent Client Processing (full concurrency aspect)
	// **Validates: Requirements 1.3, 6.1, 6.2**
	// For any set of concurrent clients, the server should process all their commands without blocking
	properties.Property("concurrent client processing", prop.ForAll(
		func(clientCount int, commands []string) bool {
			if clientCount <= 0 || len(commands) == 0 {
				return true // Skip invalid inputs
			}
			
			// Create server with test configuration
			config := &ServerConfig{
				Port:         0, // Use random available port
				MaxClients:   clientCount + 5, // Allow more than test clients
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			}
			server := NewServer(config)
			
			// Start server
			err := server.Start()
			if err != nil {
				t.Logf("Failed to start server: %v", err)
				return false
			}
			defer server.Stop()
			
			// Get the actual port the server is listening on
			addr := server.listener.Addr().(*net.TCPAddr)
			port := addr.Port
			
			// Create concurrent clients
			var wg sync.WaitGroup
			results := make([]bool, clientCount)
			
			for i := 0; i < clientCount; i++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					
					// Connect to server
					conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
					if err != nil {
						t.Logf("Client %d failed to connect: %v", clientID, err)
						results[clientID] = false
						return
					}
					defer conn.Close()
					
					reader := bufio.NewReader(conn)
					writer := bufio.NewWriter(conn)
					
					// Send commands and verify responses
					for range commands {
						// Send PING command (most reliable for testing)
						pingCmd := fmt.Sprintf("*1\r\n$4\r\nPING\r\n")
						_, err := writer.WriteString(pingCmd)
						if err != nil {
							results[clientID] = false
							return
						}
						writer.Flush()
						
						// Read response
						response, err := reader.ReadString('\n')
						if err != nil {
							results[clientID] = false
							return
						}
						
						// Verify PONG response
						if response != "+PONG\r\n" {
							results[clientID] = false
							return
						}
					}
					
					results[clientID] = true
				}(i)
			}
			
			// Wait for all clients to complete with timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			
			select {
			case <-done:
				// All clients completed
			case <-time.After(10 * time.Second):
				// Timeout - some clients may be blocked
				return false
			}
			
			// Check that all clients succeeded
			for i, result := range results {
				if !result {
					t.Logf("Client %d failed", i)
					return false
				}
			}
			
			return true
		},
		gen.IntRange(1, 5), // Test with 1-5 concurrent clients
		gen.SliceOfN(3, gen.Const("PING")), // Send 3 PING commands per client
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property-based test setup for PING command reliability
func TestPINGCommandReliability(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 5: PING Command Reliability
	// **Validates: Requirements 3.3**
	// For any server state, PING commands should always be processed and return appropriate responses
	properties.Property("PING command reliability", prop.ForAll(
		func(preCommands []string, pingMessage string) bool {
			// Create server with test configuration
			config := &ServerConfig{
				Port:         0, // Use random available port
				MaxClients:   10,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			}
			server := NewServer(config)
			
			// Start server
			err := server.Start()
			if err != nil {
				t.Logf("Failed to start server: %v", err)
				return false
			}
			defer server.Stop()
			
			// Get the actual port the server is listening on
			addr := server.listener.Addr().(*net.TCPAddr)
			port := addr.Port
			
			// Connect to server
			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
			if err != nil {
				t.Logf("Failed to connect: %v", err)
				return false
			}
			defer conn.Close()
			
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			
			// Execute pre-commands to change server state
			for _, cmd := range preCommands {
				if cmd == "" {
					continue
				}
				
				// Send SET command to modify server state
				setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$5\r\nvalue\r\n", len(cmd), cmd)
				writer.WriteString(setCmd)
				writer.Flush()
				
				// Read response (should be +OK)
				reader.ReadString('\n')
			}
			
			// Test PING command reliability
			var pingCmd string
			var expectedResponse string
			
			if pingMessage == "" {
				// PING with no arguments
				pingCmd = "*1\r\n$4\r\nPING\r\n"
				expectedResponse = "+PONG\r\n"
			} else {
				// PING with message argument
				pingCmd = fmt.Sprintf("*2\r\n$4\r\nPING\r\n$%d\r\n%s\r\n", len(pingMessage), pingMessage)
				expectedResponse = fmt.Sprintf("$%d\r\n%s\r\n", len(pingMessage), pingMessage)
			}
			
			// Send PING command
			_, err = writer.WriteString(pingCmd)
			if err != nil {
				t.Logf("Failed to send PING: %v", err)
				return false
			}
			writer.Flush()
			
			// Read response
			response, err := reader.ReadString('\n')
			if err != nil {
				t.Logf("Failed to read PING response: %v", err)
				return false
			}
			
			// For bulk string responses, we need to read the content too
			if pingMessage != "" && response == fmt.Sprintf("$%d\r\n", len(pingMessage)) {
				content, err := reader.ReadString('\n')
				if err != nil {
					t.Logf("Failed to read PING content: %v", err)
					return false
				}
				response += content
			}
			
			// Verify expected response
			if response != expectedResponse {
				t.Logf("Expected %q, got %q", expectedResponse, response)
				return false
			}
			
			return true
		},
		gen.SliceOfN(3, gen.AlphaString()), // Pre-commands to change server state
		gen.AlphaString(), // PING message
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property-based test setup for string value storage capacity
func TestStringValueStorageCapacity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 12: String Value Storage Capacity
	// **Validates: Requirements 4.4**
	// For any string value within memory constraints, the key-value store should be able to store and retrieve it correctly
	properties.Property("string value storage capacity", prop.ForAll(
		func(key string, value string, valueSize int) bool {
			if key == "" {
				return true // Skip empty keys
			}
			
			// Generate different types of values based on valueSize
			var testValue string
			switch valueSize % 4 {
			case 0:
				testValue = value // Use generated string as-is
			case 1:
				testValue = "" // Empty string
			case 2:
				testValue = "a" // Single character
			case 3:
				// Create a larger string (up to 1KB)
				size := valueSize % 1024
				if size == 0 {
					size = 1
				}
				testValue = string(make([]byte, size))
				for i := range testValue {
					testValue = testValue[:i] + "x" + testValue[i+1:]
				}
			}
			
			// Create server with test configuration
			config := &ServerConfig{
				Port:         0, // Use random available port
				MaxClients:   10,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			}
			server := NewServer(config)
			
			// Start server
			err := server.Start()
			if err != nil {
				t.Logf("Failed to start server: %v", err)
				return false
			}
			defer server.Stop()
			
			// Get the actual port the server is listening on
			addr := server.listener.Addr().(*net.TCPAddr)
			port := addr.Port
			
			// Connect to server
			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
			if err != nil {
				t.Logf("Failed to connect: %v", err)
				return false
			}
			defer conn.Close()
			
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			
			// Test SET command with the generated value
			setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", 
				len(key), key, len(testValue), testValue)
			
			_, err = writer.WriteString(setCmd)
			if err != nil {
				t.Logf("Failed to send SET command: %v", err)
				return false
			}
			writer.Flush()
			
			// Read SET response (should be +OK)
			setResponse, err := reader.ReadString('\n')
			if err != nil {
				t.Logf("Failed to read SET response: %v", err)
				return false
			}
			
			if setResponse != "+OK\r\n" {
				t.Logf("SET failed, expected +OK, got %q", setResponse)
				return false
			}
			
			// Test GET command to retrieve the value
			getCmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
			
			_, err = writer.WriteString(getCmd)
			if err != nil {
				t.Logf("Failed to send GET command: %v", err)
				return false
			}
			writer.Flush()
			
			// Read GET response header
			getResponse, err := reader.ReadString('\n')
			if err != nil {
				t.Logf("Failed to read GET response: %v", err)
				return false
			}
			
			// Verify bulk string response format
			expectedHeader := fmt.Sprintf("$%d\r\n", len(testValue))
			if getResponse != expectedHeader {
				t.Logf("GET response header mismatch, expected %q, got %q", expectedHeader, getResponse)
				return false
			}
			
			// Read the actual value content
			valueResponse, err := reader.ReadString('\n')
			if err != nil {
				t.Logf("Failed to read GET value: %v", err)
				return false
			}
			
			// Verify the retrieved value matches what was stored
			expectedValue := testValue + "\r\n"
			if valueResponse != expectedValue {
				t.Logf("Value mismatch, expected %q, got %q", expectedValue, valueResponse)
				return false
			}
			
			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 100 }), // Key with reasonable length
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) <= 1000 }), // Value up to 1KB
		gen.IntRange(0, 1023), // Size parameter for different value types
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}