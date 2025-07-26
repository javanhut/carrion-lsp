package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSONRPCRequest(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *Request
		expectError bool
	}{
		{
			name:  "valid request with params",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":12345}}`,
			expected: &Request{
				Message: Message{Jsonrpc: "2.0"},
				ID:      float64(1),
				Method:  "initialize",
				Params:  map[string]interface{}{"processId": float64(12345)},
			},
			expectError: false,
		},
		{
			name:  "valid request without params",
			input: `{"jsonrpc":"2.0","id":"test-id","method":"shutdown"}`,
			expected: &Request{
				Message: Message{Jsonrpc: "2.0"},
				ID:      "test-id",
				Method:  "shutdown",
			},
			expectError: false,
		},
		{
			name:  "notification (no id)",
			input: `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{}}`,
			expected: &Request{
				Message: Message{Jsonrpc: "2.0"},
				Method:  "textDocument/didOpen",
				Params:  map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       `{"jsonrpc":"2.0","id":1,"method":"test"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "missing jsonrpc version",
			input:       `{"id":1,"method":"test"}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "wrong jsonrpc version",
			input:       `{"jsonrpc":"1.0","id":1,"method":"test"}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "missing method",
			input:       `{"jsonrpc":"2.0","id":1}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseRequest([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, req)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Message.Jsonrpc, req.Message.Jsonrpc)
				assert.Equal(t, tt.expected.Method, req.Method)
				assert.Equal(t, tt.expected.ID, req.ID)

				if tt.expected.Params != nil {
					expectedJSON, _ := json.Marshal(tt.expected.Params)
					actualJSON, _ := json.Marshal(req.Params)
					assert.JSONEq(t, string(expectedJSON), string(actualJSON))
				}
			}
		})
	}
}

func TestSerializeJSONRPCResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		expected string
	}{
		{
			name: "success response",
			response: &Response{
				Message: Message{Jsonrpc: "2.0"},
				ID:      float64(1),
				Result: map[string]interface{}{
					"capabilities": map[string]interface{}{
						"textDocumentSync": 1,
					},
				},
			},
			expected: `{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":1}}}`,
		},
		{
			name: "error response",
			response: &Response{
				Message: Message{Jsonrpc: "2.0"},
				ID:      "test-id",
				Error: &Error{
					Code:    -32600,
					Message: "Invalid Request",
				},
			},
			expected: `{"jsonrpc":"2.0","id":"test-id","error":{"code":-32600,"message":"Invalid Request"}}`,
		},
		{
			name: "error with data",
			response: &Response{
				Message: Message{Jsonrpc: "2.0"},
				ID:      float64(1),
				Error: &Error{
					Code:    -32602,
					Message: "Invalid params",
					Data:    "processId must be a number",
				},
			},
			expected: `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params","data":"processId must be a number"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := SerializeResponse(tt.response)
			require.NoError(t, err)

			// Parse both to compare as JSON objects (ignores key ordering)
			var expected, actual interface{}
			require.NoError(t, json.Unmarshal([]byte(tt.expected), &expected))
			require.NoError(t, json.Unmarshal(data, &actual))

			assert.Equal(t, expected, actual)
		})
	}
}

func TestHandleMalformedJSON(t *testing.T) {
	malformedInputs := []string{
		``,                          // empty
		`{`,                         // incomplete
		`{"jsonrpc": }`,             // invalid value
		`{"jsonrpc":"2.0","id":1,}`, // trailing comma
		`{'jsonrpc':'2.0'}`,         // single quotes
		`null`,                      // null
		`[]`,                        // array instead of object
		`"string"`,                  // string instead of object
	}

	for _, input := range malformedInputs {
		t.Run("malformed: "+input, func(t *testing.T) {
			req, err := ParseRequest([]byte(input))
			assert.Error(t, err)
			assert.Nil(t, req)
		})
	}
}

func TestRejectOversizedRequests(t *testing.T) {
	// Create a request larger than MaxRequestSize
	largeParams := strings.Repeat("a", MaxRequestSize)
	oversizedRequest := `{"jsonrpc":"2.0","id":1,"method":"test","params":"` + largeParams + `"}`

	req, err := ParseRequest([]byte(oversizedRequest))
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "request too large")
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     *Request
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: &Request{
				Message: Message{Jsonrpc: "2.0"},
				ID:      float64(1),
				Method:  "initialize",
			},
			expectError: false,
		},
		{
			name: "invalid jsonrpc version",
			request: &Request{
				Message: Message{Jsonrpc: "1.0"},
				ID:      float64(1),
				Method:  "test",
			},
			expectError: true,
			errorMsg:    "invalid jsonrpc version",
		},
		{
			name: "empty method",
			request: &Request{
				Message: Message{Jsonrpc: "2.0"},
				ID:      float64(1),
				Method:  "",
			},
			expectError: true,
			errorMsg:    "method is required",
		},
		{
			name: "notification with valid structure",
			request: &Request{
				Message: Message{Jsonrpc: "2.0"},
				Method:  "textDocument/didChange",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationDetection(t *testing.T) {
	tests := []struct {
		name           string
		request        *Request
		isNotification bool
	}{
		{
			name: "request with numeric id",
			request: &Request{
				ID:     float64(1),
				Method: "test",
			},
			isNotification: false,
		},
		{
			name: "request with string id",
			request: &Request{
				ID:     "test-id",
				Method: "test",
			},
			isNotification: false,
		},
		{
			name: "notification without id",
			request: &Request{
				Method: "textDocument/didOpen",
			},
			isNotification: true,
		},
		{
			name: "notification with nil id",
			request: &Request{
				ID:     nil,
				Method: "exit",
			},
			isNotification: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isNotification, tt.request.IsNotification())
		})
	}
}
