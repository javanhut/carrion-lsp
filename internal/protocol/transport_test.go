package protocol

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioTransport_ReadMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid message",
			input:       "Content-Length: 46\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\"}",
			expected:    `{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
			expectError: false,
		},
		{
			name:        "message with extra headers",
			input:       "Content-Length: 24\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1}",
			expected:    `{"jsonrpc":"2.0","id":1}`,
			expectError: false,
		},
		{
			name:        "missing content-length",
			input:       "Content-Type: application/json\r\n\r\n{\"jsonrpc\":\"2.0\"}",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid content-length",
			input:       "Content-Length: abc\r\n\r\n{\"jsonrpc\":\"2.0\"}",
			expected:    "",
			expectError: true,
		},
		{
			name:        "content-length mismatch",
			input:       "Content-Length: 100\r\n\r\n{\"jsonrpc\":\"2.0\"}",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty message",
			input:       "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			transport := NewStdioTransport(reader, nil)

			msg, err := transport.ReadMessage()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(msg))
			}
		})
	}
}

func TestStdioTransport_WriteMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple message",
			message:  `{"jsonrpc":"2.0","id":1,"result":null}`,
			expected: "Content-Length: 38\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":null}",
		},
		{
			name:     "empty message",
			message:  `{}`,
			expected: "Content-Length: 2\r\n\r\n{}",
		},
		{
			name:     "message with unicode",
			message:  `{"message":"Hello 世界"}`,
			expected: "Content-Length: 26\r\n\r\n{\"message\":\"Hello 世界\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			transport := NewStdioTransport(nil, &buf)

			err := transport.WriteMessage([]byte(tt.message))
			require.NoError(t, err)

			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestStdioTransport_Concurrent(t *testing.T) {
	// Test concurrent reads and writes for race conditions
	messages := []string{
		`{"jsonrpc":"2.0","id":1,"method":"test1"}`,
		`{"jsonrpc":"2.0","id":2,"method":"test2"}`,
		`{"jsonrpc":"2.0","id":3,"method":"test3"}`,
	}

	// Create a pipe for testing
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	transport := NewStdioTransport(reader, writer)

	// Write messages concurrently
	go func() {
		for _, msg := range messages {
			err := transport.WriteMessage([]byte(msg))
			assert.NoError(t, err)
		}
	}()

	// Read messages concurrently
	readMessages := make([]string, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		msg, err := transport.ReadMessage()
		require.NoError(t, err)
		readMessages = append(readMessages, string(msg))
	}

	// Verify all messages were received
	assert.Len(t, readMessages, len(messages))
}

func TestTransportContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reader := strings.NewReader("Content-Length: 10\r\n\r\n0123456789")
	transport := NewStdioTransportWithContext(ctx, reader, nil)

	// Try to read - should fail due to cancelled context
	_, err := transport.ReadMessage()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestTransportSecurity(t *testing.T) {
	t.Run("reject oversized headers", func(t *testing.T) {
		// Create a header that's too large
		largeHeader := fmt.Sprintf("Content-Length: 100\r\nX-Custom: %s\r\n\r\n", strings.Repeat("x", 10000))
		reader := strings.NewReader(largeHeader)
		transport := NewStdioTransport(reader, nil)

		_, err := transport.ReadMessage()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "header too large")
	})

	t.Run("reject too many headers", func(t *testing.T) {
		// Create too many headers
		var headers strings.Builder
		headers.WriteString("Content-Length: 10\r\n")
		for i := 0; i < 100; i++ {
			headers.WriteString(fmt.Sprintf("X-Header-%d: value\r\n", i))
		}
		headers.WriteString("\r\n0123456789")

		reader := strings.NewReader(headers.String())
		transport := NewStdioTransport(reader, nil)

		_, err := transport.ReadMessage()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many headers")
	})

	t.Run("handle partial reads gracefully", func(t *testing.T) {
		// Simulate a reader that returns data in small chunks
		input := "Content-Length: 24\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1}"
		reader := &chunkReader{data: []byte(input), chunkSize: 5}
		transport := NewStdioTransport(reader, nil)

		msg, err := transport.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, `{"jsonrpc":"2.0","id":1}`, string(msg))
	})
}

// Helper types for testing

type chunkReader struct {
	data      []byte
	pos       int
	chunkSize int
}

func (r *chunkReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	remaining := len(r.data) - r.pos
	readSize := r.chunkSize
	if readSize > remaining {
		readSize = remaining
	}
	if readSize > len(p) {
		readSize = len(p)
	}

	copy(p, r.data[r.pos:r.pos+readSize])
	r.pos += readSize
	return readSize, nil
}
