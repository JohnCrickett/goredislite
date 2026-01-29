package store

import (
	"sync"
)

// KeyValueStore provides thread-safe key-value storage operations
type KeyValueStore interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Exists(key string) bool
	Delete(key string) bool
	DeleteMultiple(keys []string) int
}

// InMemoryStore is an in-memory implementation of KeyValueStore
type InMemoryStore struct {
	data  map[string]string
	mutex sync.RWMutex
}

// NewInMemoryStore creates a new in-memory key-value store
func NewInMemoryStore() KeyValueStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

// Set stores a key-value pair
func (s *InMemoryStore) Set(key, value string) {
	// Implementation will be added in later tasks
}

// Get retrieves a value by key
func (s *InMemoryStore) Get(key string) (string, bool) {
	// Implementation will be added in later tasks
	return "", false
}

// Exists checks if a key exists
func (s *InMemoryStore) Exists(key string) bool {
	// Implementation will be added in later tasks
	return false
}

// Delete removes a key
func (s *InMemoryStore) Delete(key string) bool {
	// Implementation will be added in later tasks
	return false
}

// DeleteMultiple removes multiple keys and returns count of deleted keys
func (s *InMemoryStore) DeleteMultiple(keys []string) int {
	// Implementation will be added in later tasks
	return 0
}