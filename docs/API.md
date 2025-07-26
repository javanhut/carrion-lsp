# Carrion LSP API Documentation

This document describes the Language Server Protocol implementation for Carrion, including supported methods, capabilities, and data structures.

## Server Capabilities

The Carrion LSP server advertises the following capabilities during initialization:

```json
{
  "capabilities": {
    "textDocumentSync": {
      "openClose": true,
      "change": 1
    },
    "completionProvider": {
      "triggerCharacters": [".", "(", "["]
    },
    "hoverProvider": true,
    "definitionProvider": true,
    "referencesProvider": true,
    "documentFormattingProvider": true,
    "documentSymbolProvider": true,
    "diagnosticProvider": {
      "identifier": "carrion-lsp",
      "interFileDependencies": false,
      "workspaceDiagnostics": false
    }
  }
}
```

## Supported LSP Methods

### Lifecycle Methods

#### `initialize`
**Request**: Initialize the language server with client capabilities and configuration.

**Parameters**:
```json
{
  "processId": 12345,
  "rootUri": "file:///path/to/workspace",
  "capabilities": { /* client capabilities */ },
  "initializationOptions": {
    "carrionPath": "/path/to/carrion/interpreter"
  }
}
```

**Response**: Server capabilities and information.

#### `initialized`
**Notification**: Sent after initialization is complete.

#### `shutdown`
**Request**: Prepare the server for shutdown.

#### `exit`
**Notification**: Terminate the server process.

### Document Synchronization

#### `textDocument/didOpen`
**Notification**: A document was opened.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.crl",
    "languageId": "carrion",
    "version": 1,
    "text": "x = 42\nspell greet():\n    print(\"Hello\")"
  }
}
```

**Behavior**: 
- Parses and analyzes the document
- Sends diagnostics if errors are found
- Builds symbol table for the document

#### `textDocument/didChange`
**Notification**: Document content changed.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.crl",
    "version": 2
  },
  "contentChanges": [
    {
      "text": "x = 100\nspell greet():\n    print(\"Hello World\")"
    }
  ]
}
```

**Behavior**:
- Re-analyzes the document
- Updates symbol table
- Sends updated diagnostics

#### `textDocument/didClose`
**Notification**: Document was closed.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  }
}
```

**Behavior**:
- Removes document from tracking
- Clears diagnostics for the document

### Language Features

#### `textDocument/completion`
**Request**: Get code completion suggestions.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  },
  "position": {
    "line": 2,
    "character": 4
  }
}
```

**Response**:
```json
{
  "isIncomplete": false,
  "items": [
    {
      "label": "greet",
      "kind": 3,
      "detail": "(name) -> unknown",
      "documentation": "Function: greet"
    },
    {
      "label": "print",
      "kind": 3,
      "detail": "function",
      "documentation": "Built-in function"
    }
  ]
}
```

**Completion Item Kinds**:
- `1`: Text
- `3`: Function
- `6`: Variable
- `7`: Class
- `9`: Module

#### `textDocument/hover`
**Request**: Get hover information for a symbol.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  },
  "position": {
    "line": 0,
    "character": 2
  }
}
```

**Response**:
```json
{
  "contents": {
    "kind": "markdown",
    "value": "**Variable**: `x`\n\n**Type**: `int`\n\n**Declared at**: line 1"
  }
}
```

#### `textDocument/definition`
**Request**: Go to symbol definition.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  },
  "position": {
    "line": 5,
    "character": 10
  }
}
```

**Response**:
```json
[
  {
    "uri": "file:///path/to/file.crl",
    "range": {
      "start": {
        "line": 2,
        "character": 6
      },
      "end": {
        "line": 2,
        "character": 11
      }
    }
  }
]
```

#### `textDocument/references`
**Request**: Find all references to a symbol.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  },
  "position": {
    "line": 2,
    "character": 6
  },
  "context": {
    "includeDeclaration": true
  }
}
```

**Response**:
```json
[
  {
    "uri": "file:///path/to/file.crl",
    "range": {
      "start": {
        "line": 2,
        "character": 6
      },
      "end": {
        "line": 2,
        "character": 11
      }
    }
  },
  {
    "uri": "file:///path/to/file.crl",
    "range": {
      "start": {
        "line": 5,
        "character": 10
      },
      "end": {
        "line": 5,
        "character": 15
      }
    }
  }
]
```

#### `textDocument/documentSymbol`
**Request**: Get document symbols for outline view.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  }
}
```

**Response**:
```json
[
  {
    "name": "x",
    "detail": "int",
    "kind": 13,
    "range": {
      "start": { "line": 0, "character": 0 },
      "end": { "line": 0, "character": 1 }
    },
    "selectionRange": {
      "start": { "line": 0, "character": 0 },
      "end": { "line": 0, "character": 1 }
    }
  },
  {
    "name": "greet",
    "detail": "(name)",
    "kind": 12,
    "range": {
      "start": { "line": 2, "character": 6 },
      "end": { "line": 2, "character": 11 }
    },
    "selectionRange": {
      "start": { "line": 2, "character": 6 },
      "end": { "line": 2, "character": 11 }
    }
  },
  {
    "name": "Person",
    "detail": "class",
    "kind": 5,
    "range": {
      "start": { "line": 5, "character": 5 },
      "end": { "line": 5, "character": 11 }
    },
    "selectionRange": {
      "start": { "line": 5, "character": 5 },
      "end": { "line": 5, "character": 11 }
    },
    "children": [
      {
        "name": "init",
        "detail": "(self, name)",
        "kind": 6,
        "range": {
          "start": { "line": 6, "character": 10 },
          "end": { "line": 6, "character": 14 }
        },
        "selectionRange": {
          "start": { "line": 6, "character": 10 },
          "end": { "line": 6, "character": 14 }
        }
      }
    ]
  }
]
```

**Symbol Kinds**:
- `5`: Class
- `6`: Method
- `12`: Function
- `13`: Variable
- `2`: Module

#### `textDocument/formatting`
**Request**: Format document.

**Parameters**:
```json
{
  "textDocument": {
    "uri": "file:///path/to/file.carrion"
  },
  "options": {
    "tabSize": 4,
    "insertSpaces": true
  }
}
```

**Response**:
```json
[
  {
    "range": {
      "start": { "line": 1, "character": 0 },
      "end": { "line": 1, "character": 15 }
    },
    "newText": "    print(\"Hello\")"
  }
]
```

### Diagnostics

The server automatically sends diagnostic notifications when documents are opened or changed:

**Notification**: `textDocument/publishDiagnostics`

**Parameters**:
```json
{
  "uri": "file:///path/to/file.crl",
  "diagnostics": [
    {
      "range": {
        "start": { "line": 3, "character": 4 },
        "end": { "line": 3, "character": 13 }
      },
      "severity": 1,
      "source": "carrion-analyzer",
      "message": "undefined variable 'undefined_var'"
    }
  ]
}
```

**Diagnostic Severities**:
- `1`: Error
- `2`: Warning
- `3`: Information
- `4`: Hint

## Data Structures

### Position
```json
{
  "line": 2,        // 0-based line number
  "character": 10   // 0-based character offset
}
```

### Range
```json
{
  "start": { "line": 2, "character": 6 },
  "end": { "line": 2, "character": 11 }
}
```

### Location
```json
{
  "uri": "file:///path/to/file.crl",
  "range": { /* Range object */ }
}
```

### Diagnostic
```json
{
  "range": { /* Range object */ },
  "severity": 1,          // 1=Error, 2=Warning, 3=Info, 4=Hint
  "source": "carrion-lsp",
  "message": "Error description"
}
```

## Error Handling

The server returns standard JSON-RPC errors:

```json
{
  "code": -32602,
  "message": "Invalid params",
  "data": "Additional error information"
}
```

**Common Error Codes**:
- `-32700`: Parse error
- `-32600`: Invalid request  
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

## Configuration

The server accepts these initialization options:

### `carrionPath`
**Type**: `string`  
**Default**: `""`  
**Description**: Path to the Carrion interpreter executable. Used for advanced analysis features.

**Example**:
```json
{
  "initializationOptions": {
    "carrionPath": "/usr/local/bin/carrion"
  }
}
```

## Limitations

Current implementation limitations:

1. **Single File Analysis**: Cross-file references not yet supported
2. **Limited Type Inference**: Basic type inference only
3. **No Import Resolution**: Module imports not resolved
4. **References**: Current implementation returns empty results (framework in place)

## Performance Considerations

- **Memory**: Symbol tables are kept in memory for open documents
- **CPU**: Analysis is performed synchronously on document changes
- **Disk**: No persistent storage of analysis results
- **Network**: All communication over stdin/stdout using JSON-RPC

## Version Compatibility

- **LSP Version**: 3.17
- **Go Version**: 1.19+
- **Carrion Version**: Compatible with current Carrion language specification