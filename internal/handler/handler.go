package handler

import (
	"fmt"
	"redis-like-server/internal/resp2"
	"redis-like-server/internal/store"
)

// CommandHandler routes and executes Redis commands
type CommandHandler interface {
	Execute(cmd *resp2.Command) *resp2.RESPValue
}

// DefaultCommandHandler is the default implementation of CommandHandler
type DefaultCommandHandler struct {
	store store.KeyValueStore
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(store store.KeyValueStore) CommandHandler {
	return &DefaultCommandHandler{
		store: store,
	}
}

// Execute processes and executes a command
func (h *DefaultCommandHandler) Execute(cmd *resp2.Command) *resp2.RESPValue {
	if cmd == nil {
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  "ERR command cannot be nil",
		}
	}
	
	switch cmd.Name {
	case "PING":
		return h.handlePing(cmd.Args)
	case "SET":
		return h.handleSet(cmd.Args)
	case "GET":
		return h.handleGet(cmd.Args)
	case "EXISTS":
		return h.handleExists(cmd.Args)
	case "DEL":
		return h.handleDel(cmd.Args)
	default:
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  fmt.Sprintf("ERR unknown command '%s'", cmd.Name),
		}
	}
}

// handlePing handles PING commands
func (h *DefaultCommandHandler) handlePing(args []string) *resp2.RESPValue {
	if len(args) == 0 {
		// PING with no arguments returns PONG
		return &resp2.RESPValue{
			Type: resp2.SimpleString,
			Str:  "PONG",
		}
	} else if len(args) == 1 {
		// PING with message argument echoes the message
		return &resp2.RESPValue{
			Type: resp2.BulkString,
			Str:  args[0],
		}
	} else {
		// Too many arguments
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  "ERR wrong number of arguments for 'PING' command",
		}
	}
}

// handleSet handles SET commands
func (h *DefaultCommandHandler) handleSet(args []string) *resp2.RESPValue {
	if len(args) != 2 {
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  "ERR wrong number of arguments for 'SET' command",
		}
	}
	
	key := args[0]
	value := args[1]
	
	// Store the key-value pair
	h.store.Set(key, value)
	
	// Return OK response
	return &resp2.RESPValue{
		Type: resp2.SimpleString,
		Str:  "OK",
	}
}

// handleGet handles GET commands
func (h *DefaultCommandHandler) handleGet(args []string) *resp2.RESPValue {
	if len(args) != 1 {
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  "ERR wrong number of arguments for 'GET' command",
		}
	}
	
	key := args[0]
	
	// Retrieve value from store
	value, exists := h.store.Get(key)
	
	if exists {
		// Return the value as bulk string
		return &resp2.RESPValue{
			Type: resp2.BulkString,
			Str:  value,
		}
	} else {
		// Return null bulk string for non-existing keys
		return &resp2.RESPValue{
			Type: resp2.NullBulkString,
			Null: true,
		}
	}
}

// handleExists handles EXISTS commands
func (h *DefaultCommandHandler) handleExists(args []string) *resp2.RESPValue {
	if len(args) == 0 {
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  "ERR wrong number of arguments for 'EXISTS' command",
		}
	}
	
	count := int64(0)
	
	// Check each key
	for _, key := range args {
		if h.store.Exists(key) {
			count++
		}
	}
	
	// Return count as integer
	return &resp2.RESPValue{
		Type: resp2.Integer,
		Int:  count,
	}
}

// handleDel handles DEL commands
func (h *DefaultCommandHandler) handleDel(args []string) *resp2.RESPValue {
	if len(args) == 0 {
		return &resp2.RESPValue{
			Type: resp2.Error,
			Str:  "ERR wrong number of arguments for 'DEL' command",
		}
	}
	
	var deletedCount int64
	
	if len(args) == 1 {
		// Single key deletion
		if h.store.Delete(args[0]) {
			deletedCount = 1
		} else {
			deletedCount = 0
		}
	} else {
		// Multiple key deletion
		deletedCount = int64(h.store.DeleteMultiple(args))
	}
	
	// Return count as integer
	return &resp2.RESPValue{
		Type: resp2.Integer,
		Int:  deletedCount,
	}
}