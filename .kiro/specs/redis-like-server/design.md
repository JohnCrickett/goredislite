# Design Document: Redis-like Server

## Overview

This document outlines the design for a Redis-like server implementation in Go that supports concurrent client connections and implements core Redis commands using the RESP2 protocol. The server will provide a simplified but functional subset of Redis capabilities, focusing on reliability, performance, and protocol compliance.

The server architecture follows a concurrent, event-driven design where each client connection is handled in its own goroutine, allowing for true concurrent processing while maintaining thread-safe access to the shared key-value store.

## Architecture

The system follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Connections                       │
│                  (Multiple TCP clients)                     │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  Connection Manager                         │
│              (Accept & manage connections)                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                    RESP2 Parser                            │
│              (Parse/Format RESP2 messages)                 │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  Command Handler                           │
│                (Route & execute commands)                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                 Key-Value Store                            │
│              (Thread-safe data storage)                    │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Concurrency**: Each client connection runs in its own goroutine
2. **Thread Safety**: Shared data structures use appropriate synchronization
3. **Protocol Compliance**: Full RESP2 protocol support for implemented commands
4. **Error Resilience**: Graceful error handling without server crashes
5. **Resource Management**: Proper cleanup of connections and resources

## Components and Interfaces

### Server Component

The main server component orchestrates the entire system:

```go
type Server struct {
    listener    net.Listener
    store       *KeyValueStore
    parser      *RESP2Parser
    handler     *CommandHandler
    connManager *ConnectionManager
    config      *ServerConfig
}

type ServerConfig struct {
    Port         int
    MaxClients   int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}
```

**Key Methods:**
- `Start()`: Initialize and start the server
- `Stop()`: Gracefully shutdown the server
- `handleConnection(conn net.Conn)`: Handle individual client connections

### Connection Manager

Manages all active client connections and their lifecycle:

```go
type ConnectionManager struct {
    connections map[string]*ClientConnection
    mutex       sync.RWMutex
    maxClients  int
}

type ClientConnection struct {
    conn       net.Conn
    id         string
    lastActive time.Time
    reader     *bufio.Reader
    writer     *bufio.Writer
}
```

**Key Methods:**
- `AddConnection(conn net.Conn) *ClientConnection`
- `RemoveConnection(id string)`
- `GetActiveCount() int`
- `CleanupStaleConnections()`

### RESP2 Parser

Handles parsing and formatting of RESP2 protocol messages:

```go
type RESP2Parser struct{}

type RESPValue struct {
    Type  RESPType
    Str   string
    Int   int64
    Array []RESPValue
    Null  bool
}

type RESPType int

const (
    SimpleString RESPType = iota
    Error
    Integer
    BulkString
    Array
    NullBulkString
)
```

**Key Methods:**
- `Parse(reader *bufio.Reader) (*RESPValue, error)`
- `Serialize(value *RESPValue) []byte`
- `ParseCommand(value *RESPValue) (*Command, error)`

### Command Handler

Routes and executes Redis commands:

```go
type CommandHandler struct {
    store *KeyValueStore
}

type Command struct {
    Name string
    Args []string
}
```

**Key Methods:**
- `Execute(cmd *Command) *RESPValue`
- `handlePing(args []string) *RESPValue`
- `handleSet(args []string) *RESPValue`
- `handleGet(args []string) *RESPValue`
- `handleExists(args []string) *RESPValue`
- `handleDel(args []string) *RESPValue`

### Key-Value Store

Thread-safe in-memory storage for key-value pairs:

```go
type KeyValueStore struct {
    data  map[string]string
    mutex sync.RWMutex
}
```

**Key Methods:**
- `Set(key, value string)`
- `Get(key string) (string, bool)`
- `Exists(key string) bool`
- `Delete(key string) bool`
- `DeleteMultiple(keys []string) int`

## Data Models

### RESP2 Protocol Data Types

The server supports the following RESP2 data types as specified in the protocol:

1. **Simple Strings** (`+`): For simple responses like "OK" and "PONG"
2. **Errors** (`-`): For error messages with proper error prefixes
3. **Integers** (`:`) For numeric responses like counts and boolean values
4. **Bulk Strings** (`$`): For string data that may contain binary content
5. **Arrays** (`*`): For commands and multi-element responses
6. **Null Bulk Strings** (`$-1`): For representing null/nil values

### Command Structure

All commands follow the RESP2 array format:
```
*<number-of-elements>\r\n
$<length-of-element-1>\r\n<element-1>\r\n
$<length-of-element-2>\r\n<element-2>\r\n
...
```

### Storage Model

The key-value store uses a simple string-to-string mapping:
- **Keys**: UTF-8 strings (any valid string)
- **Values**: UTF-8 strings (binary-safe through RESP2 bulk strings)
- **Concurrency**: Protected by read-write mutex for thread safety

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: RESP2 Protocol Round-trip Consistency
*For any* valid RESP2 data structure (Simple String, Error, Integer, Bulk String, or Array), serializing then parsing should produce an equivalent structure
**Validates: Requirements 2.1, 2.2, 2.4**

### Property 2: Connection Lifecycle Management
*For any* client connection, the server should properly accept, maintain, and clean up the connection resources when it disconnects
**Validates: Requirements 1.2, 1.4**

### Property 3: Concurrent Client Processing
*For any* set of concurrent clients, the server should process all their commands without blocking and maintain data consistency across all operations
**Validates: Requirements 1.3, 6.1, 6.2, 6.3**

### Property 4: PING Command Echo Behavior
*For any* message argument provided to PING, the server should echo back exactly the same message
**Validates: Requirements 3.2**

### Property 5: PING Command Reliability
*For any* server state, PING commands should always be processed and return appropriate responses
**Validates: Requirements 3.3**

### Property 6: SET-GET Round-trip Consistency
*For any* key-value pair, setting then getting the key should return the same value that was set
**Validates: Requirements 4.1, 4.2, 4.5**

### Property 7: GET Null Behavior
*For any* non-existing key, GET should return null (RESP2 null bulk string)
**Validates: Requirements 4.3**

### Property 8: EXISTS Count Accuracy
*For any* set of keys, EXISTS should return the exact count of keys that actually exist in the store
**Validates: Requirements 5.1, 5.2**

### Property 9: DEL Count Accuracy
*For any* set of keys, DEL should return the exact count of keys that were actually deleted from the store
**Validates: Requirements 5.3, 5.4**

### Property 10: Error Handling Robustness
*For any* invalid command, malformed input, or error condition, the server should return appropriate RESP2 error responses and continue serving other clients
**Validates: Requirements 2.3, 7.1, 7.2, 7.3, 7.4**

### Property 11: Connection Cleanup on Unexpected Disconnection
*For any* client that disconnects unexpectedly, the server should detect and clean up orphaned connection resources
**Validates: Requirements 6.4**

### Property 12: String Value Storage Capacity
*For any* string value within memory constraints, the key-value store should be able to store and retrieve it correctly
**Validates: Requirements 4.4**

## Error Handling

The server implements comprehensive error handling at multiple levels:

### Protocol Level Errors
- **Malformed RESP2**: Return `-ERR Protocol error: invalid RESP2 format`
- **Invalid Data Types**: Return `-ERR Protocol error: unsupported data type`
- **Parsing Failures**: Return `-ERR Protocol error: failed to parse command`

### Command Level Errors
- **Unknown Commands**: Return `-ERR unknown command 'COMMANDNAME'`
- **Wrong Argument Count**: Return `-ERR wrong number of arguments for 'COMMANDNAME' command`
- **Invalid Arguments**: Return `-ERR invalid argument format`

### System Level Errors
- **Memory Exhaustion**: Return `-ERR out of memory`
- **Connection Limits**: Reject new connections with proper TCP close
- **Internal Errors**: Log error and return `-ERR internal server error`

### Error Response Format
All errors follow RESP2 error format: `-ERROR_PREFIX error message\r\n`

Common error prefixes:
- `ERR`: Generic errors
- `WRONGTYPE`: Type-related errors (not applicable for our string-only store)
- `NOAUTH`: Authentication errors (not implemented)

## Testing Strategy

The testing approach combines unit tests for specific behaviors with property-based tests for comprehensive validation:

### Unit Testing
Unit tests focus on:
- **Specific Examples**: Test concrete cases like `PING` → `PONG`
- **Edge Cases**: Empty strings, special characters, boundary conditions
- **Error Conditions**: Invalid commands, malformed input, connection failures
- **Integration Points**: Component interactions and data flow

### Property-Based Testing
Property tests validate universal behaviors using Go's testing framework with a property-based testing library like `gopter` or `quick`:

- **Minimum 100 iterations** per property test to ensure comprehensive coverage
- **Random Input Generation**: Generate diverse test cases automatically
- **Universal Properties**: Validate behaviors that should hold for all valid inputs
- **Concurrency Testing**: Use goroutines to test concurrent access patterns

### Test Configuration
Each property test must:
- Reference its corresponding design document property
- Use tag format: **Feature: redis-like-server, Property {number}: {property_text}**
- Run minimum 100 iterations due to randomization
- Include proper cleanup and resource management

### Test Categories

**RESP2 Protocol Tests:**
- Round-trip parsing and serialization
- All data type handling
- Error response formatting

**Command Behavior Tests:**
- SET/GET operations with various key-value combinations
- EXISTS and DEL with multiple keys
- PING with different argument patterns

**Concurrency Tests:**
- Multiple clients performing operations simultaneously
- Data consistency under concurrent access
- Connection management under load

**Error Handling Tests:**
- Invalid command handling
- Malformed input processing
- Resource exhaustion scenarios

**Performance Baseline Tests:**
- Response time measurements for simple commands
- Memory usage monitoring
- Connection handling capacity

The dual testing approach ensures both correctness (property tests) and reliability (unit tests), providing comprehensive validation of the Redis-like server implementation.