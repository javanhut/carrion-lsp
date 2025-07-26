package lexer

import (
	"testing"

	"github.com/javanhut/carrion-lsp/internal/carrion/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_BasicTokens(t *testing.T) {
	input := `spell myFunction():
    return 42`

	expected := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
		expectedColumn  int
	}{
		{token.SPELL, "spell", 1, 1},
		{token.IDENT, "myFunction", 1, 7},
		{token.LPAREN, "(", 1, 17},
		{token.RPAREN, ")", 1, 18},
		{token.COLON, ":", 1, 19},
		{token.NEWLINE, "\n", 1, 20},
		{token.INDENT, "", 2, 1},
		{token.RETURN, "return", 2, 5},
		{token.INT, "42", 2, 12},
		{token.NEWLINE, "", 2, 14},
		{token.DEDENT, "", 2, 14},
		{token.EOF, "", 2, 14},
	}

	lexer := New(input)

	for i, tt := range expected {
		tok := lexer.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - wrong token type", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] - wrong literal", i)
		assert.Equal(t, tt.expectedLine, tok.Line, "test[%d] - wrong line", i)
		assert.Equal(t, tt.expectedColumn, tok.Column, "test[%d] - wrong column", i)
	}
}

func TestLexer_Operators(t *testing.T) {
	input := `+ - * / % ** // += -= *= /= ++ -- == != < > <= >= and or not`

	expected := []token.TokenType{
		token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.MODULO,
		token.POWER, token.FLOOR_DIV, token.PLUS_ASSIGN, token.MINUS_ASSIGN,
		token.MULTIPLY_ASSIGN, token.DIVIDE_ASSIGN, token.INCREMENT, token.DECREMENT,
		token.EQ, token.NOT_EQ, token.LT, token.GT, token.LTE, token.GTE,
		token.AND, token.OR, token.NOT,
	}

	lexer := New(input)

	for i, expectedType := range expected {
		tok := lexer.NextToken()
		assert.Equal(t, expectedType, tok.Type, "test[%d] - wrong token type", i)
	}
}

func TestLexer_Keywords(t *testing.T) {
	input := `spell grim arcane init self super if otherwise else for while in skip stop return attempt ensnare resolve raise True False None import as global match case`

	expected := []token.TokenType{
		token.SPELL, token.GRIM, token.ARCANE, token.INIT, token.SELF, token.SUPER,
		token.IF, token.OTHERWISE, token.ELSE, token.FOR, token.WHILE, token.IN,
		token.SKIP, token.STOP, token.RETURN, token.ATTEMPT, token.ENSNARE,
		token.RESOLVE, token.RAISE, token.TRUE, token.FALSE, token.NONE,
		token.IMPORT, token.AS, token.GLOBAL, token.MATCH, token.CASE,
	}

	lexer := New(input)

	for i, expectedType := range expected {
		tok := lexer.NextToken()
		assert.Equal(t, expectedType, tok.Type, "test[%d] - wrong token type for %s", i, tok.Literal)
	}
}

func TestLexer_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected []struct {
			tokenType token.TokenType
			literal   string
		}
	}{
		{
			input: "42",
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.INT, "42"},
			},
		},
		{
			input: "3.14",
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.FLOAT, "3.14"},
			},
		},
		{
			input: "123 456.789 0 0.0",
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.INT, "123"},
				{token.FLOAT, "456.789"},
				{token.INT, "0"},
				{token.FLOAT, "0.0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := New(tt.input)

			for i, expected := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, expected.tokenType, tok.Type, "test[%d] - wrong token type", i)
				assert.Equal(t, expected.literal, tok.Literal, "test[%d] - wrong literal", i)
			}
		})
	}
}

func TestLexer_Strings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			tokenType token.TokenType
			literal   string
		}
	}{
		{
			name:  "double quotes",
			input: `"hello world"`,
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.STRING, "hello world"},
			},
		},
		{
			name:  "single quotes",
			input: `'hello world'`,
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.STRING, "hello world"},
			},
		},
		{
			name:  "empty string",
			input: `""`,
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.STRING, ""},
			},
		},
		{
			name:  "string with escapes",
			input: `"hello\nworld\t!"`,
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.STRING, "hello\nworld\t!"},
			},
		},
		{
			name:  "f-string",
			input: `f"Hello {name}!"`,
			expected: []struct {
				tokenType token.TokenType
				literal   string
			}{
				{token.FSTRING, "Hello {name}!"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)

			for i, expected := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, expected.tokenType, tok.Type, "test[%d] - wrong token type", i)
				assert.Equal(t, expected.literal, tok.Literal, "test[%d] - wrong literal", i)
			}
		})
	}
}

func TestLexer_Indentation(t *testing.T) {
	input := `spell test():
    if True:
        return 42
    else:
        return 0`

	expected := []token.TokenType{
		token.SPELL, token.IDENT, token.LPAREN, token.RPAREN, token.COLON, token.NEWLINE,
		token.INDENT, token.IF, token.TRUE, token.COLON, token.NEWLINE,
		token.INDENT, token.RETURN, token.INT, token.NEWLINE,
		token.DEDENT, token.ELSE, token.COLON, token.NEWLINE,
		token.INDENT, token.RETURN, token.INT, token.NEWLINE,
		token.DEDENT, token.DEDENT, token.EOF,
	}

	lexer := New(input)

	var tokens []token.Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	require.Equal(t, len(expected), len(tokens), "wrong number of tokens")

	for i, expectedType := range expected {
		assert.Equal(t, expectedType, tokens[i].Type, "test[%d] - wrong token type. got=%s", i, tokens[i].Type)
	}
}

func TestLexer_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{
			name:     "line comment",
			input:    "# This is a comment\nspell test():",
			expected: []token.TokenType{token.COMMENT, token.NEWLINE, token.SPELL, token.IDENT, token.LPAREN, token.RPAREN, token.COLON, token.NEWLINE, token.EOF},
		},
		{
			name:     "block comment",
			input:    "/* This is a block comment */ spell test():",
			expected: []token.TokenType{token.COMMENT, token.SPELL, token.IDENT, token.LPAREN, token.RPAREN, token.COLON, token.NEWLINE, token.EOF},
		},
		{
			name:     "triple backtick comment",
			input:    "```\nThis is a\nmulti-line comment\n```\nspell test():",
			expected: []token.TokenType{token.COMMENT, token.NEWLINE, token.SPELL, token.IDENT, token.LPAREN, token.RPAREN, token.COLON, token.NEWLINE, token.EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)

			for i, expectedType := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, expectedType, tok.Type, "test[%d] - wrong token type", i)
			}
		})
	}
}

func TestLexer_WithFilename(t *testing.T) {
	input := `spell test():`
	filename := "test.crl"

	lexer := NewWithFilename(input, filename)

	tok := lexer.NextToken()
	assert.Equal(t, filename, tok.Filename)
	assert.Equal(t, token.SPELL, tok.Type)
}

func TestLexer_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError bool
		expectedToken token.TokenType
	}{
		{
			name:          "invalid character",
			input:         "$invalid",
			expectedError: true,
			expectedToken: token.ILLEGAL,
		},
		{
			name:          "unterminated string",
			input:         `"unterminated string`,
			expectedError: true,
			expectedToken: token.ILLEGAL,
		},
		{
			name:          "invalid number",
			input:         "123.456.789",
			expectedError: true,
			expectedToken: token.ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)
			tok := lexer.NextToken()

			if tt.expectedError {
				assert.Equal(t, tt.expectedToken, tok.Type)
				assert.True(t, tok.IsError())
			}
		})
	}
}

func TestLexer_Position(t *testing.T) {
	input := `spell test():
    return 42`

	lexer := New(input)

	// Test position tracking
	tests := []struct {
		expectedLine   int
		expectedColumn int
	}{
		{1, 1},  // spell
		{1, 7},  // test
		{1, 11}, // (
		{1, 12}, // )
		{1, 13}, // :
		{1, 14}, // \n
		{2, 1},  // INDENT
		{2, 5},  // return
		{2, 12}, // 42
	}

	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedLine, tok.Line, "test[%d] - wrong line", i)
		assert.Equal(t, tt.expectedColumn, tok.Column, "test[%d] - wrong column", i)
	}
}

func TestLexer_IncrementalParsing(t *testing.T) {
	// Test that lexer can handle partial/incomplete input gracefully
	input := `spell incomplete(`

	lexer := New(input)

	tokens := []token.Token{}
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF || tok.Type == token.ILLEGAL {
			break
		}
	}

	// Should get at least: spell, incomplete, (, EOF/ILLEGAL
	assert.GreaterOrEqual(t, len(tokens), 3)
	assert.Equal(t, token.SPELL, tokens[0].Type)
	assert.Equal(t, token.IDENT, tokens[1].Type)
	assert.Equal(t, token.LPAREN, tokens[2].Type)
}

func TestLexer_ComplexCarrionCode(t *testing.T) {
	input := `spell fibonacci(n):
    if n <= 1:
        return n
    else:
        return fibonacci(n-1) + fibonacci(n-2)

grim Calculator:
    init(name):
        self.name = name
    
    spell add(a, b):
        return a + b
    
    spell multiply(a, b):
        return a * b

# Create instance
calc = Calculator("MyCalc")
result = calc.add(5, 3)`

	lexer := New(input)

	var tokens []token.Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	// Verify we get reasonable number of tokens
	assert.Greater(t, len(tokens), 50)

	// Verify some key tokens are present
	tokenTypes := make(map[token.TokenType]int)
	for _, tok := range tokens {
		tokenTypes[tok.Type]++
	}

	assert.Greater(t, tokenTypes[token.SPELL], 0, "Should have SPELL tokens")
	assert.Greater(t, tokenTypes[token.GRIM], 0, "Should have GRIM tokens")
	assert.Greater(t, tokenTypes[token.INIT], 0, "Should have INIT tokens")
	assert.Greater(t, tokenTypes[token.SELF], 0, "Should have SELF tokens")
	assert.Greater(t, tokenTypes[token.IF], 0, "Should have IF tokens")
	assert.Greater(t, tokenTypes[token.RETURN], 0, "Should have RETURN tokens")
	assert.Greater(t, tokenTypes[token.IDENT], 0, "Should have IDENT tokens")
	assert.Greater(t, tokenTypes[token.INDENT], 0, "Should have INDENT tokens")
	assert.Greater(t, tokenTypes[token.DEDENT], 0, "Should have DEDENT tokens")
}

func TestLexer_SecurityLimits(t *testing.T) {
	// Test with very large input to ensure security limits
	largeInput := make([]byte, 10*1024*1024) // 10MB
	for i := range largeInput {
		largeInput[i] = 'a'
	}

	lexer := New(string(largeInput))

	// Should handle large input gracefully without crashing
	tok := lexer.NextToken()
	assert.Equal(t, token.IDENT, tok.Type)
	assert.NotEmpty(t, tok.Literal)
}

func TestLexer_UnicodeSupport(t *testing.T) {
	input := `spell 测试():
    message = "Hello 世界"
    return message`

	lexer := New(input)

	// Should handle Unicode identifiers and strings
	tok := lexer.NextToken()
	assert.Equal(t, token.SPELL, tok.Type)

	tok = lexer.NextToken()
	assert.Equal(t, token.IDENT, tok.Type)
	assert.Equal(t, "测试", tok.Literal)
}

func TestLexer_EmptyInput(t *testing.T) {
	lexer := New("")

	tok := lexer.NextToken()
	assert.Equal(t, token.EOF, tok.Type)
}

func TestLexer_WhitespaceOnly(t *testing.T) {
	lexer := New("   \n\n  \t  \n")

	// Should get newlines and EOF, skip other whitespace
	tokens := []token.Token{}
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	// Should get NEWLINE tokens and EOF
	assert.Greater(t, len(tokens), 1)
	assert.Equal(t, token.EOF, tokens[len(tokens)-1].Type)
}
