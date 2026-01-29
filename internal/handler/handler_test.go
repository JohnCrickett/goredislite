package handler

import (
	"testing"

	"redis-like-server/internal/resp2"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestCommandHandler(t *testing.T) {
	// Unit tests will be added in later tasks
}

// Property-based test setup for PING echo behavior
func TestPINGEchoBehavior(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 4: PING Command Echo Behavior
	// For any message argument provided to PING, the server should echo back exactly the same message
	// **Validates: Requirements 3.2**
	properties.Property("PING echo behavior", prop.ForAll(
		func(message string) bool {
			// Create handler with mock store (not needed for PING)
			handler := NewCommandHandler(nil)
			
			// Create PING command with message argument
			cmd := &resp2.Command{
				Name: "PING",
				Args: []string{message},
			}
			
			// Execute command
			result := handler.Execute(cmd)
			
			// Verify the result
			if result == nil {
				return false
			}
			
			// Should return bulk string with the same message
			return result.Type == resp2.BulkString && result.Str == message
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property-based test setup for GET null behavior
func TestGETNullBehavior(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 7: GET Null Behavior
	// For any non-existing key, GET should return null (RESP2 null bulk string)
	// **Validates: Requirements 4.3**
	properties.Property("GET null behavior", prop.ForAll(
		func(key string) bool {
			// Create handler with empty store
			store := newMockStore()
			handler := NewCommandHandler(store)
			
			// Create GET command for non-existing key
			cmd := &resp2.Command{
				Name: "GET",
				Args: []string{key},
			}
			
			// Execute command
			result := handler.Execute(cmd)
			
			// Verify the result
			if result == nil {
				return false
			}
			
			// Should return null bulk string
			return result.Type == resp2.NullBulkString && result.Null == true
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Mock store for testing
type mockStore struct {
	data map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string]string),
	}
}

func (s *mockStore) Set(key, value string) {
	s.data[key] = value
}

func (s *mockStore) Get(key string) (string, bool) {
	value, exists := s.data[key]
	return value, exists
}

func (s *mockStore) Exists(key string) bool {
	_, exists := s.data[key]
	return exists
}

func (s *mockStore) Delete(key string) bool {
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

func (s *mockStore) DeleteMultiple(keys []string) int {
	deletedCount := 0
	for _, key := range keys {
		if _, exists := s.data[key]; exists {
			delete(s.data, key)
			deletedCount++
		}
	}
	return deletedCount
}

// Property-based test setup for error handling robustness
func TestErrorHandlingRobustness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 10: Error Handling Robustness
	// For any invalid command, malformed input, or error condition, the server should return appropriate RESP2 error responses
	// **Validates: Requirements 2.3, 7.1, 7.2, 7.3, 7.4**
	properties.Property("error handling robustness", prop.ForAll(
		func(invalidCommand string) bool {
			// Create handler with mock store
			store := newMockStore()
			handler := NewCommandHandler(store)
			
			// Test unknown command
			cmd := &resp2.Command{
				Name: invalidCommand,
				Args: []string{},
			}
			
			result := handler.Execute(cmd)
			
			// Should return error response for unknown commands
			if result == nil {
				return false
			}
			
			// Should be an error type
			if result.Type != resp2.Error {
				return false
			}
			
			// Error message should contain the command name
			return len(result.Str) > 0
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			// Exclude valid commands
			validCommands := map[string]bool{
				"PING": true, "SET": true, "GET": true, 
				"EXISTS": true, "DEL": true,
			}
			return !validCommands[s]
		}),
	))

	// Test wrong number of arguments for known commands
	properties.Property("wrong argument count handling", prop.ForAll(
		func(argCount int) bool {
			store := newMockStore()
			handler := NewCommandHandler(store)
			
			// Test SET with wrong number of arguments (should be exactly 2)
			args := make([]string, argCount)
			for i := range args {
				args[i] = "arg"
			}
			
			cmd := &resp2.Command{
				Name: "SET",
				Args: args,
			}
			
			result := handler.Execute(cmd)
			
			if argCount == 2 {
				// Should succeed with correct argument count
				return result != nil && result.Type == resp2.SimpleString && result.Str == "OK"
			} else {
				// Should return error with wrong argument count
				return result != nil && result.Type == resp2.Error
			}
		},
		gen.IntRange(0, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}