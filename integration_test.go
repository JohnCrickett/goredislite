package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"redis-like-server/internal/server"
)

// TestFullClientServerInteraction tests complete client-server interaction flows
func TestFullClientServerInteraction(t *testing.T) {
	// Create server with test configuration
	config := &server.ServerConfig{
		Port:         0, // Use random available port
		MaxClients:   10,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	srv := server.NewServer(config)

	// Start server
	err := srv.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Get the actual port the server is listening on
	addr := srv.GetListener().Addr().(*net.TCPAddr)
	port := addr.Port

	// Connect to server
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Test PING command
	t.Run("PING command", func(t *testing.T) {
		// Send PING
		pingCmd := "*1\r\n$4\r\nPING\r\n"
		_, err := writer.WriteString(pingCmd)
		if err != nil {
			t.Fatalf("Failed to send PING: %v", err)
		}
		writer.Flush()

		// Read response
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read PING response: %v", err)
		}

		if response != "+PONG\r\n" {
			t.Errorf("Expected +PONG\\r\\n, got %q", response)
		}
	})

	// Test PING with message
	t.Run("PING with message", func(t *testing.T) {
		message := "hello"
		pingCmd := fmt.Sprintf("*2\r\n$4\r\nPING\r\n$%d\r\n%s\r\n", len(message), message)
		_, err := writer.WriteString(pingCmd)
		if err != nil {
			t.Fatalf("Failed to send PING with message: %v", err)
		}
		writer.Flush()

		// Read response header
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read PING response header: %v", err)
		}

		expectedHeader := fmt.Sprintf("$%d\r\n", len(message))
		if response != expectedHeader {
			t.Errorf("Expected %q, got %q", expectedHeader, response)
		}

		// Read response content
		content, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read PING response content: %v", err)
		}

		expectedContent := message + "\r\n"
		if content != expectedContent {
			t.Errorf("Expected %q, got %q", expectedContent, content)
		}
	})

	// Test SET and GET commands
	t.Run("SET and GET commands", func(t *testing.T) {
		key := "testkey"
		value := "testvalue"

		// Send SET command
		setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
		_, err := writer.WriteString(setCmd)
		if err != nil {
			t.Fatalf("Failed to send SET: %v", err)
		}
		writer.Flush()

		// Read SET response
		setResponse, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read SET response: %v", err)
		}

		if setResponse != "+OK\r\n" {
			t.Errorf("Expected +OK\\r\\n, got %q", setResponse)
		}

		// Send GET command
		getCmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
		_, err = writer.WriteString(getCmd)
		if err != nil {
			t.Fatalf("Failed to send GET: %v", err)
		}
		writer.Flush()

		// Read GET response header
		getResponse, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read GET response header: %v", err)
		}

		expectedHeader := fmt.Sprintf("$%d\r\n", len(value))
		if getResponse != expectedHeader {
			t.Errorf("Expected %q, got %q", expectedHeader, getResponse)
		}

		// Read GET response content
		getContent, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read GET response content: %v", err)
		}

		expectedContent := value + "\r\n"
		if getContent != expectedContent {
			t.Errorf("Expected %q, got %q", expectedContent, getContent)
		}
	})

	// Test EXISTS and DEL commands
	t.Run("EXISTS and DEL commands", func(t *testing.T) {
		key := "existskey"
		value := "existsvalue"

		// First set a key
		setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
		writer.WriteString(setCmd)
		writer.Flush()
		reader.ReadString('\n') // Read SET response

		// Test EXISTS
		existsCmd := fmt.Sprintf("*2\r\n$6\r\nEXISTS\r\n$%d\r\n%s\r\n", len(key), key)
		_, err := writer.WriteString(existsCmd)
		if err != nil {
			t.Fatalf("Failed to send EXISTS: %v", err)
		}
		writer.Flush()

		existsResponse, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read EXISTS response: %v", err)
		}

		if existsResponse != ":1\r\n" {
			t.Errorf("Expected :1\\r\\n, got %q", existsResponse)
		}

		// Test DEL
		delCmd := fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key)
		_, err = writer.WriteString(delCmd)
		if err != nil {
			t.Fatalf("Failed to send DEL: %v", err)
		}
		writer.Flush()

		delResponse, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read DEL response: %v", err)
		}

		if delResponse != ":1\r\n" {
			t.Errorf("Expected :1\\r\\n, got %q", delResponse)
		}

		// Verify key is deleted
		existsCmd2 := fmt.Sprintf("*2\r\n$6\r\nEXISTS\r\n$%d\r\n%s\r\n", len(key), key)
		writer.WriteString(existsCmd2)
		writer.Flush()

		existsResponse2, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read second EXISTS response: %v", err)
		}

		if existsResponse2 != ":0\r\n" {
			t.Errorf("Expected :0\\r\\n, got %q", existsResponse2)
		}
	})
}

// TestMultipleCommandsInSequence tests executing multiple commands in sequence
func TestMultipleCommandsInSequence(t *testing.T) {
	// Create server with test configuration
	config := &server.ServerConfig{
		Port:         0, // Use random available port
		MaxClients:   10,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	srv := server.NewServer(config)

	// Start server
	err := srv.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Get the actual port the server is listening on
	addr := srv.GetListener().Addr().(*net.TCPAddr)
	port := addr.Port

	// Connect to server
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Execute a sequence of commands
	commands := []struct {
		name     string
		cmd      string
		expected string
	}{
		{"PING", "*1\r\n$4\r\nPING\r\n", "+PONG\r\n"},
		{"SET key1", "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$6\r\nvalue1\r\n", "+OK\r\n"},
		{"SET key2", "*3\r\n$3\r\nSET\r\n$4\r\nkey2\r\n$6\r\nvalue2\r\n", "+OK\r\n"},
		{"EXISTS key1", "*2\r\n$6\r\nEXISTS\r\n$4\r\nkey1\r\n", ":1\r\n"},
		{"EXISTS key1 key2", "*3\r\n$6\r\nEXISTS\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n", ":2\r\n"},
		{"DEL key1", "*2\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n", ":1\r\n"},
		{"EXISTS key1", "*2\r\n$6\r\nEXISTS\r\n$4\r\nkey1\r\n", ":0\r\n"},
		{"EXISTS key2", "*2\r\n$6\r\nEXISTS\r\n$4\r\nkey2\r\n", ":1\r\n"},
	}

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			// Send command
			_, err := writer.WriteString(cmd.cmd)
			if err != nil {
				t.Fatalf("Failed to send %s: %v", cmd.name, err)
			}
			writer.Flush()

			// Read response
			response, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("Failed to read %s response: %v", cmd.name, err)
			}

			if response != cmd.expected {
				t.Errorf("Command %s: expected %q, got %q", cmd.name, cmd.expected, response)
			}
		})
	}

	// Test GET for remaining key
	t.Run("GET key2", func(t *testing.T) {
		getCmd := "*2\r\n$3\r\nGET\r\n$4\r\nkey2\r\n"
		writer.WriteString(getCmd)
		writer.Flush()

		// Read response header
		header, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read GET response header: %v", err)
		}

		if header != "$6\r\n" {
			t.Errorf("Expected $6\\r\\n, got %q", header)
		}

		// Read response content
		content, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read GET response content: %v", err)
		}

		if content != "value2\r\n" {
			t.Errorf("Expected value2\\r\\n, got %q", content)
		}
	})
}

// TestConcurrentClientScenarios tests multiple clients connecting and operating concurrently
func TestConcurrentClientScenarios(t *testing.T) {
	// Create server with test configuration
	config := &server.ServerConfig{
		Port:         0, // Use random available port
		MaxClients:   20,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	srv := server.NewServer(config)

	// Start server
	err := srv.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Get the actual port the server is listening on
	addr := srv.GetListener().Addr().(*net.TCPAddr)
	port := addr.Port

	// Test concurrent clients performing different operations
	t.Run("concurrent SET/GET operations", func(t *testing.T) {
		const numClients = 5
		var wg sync.WaitGroup
		errors := make(chan error, numClients)

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()

				// Connect to server
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				if err != nil {
					errors <- fmt.Errorf("client %d failed to connect: %v", clientID, err)
					return
				}
				defer conn.Close()

				reader := bufio.NewReader(conn)
				writer := bufio.NewWriter(conn)

				// Each client sets and gets its own key
				key := fmt.Sprintf("client%d", clientID)
				value := fmt.Sprintf("value%d", clientID)

				// SET command
				setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
				_, err = writer.WriteString(setCmd)
				if err != nil {
					errors <- fmt.Errorf("client %d failed to send SET: %v", clientID, err)
					return
				}
				writer.Flush()

				// Read SET response
				setResponse, err := reader.ReadString('\n')
				if err != nil {
					errors <- fmt.Errorf("client %d failed to read SET response: %v", clientID, err)
					return
				}

				if setResponse != "+OK\r\n" {
					errors <- fmt.Errorf("client %d SET failed: expected +OK\\r\\n, got %q", clientID, setResponse)
					return
				}

				// GET command
				getCmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
				_, err = writer.WriteString(getCmd)
				if err != nil {
					errors <- fmt.Errorf("client %d failed to send GET: %v", clientID, err)
					return
				}
				writer.Flush()

				// Read GET response header
				getHeader, err := reader.ReadString('\n')
				if err != nil {
					errors <- fmt.Errorf("client %d failed to read GET header: %v", clientID, err)
					return
				}

				expectedHeader := fmt.Sprintf("$%d\r\n", len(value))
				if getHeader != expectedHeader {
					errors <- fmt.Errorf("client %d GET header mismatch: expected %q, got %q", clientID, expectedHeader, getHeader)
					return
				}

				// Read GET response content
				getContent, err := reader.ReadString('\n')
				if err != nil {
					errors <- fmt.Errorf("client %d failed to read GET content: %v", clientID, err)
					return
				}

				expectedContent := value + "\r\n"
				if getContent != expectedContent {
					errors <- fmt.Errorf("client %d GET content mismatch: expected %q, got %q", clientID, expectedContent, getContent)
					return
				}
			}(i)
		}

		// Wait for all clients to complete
		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})

	// Test concurrent PING operations
	t.Run("concurrent PING operations", func(t *testing.T) {
		const numClients = 10
		var wg sync.WaitGroup
		errors := make(chan error, numClients)

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()

				// Connect to server
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				if err != nil {
					errors <- fmt.Errorf("client %d failed to connect: %v", clientID, err)
					return
				}
				defer conn.Close()

				reader := bufio.NewReader(conn)
				writer := bufio.NewWriter(conn)

				// Send multiple PING commands
				for j := 0; j < 3; j++ {
					pingCmd := "*1\r\n$4\r\nPING\r\n"
					_, err = writer.WriteString(pingCmd)
					if err != nil {
						errors <- fmt.Errorf("client %d failed to send PING %d: %v", clientID, j, err)
						return
					}
					writer.Flush()

					// Read response
					response, err := reader.ReadString('\n')
					if err != nil {
						errors <- fmt.Errorf("client %d failed to read PING %d response: %v", clientID, j, err)
						return
					}

					if response != "+PONG\r\n" {
						errors <- fmt.Errorf("client %d PING %d failed: expected +PONG\\r\\n, got %q", clientID, j, response)
						return
					}
				}
			}(i)
		}

		// Wait for all clients to complete
		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})

	// Test mixed concurrent operations
	t.Run("mixed concurrent operations", func(t *testing.T) {
		const numClients = 8
		var wg sync.WaitGroup
		errors := make(chan error, numClients)

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()

				// Connect to server
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				if err != nil {
					errors <- fmt.Errorf("client %d failed to connect: %v", clientID, err)
					return
				}
				defer conn.Close()

				reader := bufio.NewReader(conn)
				writer := bufio.NewWriter(conn)

				// Each client performs different operations based on ID
				switch clientID % 4 {
				case 0: // SET operations
					key := fmt.Sprintf("mixkey%d", clientID)
					value := fmt.Sprintf("mixvalue%d", clientID)
					setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
					writer.WriteString(setCmd)
					writer.Flush()
					reader.ReadString('\n') // Read response

				case 1: // PING operations
					pingCmd := "*1\r\n$4\r\nPING\r\n"
					writer.WriteString(pingCmd)
					writer.Flush()
					response, err := reader.ReadString('\n')
					if err != nil || response != "+PONG\r\n" {
						errors <- fmt.Errorf("client %d PING failed", clientID)
						return
					}

				case 2: // EXISTS operations
					key := fmt.Sprintf("mixkey%d", clientID-2) // Check key from client ID-2
					existsCmd := fmt.Sprintf("*2\r\n$6\r\nEXISTS\r\n$%d\r\n%s\r\n", len(key), key)
					writer.WriteString(existsCmd)
					writer.Flush()
					reader.ReadString('\n') // Read response

				case 3: // GET operations
					key := fmt.Sprintf("mixkey%d", clientID-3) // Get key from client ID-3
					getCmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
					writer.WriteString(getCmd)
					writer.Flush()
					reader.ReadString('\n') // Read header
					// May need to read content if key exists, but we'll skip for simplicity
				}
			}(i)
		}

		// Wait for all clients to complete
		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})
}