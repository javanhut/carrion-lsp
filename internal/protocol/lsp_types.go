package protocol

// LSP Method names
const (
	MethodInitialize             = "initialize"
	MethodInitialized            = "initialized"
	MethodShutdown               = "shutdown"
	MethodExit                   = "exit"
	MethodTextDocumentDidOpen    = "textDocument/didOpen"
	MethodTextDocumentDidChange  = "textDocument/didChange"
	MethodTextDocumentDidClose   = "textDocument/didClose"
	MethodTextDocumentCompletion = "textDocument/completion"
	MethodTextDocumentHover      = "textDocument/hover"
	MethodTextDocumentDefinition = "textDocument/definition"
	MethodTextDocumentReferences = "textDocument/references"
	MethodTextDocumentFormatting = "textDocument/formatting"
	MethodWorkspaceSymbol        = "workspace/symbol"
	MethodTextDocumentSymbol     = "textDocument/documentSymbol"
	MethodTextDocumentDiagnostic = "textDocument/diagnostic"
)

// Initialize request parameters
type InitializeParams struct {
	ProcessID             *int               `json:"processId"`
	ClientInfo            *ClientInfo        `json:"clientInfo,omitempty"`
	Locale                string             `json:"locale,omitempty"`
	RootPath              *string            `json:"rootPath,omitempty"`
	RootURI               *string            `json:"rootUri"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	WorkspaceFolders      []WorkspaceFolder  `json:"workspaceFolders"`
}

// Client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Workspace folder
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// Client capabilities (simplified for now)
type ClientCapabilities struct {
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Workspace    *WorkspaceClientCapabilities    `json:"workspace,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Synchronization *TextDocumentSyncClientCapabilities   `json:"synchronization,omitempty"`
	Completion      *CompletionClientCapabilities         `json:"completion,omitempty"`
	Hover           *HoverClientCapabilities              `json:"hover,omitempty"`
	Definition      *DefinitionClientCapabilities         `json:"definition,omitempty"`
	References      *ReferenceClientCapabilities          `json:"references,omitempty"`
	Formatting      *DocumentFormattingClientCapabilities `json:"formatting,omitempty"`
	Diagnostic      *DiagnosticClientCapabilities         `json:"diagnostic,omitempty"`
}

type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
	WillSave            *bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   *bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             *bool `json:"didSave,omitempty"`
}

type CompletionClientCapabilities struct {
	DynamicRegistration *bool                       `json:"dynamicRegistration,omitempty"`
	CompletionItem      *CompletionItemCapabilities `json:"completionItem,omitempty"`
}

type CompletionItemCapabilities struct {
	SnippetSupport          *bool       `json:"snippetSupport,omitempty"`
	CommitCharactersSupport *bool       `json:"commitCharactersSupport,omitempty"`
	DocumentationFormat     []string    `json:"documentationFormat,omitempty"`
	DeprecatedSupport       *bool       `json:"deprecatedSupport,omitempty"`
	PreselectSupport        *bool       `json:"preselectSupport,omitempty"`
	TagSupport              *TagSupport `json:"tagSupport,omitempty"`
}

type TagSupport struct {
	ValueSet []int `json:"valueSet"`
}

type HoverClientCapabilities struct {
	DynamicRegistration *bool    `json:"dynamicRegistration,omitempty"`
	ContentFormat       []string `json:"contentFormat,omitempty"`
}

type DefinitionClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
	LinkSupport         *bool `json:"linkSupport,omitempty"`
}

type ReferenceClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

type DocumentFormattingClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

type DiagnosticClientCapabilities struct {
	DynamicRegistration    *bool `json:"dynamicRegistration,omitempty"`
	RelatedDocumentSupport *bool `json:"relatedDocumentSupport,omitempty"`
}

type WorkspaceClientCapabilities struct {
	ApplyEdit              *bool                                     `json:"applyEdit,omitempty"`
	WorkspaceEdit          *WorkspaceEditClientCapabilities          `json:"workspaceEdit,omitempty"`
	DidChangeConfiguration *DidChangeConfigurationClientCapabilities `json:"didChangeConfiguration,omitempty"`
	DidChangeWatchedFiles  *DidChangeWatchedFilesClientCapabilities  `json:"didChangeWatchedFiles,omitempty"`
	Symbol                 *WorkspaceSymbolClientCapabilities        `json:"symbol,omitempty"`
	ExecuteCommand         *ExecuteCommandClientCapabilities         `json:"executeCommand,omitempty"`
	WorkspaceFolders       *bool                                     `json:"workspaceFolders,omitempty"`
	Configuration          *bool                                     `json:"configuration,omitempty"`
}

type WorkspaceEditClientCapabilities struct {
	DocumentChanges    *bool    `json:"documentChanges,omitempty"`
	ResourceOperations []string `json:"resourceOperations,omitempty"`
	FailureHandling    *string  `json:"failureHandling,omitempty"`
}

type DidChangeConfigurationClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

type DidChangeWatchedFilesClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

type WorkspaceSymbolClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

type ExecuteCommandClientCapabilities struct {
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

// Initialize result
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// Server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Server capabilities
type ServerCapabilities struct {
	TextDocumentSync                *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	CompletionProvider              *CompletionOptions       `json:"completionProvider,omitempty"`
	HoverProvider                   *bool                    `json:"hoverProvider,omitempty"`
	DefinitionProvider              *bool                    `json:"definitionProvider,omitempty"`
	ReferencesProvider              *bool                    `json:"referencesProvider,omitempty"`
	DocumentFormattingProvider      *bool                    `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider *bool                    `json:"documentRangeFormattingProvider,omitempty"`
	DocumentSymbolProvider          *bool                    `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider         *bool                    `json:"workspaceSymbolProvider,omitempty"`
	DiagnosticProvider              *DiagnosticOptions       `json:"diagnosticProvider,omitempty"`
}

// Text document sync options
type TextDocumentSyncOptions struct {
	OpenClose         *bool                `json:"openClose,omitempty"`
	Change            TextDocumentSyncKind `json:"change,omitempty"`
	WillSave          *bool                `json:"willSave,omitempty"`
	WillSaveWaitUntil *bool                `json:"willSaveWaitUntil,omitempty"`
	Save              *SaveOptions         `json:"save,omitempty"`
}

// Text document sync kinds
type TextDocumentSyncKind int

const (
	TextDocumentSyncKindNone        TextDocumentSyncKind = 0
	TextDocumentSyncKindFull        TextDocumentSyncKind = 1
	TextDocumentSyncKindIncremental TextDocumentSyncKind = 2
)

// Save options
type SaveOptions struct {
	IncludeText *bool `json:"includeText,omitempty"`
}

// Completion options
type CompletionOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	AllCommitCharacters []string `json:"allCommitCharacters,omitempty"`
	ResolveProvider     *bool    `json:"resolveProvider,omitempty"`
}

// Diagnostic options
type DiagnosticOptions struct {
	Identifier            string `json:"identifier,omitempty"`
	InterFileDependencies bool   `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool   `json:"workspaceDiagnostics"`
}

// Text document item
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// Versioned text document identifier
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// Text document identifier
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// Position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a diagnostic, such as a compiler error
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           *DiagnosticSeverity            `json:"severity,omitempty"`
	Code               interface{}                    `json:"code,omitempty"`
	CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	Tags               []DiagnosticTag                `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
	Data               interface{}                    `json:"data,omitempty"`
}

// Diagnostic severity
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// Diagnostic tag
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// Code description
type CodeDescription struct {
	Href string `json:"href"`
}

// Diagnostic related information
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// Document symbol
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Tags           []SymbolTag      `json:"tags,omitempty"`
	Deprecated     *bool            `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// Symbol kind
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// Symbol tag
type SymbolTag int

const (
	SymbolTagDeprecated SymbolTag = 1
)

// Completion item
type CompletionItem struct {
	Label               string                      `json:"label"`
	LabelDetails        *CompletionItemLabelDetails `json:"labelDetails,omitempty"`
	Kind                *CompletionItemKind         `json:"kind,omitempty"`
	Tags                []CompletionItemTag         `json:"tags,omitempty"`
	Detail              string                      `json:"detail,omitempty"`
	Documentation       interface{}                 `json:"documentation,omitempty"`
	Deprecated          *bool                       `json:"deprecated,omitempty"`
	Preselect           *bool                       `json:"preselect,omitempty"`
	SortText            string                      `json:"sortText,omitempty"`
	FilterText          string                      `json:"filterText,omitempty"`
	InsertText          string                      `json:"insertText,omitempty"`
	InsertTextFormat    *InsertTextFormat           `json:"insertTextFormat,omitempty"`
	InsertTextMode      *InsertTextMode             `json:"insertTextMode,omitempty"`
	TextEdit            interface{}                 `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit                  `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string                    `json:"commitCharacters,omitempty"`
	Command             *Command                    `json:"command,omitempty"`
	Data                interface{}                 `json:"data,omitempty"`
}

// Completion item label details
type CompletionItemLabelDetails struct {
	Detail      string `json:"detail,omitempty"`
	Description string `json:"description,omitempty"`
}

// Completion item kind
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// Completion item tag
type CompletionItemTag int

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// Insert text format
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// Insert text mode
type InsertTextMode int

const (
	InsertTextModeAdjustIndentation InsertTextMode = 1
)

// Text edit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Command
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// Hover result
type Hover struct {
	Contents interface{} `json:"contents"`
	Range    *Range      `json:"range,omitempty"`
}

// Markup content
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// Markup kind
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// Document synchronization notification parameters

// DidOpenTextDocumentParams represents the parameters for textDocument/didOpen notification
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams represents the parameters for textDocument/didChange notification
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseTextDocumentParams represents the parameters for textDocument/didClose notification
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentContentChangeEvent represents a change to a text document
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`       // The range of the document that changed
	RangeLength *int   `json:"rangeLength,omitempty"` // The optional length of the range that got replaced
	Text        string `json:"text"`                  // The new text for the provided range or entire document
}

// CompletionParams represents the parameters for textDocument/completion request
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
}

// CompletionContext provides additional information about the completion request
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter *string               `json:"triggerCharacter,omitempty"`
}

// CompletionTriggerKind defines how a completion was triggered
type CompletionTriggerKind int

const (
	CompletionTriggerKindInvoked                         CompletionTriggerKind = 1 // Completion was triggered by typing an identifier
	CompletionTriggerKindTriggerCharacter                CompletionTriggerKind = 2 // Completion was triggered by a trigger character
	CompletionTriggerKindTriggerForIncompleteCompletions CompletionTriggerKind = 3 // Completion was re-triggered
)

// CompletionList represents a collection of completion items
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// HoverParams represents the parameters for textDocument/hover request
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DefinitionParams represents the parameters for textDocument/definition request
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DocumentFormattingParams represents the parameters for textDocument/formatting request
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions defines formatting options
type FormattingOptions struct {
	TabSize                int                    `json:"tabSize"`
	InsertSpaces           bool                   `json:"insertSpaces"`
	TrimTrailingWhitespace *bool                  `json:"trimTrailingWhitespace,omitempty"`
	InsertFinalNewline     *bool                  `json:"insertFinalNewline,omitempty"`
	TrimFinalNewlines      *bool                  `json:"trimFinalNewlines,omitempty"`
	AdditionalProperties   map[string]interface{} `json:"-"`
}

// ReferenceParams represents the parameters for textDocument/references request
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext provides additional context for reference requests
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// DocumentSymbolParams represents the parameters for textDocument/documentSymbol request
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}
