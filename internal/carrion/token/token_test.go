package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{ILLEGAL, "ILLEGAL"},
		{EOF, "EOF"},
		{IDENT, "IDENT"},
		{INT, "INT"},
		{FLOAT, "FLOAT"},
		{STRING, "STRING"},
		{SPELL, "SPELL"},
		{GRIM, "GRIM"},
		{IF, "IF"},
		{ELSE, "ELSE"},
		{RETURN, "RETURN"},
		{ASSIGN, "ASSIGN"},
		{PLUS, "PLUS"},
		{MINUS, "MINUS"},
		{ASTERISK, "ASTERISK"},
		{SLASH, "SLASH"},
		{EQ, "EQ"},
		{NOT_EQ, "NOT_EQ"},
		{LT, "LT"},
		{GT, "GT"},
		{LPAREN, "LPAREN"},
		{RPAREN, "RPAREN"},
		{LBRACE, "LBRACE"},
		{RBRACE, "RBRACE"},
		{LBRACKET, "LBRACKET"},
		{RBRACKET, "RBRACKET"},
		{COMMA, "COMMA"},
		{SEMICOLON, "SEMICOLON"},
		{COLON, "COLON"},
		{DOT, "DOT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tokenType.String())
		})
	}
}

func TestToken_Position(t *testing.T) {
	token := Token{
		Type:     IDENT,
		Literal:  "test",
		Filename: "test.crl",
		Line:     10,
		Column:   5,
	}

	assert.Equal(t, IDENT, token.Type)
	assert.Equal(t, "test", token.Literal)
	assert.Equal(t, "test.crl", token.Filename)
	assert.Equal(t, 10, token.Line)
	assert.Equal(t, 5, token.Column)
}

func TestToken_LSPPosition(t *testing.T) {
	// Test conversion from 1-based Carrion positions to 0-based LSP positions
	token := Token{
		Type:    IDENT,
		Literal: "test",
		Line:    10, // 1-based
		Column:  5,  // 1-based
	}

	lspLine, lspChar := token.LSPPosition()
	assert.Equal(t, 9, lspLine) // 0-based
	assert.Equal(t, 4, lspChar) // 0-based
}

func TestToken_Range(t *testing.T) {
	token := Token{
		Type:    IDENT,
		Literal: "myFunction",
		Line:    5,
		Column:  10,
	}

	startLine, startChar, endLine, endChar := token.Range()
	assert.Equal(t, 4, startLine) // 0-based
	assert.Equal(t, 9, startChar) // 0-based
	assert.Equal(t, 4, endLine)   // Same line
	assert.Equal(t, 19, endChar)  // Start + length
}

func TestToken_IsKeyword(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{SPELL, true},
		{GRIM, true},
		{IF, true},
		{ELSE, true},
		{RETURN, true},
		{TRUE, true},
		{FALSE, true},
		{IDENT, false},
		{INT, false},
		{STRING, false},
		{PLUS, false},
		{LPAREN, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			token := Token{Type: tt.tokenType}
			assert.Equal(t, tt.expected, token.IsKeyword())
		})
	}
}

func TestToken_IsOperator(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{PLUS, true},
		{MINUS, true},
		{ASTERISK, true},
		{SLASH, true},
		{EQ, true},
		{NOT_EQ, true},
		{LT, true},
		{GT, true},
		{AND, true},
		{OR, true},
		{ASSIGN, true},
		{PLUS_ASSIGN, true},
		{INCREMENT, true},
		{SPELL, false},
		{IDENT, false},
		{INT, false},
		{LPAREN, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			token := Token{Type: tt.tokenType}
			assert.Equal(t, tt.expected, token.IsOperator())
		})
	}
}

func TestToken_IsLiteral(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{INT, true},
		{FLOAT, true},
		{STRING, true},
		{TRUE, true},
		{FALSE, true},
		{NONE, true},
		{IDENT, false}, // Identifiers are not literals
		{SPELL, false},
		{PLUS, false},
		{LPAREN, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			token := Token{Type: tt.tokenType}
			assert.Equal(t, tt.expected, token.IsLiteral())
		})
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		ident    string
		expected TokenType
	}{
		// Carrion keywords
		{"spell", SPELL},
		{"grim", GRIM},
		{"arcane", ARCANE},
		{"init", INIT},
		{"self", SELF},
		{"super", SUPER},
		{"if", IF},
		{"otherwise", OTHERWISE},
		{"else", ELSE},
		{"for", FOR},
		{"while", WHILE},
		{"in", IN},
		{"skip", SKIP},
		{"stop", STOP},
		{"return", RETURN},
		{"attempt", ATTEMPT},
		{"ensnare", ENSNARE},
		{"resolve", RESOLVE},
		{"raise", RAISE},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"True", TRUE},
		{"False", FALSE},
		{"None", NONE},
		{"import", IMPORT},
		{"as", AS},
		{"global", GLOBAL},
		{"match", MATCH},
		{"case", CASE},

		// Non-keywords should return IDENT
		{"myVariable", IDENT},
		{"myFunction", IDENT},
		{"MyClass", IDENT},
		{"test123", IDENT},
		{"_private", IDENT},
		{"", IDENT}, // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.ident, func(t *testing.T) {
			result := LookupIdent(tt.ident)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewToken(t *testing.T) {
	token := NewToken(SPELL, "spell", "test.crl", 5, 10)

	assert.Equal(t, SPELL, token.Type)
	assert.Equal(t, "spell", token.Literal)
	assert.Equal(t, "test.crl", token.Filename)
	assert.Equal(t, 5, token.Line)
	assert.Equal(t, 10, token.Column)
}

func TestToken_String(t *testing.T) {
	token := Token{
		Type:    SPELL,
		Literal: "spell",
		Line:    5,
		Column:  10,
	}

	result := token.String()
	expected := "Token{Type: SPELL, Literal: 'spell', Position: 5:10}"
	assert.Equal(t, expected, result)
}

func TestToken_IsError(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{ILLEGAL, true},
		{SPELL, false},
		{IDENT, false},
		{EOF, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			token := Token{Type: tt.tokenType}
			assert.Equal(t, tt.expected, token.IsError())
		})
	}
}

func TestToken_IsEOF(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{EOF, true},
		{SPELL, false},
		{IDENT, false},
		{ILLEGAL, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			token := Token{Type: tt.tokenType}
			assert.Equal(t, tt.expected, token.IsEOF())
		})
	}
}
