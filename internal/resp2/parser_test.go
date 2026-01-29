package resp2

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestRESP2Parser(t *testing.T) {
	// Unit tests will be added in later tasks
}

// Property-based test setup for RESP2 round-trip consistency
func TestRESP2RoundTripConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: RESP2 Protocol Round-trip Consistency
	// For any valid RESP2 data structure, serializing then parsing should produce an equivalent structure
	properties.Property("RESP2 round-trip consistency", prop.ForAll(
		func(respValue *RESPValue) bool {
			// This test will be implemented in task 2.2
			return true
		},
		genRESPValue(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Generator for RESP2 values - will be implemented in task 2.2
func genRESPValue() gopter.Gen {
	return gen.OneConstOf(&RESPValue{Type: SimpleString, Str: "test"})
}