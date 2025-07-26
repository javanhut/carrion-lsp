package protocol

import (
	"encoding/json"
	"fmt"
)

// Constants for JSON-RPC and security limits
const (
	JSONRPCVersion = "2.0"
	MaxRequestSize = 1 * 1024 * 1024 // 1MB max request size
	MaxHeaderSize  = 8 * 1024        // 8KB max headers
	MaxHeaderCount = 50              // Maximum number of headers
)

// JSON-RPC 2.0 message types

// Message represents the base JSON-RPC message
type Message struct {
	Jsonrpc string `json:"jsonrpc"`
}

// Request represents a JSON-RPC request
type Request struct {
	Message
	ID     interface{} `json:"id,omitempty"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	Message
	ID     interface{} `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Standard errors
var (
	ErrParseError = &Error{
		Code:    ParseError,
		Message: "Parse error",
	}
	ErrInvalidRequest = &Error{
		Code:    InvalidRequest,
		Message: "Invalid Request",
	}
	ErrMethodNotFound = &Error{
		Code:    MethodNotFound,
		Message: "Method not found",
	}
	ErrInvalidParams = &Error{
		Code:    InvalidParams,
		Message: "Invalid params",
	}
	ErrInternalError = &Error{
		Code:    InternalError,
		Message: "Internal error",
	}
)

// ParseRequest parses a JSON-RPC request from bytes
func ParseRequest(data []byte) (*Request, error) {
	// Security check: prevent oversized requests
	if len(data) > MaxRequestSize {
		return nil, fmt.Errorf("request too large: %d bytes exceeds limit of %d", len(data), MaxRequestSize)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return &req, nil
}

// SerializeResponse serializes a JSON-RPC response to bytes
func SerializeResponse(resp *Response) ([]byte, error) {
	return json.Marshal(resp)
}

// Validate validates a JSON-RPC request
func (r *Request) Validate() error {
	if r.Jsonrpc != JSONRPCVersion {
		return fmt.Errorf("invalid jsonrpc version: %s, expected %s", r.Jsonrpc, JSONRPCVersion)
	}

	if r.Method == "" {
		return fmt.Errorf("method is required")
	}

	return nil
}

// IsNotification returns true if this is a notification (no ID)
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// Error implements the error interface for Error
func (e *Error) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (%v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewErrorResponse creates an error response for a request
func NewErrorResponse(id interface{}, err *Error) *Response {
	return &Response{
		Message: Message{Jsonrpc: JSONRPCVersion},
		ID:      id,
		Error:   err,
	}
}

// NewSuccessResponse creates a success response for a request
func NewSuccessResponse(id interface{}, result interface{}) *Response {
	return &Response{
		Message: Message{Jsonrpc: JSONRPCVersion},
		ID:      id,
		Result:  result,
	}
}
