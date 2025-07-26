package server

import (
	"fmt"
	"strings"
	"sync"

	"github.com/javanhut/carrion-lsp/internal/carrion/analyzer"
	"github.com/javanhut/carrion-lsp/internal/carrion/lexer"
	"github.com/javanhut/carrion-lsp/internal/carrion/parser"
	"github.com/javanhut/carrion-lsp/internal/carrion/symbol"
	"github.com/javanhut/carrion-lsp/internal/protocol"
)

// Document represents a text document managed by the LSP server
type Document struct {
	URI         string
	LanguageID  string
	Version     int
	Text        string
	Analyzer    *analyzer.Analyzer
	Diagnostics []protocol.Diagnostic
}

// DocumentManager manages text documents and their analysis
type DocumentManager struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

// NewDocumentManager creates a new document manager
func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[string]*Document),
	}
}

// OpenDocument handles opening a document
func (dm *DocumentManager) OpenDocument(params *protocol.DidOpenTextDocumentParams) (*Document, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	uri := params.TextDocument.URI
	if _, exists := dm.documents[uri]; exists {
		return nil, fmt.Errorf("document %s is already open", uri)
	}

	doc := &Document{
		URI:        uri,
		LanguageID: params.TextDocument.LanguageID,
		Version:    params.TextDocument.Version,
		Text:       params.TextDocument.Text,
	}

	// Analyze the document
	if err := dm.analyzeDocument(doc); err != nil {
		// Don't fail on analysis errors, just log them
		doc.Diagnostics = []protocol.Diagnostic{
			{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 0},
				},
				Severity: &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityError}[0],
				Source:   "carrion-lsp",
				Message:  fmt.Sprintf("Analysis failed: %s", err.Error()),
			},
		}
	}

	dm.documents[uri] = doc
	return doc, nil
}

// ChangeDocument handles document changes
func (dm *DocumentManager) ChangeDocument(params *protocol.DidChangeTextDocumentParams) (*Document, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	uri := params.TextDocument.URI
	doc, exists := dm.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	// Update document version
	doc.Version = params.TextDocument.Version

	// Apply content changes
	for _, change := range params.ContentChanges {
		if change.Range == nil {
			// Full document update
			doc.Text = change.Text
		} else {
			// Incremental update (for now, we'll implement full document sync)
			// In a production implementation, you'd want to handle incremental changes
			doc.Text = change.Text
		}
	}

	// Re-analyze the document
	if err := dm.analyzeDocument(doc); err != nil {
		// Don't fail on analysis errors, just create diagnostic
		doc.Diagnostics = []protocol.Diagnostic{
			{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 0},
				},
				Severity: &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityError}[0],
				Source:   "carrion-lsp",
				Message:  fmt.Sprintf("Analysis failed: %s", err.Error()),
			},
		}
	}

	return doc, nil
}

// CloseDocument handles closing a document
func (dm *DocumentManager) CloseDocument(params *protocol.DidCloseTextDocumentParams) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	uri := params.TextDocument.URI
	if _, exists := dm.documents[uri]; !exists {
		return fmt.Errorf("document %s is not open", uri)
	}

	delete(dm.documents, uri)
	return nil
}

// GetDocument retrieves a document by URI
func (dm *DocumentManager) GetDocument(uri string) (*Document, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	doc, exists := dm.documents[uri]
	return doc, exists
}

// GetAllDocuments returns all open documents
func (dm *DocumentManager) GetAllDocuments() map[string]*Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*Document)
	for uri, doc := range dm.documents {
		result[uri] = doc
	}
	return result
}

// analyzeDocument performs semantic analysis on a document
func (dm *DocumentManager) analyzeDocument(doc *Document) error {
	// Only analyze Carrion files
	if doc.LanguageID != "carrion" && !strings.HasSuffix(doc.URI, ".crl") {
		doc.Analyzer = nil
		doc.Diagnostics = nil
		return nil
	}

	// Create lexer and parser
	l := lexer.New(doc.Text)
	p := parser.New(l)
	program := p.ParseProgram()

	// Create analyzer
	a := analyzer.New()

	// Analyze the program
	_ = a.Analyze(program) // Ignore the error - we'll use diagnostics instead
	doc.Analyzer = a

	// Convert analyzer diagnostics to LSP diagnostics
	doc.Diagnostics = convertAnalyzerDiagnostics(a.GetDiagnostics())

	// Add parser errors as diagnostics
	for _, parseError := range p.Errors() {
		diagnostic := protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 0},
			},
			Severity: &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityError}[0],
			Source:   "carrion-parser",
			Message:  parseError,
		}
		doc.Diagnostics = append(doc.Diagnostics, diagnostic)
	}

	// Don't return the analysis error - we've converted all errors to diagnostics
	// This allows the LSP to show detailed diagnostics instead of a generic error
	return nil
}

// convertAnalyzerDiagnostics converts analyzer diagnostics to LSP diagnostics
func convertAnalyzerDiagnostics(analyzerDiags []analyzer.Diagnostic) []protocol.Diagnostic {
	var diagnostics []protocol.Diagnostic

	for _, diag := range analyzerDiags {
		lspDiag := protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      diag.Range.Start.Line,
					Character: diag.Range.Start.Character,
				},
				End: protocol.Position{
					Line:      diag.Range.End.Line,
					Character: diag.Range.End.Character,
				},
			},
			Source:  diag.Source,
			Message: diag.Message,
		}

		// Convert severity
		switch diag.Severity {
		case analyzer.DiagnosticError:
			lspDiag.Severity = &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityError}[0]
		case analyzer.DiagnosticWarning:
			lspDiag.Severity = &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityWarning}[0]
		case analyzer.DiagnosticInformation:
			lspDiag.Severity = &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityInformation}[0]
		case analyzer.DiagnosticHint:
			lspDiag.Severity = &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityHint}[0]
		}

		diagnostics = append(diagnostics, lspDiag)
	}

	return diagnostics
}

// GetCompletionItems returns completion items for a position in a document
func (dm *DocumentManager) GetCompletionItems(uri string, position protocol.Position) ([]protocol.CompletionItem, error) {
	doc, exists := dm.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get prefix at position (simplified implementation)
	prefix := dm.getPrefixAtPosition(doc.Text, position)

	// Get completion items from analyzer
	symbols := doc.Analyzer.GetCompletionItems(position.Line, position.Character, prefix)

	var items []protocol.CompletionItem
	for _, sym := range symbols {
		kind := getCompletionItemKind(sym.Type)
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

// GetHoverInformation returns hover information for a position in a document
func (dm *DocumentManager) GetHoverInformation(uri string, position protocol.Position) (*protocol.Hover, error) {
	doc, exists := dm.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get the identifier at the position
	identifier := dm.getIdentifierAtPosition(doc.Text, position)
	if identifier == "" {
		return nil, nil // No identifier at position
	}

	// Try to get symbol at specific position first (for scope-aware lookup)
	symbol := doc.Analyzer.GetSymbolAtPosition(position.Line+1, position.Character) // Convert 0-based to 1-based
	if symbol == nil {
		// Fall back to global lookup
		var exists bool
		symbol, exists = doc.Analyzer.GetSymbolTable().Lookup(identifier)
		if !exists {
			return nil, nil // Symbol not found
		}
	}

	// Create hover content based on symbol type
	content := dm.createHoverContent(symbol)
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

// getIdentifierAtPosition extracts the identifier at the given position
func (dm *DocumentManager) getIdentifierAtPosition(text string, position protocol.Position) string {
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
	for start > 0 && isIdentifierChar(rune(line[start-1])) {
		start--
	}

	// Move end forward to find the end of the identifier
	for end < len(line) && isIdentifierChar(rune(line[end])) {
		end++
	}

	// Return the identifier if we found one
	if start < end && isIdentifierChar(rune(line[start])) {
		return line[start:end]
	}

	return ""
}

// createHoverContent creates markdown content for hover information
func (dm *DocumentManager) createHoverContent(sym *symbol.Symbol) string {
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

// GetDefinitionLocation returns the definition location for a symbol at a position
func (dm *DocumentManager) GetDefinitionLocation(uri string, position protocol.Position) ([]protocol.Location, error) {
	doc, exists := dm.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get the identifier at the position
	identifier := dm.getIdentifierAtPosition(doc.Text, position)
	if identifier == "" {
		return []protocol.Location{}, nil // No identifier at position
	}

	// Try to get symbol at specific position first (for scope-aware lookup)
	sym := doc.Analyzer.GetSymbolAtPosition(position.Line+1, position.Character)
	if sym == nil {
		// Fall back to global lookup
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

	// Create location from symbol's token position
	location := protocol.Location{
		URI: uri, // For now, assume all symbols are in the same file
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

// FormatDocument formats a document and returns the text edits
func (dm *DocumentManager) FormatDocument(uri string, options protocol.FormattingOptions) ([]protocol.TextEdit, error) {
	doc, exists := dm.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	// Only format Carrion files
	if doc.LanguageID != "carrion" && !strings.HasSuffix(doc.URI, ".crl") {
		return []protocol.TextEdit{}, nil
	}

	formatter := NewCarrionFormatter(options)
	edits := formatter.FormatDocument(doc.Text)

	return edits, nil
}

// GetReferences returns all references to a symbol at the given position
func (dm *DocumentManager) GetReferences(uri string, position protocol.Position, includeDeclaration bool) ([]protocol.Location, error) {
	doc, exists := dm.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get the identifier at the position
	identifier := dm.getIdentifierAtPosition(doc.Text, position)
	if identifier == "" {
		return []protocol.Location{}, nil // No identifier at position
	}

	// Find references using the analyzer
	references := doc.Analyzer.FindReferences(position.Line+1, position.Character, includeDeclaration)

	// Convert analyzer references to LSP locations
	var locations []protocol.Location
	for _, ref := range references {
		location := protocol.Location{
			URI: uri, // For now, assume all references are in the same file
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      ref.Line - 1, // Convert 1-based to 0-based
					Character: ref.Column - 1,
				},
				End: protocol.Position{
					Line:      ref.Line - 1,
					Character: ref.Column - 1 + ref.Length,
				},
			},
		}
		locations = append(locations, location)
	}

	return locations, nil
}

// GetDocumentSymbols returns all symbols in a document for outline view
func (dm *DocumentManager) GetDocumentSymbols(uri string) ([]protocol.DocumentSymbol, error) {
	doc, exists := dm.GetDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}

	if doc.Analyzer == nil {
		return nil, fmt.Errorf("document %s has no analyzer", uri)
	}

	// Get all symbols from the analyzer
	symbols := doc.Analyzer.GetSymbolTable().GetAllSymbols()

	var documentSymbols []protocol.DocumentSymbol
	for name, sym := range symbols {
		if sym.Token.Line <= 0 {
			continue // Skip symbols without valid positions (like built-ins)
		}

		symbolKind := dm.getSymbolKind(sym.Type)

		documentSymbol := protocol.DocumentSymbol{
			Name:   name,
			Detail: dm.getSymbolDetail(sym),
			Kind:   symbolKind,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      sym.Token.Line - 1, // Convert 1-based to 0-based
					Character: sym.Token.Column - 1,
				},
				End: protocol.Position{
					Line:      sym.Token.Line - 1,
					Character: sym.Token.Column - 1 + len(name),
				},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{
					Line:      sym.Token.Line - 1,
					Character: sym.Token.Column - 1,
				},
				End: protocol.Position{
					Line:      sym.Token.Line - 1,
					Character: sym.Token.Column - 1 + len(name),
				},
			},
		}

		// Add children for classes (methods)
		if sym.Type == symbol.ClassSymbol && len(sym.Members) > 0 {
			for memberName, member := range sym.Members {
				if member.Token.Line > 0 {
					childSymbol := protocol.DocumentSymbol{
						Name:   memberName,
						Detail: dm.getSymbolDetail(member),
						Kind:   dm.getSymbolKind(member.Type),
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      member.Token.Line - 1,
								Character: member.Token.Column - 1,
							},
							End: protocol.Position{
								Line:      member.Token.Line - 1,
								Character: member.Token.Column - 1 + len(memberName),
							},
						},
						SelectionRange: protocol.Range{
							Start: protocol.Position{
								Line:      member.Token.Line - 1,
								Character: member.Token.Column - 1,
							},
							End: protocol.Position{
								Line:      member.Token.Line - 1,
								Character: member.Token.Column - 1 + len(memberName),
							},
						},
					}
					documentSymbol.Children = append(documentSymbol.Children, childSymbol)
				}
			}
		}

		documentSymbols = append(documentSymbols, documentSymbol)
	}

	return documentSymbols, nil
}

// getSymbolKind converts analyzer symbol type to LSP symbol kind
func (dm *DocumentManager) getSymbolKind(symType symbol.SymbolType) protocol.SymbolKind {
	switch symType {
	case symbol.VariableSymbol:
		return protocol.SymbolKindVariable
	case symbol.FunctionSymbol:
		return protocol.SymbolKindFunction
	case symbol.ClassSymbol:
		return protocol.SymbolKindClass
	case symbol.ParameterSymbol:
		return protocol.SymbolKindVariable
	case symbol.ModuleSymbol:
		return protocol.SymbolKindModule
	case symbol.BuiltinSymbol:
		return protocol.SymbolKindFunction
	default:
		return protocol.SymbolKindVariable
	}
}

// getSymbolDetail returns a detail string for a symbol
func (dm *DocumentManager) getSymbolDetail(sym *symbol.Symbol) string {
	switch sym.Type {
	case symbol.FunctionSymbol:
		var params []string
		for _, param := range sym.Parameters {
			params = append(params, param.Name)
		}
		detail := fmt.Sprintf("(%s)", strings.Join(params, ", "))
		if sym.ReturnType != "" && sym.ReturnType != "unknown" {
			detail += fmt.Sprintf(" -> %s", sym.ReturnType)
		}
		return detail
	case symbol.VariableSymbol:
		if sym.DataType != "" && sym.DataType != "unknown" {
			return sym.DataType
		}
		return "variable"
	case symbol.ClassSymbol:
		if sym.Parent != nil {
			return fmt.Sprintf("extends %s", sym.Parent.Name)
		}
		return "class"
	case symbol.ParameterSymbol:
		if sym.DataType != "" && sym.DataType != "unknown" {
			return sym.DataType
		}
		return "parameter"
	case symbol.ModuleSymbol:
		return "module"
	case symbol.BuiltinSymbol:
		return "built-in"
	default:
		return ""
	}
}

// getPrefixAtPosition extracts the word prefix at the given position
func (dm *DocumentManager) getPrefixAtPosition(text string, position protocol.Position) string {
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
	for start > 0 && isIdentifierChar(rune(line[start-1])) {
		start--
	}

	return line[start:position.Character]
}

// isIdentifierChar checks if a character can be part of an identifier
func isIdentifierChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_'
}

// getCompletionItemKind converts symbol type to LSP completion item kind
func getCompletionItemKind(symType symbol.SymbolType) protocol.CompletionItemKind {
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

// GetDiagnostics returns diagnostics for a document
func (dm *DocumentManager) GetDiagnostics(uri string) ([]protocol.Diagnostic, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	doc, exists := dm.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document %s not found", uri)
	}

	return doc.Diagnostics, nil
}
