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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
}

// Get retrieves a value by key
func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

// Exists checks if a key exists
func (s *InMemoryStore) Exists(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.data[key]
	return exists
}

// Delete removes a key
func (s *InMemoryStore) Delete(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

// DeleteMultiple removes multiple keys and returns count of deleted keys
func (s *InMemoryStore) DeleteMultiple(keys []string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	deletedCount := 0
	for _, key := range keys {
		if _, exists := s.data[key]; exists {
			delete(s.data, key)
			deletedCount++
		}
	}
	return deletedCount
}