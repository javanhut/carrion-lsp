package protocol

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Transport defines the interface for message transport
type Transport interface {
	ReadMessage() ([]byte, error)
	WriteMessage(data []byte) error
	Close() error
}

// StdioTransport implements Transport using stdio
type StdioTransport struct {
	reader io.Reader
	writer io.Writer
	ctx    context.Context
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: reader,
		writer: writer,
		ctx:    context.Background(),
	}
}

// NewStdioTransportWithContext creates a new stdio transport with context
func NewStdioTransportWithContext(ctx context.Context, reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: reader,
		writer: writer,
		ctx:    ctx,
	}
}

// ReadMessage reads a message from the transport using LSP protocol
func (t *StdioTransport) ReadMessage() ([]byte, error) {
	// Check context cancellation
	select {
	case <-t.ctx.Done():
		return nil, t.ctx.Err()
	default:
	}

	reader := bufio.NewReader(t.reader)
	headers := make(map[string]string)
	headerCount := 0

	// Read headers line by line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading headers: %w", err)
		}

		// Remove CRLF or LF
		line = strings.TrimRight(line, "\r\n")

		// Empty line indicates end of headers
		if line == "" {
			break
		}

		// Security check: prevent too many headers
		headerCount++
		if headerCount > MaxHeaderCount {
			return nil, fmt.Errorf("too many headers: %d exceeds limit of %d", headerCount, MaxHeaderCount)
		}

		// Security check: prevent oversized headers
		if len(line) > MaxHeaderSize {
			return nil, fmt.Errorf("header too large: %d bytes exceeds limit of %d", len(line), MaxHeaderSize)
		}

		// Parse header
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed header: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}

	// Get content length
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
	}

	// Security check: prevent oversized content
	if contentLength > MaxRequestSize {
		return nil, fmt.Errorf("content too large: %d bytes exceeds limit of %d", contentLength, MaxRequestSize)
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("invalid Content-Length: %d", contentLength)
	}

	// Read the content
	content := make([]byte, contentLength)
	_, err = io.ReadFull(reader, content)
	if err != nil {
		return nil, fmt.Errorf("error reading content: %w", err)
	}

	return content, nil
}

// WriteMessage writes a message to the transport using LSP protocol
func (t *StdioTransport) WriteMessage(data []byte) error {
	// Check context cancellation
	select {
	case <-t.ctx.Done():
		return t.ctx.Err()
	default:
	}

	// Write headers
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := t.writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write content
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("error writing content: %w", err)
	}

	return nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	// For stdio, we don't actually close the streams
	// They are managed by the OS
	return nil
}
