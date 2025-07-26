package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/javanhut/carrion-lsp/internal/carrion/symbol"
	"github.com/javanhut/carrion-lsp/internal/protocol"
)

// Server state constants
type ServerState int

const (
	ServerStateUninitialized ServerState = iota
	ServerStateInitializing
	ServerStateInitialized
	ServerStateShuttingDown
	ServerStateExited
)

// Server represents the LSP server
type Server struct {
	mu               sync.RWMutex
	state            ServerState
	transport        protocol.Transport
	options          ServerOptions
	rootURI          string
	clientInfo       *protocol.ClientInfo
	capabilities     protocol.ClientCapabilities
	logger           *log.Logger
	workspaceManager *WorkspaceManager
	docManager       *DocumentManager // Fallback for non-workspace operations
}

// ServerOptions contains server configuration
type ServerOptions struct {
	CarrionPath string
	Logger      *log.Logger
}

// Version information
const (
	ServerName    = "carrion-lsp"
	ServerVersion = "0.1.0"
)

// NewServer creates a new LSP server with default options
func NewServer() *Server {
	return NewServerWithOptions(ServerOptions{})
}

// NewServerWithOptions creates a new LSP server with custom options
func NewServerWithOptions(opts ServerOptions) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "[carrion-lsp] ", log.LstdFlags)
	}

	return &Server{
		state:      ServerStateUninitialized,
		options:    opts,
		logger:     logger,
		docManager: NewDocumentManager(), // Fallback for basic operations
	}
}

// NewServerWithTransport creates a new LSP server with a specific transport
func NewServerWithTransport(transport protocol.Transport) *Server {
	server := NewServer()
	server.transport = transport
	return server
}

// SetTransport sets the transport for an existing server
func (s *Server) SetTransport(transport protocol.Transport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transport = transport
}

// Initialize handles the initialize request
func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already initialized
	if s.state != ServerStateUninitialized {
		return nil, fmt.Errorf("server already initialized")
	}

	s.state = ServerStateInitializing
	s.logger.Printf("Initializing server with client: %s", getClientName(params.ClientInfo))

	// Store client information
	if params.RootURI != nil {
		s.rootURI = *params.RootURI
	}
	s.clientInfo = params.ClientInfo
	s.capabilities = params.Capabilities

	// Handle initialization options
	if params.InitializationOptions != nil {
		if opts, ok := params.InitializationOptions.(map[string]interface{}); ok {
			if carrionPath, exists := opts["carrionPath"]; exists {
				if path, ok := carrionPath.(string); ok && path != "" {
					s.options.CarrionPath = path
				}
			}
		}
	}

	// Validate Carrion path if provided
	if s.options.CarrionPath != "" {
		if _, err := os.Stat(s.options.CarrionPath); os.IsNotExist(err) {
			s.logger.Printf("Warning: Carrion path does not exist: %s", s.options.CarrionPath)
			// Don't fail, just warn
		}
	}

	// Initialize workspace manager if we have a root URI
	if s.rootURI != "" {
		workspaceRoot := s.rootURI
		// Convert URI to file path if needed
		if strings.HasPrefix(workspaceRoot, "file://") {
			workspaceRoot = strings.TrimPrefix(workspaceRoot, "file://")
		}
		s.workspaceManager = NewWorkspaceManager(workspaceRoot, s.options.CarrionPath)
		s.logger.Printf("Initialized workspace manager for: %s", workspaceRoot)
	}

	// Build server capabilities based on client capabilities
	serverCapabilities := s.buildServerCapabilities()

	result := &protocol.InitializeResult{
		Capabilities: serverCapabilities,
		ServerInfo: &protocol.ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	s.logger.Printf("Server initialized successfully")
	return result, nil
}

// Initialized handles the initialized notification
func (s *Server) Initialized(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != ServerStateInitializing {
		return fmt.Errorf("server not in initializing state")
	}

	s.state = ServerStateInitialized
	s.logger.Printf("Server is now ready to handle requests")
	return nil
}

// Shutdown handles the shutdown request
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != ServerStateInitialized {
		return fmt.Errorf("server not initialized")
	}

	s.state = ServerStateShuttingDown
	s.logger.Printf("Server shutting down")
	return nil
}

// Exit handles the exit notification
func (s *Server) Exit() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = ServerStateExited
	s.logger.Printf("Server exited")
}

// ProcessRequest processes a single request from the transport
func (s *Server) ProcessRequest(ctx context.Context) error {
	if s.transport == nil {
		return fmt.Errorf("no transport configured")
	}

	// Read message from transport
	data, err := s.transport.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	// Parse JSON-RPC request
	req, err := protocol.ParseRequest(data)
	if err != nil {
		// Send error response if we can parse the ID
		s.sendErrorResponse(nil, protocol.ErrParseError)
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// Handle the request
	if req.IsNotification() {
		return s.handleNotification(ctx, req)
	} else {
		return s.handleRequest(ctx, req)
	}
}

// handleRequest handles a request that expects a response
func (s *Server) handleRequest(ctx context.Context, req *protocol.Request) error {
	var result interface{}
	var err error

	switch req.Method {
	case protocol.MethodInitialize:
		result, err = s.handleInitializeRequest(ctx, req)
	case protocol.MethodShutdown:
		result, err = s.handleShutdownRequest(ctx, req)
	case protocol.MethodTextDocumentCompletion:
		result, err = s.handleCompletionRequest(ctx, req)
	case protocol.MethodTextDocumentHover:
		result, err = s.handleHoverRequest(ctx, req)
	case protocol.MethodTextDocumentDefinition:
		result, err = s.handleDefinitionRequest(ctx, req)
	case protocol.MethodTextDocumentReferences:
		result, err = s.handleReferencesRequest(ctx, req)
	case protocol.MethodTextDocumentSymbol:
		result, err = s.handleDocumentSymbolRequest(ctx, req)
	case protocol.MethodTextDocumentFormatting:
		result, err = s.handleFormattingRequest(ctx, req)
	default:
		err = fmt.Errorf("method not found: %s", req.Method)
	}

	// Send response
	if err != nil {
		s.sendErrorResponse(req.ID, &protocol.Error{
			Code:    protocol.MethodNotFound,
			Message: err.Error(),
		})
	} else {
		s.sendSuccessResponse(req.ID, result)
	}

	return nil
}

// handleNotification handles a notification that doesn't expect a response
func (s *Server) handleNotification(ctx context.Context, req *protocol.Request) error {
	switch req.Method {
	case protocol.MethodInitialized:
		return s.handleInitializedNotification(ctx, req)
	case protocol.MethodExit:
		s.handleExitNotification(ctx, req)
		return nil
	case protocol.MethodTextDocumentDidOpen:
		return s.handleDidOpenNotification(ctx, req)
	case protocol.MethodTextDocumentDidChange:
		return s.handleDidChangeNotification(ctx, req)
	case protocol.MethodTextDocumentDidClose:
		return s.handleDidCloseNotification(ctx, req)
	default:
		s.logger.Printf("Unknown notification: %s", req.Method)
		return nil
	}
}

// Request handlers

func (s *Server) handleInitializeRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	var params protocol.InitializeParams
	if req.Params != nil {
		// Convert params to InitializeParams
		// This is a simplified approach - in production you'd use proper JSON unmarshaling
		if paramsMap, ok := req.Params.(map[string]interface{}); ok {
			// Parse processId
			if processId, exists := paramsMap["processId"]; exists {
				if pid, ok := processId.(float64); ok {
					pidInt := int(pid)
					params.ProcessID = &pidInt
				}
			}

			// Parse rootUri
			if rootUri, exists := paramsMap["rootUri"]; exists {
				if uri, ok := rootUri.(string); ok {
					params.RootURI = &uri
				}
			}

			// Parse clientInfo
			if clientInfo, exists := paramsMap["clientInfo"]; exists {
				if info, ok := clientInfo.(map[string]interface{}); ok {
					params.ClientInfo = &protocol.ClientInfo{}
					if name, exists := info["name"]; exists {
						if n, ok := name.(string); ok {
							params.ClientInfo.Name = n
						}
					}
					if version, exists := info["version"]; exists {
						if v, ok := version.(string); ok {
							params.ClientInfo.Version = v
						}
					}
				}
			}

			// Parse capabilities (simplified)
			if capabilities, exists := paramsMap["capabilities"]; exists {
				if caps, ok := capabilities.(map[string]interface{}); ok {
					params.Capabilities = protocol.ClientCapabilities{}
					// Parse textDocument capabilities
					if textDoc, exists := caps["textDocument"]; exists {
						if _, ok := textDoc.(map[string]interface{}); ok {
							params.Capabilities.TextDocument = &protocol.TextDocumentClientCapabilities{}
							// Add more parsing as needed
						}
					}
				}
			}

			// Parse initializationOptions
			if initOpts, exists := paramsMap["initializationOptions"]; exists {
				params.InitializationOptions = initOpts
			}
		}
	}

	return s.Initialize(ctx, &params)
}

func (s *Server) handleShutdownRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	err := s.Shutdown(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil // Return null result for shutdown
}

func (s *Server) handleInitializedNotification(ctx context.Context, req *protocol.Request) error {
	return s.Initialized(ctx)
}

func (s *Server) handleExitNotification(ctx context.Context, req *protocol.Request) {
	s.Exit()
}

// Document synchronization handlers

func (s *Server) handleDidOpenNotification(ctx context.Context, req *protocol.Request) error {
	if !s.IsInitialized() {
		return fmt.Errorf("server not initialized")
	}

	var params protocol.DidOpenTextDocumentParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse didOpen params: %w", err)
	}

	s.logger.Printf("Opening document: %s", params.TextDocument.URI)

	var doc *Document
	var err error

	// Use workspace manager if available, otherwise fall back to document manager
	if s.workspaceManager != nil {
		doc, err = s.workspaceManager.OpenDocument(&params)
	} else {
		doc, err = s.docManager.OpenDocument(&params)
	}

	if err != nil {
		s.logger.Printf("Error opening document %s: %v", params.TextDocument.URI, err)
		return err
	}

	// Send diagnostics
	s.sendDiagnostics(params.TextDocument.URI, doc.Diagnostics)

	return nil
}

func (s *Server) handleDidChangeNotification(ctx context.Context, req *protocol.Request) error {
	if !s.IsInitialized() {
		return fmt.Errorf("server not initialized")
	}

	var params protocol.DidChangeTextDocumentParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse didChange params: %w", err)
	}

	s.logger.Printf("Document changed: %s (version %d)", params.TextDocument.URI, params.TextDocument.Version)

	var doc *Document
	var err error

	// Use workspace manager if available, otherwise fall back to document manager
	if s.workspaceManager != nil {
		doc, err = s.workspaceManager.ChangeDocument(&params)
	} else {
		doc, err = s.docManager.ChangeDocument(&params)
	}

	if err != nil {
		s.logger.Printf("Error changing document %s: %v", params.TextDocument.URI, err)
		return err
	}

	// Send updated diagnostics
	s.sendDiagnostics(params.TextDocument.URI, doc.Diagnostics)

	return nil
}

func (s *Server) handleDidCloseNotification(ctx context.Context, req *protocol.Request) error {
	if !s.IsInitialized() {
		return fmt.Errorf("server not initialized")
	}

	var params protocol.DidCloseTextDocumentParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse didClose params: %w", err)
	}

	s.logger.Printf("Closing document: %s", params.TextDocument.URI)

	var err error

	// Use workspace manager if available, otherwise fall back to document manager
	if s.workspaceManager != nil {
		err = s.workspaceManager.CloseDocument(&params)
	} else {
		err = s.docManager.CloseDocument(&params)
	}

	if err != nil {
		s.logger.Printf("Error closing document %s: %v", params.TextDocument.URI, err)
		return err
	}

	// Clear diagnostics
	s.sendDiagnostics(params.TextDocument.URI, nil)

	return nil
}

// Language feature handlers

func (s *Server) handleCompletionRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server not initialized")
	}

	var params protocol.CompletionParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse completion params: %w", err)
	}

	var items []protocol.CompletionItem
	var err error

	// Use workspace manager if available (includes imported symbols), otherwise fall back to document manager
	if s.workspaceManager != nil {
		items, err = s.getWorkspaceCompletionItems(params.TextDocument.URI, params.Position)
	} else {
		items, err = s.docManager.GetCompletionItems(params.TextDocument.URI, params.Position)
	}

	if err != nil {
		s.logger.Printf("Error getting completion items for %s: %v", params.TextDocument.URI, err)
		return []protocol.CompletionItem{}, nil
	}

	return protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

func (s *Server) handleHoverRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server not initialized")
	}

	var params protocol.HoverParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse hover params: %w", err)
	}

	s.logger.Printf("Hover request for %s at line %d, char %d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	var hover *protocol.Hover
	var err error

	// Use workspace manager if available (includes imported symbols), otherwise fall back to document manager
	if s.workspaceManager != nil {
		hover, err = s.getWorkspaceHoverInformation(params.TextDocument.URI, params.Position)
	} else {
		hover, err = s.docManager.GetHoverInformation(params.TextDocument.URI, params.Position)
	}

	if err != nil {
		s.logger.Printf("Error getting hover information for %s: %v", params.TextDocument.URI, err)
		return nil, nil // Return null on error rather than failing
	}

	return hover, nil
}

func (s *Server) handleDefinitionRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server not initialized")
	}

	var params protocol.DefinitionParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse definition params: %w", err)
	}

	s.logger.Printf("Definition request for %s at line %d, char %d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	var locations []protocol.Location
	var err error

	// Use workspace manager if available (supports cross-file go-to-definition), otherwise fall back to document manager
	if s.workspaceManager != nil {
		locations, err = s.getWorkspaceDefinitionLocation(params.TextDocument.URI, params.Position)
	} else {
		locations, err = s.docManager.GetDefinitionLocation(params.TextDocument.URI, params.Position)
	}

	if err != nil {
		s.logger.Printf("Error getting definition location for %s: %v", params.TextDocument.URI, err)
		return []protocol.Location{}, nil // Return empty array on error
	}

	return locations, nil
}

func (s *Server) handleFormattingRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server not initialized")
	}

	var params protocol.DocumentFormattingParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse formatting params: %w", err)
	}

	s.logger.Printf("Formatting request for %s", params.TextDocument.URI)

	edits, err := s.docManager.FormatDocument(params.TextDocument.URI, params.Options)
	if err != nil {
		s.logger.Printf("Error formatting document %s: %v", params.TextDocument.URI, err)
		return []protocol.TextEdit{}, nil // Return empty array on error
	}

	return edits, nil
}

func (s *Server) handleReferencesRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server not initialized")
	}

	var params protocol.ReferenceParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse references params: %w", err)
	}

	s.logger.Printf("References request for %s at line %d, char %d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	locations, err := s.docManager.GetReferences(params.TextDocument.URI, params.Position, params.Context.IncludeDeclaration)
	if err != nil {
		s.logger.Printf("Error getting references for %s: %v", params.TextDocument.URI, err)
		return []protocol.Location{}, nil // Return empty array on error
	}

	return locations, nil
}

func (s *Server) handleDocumentSymbolRequest(ctx context.Context, req *protocol.Request) (interface{}, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server not initialized")
	}

	var params protocol.DocumentSymbolParams
	if err := s.parseParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse document symbol params: %w", err)
	}

	s.logger.Printf("Document symbol request for %s", params.TextDocument.URI)

	symbols, err := s.docManager.GetDocumentSymbols(params.TextDocument.URI)
	if err != nil {
		s.logger.Printf("Error getting document symbols for %s: %v", params.TextDocument.URI, err)
		return []protocol.DocumentSymbol{}, nil // Return empty array on error
	}

	return symbols, nil
}

// Response helpers

func (s *Server) sendSuccessResponse(id interface{}, result interface{}) error {
	if s.transport == nil {
		return fmt.Errorf("no transport configured")
	}

	response := protocol.NewSuccessResponse(id, result)
	data, err := protocol.SerializeResponse(response)
	if err != nil {
		return fmt.Errorf("failed to serialize response: %w", err)
	}

	return s.transport.WriteMessage(data)
}

func (s *Server) sendErrorResponse(id interface{}, err *protocol.Error) error {
	if s.transport == nil {
		return fmt.Errorf("no transport configured")
	}

	response := protocol.NewErrorResponse(id, err)
	data, err2 := protocol.SerializeResponse(response)
	if err2 != nil {
		return fmt.Errorf("failed to serialize error response: %w", err2)
	}

	return s.transport.WriteMessage(data)
}

// State queries

func (s *Server) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == ServerStateInitialized || s.state == ServerStateShuttingDown
}

func (s *Server) IsShuttingDown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == ServerStateShuttingDown
}

func (s *Server) IsExited() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == ServerStateExited
}

// buildServerCapabilities builds server capabilities based on client capabilities
func (s *Server) buildServerCapabilities() protocol.ServerCapabilities {
	capabilities := protocol.ServerCapabilities{
		TextDocumentSync: &protocol.TextDocumentSyncOptions{
			OpenClose: boolPtr(true),
			Change:    protocol.TextDocumentSyncKindFull,
		},
		DiagnosticProvider: &protocol.DiagnosticOptions{
			Identifier:            "carrion-lsp",
			InterFileDependencies: false,
			WorkspaceDiagnostics:  false,
		},
	}

	// Enable features based on client capabilities
	if s.capabilities.TextDocument != nil {
		if s.capabilities.TextDocument.Completion != nil {
			capabilities.CompletionProvider = &protocol.CompletionOptions{
				TriggerCharacters: []string{".", "(", "["},
			}
		}

		if s.capabilities.TextDocument.Hover != nil {
			capabilities.HoverProvider = boolPtr(true)
		}

		if s.capabilities.TextDocument.Definition != nil {
			capabilities.DefinitionProvider = boolPtr(true)
		}

		if s.capabilities.TextDocument.References != nil {
			capabilities.ReferencesProvider = boolPtr(true)
		}

		if s.capabilities.TextDocument.Formatting != nil {
			capabilities.DocumentFormattingProvider = boolPtr(true)
		}
	}

	// Always enable basic features for now (TODO: make this configurable)
	if capabilities.CompletionProvider == nil {
		capabilities.CompletionProvider = &protocol.CompletionOptions{
			TriggerCharacters: []string{".", "(", "["},
		}
	}
	if capabilities.HoverProvider == nil {
		capabilities.HoverProvider = boolPtr(true)
	}
	if capabilities.DefinitionProvider == nil {
		capabilities.DefinitionProvider = boolPtr(true)
	}
	if capabilities.ReferencesProvider == nil {
		capabilities.ReferencesProvider = boolPtr(true)
	}
	if capabilities.DocumentFormattingProvider == nil {
		capabilities.DocumentFormattingProvider = boolPtr(true)
	}
	if capabilities.DocumentSymbolProvider == nil {
		capabilities.DocumentSymbolProvider = boolPtr(true)
	}

	return capabilities
}

// Helper functions

func getClientName(clientInfo *protocol.ClientInfo) string {
	if clientInfo == nil {
		return "unknown"
	}
	if clientInfo.Version != "" {
		return fmt.Sprintf("%s %s", clientInfo.Name, clientInfo.Version)
	}
	return clientInfo.Name
}

func boolPtr(b bool) *bool {
	return &b
}

// parseParams parses request parameters into the given struct
func (s *Server) parseParams(params interface{}, target interface{}) error {
	if params == nil {
		return fmt.Errorf("params is nil")
	}

	// Convert to JSON and back to properly deserialize
	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	err = json.Unmarshal(jsonData, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return nil
}

// sendDiagnostics sends diagnostic information to the client
func (s *Server) sendDiagnostics(uri string, diagnostics []protocol.Diagnostic) {
	if s.transport == nil {
		return
	}

	if diagnostics == nil {
		diagnostics = []protocol.Diagnostic{}
	}

	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]interface{}{
			"uri":         uri,
			"diagnostics": diagnostics,
		},
	}

	data, err := json.Marshal(notification)
	if err != nil {
		s.logger.Printf("Failed to marshal diagnostics notification: %v", err)
		return
	}

	err = s.transport.WriteMessage(data)
	if err != nil {
		s.logger.Printf("Failed to send diagnostics notification: %v", err)
	}
}

// getWorkspaceCompletionItems returns completion items using the workspace manager (includes imported symbols)
func (s *Server) getWorkspaceCompletionItems(uri string, position protocol.Position) ([]protocol.CompletionItem, error) {
	doc, exists := s.workspaceManager.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Check if this is member access completion (obj.member)
	memberContext := s.getMemberAccessContext(doc.Text, position)

	var symbols []*symbol.Symbol
	if memberContext.IsMemberAccess {
		// Get member completion items
		symbols = doc.Analyzer.GetMemberCompletionItems(memberContext.ObjectName, memberContext.MemberPrefix, position.Line, position.Character)
	} else {
		// Regular completion
		prefix := s.getPrefixAtPosition(doc.Text, position)
		symbols = doc.Analyzer.GetCompletionItems(position.Line, position.Character, prefix)
	}

	var items []protocol.CompletionItem
	for _, sym := range symbols {
		kind := s.getCompletionItemKind(sym.Type)
		detail := sym.DataType
		if sym.Type == symbol.FunctionSymbol && len(sym.Parameters) > 0 {
			var params []string
			for _, param := range sym.Parameters {
				params = append(params, param.Name)
			}
			detail = fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), sym.ReturnType)
		}

		items = append(items, protocol.CompletionItem{
			Label:  sym.Name,
			Kind:   &kind,
			Detail: detail,
		})
	}

	return items, nil
}

// getWorkspaceHoverInformation returns hover information using the workspace manager (includes imported symbols)
func (s *Server) getWorkspaceHoverInformation(uri string, position protocol.Position) (*protocol.Hover, error) {
	doc, exists := s.workspaceManager.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get the identifier at the position
	identifier := s.getIdentifierAtPosition(doc.Text, position)
	if identifier == "" {
		return nil, nil // No identifier at position
	}

	// Try to get symbol at specific position first (for scope-aware lookup)
	symbol := doc.Analyzer.GetSymbolAtPosition(position.Line+1, position.Character) // Convert 0-based to 1-based
	if symbol == nil {
		// Fall back to global lookup (this now includes imported symbols from workspace manager)
		var exists bool
		symbol, exists = doc.Analyzer.GetSymbolTable().Lookup(identifier)
		if !exists {
			return nil, nil // Symbol not found
		}
	}

	// Create hover content based on symbol type
	content := s.createHoverContent(symbol)
	if content == "" {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: content,
		},
	}, nil
}

// Helper methods for workspace-aware completion and hover

// getPrefixAtPosition extracts the word prefix at the given position
func (s *Server) getPrefixAtPosition(text string, position protocol.Position) string {
	lines := strings.Split(text, "\n")
	if position.Line >= len(lines) {
		return ""
	}

	line := lines[position.Line]
	if position.Character > len(line) {
		return ""
	}

	// Find the start of the current word
	start := position.Character
	for start > 0 && s.isIdentifierChar(rune(line[start-1])) {
		start--
	}

	return line[start:position.Character]
}

// getIdentifierAtPosition extracts the identifier at the given position
func (s *Server) getIdentifierAtPosition(text string, position protocol.Position) string {
	lines := strings.Split(text, "\n")
	if position.Line >= len(lines) {
		return ""
	}

	line := lines[position.Line]
	if position.Character >= len(line) {
		return ""
	}

	// Find the bounds of the identifier at the cursor position
	start := position.Character
	end := position.Character

	// Move start backward to find the beginning of the identifier
	for start > 0 && s.isIdentifierChar(rune(line[start-1])) {
		start--
	}

	// Move end forward to find the end of the identifier
	for end < len(line) && s.isIdentifierChar(rune(line[end])) {
		end++
	}

	// Return the identifier if we found one
	if start < end && s.isIdentifierChar(rune(line[start])) {
		return line[start:end]
	}

	return ""
}

// createHoverContent creates markdown content for hover information
func (s *Server) createHoverContent(sym *symbol.Symbol) string {
	var content strings.Builder

	switch sym.Type {
	case symbol.VariableSymbol:
		content.WriteString(fmt.Sprintf("**Variable**: `%s`\n\n", sym.Name))
		content.WriteString(fmt.Sprintf("**Type**: `%s`\n\n", sym.DataType))
		if sym.Token.Line > 0 {
			content.WriteString(fmt.Sprintf("**Declared at**: line %d\n", sym.Token.Line))
		}

	case symbol.FunctionSymbol:
		content.WriteString(fmt.Sprintf("**Function**: `%s`\n\n", sym.Name))

		// Function signature
		var params []string
		for _, param := range sym.Parameters {
			params = append(params, param.Name)
		}
		signature := fmt.Sprintf("spell %s(%s)", sym.Name, strings.Join(params, ", "))
		if sym.ReturnType != "" && sym.ReturnType != "unknown" {
			signature += fmt.Sprintf(" -> %s", sym.ReturnType)
		}
		content.WriteString(fmt.Sprintf("```carrion\n%s\n```\n\n", signature))

		if sym.Token.Line > 0 {
			content.WriteString(fmt.Sprintf("**Declared at**: line %d\n", sym.Token.Line))
		}

	case symbol.ClassSymbol:
		content.WriteString(fmt.Sprintf("**Class**: `%s`\n\n", sym.Name))
		content.WriteString(fmt.Sprintf("```carrion\ngrim %s\n```\n\n", sym.Name))

		// Show inheritance
		if sym.Parent != nil {
			content.WriteString(fmt.Sprintf("**Inherits from**: `%s`\n\n", sym.Parent.Name))
		}

		// Show methods
		if len(sym.Members) > 0 {
			content.WriteString("**Methods**:\n")
			for name, member := range sym.Members {
				if member.Type == symbol.FunctionSymbol {
					var params []string
					for _, param := range member.Parameters {
						params = append(params, param.Name)
					}
					content.WriteString(fmt.Sprintf("- `%s(%s)`\n", name, strings.Join(params, ", ")))
				}
			}
			content.WriteString("\n")
		}

		if sym.Token.Line > 0 {
			content.WriteString(fmt.Sprintf("**Declared at**: line %d\n", sym.Token.Line))
		}

	case symbol.ParameterSymbol:
		content.WriteString(fmt.Sprintf("**Parameter**: `%s`\n\n", sym.Name))
		content.WriteString(fmt.Sprintf("**Type**: `%s`\n\n", sym.DataType))

	case symbol.ModuleSymbol:
		content.WriteString(fmt.Sprintf("**Module**: `%s`\n\n", sym.Name))
		if sym.Token.Line > 0 {
			content.WriteString(fmt.Sprintf("**Imported at**: line %d\n", sym.Token.Line))
		}

		// Show module members
		if len(sym.Members) > 0 {
			content.WriteString("**Available symbols**:\n")
			for name, member := range sym.Members {
				switch member.Type {
				case symbol.FunctionSymbol:
					content.WriteString(fmt.Sprintf("- `%s()` - function\n", name))
				case symbol.ClassSymbol:
					content.WriteString(fmt.Sprintf("- `%s` - class\n", name))
				case symbol.VariableSymbol:
					content.WriteString(fmt.Sprintf("- `%s` - variable\n", name))
				}
			}
			content.WriteString("\n")
		}

	case symbol.BuiltinSymbol:
		content.WriteString(fmt.Sprintf("**Built-in Function**: `%s`\n\n", sym.Name))
		content.WriteString(fmt.Sprintf("**Type**: `%s`\n\n", sym.DataType))

		// Add documentation for common built-ins
		switch sym.Name {
		case "print":
			content.WriteString("Prints values to standard output\n")
		case "len":
			content.WriteString("Returns the length of a sequence or collection\n")
		case "str":
			content.WriteString("Converts a value to its string representation\n")
		case "int":
			content.WriteString("Converts a value to an integer\n")
		case "float":
			content.WriteString("Converts a value to a floating-point number\n")
		case "bool":
			content.WriteString("Converts a value to a boolean\n")
		}

	default:
		return ""
	}

	return content.String()
}

// isIdentifierChar checks if a character can be part of an identifier
func (s *Server) isIdentifierChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_'
}

// getCompletionItemKind converts symbol type to LSP completion item kind
func (s *Server) getCompletionItemKind(symType symbol.SymbolType) protocol.CompletionItemKind {
	switch symType {
	case symbol.VariableSymbol:
		return protocol.CompletionItemKindVariable
	case symbol.FunctionSymbol:
		return protocol.CompletionItemKindFunction
	case symbol.ClassSymbol:
		return protocol.CompletionItemKindClass
	case symbol.ParameterSymbol:
		return protocol.CompletionItemKindVariable
	case symbol.ModuleSymbol:
		return protocol.CompletionItemKindModule
	case symbol.BuiltinSymbol:
		return protocol.CompletionItemKindFunction
	default:
		return protocol.CompletionItemKindText
	}
}

// MemberAccessContext represents context for member access completion
type MemberAccessContext struct {
	IsMemberAccess bool
	ObjectName     string
	MemberPrefix   string
}

// getMemberAccessContext analyzes if the current position is for member access completion
func (s *Server) getMemberAccessContext(text string, position protocol.Position) MemberAccessContext {
	lines := strings.Split(text, "\n")
	if position.Line >= len(lines) {
		return MemberAccessContext{IsMemberAccess: false}
	}

	line := lines[position.Line]
	if position.Character > len(line) {
		return MemberAccessContext{IsMemberAccess: false}
	}

	// Look for pattern: identifier.partial_member
	// Find the position of the dot
	dotPos := -1
	for i := position.Character - 1; i >= 0; i-- {
		if line[i] == '.' {
			dotPos = i
			break
		}
		// If we hit whitespace or other non-identifier chars without finding a dot, it's not member access
		if !s.isIdentifierChar(rune(line[i])) {
			break
		}
	}

	if dotPos == -1 {
		return MemberAccessContext{IsMemberAccess: false}
	}

	// Extract object name (before the dot)
	objectStart := dotPos - 1
	for objectStart >= 0 && s.isIdentifierChar(rune(line[objectStart])) {
		objectStart--
	}
	objectStart++ // Move to the first character of the identifier

	if objectStart >= dotPos {
		return MemberAccessContext{IsMemberAccess: false}
	}

	objectName := line[objectStart:dotPos]

	// Extract member prefix (after the dot)
	memberPrefix := line[dotPos+1 : position.Character]

	return MemberAccessContext{
		IsMemberAccess: true,
		ObjectName:     objectName,
		MemberPrefix:   memberPrefix,
	}
}

// getWorkspaceDefinitionLocation returns definition locations using the workspace manager (supports cross-file definitions)
func (s *Server) getWorkspaceDefinitionLocation(uri string, position protocol.Position) ([]protocol.Location, error) {
	doc, exists := s.workspaceManager.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get the identifier at the position
	identifier := s.getIdentifierAtPosition(doc.Text, position)
	if identifier == "" {
		return []protocol.Location{}, nil // No identifier at position
	}

	// Try to get symbol at specific position first (for scope-aware lookup)
	sym := doc.Analyzer.GetSymbolAtPosition(position.Line+1, position.Character)
	if sym == nil {
		// Fall back to global lookup (this now includes imported symbols from workspace manager)
		var exists bool
		sym, exists = doc.Analyzer.GetSymbolTable().Lookup(identifier)
		if !exists {
			return []protocol.Location{}, nil // Symbol not found
		}
	}

	// Don't provide definitions for built-in symbols
	if sym.Type == symbol.BuiltinSymbol {
		return []protocol.Location{}, nil
	}

	// For module symbols, try to find the actual import statement or module file
	if sym.Type == symbol.ModuleSymbol {
		return s.getModuleDefinitionLocation(sym, uri)
	}

	// Check if this symbol is from an imported module
	if sym.Token.Line <= 0 || sym.Token.Line == 1 {
		// This might be an imported symbol - try to find it in the module cache
		return s.findSymbolInImportedModules(identifier, uri)
	}

	// Create location from symbol's token position
	// First, determine which file the symbol is in
	var symbolURI string
	if sym.Token.Line > 0 {
		// Symbol is in current file
		symbolURI = uri
	} else {
		// For imported symbols, we would need to track which file they came from
		// For now, assume same file
		symbolURI = uri
	}

	location := protocol.Location{
		URI: symbolURI,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      sym.Token.Line - 1, // Convert 1-based to 0-based
				Character: sym.Token.Column - 1,
			},
			End: protocol.Position{
				Line:      sym.Token.Line - 1,
				Character: sym.Token.Column - 1 + len(sym.Name),
			},
		},
	}

	return []protocol.Location{location}, nil
}

// getModuleDefinitionLocation finds the definition location for a module import
func (s *Server) getModuleDefinitionLocation(moduleSymbol *symbol.Symbol, currentURI string) ([]protocol.Location, error) {
	// For module symbols, the definition is the import statement itself
	if moduleSymbol.Token.Line > 0 {
		location := protocol.Location{
			URI: currentURI,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      moduleSymbol.Token.Line - 1,
					Character: moduleSymbol.Token.Column - 1,
				},
				End: protocol.Position{
					Line:      moduleSymbol.Token.Line - 1,
					Character: moduleSymbol.Token.Column - 1 + len(moduleSymbol.Name),
				},
			},
		}
		return []protocol.Location{location}, nil
	}

	return []protocol.Location{}, nil
}

// findSymbolInImportedModules searches for a symbol across all imported modules
func (s *Server) findSymbolInImportedModules(symbolName, currentURI string) ([]protocol.Location, error) {
	// Get current document to access its imports
	_, exists := s.workspaceManager.GetDocument(currentURI)
	if !exists {
		return []protocol.Location{}, nil
	}

	// Check if this is a member access (e.g., module.symbol)
	// For now, we'll implement basic symbol lookup
	// A more sophisticated implementation would parse member expressions

	// Search through the workspace's module cache for symbols
	s.workspaceManager.mu.RLock()
	defer s.workspaceManager.mu.RUnlock()

	for filePath, cachedModule := range s.workspaceManager.moduleCache {
		if exportedSymbol, exists := cachedModule.ExportedSymbols[symbolName]; exists {
			// Convert file path to URI
			moduleURI := "file://" + filePath

			location := protocol.Location{
				URI: moduleURI,
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      exportedSymbol.Token.Line - 1,
						Character: exportedSymbol.Token.Column - 1,
					},
					End: protocol.Position{
						Line:      exportedSymbol.Token.Line - 1,
						Character: exportedSymbol.Token.Column - 1 + len(exportedSymbol.Name),
					},
				},
			}
			return []protocol.Location{location}, nil
		}
	}

	// Symbol not found in imported modules
	return []protocol.Location{}, nil
}
