package handler

import (
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
	// Implementation will be added in later tasks
	return nil
}

// handlePing handles PING commands
func (h *DefaultCommandHandler) handlePing(args []string) *resp2.RESPValue {
	// Implementation will be added in later tasks
	return nil
}

// handleSet handles SET commands
func (h *DefaultCommandHandler) handleSet(args []string) *resp2.RESPValue {
	// Implementation will be added in later tasks
	return nil
}

// handleGet handles GET commands
func (h *DefaultCommandHandler) handleGet(args []string) *resp2.RESPValue {
	// Implementation will be added in later tasks
	return nil
}

// handleExists handles EXISTS commands
func (h *DefaultCommandHandler) handleExists(args []string) *resp2.RESPValue {
	// Implementation will be added in later tasks
	return nil
}

// handleDel handles DEL commands
func (h *DefaultCommandHandler) handleDel(args []string) *resp2.RESPValue {
	// Implementation will be added in later tasks
	return nil
}