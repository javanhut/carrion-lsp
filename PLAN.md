# Carrion Language Server Protocol (LSP) Implementation Plan

## Project Overview
Building a Language Server Protocol implementation for the Carrion Programming Language to provide IDE features like IntelliSense, diagnostics, code formatting, and navigation.

**Repository**: `/home/javanhut/LSP-Carrion/carrion-lsp/`  
**Language**: Go  
**Reference**: `/home/javanhut/LSP-Carrion/TheCarrionLanguage/` (read-only)

## Overall Architecture

### Directory Structure
```
carrion-lsp/
├── cmd/
│   └── carrion-lsp/
│       └── main.go          # LSP server entry point
├── internal/
│   ├── server/
│   │   ├── server.go        # Main LSP server implementation
│   │   ├── handler.go       # Request handler routing
│   │   └── state.go         # Server state management
│   ├── protocol/
│   │   ├── types.go         # LSP protocol type definitions
│   │   ├── methods.go       # LSP method constants
│   │   └── jsonrpc.go       # JSON-RPC 2.0 implementation
│   ├── analyzer/
│   │   ├── lexer.go         # Carrion lexer (adapted)
│   │   ├── parser.go        # Carrion parser (adapted)
│   │   ├── ast.go           # AST definitions
│   │   └── semantic.go      # Semantic analysis
│   ├── features/
│   │   ├── completion.go    # Auto-completion
│   │   ├── hover.go         # Hover information
│   │   ├── definition.go    # Go to definition
│   │   ├── references.go    # Find references
│   │   ├── diagnostics.go   # Error/warning detection
│   │   ├── formatting.go    # Code formatting
│   │   ├── symbols.go       # Document/workspace symbols
│   │   └── signature.go     # Signature help
│   ├── workspace/
│   │   ├── document.go      # Document management
│   │   ├── cache.go         # AST/analysis caching
│   │   └── index.go         # Symbol indexing
│   └── config/
│       └── config.go        # Server configuration
├── pkg/
│   └── carrion/
│       ├── stdlib.go        # Munin standard library definitions
│       └── builtins.go      # Built-in function signatures
├── tests/
├── go.mod
├── go.sum
├── README.md
└── PLAN.md                  # This file
```

## Implementation Phases

### Phase 1: Foundation (Essential)  APPROVED
1. **Basic LSP Server Infrastructure**
   - JSON-RPC 2.0 communication
   - LSP protocol lifecycle
   - Document synchronization
   - Basic syntax diagnostics

2. **Syntax Highlighting & Diagnostics**
   - Real-time syntax error detection
   - Lexical and parse error reporting
   - Enhanced error messages

3. **Code Completion (IntelliSense)**
   - Keywords completion
   - Variable/function completion
   - Grimoire methods and properties
   - Built-in functions and stdlib

### Phase 2: Navigation (High Priority)
4. **Go to Definition**
   - Jump to spell definitions
   - Jump to grimoire definitions
   - Jump to variable declarations

5. **Find References**
   - Find variable uses
   - Find spell calls
   - Find grimoire instantiations

6. **Document Symbols**
   - Outline view
   - List spells and grimoires
   - Hierarchical navigation

### Phase 3: Developer Experience
7. **Hover Information**
   - Type information
   - Function signatures
   - Documentation

8. **Signature Help**
   - Parameter hints
   - Overloaded methods

9. **Code Formatting**
   - Auto-indent
   - Format document
   - Format selection

### Phase 4: Advanced Features
10. **Refactoring**
    - Rename symbols
    - Extract spell
    - Extract variable

11. **Code Actions**
    - Quick fixes
    - Import suggestions
    - Add missing methods

12. **Workspace Features**
    - Multi-file search
    - Project refactoring
    - Import resolution

## Feature 1: Basic LSP Server Infrastructure (APPROVED)

### Status: Ready for Implementation

### Objectives
- Accept LSP client connections
- Handle JSON-RPC 2.0 over stdin/stdout
- Implement LSP lifecycle
- Synchronize documents
- Provide basic diagnostics

### Technical Details
1. **JSON-RPC 2.0 Layer**
   - Request/Response types
   - Message correlation
   - Error handling

2. **LSP Capabilities**
   - TextDocumentSync (full)
   - DiagnosticProvider
   - Future capability expansion

3. **Document Management**
   - In-memory storage
   - Version tracking
   - AST caching

4. **Diagnostics**
   - Carrion lexer/parser integration
   - Syntax error reporting
   - Enhanced error formatting

### Implementation Steps
1. Set up project structure
2. Implement JSON-RPC transport
3. Create LSP server skeleton
4. Integrate Carrion lexer/parser
5. Implement document synchronization
6. Add diagnostic publishing

### Testing Strategy
- Unit tests for JSON-RPC
- Integration tests with mock client
- Manual VS Code testing
- Comprehensive syntax coverage

### Success Criteria
- Server starts successfully
- VS Code recognition
- Document parsing on open
- Error squiggles display
- Enhanced error messages

### Timeline
- Setup and JSON-RPC: 2 days
- LSP server skeleton: 2 days
- Carrion integration: 3 days
- Document sync & diagnostics: 3 days
- Testing and refinement: 2 days
- **Total: ~2 weeks**

## Carrion Language Reference

### Keywords
- Control: `if`, `otherwise`, `else`, `for`, `while`, `skip`, `stop`, `match`, `case`, `return`
- OOP: `grim`, `spell`, `init`, `self`, `super`, `arcane`, `arcanespell`
- Errors: `attempt`, `ensnare`, `resolve`, `raise`, `check`
- Logic: `and`, `or`, `not`, `True`, `False`, `None`
- Module: `import`, `as`
- Other: `var`, `ignore`, `main`, `global`, `autoclose`

### Built-in Functions
- Type conversion: `int()`, `float()`, `str()`, `bool()`, `list()`, `tuple()`
- Utilities: `len()`, `type()`, `print()`, `input()`, `range()`, `max()`, `abs()`
- Collections: `enumerate()`, `pairs()`
- File: `open()`
- JSON: `parseHash()`
- Error: `Error()`

### Standard Library (Munin)
- Primitive Grimoires: Integer, Float, String, Boolean, Array
- System Grimoires: File, OS
- Modules: http, time, sockets

## Development Guidelines
1. Maintain separation from Carrion repo
2. Follow Go best practices
3. Document all public APIs
4. Write comprehensive tests
5. Use Carrion's error message style
6. Support cross-platform operation

## Progress Tracking

### Completed
- [x] Analyze Carrion language features
- [x] Create overall architecture plan
- [x] Define feature roadmap
- [x] Create Feature 1 detailed plan
- [x] Get approval for Feature 1
- [x] **Feature 1: Basic LSP Server Infrastructure COMPLETED**
  - [x] JSON-RPC 2.0 transport layer
  - [x] LSP protocol lifecycle (initialize/initialized/shutdown/exit)
  - [x] Server state management
  - [x] Command-line interface
  - [x] Basic security measures
  - [x] Comprehensive test suite

### In Progress
- [ ] Feature 2: Document synchronization and Carrion language integration

### Pending
- [ ] Feature 3: Real-time diagnostics (syntax errors)
- [ ] Feature 4: Code completion (IntelliSense)
- [ ] Phase 2-4 Features

## Notes
- The Carrion repository is treated as read-only reference
- All LSP code lives in the separate carrion-lsp directory
- We'll adapt Carrion's lexer/parser rather than importing directly
- Enhanced error messages from Carrion will be preserved