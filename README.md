# Redis-like Server

A Redis-like server implementation in Go that supports concurrent client connections and implements core Redis commands using the RESP2 protocol.

## Project Structure

```
redis-like-server/
├── main.go                           # Application entry point
├── go.mod                           # Go module definition
├── internal/
│   ├── server/                      # Main server component
│   │   ├── server.go               # Server implementation
│   │   └── server_test.go          # Server tests
│   ├── resp2/                       # RESP2 protocol parser
│   │   ├── parser.go               # Parser implementation
│   │   └── parser_test.go          # Parser tests
│   ├── store/                       # Key-value store
│   │   ├── store.go                # Store implementation
│   │   └── store_test.go           # Store tests
│   ├── handler/                     # Command handler
│   │   ├── handler.go              # Handler implementation
│   │   └── handler_test.go         # Handler tests
│   └── connection/                  # Connection management
│       ├── manager.go              # Connection manager implementation
│       └── manager_test.go         # Connection manager tests
└── .kiro/specs/redis-like-server/   # Specification documents
    ├── requirements.md              # Requirements document
    ├── design.md                   # Design document
    └── tasks.md                    # Implementation tasks
```

## Features

- **Concurrent Client Support**: Handle multiple clients simultaneously
- **RESP2 Protocol**: Full Redis Serialization Protocol v2 support
- **Core Commands**: PING, SET, GET, EXISTS, DEL
- **Thread-Safe Storage**: Concurrent access to key-value store
- **Property-Based Testing**: Comprehensive correctness validation
- **Graceful Shutdown**: Clean resource management

## Building and Running

```bash
# Build the server
go build -o redis-server

# Run the server (default port 6379)
./redis-server

# Run with custom port
./redis-server -port 8080

# Run tests
go test -v ./...
```

## Command Line Options

- `-port`: Port to listen on (default: 6379)
- `-max-clients`: Maximum concurrent clients (default: 1000)
- `-read-timeout`: Read timeout for connections (default: 30s)
- `-write-timeout`: Write timeout for connections (default: 30s)

## Development Status

This project follows a spec-driven development approach. See `.kiro/specs/redis-like-server/tasks.md` for the current implementation status and next steps.
