package server

import (
	"context"
	"testing"

	"github.com/javanhut/carrion-lsp/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Initialize(t *testing.T) {
	tests := []struct {
		name                 string
		params               protocol.InitializeParams
		expectedCapabilities protocol.ServerCapabilities
		expectError          bool
	}{
		{
			name: "basic initialization",
			params: protocol.InitializeParams{
				ProcessID: intPtr(12345),
				ClientInfo: &protocol.ClientInfo{
					Name:    "vscode",
					Version: "1.74.0",
				},
				RootURI: stringPtr("file:///path/to/workspace"),
				Capabilities: protocol.ClientCapabilities{
					TextDocument: &protocol.TextDocumentClientCapabilities{
						Synchronization: &protocol.TextDocumentSyncClientCapabilities{
							DynamicRegistration: testBoolPtr(true),
						},
						Completion: &protocol.CompletionClientCapabilities{
							DynamicRegistration: testBoolPtr(true),
						},
					},
				},
			},
			expectedCapabilities: protocol.ServerCapabilities{
				TextDocumentSync: &protocol.TextDocumentSyncOptions{
					OpenClose: testBoolPtr(true),
					Change:    protocol.TextDocumentSyncKindFull,
				},
				CompletionProvider: &protocol.CompletionOptions{
					TriggerCharacters: []string{".", "(", "["},
				},
				HoverProvider:              testBoolPtr(true),
				DefinitionProvider:         testBoolPtr(true),
				ReferencesProvider:         testBoolPtr(true),
				DocumentFormattingProvider: testBoolPtr(true),
				DocumentSymbolProvider:     testBoolPtr(true),
				DiagnosticProvider: &protocol.DiagnosticOptions{
					Identifier:            "carrion-lsp",
					InterFileDependencies: false,
					WorkspaceDiagnostics:  false,
				},
			},
			expectError: false,
		},
		{
			name: "minimal client capabilities",
			params: protocol.InitializeParams{
				ProcessID:    intPtr(12345),
				RootURI:      stringPtr("file:///minimal"),
				Capabilities: protocol.ClientCapabilities{},
			},
			expectedCapabilities: protocol.ServerCapabilities{
				TextDocumentSync: &protocol.TextDocumentSyncOptions{
					OpenClose: testBoolPtr(true),
					Change:    protocol.TextDocumentSyncKindFull,
				},
				CompletionProvider: &protocol.CompletionOptions{
					TriggerCharacters: []string{".", "(", "["},
				},
				HoverProvider:              testBoolPtr(true),
				DefinitionProvider:         testBoolPtr(true),
				ReferencesProvider:         testBoolPtr(true),
				DocumentFormattingProvider: testBoolPtr(true),
				DocumentSymbolProvider:     testBoolPtr(true),
				DiagnosticProvider: &protocol.DiagnosticOptions{
					Identifier:            "carrion-lsp",
					InterFileDependencies: false,
					WorkspaceDiagnostics:  false,
				},
			},
			expectError: false,
		},
		{
			name: "null process id (allowed)",
			params: protocol.InitializeParams{
				ProcessID:    nil,
				RootURI:      stringPtr("file:///test"),
				Capabilities: protocol.ClientCapabilities{},
			},
			expectedCapabilities: protocol.ServerCapabilities{
				TextDocumentSync: &protocol.TextDocumentSyncOptions{
					OpenClose: testBoolPtr(true),
					Change:    protocol.TextDocumentSyncKindFull,
				},
				CompletionProvider: &protocol.CompletionOptions{
					TriggerCharacters: []string{".", "(", "["},
				},
				HoverProvider:              testBoolPtr(true),
				DefinitionProvider:         testBoolPtr(true),
				ReferencesProvider:         testBoolPtr(true),
				DocumentFormattingProvider: testBoolPtr(true),
				DocumentSymbolProvider:     testBoolPtr(true),
				DiagnosticProvider: &protocol.DiagnosticOptions{
					Identifier:            "carrion-lsp",
					InterFileDependencies: false,
					WorkspaceDiagnostics:  false,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()

			result, err := server.Initialize(context.Background(), &tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Check server info
				assert.Equal(t, "carrion-lsp", result.ServerInfo.Name)
				assert.NotEmpty(t, result.ServerInfo.Version)

				// Check capabilities
				assertCapabilitiesEqual(t, tt.expectedCapabilities, result.Capabilities)
			}
		})
	}
}

func TestServer_InitializeLifecycle(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	// Server should not be initialized initially
	assert.False(t, server.IsInitialized())

	// Initialize
	params := &protocol.InitializeParams{
		ProcessID:    intPtr(12345),
		RootURI:      stringPtr("file:///test"),
		Capabilities: protocol.ClientCapabilities{},
	}

	result, err := server.Initialize(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Server should still not be initialized until 'initialized' notification
	assert.False(t, server.IsInitialized())

	// Send initialized notification
	err = server.Initialized(ctx)
	require.NoError(t, err)

	// Now server should be initialized
	assert.True(t, server.IsInitialized())

	// Shutdown
	err = server.Shutdown(ctx)
	require.NoError(t, err)

	// Server should be shutting down but still initialized
	assert.True(t, server.IsInitialized())
	assert.True(t, server.IsShuttingDown())

	// Exit
	server.Exit()

	// Server should be exited
	assert.True(t, server.IsExited())
}

func TestServer_InitializeWithCarrionPath(t *testing.T) {
	tests := []struct {
		name        string
		carrionPath string
		expectError bool
	}{
		{
			name:        "valid carrion path",
			carrionPath: "../TheCarrionLanguage",
			expectError: false,
		},
		{
			name:        "empty carrion path (should default)",
			carrionPath: "",
			expectError: false,
		},
		{
			name:        "invalid carrion path",
			carrionPath: "/nonexistent/path",
			expectError: false, // Should not error, just warn
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ServerOptions{
				CarrionPath: tt.carrionPath,
			}
			server := NewServerWithOptions(opts)

			params := &protocol.InitializeParams{
				ProcessID:    intPtr(12345),
				RootURI:      stringPtr("file:///test"),
				Capabilities: protocol.ClientCapabilities{},
				InitializationOptions: map[string]interface{}{
					"carrionPath": tt.carrionPath,
				},
			}

			result, err := server.Initialize(context.Background(), params)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestServer_DoubleInitialize(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	params := &protocol.InitializeParams{
		ProcessID:    intPtr(12345),
		RootURI:      stringPtr("file:///test"),
		Capabilities: protocol.ClientCapabilities{},
	}

	// First initialize should succeed
	result1, err := server.Initialize(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result1)

	// Second initialize should fail
	result2, err := server.Initialize(ctx, params)
	assert.Error(t, err)
	assert.Nil(t, result2)
	assert.Contains(t, err.Error(), "already initialized")
}

func TestServer_ShutdownBeforeInitialize(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	// Shutdown before initialize should fail
	err := server.Shutdown(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestServer_Integration_FullFlow(t *testing.T) {
	// This test will verify the full message flow when we implement the main server loop
	// For now, let's test the individual components directly
	server := NewServer()
	ctx := context.Background()

	// Prepare initialize request
	initParams := protocol.InitializeParams{
		ProcessID: intPtr(12345),
		ClientInfo: &protocol.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
		RootURI: stringPtr("file:///test"),
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Synchronization: &protocol.TextDocumentSyncClientCapabilities{
					DynamicRegistration: func() *bool { b := true; return &b }(),
				},
			},
		},
	}

	// Test direct method calls
	result, err := server.Initialize(ctx, &initParams)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result structure
	assert.Equal(t, "carrion-lsp", result.ServerInfo.Name)
	assert.NotNil(t, result.Capabilities.TextDocumentSync)

	// Test initialized notification
	err = server.Initialized(ctx)
	require.NoError(t, err)
	assert.True(t, server.IsInitialized())
}

// Helper functions for tests

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func testBoolPtr(b bool) *bool {
	return &b
}

func assertCapabilitiesEqual(t *testing.T, expected, actual protocol.ServerCapabilities) {
	// Compare each capability field
	if expected.TextDocumentSync != nil {
		require.NotNil(t, actual.TextDocumentSync)
		assert.Equal(t, expected.TextDocumentSync.OpenClose, actual.TextDocumentSync.OpenClose)
		assert.Equal(t, expected.TextDocumentSync.Change, actual.TextDocumentSync.Change)
	}

	if expected.CompletionProvider != nil {
		require.NotNil(t, actual.CompletionProvider)
		assert.Equal(t, expected.CompletionProvider.TriggerCharacters, actual.CompletionProvider.TriggerCharacters)
	}

	assert.Equal(t, expected.HoverProvider, actual.HoverProvider)
	assert.Equal(t, expected.DefinitionProvider, actual.DefinitionProvider)
	assert.Equal(t, expected.ReferencesProvider, actual.ReferencesProvider)
	assert.Equal(t, expected.DocumentFormattingProvider, actual.DocumentFormattingProvider)
	assert.Equal(t, expected.DocumentSymbolProvider, actual.DocumentSymbolProvider)

	if expected.DiagnosticProvider != nil {
		require.NotNil(t, actual.DiagnosticProvider)
		assert.Equal(t, expected.DiagnosticProvider.Identifier, actual.DiagnosticProvider.Identifier)
		assert.Equal(t, expected.DiagnosticProvider.InterFileDependencies, actual.DiagnosticProvider.InterFileDependencies)
		assert.Equal(t, expected.DiagnosticProvider.WorkspaceDiagnostics, actual.DiagnosticProvider.WorkspaceDiagnostics)
	}
}
