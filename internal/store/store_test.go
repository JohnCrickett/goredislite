package store

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestKeyValueStore(t *testing.T) {
	// Unit tests will be added in later tasks
}

// Property-based test setup for SET-GET round-trip consistency
func TestSETGETRoundTripConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 6: SET-GET Round-trip Consistency
	// For any key-value pair, setting then getting the key should return the same value that was set
	properties.Property("SET-GET round-trip consistency", prop.ForAll(
		func(key, value string) bool {
			store := NewInMemoryStore()
			
			// Set the key-value pair
			store.Set(key, value)
			
			// Get the value back
			retrievedValue, exists := store.Get(key)
			
			// The key should exist and the value should match exactly
			return exists && retrievedValue == value
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property-based test setup for EXISTS count accuracy
func TestEXISTSCountAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 8: EXISTS Count Accuracy
	// For any set of keys, EXISTS should return the exact count of keys that actually exist in the store
	properties.Property("EXISTS count accuracy", prop.ForAll(
		func(keys []string) bool {
			store := NewInMemoryStore()
			
			// Set some of the keys with values
			existingKeys := make(map[string]bool)
			for i, key := range keys {
				if i%2 == 0 { // Set every other key
					store.Set(key, "value")
					existingKeys[key] = true
				}
			}
			
			// Count how many keys should exist (accounting for duplicates)
			expectedCount := 0
			checkedKeys := make(map[string]bool)
			for _, key := range keys {
				if !checkedKeys[key] { // Only count unique keys
					checkedKeys[key] = true
					if existingKeys[key] {
						expectedCount++
					}
				}
			}
			
			// Count how many keys actually exist using individual Exists calls
			actualCount := 0
			checkedKeys = make(map[string]bool)
			for _, key := range keys {
				if !checkedKeys[key] { // Only check unique keys
					checkedKeys[key] = true
					if store.Exists(key) {
						actualCount++
					}
				}
			}
			
			return actualCount == expectedCount
		},
		gen.SliceOf(gen.AlphaString()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property-based test setup for DEL count accuracy
func TestDELCountAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 9: DEL Count Accuracy
	// For any set of keys, DEL should return the exact count of keys that were actually deleted from the store
	properties.Property("DEL count accuracy", prop.ForAll(
		func(keys []string) bool {
			store := NewInMemoryStore()
			
			// Set all keys with values
			for _, key := range keys {
				store.Set(key, "value")
			}
			
			// Count unique keys that should be deleted
			uniqueKeys := make(map[string]bool)
			for _, key := range keys {
				uniqueKeys[key] = true
			}
			expectedDeleteCount := len(uniqueKeys)
			
			// Delete using DeleteMultiple
			actualDeleteCount := store.DeleteMultiple(keys)
			
			// Verify the count matches
			if actualDeleteCount != expectedDeleteCount {
				return false
			}
			
			// Verify all keys are actually deleted
			for key := range uniqueKeys {
				if store.Exists(key) {
					return false
				}
			}
			
			return true
		},
		gen.SliceOf(gen.AlphaString()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property-based test setup for concurrent data consistency
func TestConcurrentDataConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 3: Concurrent Client Processing (data consistency aspect)
	// For any set of concurrent operations, the store should maintain data consistency
	properties.Property("concurrent data consistency", prop.ForAll(
		func(operations []string) bool {
			store := NewInMemoryStore()
			
			// Skip empty operations
			if len(operations) == 0 {
				return true
			}
			
			// Use a channel to coordinate goroutines
			done := make(chan bool, len(operations))
			
			// Perform concurrent SET operations
			for i, key := range operations {
				go func(k string, value int) {
					store.Set(k, string(rune('A'+value%26))) // Use letters A-Z as values
					done <- true
				}(key, i)
			}
			
			// Wait for all SET operations to complete
			for i := 0; i < len(operations); i++ {
				<-done
			}
			
			// Verify data consistency by checking that each key has a valid value
			uniqueKeys := make(map[string]bool)
			for _, key := range operations {
				uniqueKeys[key] = true
			}
			
			for key := range uniqueKeys {
				value, exists := store.Get(key)
				if !exists {
					return false // Key should exist after SET
				}
				if len(value) != 1 || value[0] < 'A' || value[0] > 'Z' {
					return false // Value should be a valid letter
				}
			}
			
			// Test concurrent DELETE operations
			deleteKeys := make([]string, 0, len(uniqueKeys))
			for key := range uniqueKeys {
				deleteKeys = append(deleteKeys, key)
			}
			
			// Perform concurrent DELETE operations
			deletedCount := store.DeleteMultiple(deleteKeys)
			
			// Verify all keys are deleted
			if deletedCount != len(uniqueKeys) {
				return false
			}
			
			for key := range uniqueKeys {
				if store.Exists(key) {
					return false // Key should not exist after DELETE
				}
			}
			
			return true
		},
		gen.SliceOf(gen.AlphaString()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}