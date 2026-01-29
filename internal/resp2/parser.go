package resp2

import (
	"bufio"
)

// RESPType represents the type of RESP2 data
type RESPType int

const (
	SimpleString RESPType = iota
	Error
	Integer
	BulkString
	Array
	NullBulkString
)

// RESPValue represents a RESP2 value
type RESPValue struct {
	Type  RESPType
	Str   string
	Int   int64
	Array []RESPValue
	Null  bool
}

// RESP2Parser handles parsing and formatting of RESP2 protocol messages
type RESP2Parser interface {
	Parse(reader *bufio.Reader) (*RESPValue, error)
	Serialize(value *RESPValue) []byte
	ParseCommand(value *RESPValue) (*Command, error)
}

// Command represents a parsed Redis command
type Command struct {
	Name string
	Args []string
}

// DefaultRESP2Parser is the default implementation of RESP2Parser
type DefaultRESP2Parser struct{}

// NewRESP2Parser creates a new RESP2 parser
func NewRESP2Parser() RESP2Parser {
	return &DefaultRESP2Parser{}
}

// Parse parses a RESP2 message from a reader
func (p *DefaultRESP2Parser) Parse(reader *bufio.Reader) (*RESPValue, error) {
	// Implementation will be added in later tasks
	return nil, nil
}

// Serialize converts a RESPValue to bytes
func (p *DefaultRESP2Parser) Serialize(value *RESPValue) []byte {
	// Implementation will be added in later tasks
	return nil
}

// ParseCommand converts a RESPValue to a Command
func (p *DefaultRESP2Parser) ParseCommand(value *RESPValue) (*Command, error) {
	// Implementation will be added in later tasks
	return nil, nil
}