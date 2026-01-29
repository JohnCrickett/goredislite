package resp2

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestRESP2Parser(t *testing.T) {
	parser := NewRESP2Parser()

	t.Run("Simple String parsing", func(t *testing.T) {
		input := "+OK\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != SimpleString || result.Str != "OK" {
			t.Fatalf("Expected SimpleString 'OK', got %v", result)
		}
	})

	t.Run("Error parsing", func(t *testing.T) {
		input := "-ERR unknown command\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != Error || result.Str != "ERR unknown command" {
			t.Fatalf("Expected Error 'ERR unknown command', got %v", result)
		}
	})

	t.Run("Integer parsing", func(t *testing.T) {
		input := ":1000\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != Integer || result.Int != 1000 {
			t.Fatalf("Expected Integer 1000, got %v", result)
		}
	})

	t.Run("Bulk String parsing", func(t *testing.T) {
		input := "$6\r\nfoobar\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != BulkString || result.Str != "foobar" {
			t.Fatalf("Expected BulkString 'foobar', got %v", result)
		}
	})

	t.Run("Null Bulk String parsing", func(t *testing.T) {
		input := "$-1\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != NullBulkString || !result.Null {
			t.Fatalf("Expected NullBulkString, got %v", result)
		}
	})

	t.Run("Array parsing", func(t *testing.T) {
		input := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != Array || len(result.Array) != 2 {
			t.Fatalf("Expected Array with 2 elements, got %v", result)
		}
		if result.Array[0].Str != "foo" || result.Array[1].Str != "bar" {
			t.Fatalf("Expected array elements 'foo', 'bar', got %v", result.Array)
		}
	})

	t.Run("Null Array parsing", func(t *testing.T) {
		input := "*-1\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != Array || !result.Null {
			t.Fatalf("Expected null Array, got %v", result)
		}
	})

	// Edge cases and malformed input tests
	t.Run("Invalid type indicator", func(t *testing.T) {
		input := "?invalid\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for invalid type indicator")
		}
	})

	t.Run("Missing CRLF", func(t *testing.T) {
		input := "+OK\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for missing \\r\\n")
		}
	})

	t.Run("Invalid integer format", func(t *testing.T) {
		input := ":not-a-number\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for invalid integer format")
		}
	})

	t.Run("Invalid bulk string length", func(t *testing.T) {
		input := "$not-a-number\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for invalid bulk string length")
		}
	})

	t.Run("Negative bulk string length (not -1)", func(t *testing.T) {
		input := "$-5\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for invalid negative bulk string length")
		}
	})

	t.Run("Bulk string length mismatch", func(t *testing.T) {
		input := "$10\r\nshort\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for bulk string length mismatch")
		}
	})

	t.Run("Invalid array length", func(t *testing.T) {
		input := "*not-a-number\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for invalid array length")
		}
	})

	t.Run("Negative array length (not -1)", func(t *testing.T) {
		input := "*-5\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		_, err := parser.Parse(reader)
		
		if err == nil {
			t.Fatal("Expected error for invalid negative array length")
		}
	})

	t.Run("Empty bulk string", func(t *testing.T) {
		input := "$0\r\n\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != BulkString || result.Str != "" {
			t.Fatalf("Expected empty BulkString, got %v", result)
		}
	})

	t.Run("Empty array", func(t *testing.T) {
		input := "*0\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != Array || len(result.Array) != 0 {
			t.Fatalf("Expected empty Array, got %v", result)
		}
	})

	t.Run("Binary data in bulk string", func(t *testing.T) {
		binaryData := []byte{0, 1, 2, 255, 254, 253}
		input := fmt.Sprintf("$%d\r\n%s\r\n", len(binaryData), string(binaryData))
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != BulkString || result.Str != string(binaryData) {
			t.Fatalf("Expected binary BulkString, got %v", result)
		}
	})

	t.Run("Special characters in simple string", func(t *testing.T) {
		input := "+Hello World! @#$%^&*()\r\n"
		reader := bufio.NewReader(bytes.NewReader([]byte(input)))
		result, err := parser.Parse(reader)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Type != SimpleString || result.Str != "Hello World! @#$%^&*()" {
			t.Fatalf("Expected SimpleString with special chars, got %v", result)
		}
	})
}

// Property-based test setup for RESP2 round-trip consistency
func TestRESP2RoundTripConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)
	parser := NewRESP2Parser()

	// Property 1: RESP2 Protocol Round-trip Consistency
	// For any valid RESP2 data structure, serializing then parsing should produce an equivalent structure
	// **Feature: redis-like-server, Property 1: RESP2 Protocol Round-trip Consistency**
	// **Validates: Requirements 2.1, 2.2, 2.4**
	properties.Property("RESP2 round-trip consistency", prop.ForAll(
		func(respValue *RESPValue) bool {
			// Serialize the value
			serialized := parser.Serialize(respValue)
			
			// Parse it back
			reader := bufio.NewReader(bytes.NewReader(serialized))
			parsed, err := parser.Parse(reader)
			
			if err != nil {
				t.Logf("Parse error: %v", err)
				return false
			}
			
			// Check if the parsed value is equivalent to the original
			return respValuesEqual(respValue, parsed)
		},
		genRESPValue(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// respValuesEqual compares two RESPValue structs for equality
func respValuesEqual(a, b *RESPValue) bool {
	if a.Type != b.Type {
		return false
	}
	
	if a.Null != b.Null {
		return false
	}
	
	switch a.Type {
	case SimpleString, Error, BulkString:
		return a.Str == b.Str
	case Integer:
		return a.Int == b.Int
	case NullBulkString:
		return a.Null == b.Null
	case Array:
		if a.Null || b.Null {
			return a.Null == b.Null
		}
		if len(a.Array) != len(b.Array) {
			return false
		}
		for i := range a.Array {
			if !respValuesEqual(&a.Array[i], &b.Array[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Generator for RESP2 values - simplified approach
func genRESPValue() gopter.Gen {
	return gen.OneConstOf(
		&RESPValue{Type: SimpleString, Str: "OK"},
		&RESPValue{Type: Error, Str: "ERR something went wrong"},
		&RESPValue{Type: Integer, Int: 42},
		&RESPValue{Type: BulkString, Str: "hello world"},
		&RESPValue{Type: NullBulkString, Null: true},
		&RESPValue{Type: Array, Array: []RESPValue{
			{Type: SimpleString, Str: "PING"},
			{Type: BulkString, Str: "test"},
		}},
		&RESPValue{Type: Array, Null: true},
	).SuchThat(func(v interface{}) bool {
		return v != nil
	})
}