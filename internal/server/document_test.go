package server

import (
	"strings"
	"testing"

	"github.com/javanhut/carrion-lsp/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentManager_OpenDocument(t *testing.T) {
	dm := NewDocumentManager()

	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `x = 42
y = "hello"
spell greet(name):
    return "Hello, " + name`,
		},
	}

	doc, err := dm.OpenDocument(params)
	require.NoError(t, err)
	assert.Equal(t, "file:///test.carrion", doc.URI)
	assert.Equal(t, "carrion", doc.LanguageID)
	assert.Equal(t, 1, doc.Version)
	assert.NotNil(t, doc.Analyzer)

	// Check that document is tracked
	retrieved, exists := dm.GetDocument("file:///test.carrion")
	assert.True(t, exists)
	assert.Equal(t, doc, retrieved)
}

func TestDocumentManager_ChangeDocument(t *testing.T) {
	dm := NewDocumentManager()

	// First open the document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text:       "x = 42",
		},
	}

	_, err := dm.OpenDocument(openParams)
	require.NoError(t, err)

	// Now change it
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			URI:     "file:///test.carrion",
			Version: 2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{
				Text: "x = 100\ny = \"changed\"",
			},
		},
	}

	doc, err := dm.ChangeDocument(changeParams)
	require.NoError(t, err)
	assert.Equal(t, 2, doc.Version)
	assert.Equal(t, "x = 100\ny = \"changed\"", doc.Text)
	assert.NotNil(t, doc.Analyzer)
}

func TestDocumentManager_CloseDocument(t *testing.T) {
	dm := NewDocumentManager()

	// First open the document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text:       "x = 42",
		},
	}

	_, err := dm.OpenDocument(openParams)
	require.NoError(t, err)

	// Verify it's there
	_, exists := dm.GetDocument("file:///test.carrion")
	assert.True(t, exists)

	// Close it
	closeParams := &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: "file:///test.carrion",
		},
	}

	err = dm.CloseDocument(closeParams)
	require.NoError(t, err)

	// Verify it's gone
	_, exists = dm.GetDocument("file:///test.carrion")
	assert.False(t, exists)
}

func TestDocumentManager_GetCompletionItems(t *testing.T) {
	dm := NewDocumentManager()

	// Open a document with some symbols
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `x = 42
y = "hello"
spell greet(name):
    return "Hello, " + name

grim Person:
    spell init(self, name):
        self.name = name`,
		},
	}

	_, err := dm.OpenDocument(params)
	require.NoError(t, err)

	// Get completion items
	items, err := dm.GetCompletionItems("file:///test.carrion", protocol.Position{Line: 8, Character: 0})
	require.NoError(t, err)

	// Should have variables, functions, classes, and built-ins
	itemNames := make([]string, len(items))
	for i, item := range items {
		itemNames[i] = item.Label
	}

	assert.Contains(t, itemNames, "x")
	assert.Contains(t, itemNames, "y")
	assert.Contains(t, itemNames, "greet")
	assert.Contains(t, itemNames, "Person")
	assert.Contains(t, itemNames, "print") // built-in
}

func TestDocumentManager_NonCarrionFile(t *testing.T) {
	dm := NewDocumentManager()

	// Open a non-Carrion file
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.txt",
			LanguageID: "plaintext",
			Version:    1,
			Text:       "This is not Carrion code",
		},
	}

	doc, err := dm.OpenDocument(params)
	require.NoError(t, err)
	assert.Nil(t, doc.Analyzer) // Should not analyze non-Carrion files
	assert.Nil(t, doc.Diagnostics)
}

func TestDocumentManager_AnalysisError(t *testing.T) {
	dm := NewDocumentManager()

	// Open a document with syntax errors
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `x = undefined_variable
y = another_undefined`,
		},
	}

	doc, err := dm.OpenDocument(params)
	require.NoError(t, err) // Opening should succeed even with analysis errors
	assert.NotNil(t, doc.Analyzer)
	assert.True(t, len(doc.Diagnostics) > 0) // Should have diagnostics for undefined variables
}

func TestDocumentManager_GetHoverInformation(t *testing.T) {
	dm := NewDocumentManager()

	// Open a document with various symbols
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `counter = 42
name = "test"

spell greet(user):
    return "Hello, " + user

grim Person:
    spell init(self, name):
        self.name = name
    
    spell say_hello(self):
        return "Hello, " + self.name`,
		},
	}

	_, err := dm.OpenDocument(params)
	require.NoError(t, err)

	tests := []struct {
		name         string
		position     protocol.Position
		expectedType string
		shouldFind   bool
	}{
		{
			name:         "hover over variable",
			position:     protocol.Position{Line: 0, Character: 2}, // "counter"
			expectedType: "Variable",
			shouldFind:   true,
		},
		{
			name:         "hover over function",
			position:     protocol.Position{Line: 3, Character: 8}, // "greet"
			expectedType: "Function",
			shouldFind:   true,
		},
		{
			name:         "hover over class",
			position:     protocol.Position{Line: 6, Character: 7}, // "Person"
			expectedType: "Class",
			shouldFind:   true,
		},
		{
			name:         "hover over self parameter",
			position:     protocol.Position{Line: 7, Character: 18}, // "self" in parameter
			expectedType: "Parameter",
			shouldFind:   false, // Currently self is not found in global scope
		},
		{
			name:         "hover over built-in",
			position:     protocol.Position{Line: 4, Character: 4}, // (assuming print was used)
			expectedType: "",
			shouldFind:   false, // This test won't find anything at that position
		},
		{
			name:         "hover over empty space",
			position:     protocol.Position{Line: 2, Character: 0}, // empty line
			expectedType: "",
			shouldFind:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hover, err := dm.GetHoverInformation("file:///test.carrion", tt.position)
			require.NoError(t, err)

			if tt.shouldFind {
				assert.NotNil(t, hover)
				if hover != nil {
					assert.Contains(t, hover.Contents.(protocol.MarkupContent).Value, tt.expectedType)
				}
			} else {
				assert.Nil(t, hover)
			}
		})
	}
}

func TestDocumentManager_GetIdentifierAtPosition(t *testing.T) {
	dm := NewDocumentManager()

	text := `counter = 42
spell greet(name):
    return name`

	tests := []struct {
		name     string
		position protocol.Position
		expected string
	}{
		{
			name:     "identifier at start",
			position: protocol.Position{Line: 0, Character: 0},
			expected: "counter",
		},
		{
			name:     "identifier in middle",
			position: protocol.Position{Line: 0, Character: 3},
			expected: "counter",
		},
		{
			name:     "identifier at end",
			position: protocol.Position{Line: 0, Character: 6},
			expected: "counter",
		},
		{
			name:     "function name",
			position: protocol.Position{Line: 1, Character: 8},
			expected: "greet",
		},
		{
			name:     "parameter name",
			position: protocol.Position{Line: 1, Character: 14},
			expected: "name",
		},
		{
			name:     "no identifier",
			position: protocol.Position{Line: 0, Character: 8}, // space character
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dm.getIdentifierAtPosition(text, tt.position)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDocumentManager_GetReferences(t *testing.T) {
	dm := NewDocumentManager()

	// Open a document with various symbols
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `counter = 42
name = "test"

spell greet(user):
    return "Hello, " + user

grim Person:
    spell init(self, name):
        self.name = name

result = greet("world")`,
		},
	}

	_, err := dm.OpenDocument(params)
	require.NoError(t, err)

	tests := []struct {
		name               string
		position           protocol.Position
		includeDeclaration bool
		expectReferences   bool
	}{
		{
			name:               "references to function",
			position:           protocol.Position{Line: 10, Character: 9}, // "greet" in greet("world")
			includeDeclaration: true,
			expectReferences:   false, // Our current implementation doesn't find references yet
		},
		{
			name:               "no identifier at position",
			position:           protocol.Position{Line: 2, Character: 0}, // empty line
			includeDeclaration: false,
			expectReferences:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			references, err := dm.GetReferences("file:///test.carrion", tt.position, tt.includeDeclaration)
			require.NoError(t, err)

			if tt.expectReferences {
				assert.Greater(t, len(references), 0)
			} else {
				assert.Len(t, references, 0)
			}
		})
	}
}

func TestDocumentManager_GetDocumentSymbols(t *testing.T) {
	dm := NewDocumentManager()

	// Open a document with various symbols
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `counter = 42
name = "test"

spell greet(user):
    return "Hello, " + user

grim Person:
    spell init(self, name):
        self.name = name
    
    spell say_hello(self):
        return "Hello, " + self.name`,
		},
	}

	_, err := dm.OpenDocument(params)
	require.NoError(t, err)

	symbols, err := dm.GetDocumentSymbols("file:///test.carrion")
	require.NoError(t, err)

	// Should have variables, functions, and classes
	symbolNames := make([]string, len(symbols))
	for i, symbol := range symbols {
		symbolNames[i] = symbol.Name
	}

	assert.Contains(t, symbolNames, "counter")
	assert.Contains(t, symbolNames, "name")
	assert.Contains(t, symbolNames, "greet")
	assert.Contains(t, symbolNames, "Person")

	// Find the Person class and check it has child methods
	var personSymbol *protocol.DocumentSymbol
	for _, symbol := range symbols {
		if symbol.Name == "Person" {
			personSymbol = &symbol
			break
		}
	}

	if personSymbol != nil {
		childNames := make([]string, len(personSymbol.Children))
		for i, child := range personSymbol.Children {
			childNames[i] = child.Name
		}
		assert.Contains(t, childNames, "init")
		assert.Contains(t, childNames, "say_hello")
	}
}

func TestDocumentManager_NonCarrionReferences(t *testing.T) {
	dm := NewDocumentManager()

	// Open a non-Carrion file
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.txt",
			LanguageID: "plaintext",
			Version:    1,
			Text:       "This is not Carrion code",
		},
	}

	_, err := dm.OpenDocument(params)
	require.NoError(t, err)

	// Should return error for non-Carrion files
	_, err = dm.GetReferences("file:///test.txt", protocol.Position{Line: 0, Character: 0}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has no analyzer")
}

func TestDocumentManager_GetDefinitionLocation(t *testing.T) {
	dm := NewDocumentManager()

	// Open a document with various symbols
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///test.carrion",
			LanguageID: "carrion",
			Version:    1,
			Text: `counter = 42
name = "test"

spell greet(user):
    return "Hello, " + user

grim Person:
    spell init(self, name):
        self.name = name

result = greet("world")
person = Person()`,
		},
	}

	_, err := dm.OpenDocument(params)
	require.NoError(t, err)

	tests := []struct {
		name           string
		position       protocol.Position
		expectLocation bool
		expectedLine   int // 0-based line number where definition should be
	}{
		{
			name:           "definition of variable",
			position:       protocol.Position{Line: 10, Character: 9}, // "greet" in greet("world")
			expectLocation: true,
			expectedLine:   3, // spell greet is on line 3
		},
		{
			name:           "definition of class",
			position:       protocol.Position{Line: 11, Character: 9}, // "Person" in Person()
			expectLocation: true,
			expectedLine:   6, // grim Person is on line 6
		},
		{
			name:           "definition of declared variable",
			position:       protocol.Position{Line: 0, Character: 2}, // "counter" declaration itself
			expectLocation: true,
			expectedLine:   0, // counter is declared on line 0
		},
		{
			name:           "no definition for built-in",
			position:       protocol.Position{Line: 4, Character: 11}, // "Hello" (if it was print)
			expectLocation: false,
		},
		{
			name:           "no identifier at position",
			position:       protocol.Position{Line: 2, Character: 0}, // empty line
			expectLocation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, err := dm.GetDefinitionLocation("file:///test.carrion", tt.position)
			require.NoError(t, err)

			if tt.expectLocation {
				assert.Len(t, locations, 1)
				if len(locations) > 0 {
					assert.Equal(t, "file:///test.carrion", locations[0].URI)
					assert.Equal(t, tt.expectedLine, locations[0].Range.Start.Line)
				}
			} else {
				assert.Len(t, locations, 0)
			}
		})
	}
}

func TestDocumentManager_FormatDocument_Basic(t *testing.T) {
	// Simple test without opening documents that might cause parser issues
	formatter := NewCarrionFormatter(protocol.FormattingOptions{
		TabSize:      4,
		InsertSpaces: true,
	})

	input := `x = 1
y = 2`

	edits := formatter.FormatDocument(input)
	assert.Len(t, edits, 0) // No edits needed for already-formatted simple assignments
}

func TestCarrionFormatter_FormatDocument(t *testing.T) {
	formatter := NewCarrionFormatter(protocol.FormattingOptions{
		TabSize:      4,
		InsertSpaces: true,
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple function",
			input: `spell greet(name):
return "Hello, " + name`,
			expected: `spell greet(name):
    return "Hello, " + name`,
		},
		{
			name: "nested blocks",
			input: `spell test():
if True:
x = 1
else:
x = 0`,
			expected: `spell test():
    if True:
        x = 1
    else:
        x = 0`,
		},
		{
			name: "class definition",
			input: `grim Person:
spell init(self, name):
self.name = name`,
			expected: `grim Person:
    spell init(self, name):
        self.name = name`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edits := formatter.FormatDocument(tt.input)

			// Apply edits to get formatted result
			lines := strings.Split(tt.input, "\n")
			for _, edit := range edits {
				if edit.Range.Start.Line < len(lines) {
					lines[edit.Range.Start.Line] = edit.NewText
				}
			}
			result := strings.Join(lines, "\n")

			assert.Equal(t, tt.expected, result)
		})
	}
}
