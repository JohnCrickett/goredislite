package resp2

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+': // Simple String
		return p.parseSimpleString(reader)
	case '-': // Error
		return p.parseError(reader)
	case ':': // Integer
		return p.parseInteger(reader)
	case '$': // Bulk String
		return p.parseBulkString(reader)
	case '*': // Array
		return p.parseArray(reader)
	default:
		return nil, fmt.Errorf("invalid RESP2 type indicator: %c", b)
	}
}

// parseSimpleString parses a simple string (+OK\r\n)
func (p *DefaultRESP2Parser) parseSimpleString(reader *bufio.Reader) (*RESPValue, error) {
	line, err := p.readLine(reader)
	if err != nil {
		return nil, err
	}
	return &RESPValue{Type: SimpleString, Str: line}, nil
}

// parseError parses an error (-ERR message\r\n)
func (p *DefaultRESP2Parser) parseError(reader *bufio.Reader) (*RESPValue, error) {
	line, err := p.readLine(reader)
	if err != nil {
		return nil, err
	}
	return &RESPValue{Type: Error, Str: line}, nil
}

// parseInteger parses an integer (:123\r\n)
func (p *DefaultRESP2Parser) parseInteger(reader *bufio.Reader) (*RESPValue, error) {
	line, err := p.readLine(reader)
	if err != nil {
		return nil, err
	}
	
	num, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer format: %s", line)
	}
	
	return &RESPValue{Type: Integer, Int: num}, nil
}

// parseBulkString parses a bulk string ($6\r\nfoobar\r\n or $-1\r\n for null)
func (p *DefaultRESP2Parser) parseBulkString(reader *bufio.Reader) (*RESPValue, error) {
	line, err := p.readLine(reader)
	if err != nil {
		return nil, err
	}
	
	length, err := strconv.Atoi(line)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %s", line)
	}
	
	// Handle null bulk string
	if length == -1 {
		return &RESPValue{Type: NullBulkString, Null: true}, nil
	}
	
	if length < 0 {
		return nil, fmt.Errorf("invalid bulk string length: %d", length)
	}
	
	// Read the string data
	data := make([]byte, length)
	_, err = io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}
	
	// Read the trailing \r\n
	_, err = p.readLine(reader)
	if err != nil {
		return nil, err
	}
	
	return &RESPValue{Type: BulkString, Str: string(data)}, nil
}

// parseArray parses an array (*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n or *-1\r\n for null)
func (p *DefaultRESP2Parser) parseArray(reader *bufio.Reader) (*RESPValue, error) {
	line, err := p.readLine(reader)
	if err != nil {
		return nil, err
	}
	
	length, err := strconv.Atoi(line)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %s", line)
	}
	
	// Handle null array
	if length == -1 {
		return &RESPValue{Type: Array, Null: true}, nil
	}
	
	if length < 0 {
		return nil, fmt.Errorf("invalid array length: %d", length)
	}
	
	// Parse array elements
	elements := make([]RESPValue, length)
	for i := 0; i < length; i++ {
		element, err := p.Parse(reader)
		if err != nil {
			return nil, err
		}
		elements[i] = *element
	}
	
	return &RESPValue{Type: Array, Array: elements}, nil
}

// readLine reads a line ending with \r\n and returns the content without the CRLF
func (p *DefaultRESP2Parser) readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	// Remove \r\n
	if len(line) < 2 || line[len(line)-2:] != "\r\n" {
		return "", errors.New("invalid line ending, expected \\r\\n")
	}
	
	return line[:len(line)-2], nil
}

// Serialize converts a RESPValue to bytes
func (p *DefaultRESP2Parser) Serialize(value *RESPValue) []byte {
	switch value.Type {
	case SimpleString:
		return []byte(fmt.Sprintf("+%s\r\n", value.Str))
	case Error:
		return []byte(fmt.Sprintf("-%s\r\n", value.Str))
	case Integer:
		return []byte(fmt.Sprintf(":%d\r\n", value.Int))
	case BulkString:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value.Str), value.Str))
	case NullBulkString:
		return []byte("$-1\r\n")
	case Array:
		if value.Null {
			return []byte("*-1\r\n")
		}
		
		var result strings.Builder
		result.WriteString(fmt.Sprintf("*%d\r\n", len(value.Array)))
		
		for _, element := range value.Array {
			result.Write(p.Serialize(&element))
		}
		
		return []byte(result.String())
	default:
		return []byte("-ERR unknown RESP2 type\r\n")
	}
}

// ParseCommand converts a RESPValue to a Command
func (p *DefaultRESP2Parser) ParseCommand(value *RESPValue) (*Command, error) {
	if value == nil {
		return nil, errors.New("command value cannot be nil")
	}
	
	if value.Type != Array {
		return nil, errors.New("command must be an array")
	}
	
	if value.Null || len(value.Array) == 0 {
		return nil, errors.New("command array cannot be empty")
	}
	
	// First element should be the command name
	if value.Array[0].Type != BulkString {
		return nil, errors.New("command name must be a bulk string")
	}
	
	cmd := &Command{
		Name: strings.ToUpper(value.Array[0].Str),
		Args: make([]string, len(value.Array)-1),
	}
	
	// Extract arguments
	for i := 1; i < len(value.Array); i++ {
		if value.Array[i].Type != BulkString {
			return nil, fmt.Errorf("command argument %d must be a bulk string", i)
		}
		cmd.Args[i-1] = value.Array[i].Str
	}
	
	return cmd, nil
}