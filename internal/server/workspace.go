package server

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/javanhut/carrion-lsp/internal/carrion/analyzer"
	"github.com/javanhut/carrion-lsp/internal/carrion/ast"
	"github.com/javanhut/carrion-lsp/internal/carrion/lexer"
	"github.com/javanhut/carrion-lsp/internal/carrion/parser"
	"github.com/javanhut/carrion-lsp/internal/carrion/symbol"
	"github.com/javanhut/carrion-lsp/internal/carrion/token"
	"github.com/javanhut/carrion-lsp/internal/protocol"
)

// WorkspaceManager handles multi-file analysis and dependency tracking
type WorkspaceManager struct {
	mu            sync.RWMutex
	documents     sync.Map                      // URI -> Document (thread-safe map)
	dependencies  sync.Map                      // file -> []string (thread-safe map)
	dependents    sync.Map                      // file -> []string (thread-safe map)
	moduleCache   sync.Map                      // module path -> CachedModule (thread-safe map)
	resolver      *ModuleResolver
	analysisQueue chan string // Files that need re-analysis
	isAnalyzing   bool
	symbolIndex   sync.Map                      // symbol name -> GlobalSymbolEntry (thread-safe map)
	shutdownCh    chan struct{}                 // Signal shutdown to worker
	workerDone    chan struct{}                 // Signal when worker is done
}

// CachedModule represents a cached analysis result for a module
type CachedModule struct {
	FilePath        string
	LastModified    time.Time
	Analyzer        *analyzer.Analyzer
	ExportedSymbols map[string]*symbol.Symbol // Symbols available for import
	Imports         []ImportInfo
	Errors          []string
}

// ImportInfo represents information about an import statement
type ImportInfo struct {
	ModuleName      string
	Alias           string                    // Empty if no alias
	ModuleInfo      *ModuleInfo               // Resolved module information
	ImportedSymbols map[string]*symbol.Symbol // Symbols imported from this module
}

// GlobalSymbolEntry represents a symbol that can be found across the workspace
type GlobalSymbolEntry struct {
	Symbol   *symbol.Symbol
	FilePath string
	Module   string
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(workspaceRoot, carrionPath string) *WorkspaceManager {
	wm := &WorkspaceManager{
		resolver:      NewModuleResolver(workspaceRoot, carrionPath),
		analysisQueue: make(chan string, 1000), // Increased buffer size to reduce blocking
		shutdownCh:    make(chan struct{}),
		workerDone:    make(chan struct{}),
	}

	// Start background analysis worker
	go wm.analysisWorker()

	return wm
}

// OpenDocument handles opening a document with workspace-aware analysis
func (wm *WorkspaceManager) OpenDocument(params *protocol.DidOpenTextDocumentParams) (*Document, error) {
	uri := params.TextDocument.URI
	if _, exists := wm.documents.Load(uri); exists {
		return nil, fmt.Errorf("document %s is already open", uri)
	}

	doc := &Document{
		URI:        uri,
		LanguageID: params.TextDocument.LanguageID,
		Version:    params.TextDocument.Version,
		Text:       params.TextDocument.Text,
	}

	// Analyze the document with workspace context
	if err := wm.analyzeDocumentWithWorkspace(doc); err != nil {
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

	wm.documents.Store(uri, doc)

	// Queue dependent files for re-analysis
	wm.queueDependentsForAnalysis(uri)

	return doc, nil
}

// ChangeDocument handles document changes with dependency tracking
func (wm *WorkspaceManager) ChangeDocument(params *protocol.DidChangeTextDocumentParams) (*Document, error) {
	uri := params.TextDocument.URI
	docInterface, exists := wm.documents.Load(uri)
	if !exists {
		return nil, fmt.Errorf("document %s is not open", uri)
	}
	doc := docInterface.(*Document)

	// Update document version and content
	doc.Version = params.TextDocument.Version
	for _, change := range params.ContentChanges {
		if change.Range == nil {
			doc.Text = change.Text
		} else {
			doc.Text = change.Text
		}
	}

	// Re-analyze with workspace context
	if err := wm.analyzeDocumentWithWorkspace(doc); err != nil {
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

	// Queue dependent files for re-analysis
	wm.queueDependentsForAnalysis(uri)

	return doc, nil
}

// CloseDocument handles closing a document
func (wm *WorkspaceManager) CloseDocument(params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI
	if _, exists := wm.documents.Load(uri); !exists {
		return fmt.Errorf("document %s is not open", uri)
	}

	// Remove from documents but keep in cache for dependencies
	wm.documents.Delete(uri)

	return nil
}

// analyzeDocumentWithWorkspace performs workspace-aware analysis
func (wm *WorkspaceManager) analyzeDocumentWithWorkspace(doc *Document) error {
	// Only analyze Carrion files
	if doc.LanguageID != "carrion" && !strings.HasSuffix(doc.URI, ".crl") {
		doc.Analyzer = nil
		doc.Diagnostics = nil
		return nil
	}

	// Parse the document
	l := lexer.New(doc.Text)
	p := parser.New(l)
	program := p.ParseProgram()

	// Create analyzer
	a := analyzer.New()

	// Process imports before analyzing
	importInfos, err := wm.processImports(program, doc.URI)
	if err != nil {
		// Add import errors as diagnostics but continue analysis
		doc.Diagnostics = append(doc.Diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 0},
			},
			Severity: &[]protocol.DiagnosticSeverity{protocol.DiagnosticSeverityWarning}[0],
			Source:   "carrion-import",
			Message:  err.Error(),
		})
	}

	// Add imported symbols to the analyzer's symbol table
	for _, importInfo := range importInfos {
		wm.addImportedSymbols(a, importInfo)
	}

	// Analyze the program
	_ = a.Analyze(program) // Ignore error - we use diagnostics instead
	doc.Analyzer = a

	// Convert analyzer diagnostics to LSP diagnostics
	doc.Diagnostics = append(doc.Diagnostics, convertAnalyzerDiagnostics(a.GetDiagnostics())...)

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

	// Update dependency tracking
	wm.updateDependencies(doc.URI, importInfos)

	// Cache the analysis result
	wm.cacheModuleAnalysis(doc.URI, a, importInfos)

	return nil
}

// processImports resolves and loads all imports for a document
func (wm *WorkspaceManager) processImports(program *ast.Program, currentURI string) ([]ImportInfo, error) {
	var imports []ImportInfo
	var errors []string

	// Extract import statements from the AST
	for _, stmt := range program.Statements {
		if importStmt, ok := stmt.(*ast.ImportStatement); ok {
			moduleName := importStmt.Module.Value
			alias := ""
			if importStmt.Alias != nil {
				alias = importStmt.Alias.Value
			}

			// Resolve the import
			moduleInfo, err := wm.resolver.ResolveImport(moduleName, currentURI)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to resolve import '%s': %s", moduleName, err.Error()))
				continue
			}

			// Load symbols from the module
			importedSymbols, err := wm.loadModuleSymbols(moduleInfo)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to load symbols from '%s': %s", moduleName, err.Error()))
				continue
			}

			imports = append(imports, ImportInfo{
				ModuleName:      moduleName,
				Alias:           alias,
				ModuleInfo:      moduleInfo,
				ImportedSymbols: importedSymbols,
			})
		}
	}

	var finalError error
	if len(errors) > 0 {
		finalError = fmt.Errorf("import errors: %s", strings.Join(errors, "; "))
	}

	return imports, finalError
}

// loadModuleSymbols loads symbols from a module
func (wm *WorkspaceManager) loadModuleSymbols(moduleInfo *ModuleInfo) (map[string]*symbol.Symbol, error) {
	if moduleInfo.IsBuiltin {
		return wm.getBuiltinModuleSymbols(moduleInfo.Name), nil
	}

	// Check cache first
	if cachedInterface, exists := wm.moduleCache.Load(moduleInfo.FilePath); exists {
		cached := cachedInterface.(*CachedModule)
		// TODO: Check if file has been modified
		return cached.ExportedSymbols, nil
	}

	// Load and analyze the module file
	return wm.analyzeModuleFile(moduleInfo.FilePath)
}

// analyzeModuleFile analyzes a module file and extracts exported symbols
func (wm *WorkspaceManager) analyzeModuleFile(filePath string) (map[string]*symbol.Symbol, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse and analyze
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	a := analyzer.New()
	_ = a.Analyze(program)

	// Extract top-level symbols (these are exportable)
	exportedSymbols := make(map[string]*symbol.Symbol)
	for name, sym := range a.GetSymbolTable().GetAllSymbols() {
		// Only export top-level symbols
		if sym.Type == symbol.FunctionSymbol || sym.Type == symbol.ClassSymbol || sym.Type == symbol.VariableSymbol {
			exportedSymbols[name] = sym
		}
	}

	return exportedSymbols, nil
}

// getBuiltinModuleSymbols returns symbols for built-in modules
func (wm *WorkspaceManager) getBuiltinModuleSymbols(moduleName string) map[string]*symbol.Symbol {
	symbols := make(map[string]*symbol.Symbol)

	// Define built-in module symbols based on module name
	switch moduleName {
	case "os":
		symbols["listdir"] = &symbol.Symbol{Name: "listdir", Type: symbol.FunctionSymbol, DataType: "function"}
		symbols["getcwd"] = &symbol.Symbol{Name: "getcwd", Type: symbol.FunctionSymbol, DataType: "function"}
		symbols["chdir"] = &symbol.Symbol{Name: "chdir", Type: symbol.FunctionSymbol, DataType: "function"}
	case "file":
		symbols["open"] = &symbol.Symbol{Name: "open", Type: symbol.FunctionSymbol, DataType: "function"}
		symbols["read"] = &symbol.Symbol{Name: "read", Type: symbol.FunctionSymbol, DataType: "function"}
		symbols["write"] = &symbol.Symbol{Name: "write", Type: symbol.FunctionSymbol, DataType: "function"}
	case "http":
		symbols["get"] = &symbol.Symbol{Name: "get", Type: symbol.FunctionSymbol, DataType: "function"}
		symbols["post"] = &symbol.Symbol{Name: "post", Type: symbol.FunctionSymbol, DataType: "function"}
	case "time":
		symbols["now"] = &symbol.Symbol{Name: "now", Type: symbol.FunctionSymbol, DataType: "function"}
		symbols["sleep"] = &symbol.Symbol{Name: "sleep", Type: symbol.FunctionSymbol, DataType: "function"}
	}

	return symbols
}

// addImportedSymbols adds imported symbols to the analyzer's symbol table
func (wm *WorkspaceManager) addImportedSymbols(a *analyzer.Analyzer, importInfo ImportInfo) {
	symbolName := importInfo.ModuleName
	if importInfo.Alias != "" {
		symbolName = importInfo.Alias
	}

	// Create a module symbol that contains all imported symbols
	moduleSymbol := &symbol.Symbol{
		Name:     symbolName,
		Type:     symbol.ModuleSymbol,
		DataType: "module",
		Members:  importInfo.ImportedSymbols,
		Token:    token.Token{Type: token.IDENT, Literal: symbolName, Line: 1, Column: 1},
	}

	// Add to global scope
	err := a.GetSymbolTable().GlobalScope.Define(moduleSymbol)
	if err != nil {
		// Log the error but continue - don't fail the entire import process
		fmt.Printf("Warning: failed to add imported module '%s': %s\n", symbolName, err.Error())
	}
}

// updateDependencies updates the dependency tracking
func (wm *WorkspaceManager) updateDependencies(uri string, imports []ImportInfo) {
	// Clear old dependencies
	if oldDepsInterface, exists := wm.dependencies.Load(uri); exists {
		oldDeps := oldDepsInterface.([]string)
		for _, dep := range oldDeps {
			wm.removeDependency(dep, uri)
		}
	}

	// Add new dependencies
	var newDeps []string
	for _, importInfo := range imports {
		if !importInfo.ModuleInfo.IsBuiltin && importInfo.ModuleInfo.FilePath != "" {
			newDeps = append(newDeps, importInfo.ModuleInfo.FilePath)
			wm.addDependency(importInfo.ModuleInfo.FilePath, uri)
		}
	}

	wm.dependencies.Store(uri, newDeps)
}

// addDependency adds a dependency relationship
func (wm *WorkspaceManager) addDependency(dependency, dependent string) {
	for {
		dependentsInterface, _ := wm.dependents.LoadOrStore(dependency, []string{})
		dependents := dependentsInterface.([]string)

		// Add if not already present
		for _, existing := range dependents {
			if existing == dependent {
				return
			}
		}

		updatedDependents := append(dependents, dependent)
		// Use compare-and-swap to handle concurrent modifications
		if wm.dependents.CompareAndSwap(dependency, dependents, updatedDependents) {
			break
		}
		// If CAS failed, retry the operation
	}
}

// removeDependency removes a dependency relationship
func (wm *WorkspaceManager) removeDependency(dependency, dependent string) {
	for {
		depsInterface, exists := wm.dependents.Load(dependency)
		if !exists {
			return
		}
		
		deps := depsInterface.([]string)
		found := false
		var updatedDeps []string
		
		for i, dep := range deps {
			if dep == dependent {
				updatedDeps = append(deps[:i], deps[i+1:]...)
				found = true
				break
			}
		}
		
		if !found {
			return
		}
		
		// Use compare-and-swap to handle concurrent modifications
		if wm.dependents.CompareAndSwap(dependency, deps, updatedDeps) {
			break
		}
		// If CAS failed, retry the operation
	}
}

// cacheModuleAnalysis caches the analysis result for a module
func (wm *WorkspaceManager) cacheModuleAnalysis(filePath string, a *analyzer.Analyzer, imports []ImportInfo) {
	exportedSymbols := make(map[string]*symbol.Symbol)
	for name, sym := range a.GetSymbolTable().GetAllSymbols() {
		if sym.Type == symbol.FunctionSymbol || sym.Type == symbol.ClassSymbol || sym.Type == symbol.VariableSymbol {
			exportedSymbols[name] = sym
		}
	}

	cachedModule := &CachedModule{
		FilePath:        filePath,
		LastModified:    time.Now(),
		Analyzer:        a,
		ExportedSymbols: exportedSymbols,
		Imports:         imports,
		Errors:          a.GetErrors(),
	}
	wm.moduleCache.Store(filePath, cachedModule)
}

// queueDependentsForAnalysis queues dependent files for re-analysis
func (wm *WorkspaceManager) queueDependentsForAnalysis(uri string) {
	if dependentsInterface, exists := wm.dependents.Load(uri); exists {
		dependents := dependentsInterface.([]string)
		for _, dependent := range dependents {
			select {
			case wm.analysisQueue <- dependent:
				// Successfully queued
			default:
				// Queue is full, implement priority handling
				// Remove oldest item and add new one to prevent queue overflow
				select {
				case <-wm.analysisQueue:
					// Removed oldest item
					select {
					case wm.analysisQueue <- dependent:
						// Successfully added new item
					default:
						// Still full, skip this one
					}
				default:
					// Queue cleared in between, skip
				}
			}
		}
	}
}

// analysisWorker processes the analysis queue in the background
func (wm *WorkspaceManager) analysisWorker() {
	defer close(wm.workerDone)
	
	for {
		select {
		case uri := <-wm.analysisQueue:
			if docInterface, exists := wm.documents.Load(uri); exists {
				doc := docInterface.(*Document)
				wm.analyzeDocumentWithWorkspace(doc)
			}
		case <-wm.shutdownCh:
			return
		}
	}
}

// GetDocument retrieves a document by URI
func (wm *WorkspaceManager) GetDocument(uri string) (*Document, bool) {
	docInterface, exists := wm.documents.Load(uri)
	if !exists {
		return nil, false
	}
	return docInterface.(*Document), true
}

// GetAllDocuments returns all open documents
func (wm *WorkspaceManager) GetAllDocuments() map[string]*Document {
	result := make(map[string]*Document)
	wm.documents.Range(func(key, value interface{}) bool {
		uri := key.(string)
		doc := value.(*Document)
		result[uri] = doc
		return true
	})
	return result
}

// Shutdown gracefully shuts down the workspace manager
func (wm *WorkspaceManager) Shutdown() error {
	// Signal the worker to stop
	close(wm.shutdownCh)
	
	// Wait for worker to finish
	<-wm.workerDone
	
	return nil
}
