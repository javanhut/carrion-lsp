package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/javanhut/carrion-lsp/internal/carrion/token"
)

// Lexer represents the lexical analyzer for Carrion
type Lexer struct {
	input        string
	position     int  // current byte position in input (for slicing)
	readPosition int  // current reading byte position in input (after current char)
	ch           rune // current char under examination
	sourceFile   string

	// Indentation tracking
	indentStack        []int
	tokenQueue         []token.Token
	atLineStart        bool
	implicitNewlineGen bool // tracks if we've generated the implicit EOF newline
}

// New creates a new lexer instance
func New(input string) *Lexer {
	l := &Lexer{
		input:       input,
		sourceFile:  "",
		indentStack: []int{0},
		atLineStart: true,
	}
	l.readChar()
	return l
}

// NewWithFilename creates a new lexer instance with a filename
func NewWithFilename(input, sourceFile string) *Lexer {
	l := &Lexer{
		input:       input,
		sourceFile:  sourceFile,
		indentStack: []int{0},
		atLineStart: true,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL character represents EOF
		l.position = l.readPosition
	} else {
		var size int
		l.ch, size = utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.position = l.readPosition
		l.readPosition += size
	}
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return ch
}

// peekCharN returns the character N positions ahead without advancing
func (l *Lexer) peekCharN(n int) rune {
	pos := l.readPosition
	for i := 0; i < n-1; i++ {
		if pos >= len(l.input) {
			return 0
		}
		_, size := utf8.DecodeRuneInString(l.input[pos:])
		pos += size
	}
	if pos >= len(l.input) {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[pos:])
	return ch
}

// getCurrentPosition returns current line and column (1-based)
func (l *Lexer) getCurrentPosition() (int, int) {
	line := 1
	col := 1

	// Count bytes up to current position
	for i := 0; i < l.position; {
		if i >= len(l.input) {
			break
		}
		r, size := utf8.DecodeRuneInString(l.input[i:])
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		i += size
	}

	return line, col
}

// NextToken scans and returns the next token
func (l *Lexer) NextToken() token.Token {
	// Return queued tokens first
	if len(l.tokenQueue) > 0 {
		tok := l.tokenQueue[0]
		l.tokenQueue = l.tokenQueue[1:]
		return tok
	}

	var tok token.Token

	// Handle indentation at start of line
	if l.atLineStart {
		l.atLineStart = false
		if indentTok := l.handleIndentation(); indentTok != nil {
			return *indentTok
		}
	}

	l.skipWhitespace()

	line, col := l.getCurrentPosition()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.EQ, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.ASSIGN, string(l.ch), line, col)
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.INCREMENT, string(ch)+string(l.ch), line, col)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.PLUS_ASSIGN, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.PLUS, string(l.ch), line, col)
		}
	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.DECREMENT, string(ch)+string(l.ch), line, col)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.MINUS_ASSIGN, string(ch)+string(l.ch), line, col)
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.ARROW, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.MINUS, string(l.ch), line, col)
		}
	case '*':
		if l.peekChar() == '*' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				secondCh := l.ch
				l.readChar()
				tok = l.newToken(token.POWER_ASSIGN, string(ch)+string(secondCh)+string(l.ch), line, col)
			} else {
				tok = l.newToken(token.POWER, string(ch)+string(l.ch), line, col)
			}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.MULTIPLY_ASSIGN, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.ASTERISK, string(l.ch), line, col)
		}
	case '/':
		if l.peekChar() == '/' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				secondCh := l.ch
				l.readChar()
				tok = l.newToken(token.FLOOR_ASSIGN, string(ch)+string(secondCh)+string(l.ch), line, col)
			} else {
				tok = l.newToken(token.FLOOR_DIV, string(ch)+string(l.ch), line, col)
			}
		} else if l.peekChar() == '*' {
			// Block comment
			return l.readBlockComment(line, col)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.DIVIDE_ASSIGN, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.SLASH, string(l.ch), line, col)
		}
	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.MODULO_ASSIGN, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.MODULO, string(l.ch), line, col)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.NOT_EQ, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.ILLEGAL, string(l.ch), line, col)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.LTE, string(ch)+string(l.ch), line, col)
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.LEFT_SHIFT, string(ch)+string(l.ch), line, col)
		} else if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.UNPACK, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.LT, string(l.ch), line, col)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.GTE, string(ch)+string(l.ch), line, col)
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.RIGHT_SHIFT, string(ch)+string(l.ch), line, col)
		} else {
			tok = l.newToken(token.GT, string(l.ch), line, col)
		}
	case '&':
		tok = l.newToken(token.BITWISE_AND, string(l.ch), line, col)
	case '|':
		tok = l.newToken(token.BITWISE_OR, string(l.ch), line, col)
	case '^':
		tok = l.newToken(token.BITWISE_XOR, string(l.ch), line, col)
	case '~':
		tok = l.newToken(token.BITWISE_NOT, string(l.ch), line, col)
	case '(':
		tok = l.newToken(token.LPAREN, string(l.ch), line, col)
	case ')':
		tok = l.newToken(token.RPAREN, string(l.ch), line, col)
	case '{':
		tok = l.newToken(token.LBRACE, string(l.ch), line, col)
	case '}':
		tok = l.newToken(token.RBRACE, string(l.ch), line, col)
	case '[':
		tok = l.newToken(token.LBRACKET, string(l.ch), line, col)
	case ']':
		tok = l.newToken(token.RBRACKET, string(l.ch), line, col)
	case ',':
		tok = l.newToken(token.COMMA, string(l.ch), line, col)
	case ';':
		tok = l.newToken(token.SEMICOLON, string(l.ch), line, col)
	case ':':
		tok = l.newToken(token.COLON, string(l.ch), line, col)
	case '.':
		tok = l.newToken(token.DOT, string(l.ch), line, col)
	case '#':
		return l.readLineComment(line, col)
	case '@':
		tok = l.newToken(token.AT, string(l.ch), line, col)
	case '\n':
		l.atLineStart = true
		tok = l.newToken(token.NEWLINE, string(l.ch), line, col)
	case '"':
		return l.readString('"', line, col)
	case '\'':
		return l.readString('\'', line, col)
	case '`':
		if l.peekChar() == '`' && l.peekCharN(2) == '`' {
			return l.readTripleBacktickComment(line, col)
		}
		tok = l.newToken(token.ILLEGAL, string(l.ch), line, col)
	case 0:
		// Add implicit NEWLINE at EOF if not at start of line and input is not empty
		if !l.atLineStart && !l.implicitNewlineGen && l.position > 0 {
			l.implicitNewlineGen = true
			l.atLineStart = true
			return l.newToken(token.NEWLINE, "", line, col)
		}
		// Generate remaining DEDENT tokens at EOF
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			for len(l.indentStack) > 1 {
				l.tokenQueue = append(l.tokenQueue, l.newToken(token.DEDENT, "", line, col))
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
			}
			return l.newToken(token.DEDENT, "", line, col)
		}
		tok = l.newToken(token.EOF, "", line, col)
	default:
		if isLetter(l.ch) {
			return l.readIdentifier(line, col)
		} else if isDigit(l.ch) {
			return l.readNumber(line, col)
		} else {
			tok = l.newToken(token.ILLEGAL, string(l.ch), line, col)
		}
	}

	l.readChar()
	return tok
}

// handleIndentation handles indentation and generates INDENT/DEDENT tokens
func (l *Lexer) handleIndentation() *token.Token {
	if l.ch == 0 {
		return nil
	}

	line, col := l.getCurrentPosition()
	indent := 0

	// Count leading whitespace
	for l.ch == ' ' || l.ch == '\t' {
		if l.ch == ' ' {
			indent++
		} else {
			indent += 4 // Tab counts as 4 spaces
		}
		l.readChar()
	}

	// Skip empty lines and comments
	if l.ch == '\n' || l.ch == '#' || l.ch == 0 {
		l.atLineStart = true
		return nil
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if indent > currentIndent {
		// Increased indentation
		l.indentStack = append(l.indentStack, indent)
		tok := l.newToken(token.INDENT, "", line, col)
		return &tok
	} else if indent < currentIndent {
		// Decreased indentation - may need multiple DEDENT tokens
		dedentCount := 0
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			dedentCount++
		}

		if dedentCount > 0 {
			// Queue additional DEDENT tokens
			for i := 1; i < dedentCount; i++ {
				l.tokenQueue = append(l.tokenQueue, l.newToken(token.DEDENT, "", line, col))
			}
			tok := l.newToken(token.DEDENT, "", line, col)
			return &tok
		}
	}

	return nil
}

// skipWhitespace skips spaces and tabs (but not newlines)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier(line, col int) token.Token {
	start := l.position

	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	literal := l.input[start:l.position]
	tokType := token.LookupIdent(literal)

	// Handle f-strings
	if literal == "f" && l.ch == '"' {
		return l.readFString(line, col)
	}

	return l.newToken(tokType, literal, line, col)
}

// readNumber reads integer and floating point numbers
func (l *Lexer) readNumber(line, col int) token.Token {
	start := l.position
	var tokType token.TokenType = token.INT

	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for floating point
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokType = token.FLOAT
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}

		// Check for invalid second decimal point
		if l.ch == '.' && isDigit(l.peekChar()) {
			// Invalid number with multiple decimal points
			for l.ch != 0 && (isDigit(l.ch) || l.ch == '.') {
				l.readChar()
			}
			literal := l.input[start:l.position]
			return l.newToken(token.ILLEGAL, literal, line, col)
		}
	}

	literal := l.input[start:l.position]
	return l.newToken(tokType, literal, line, col)
}

// readString reads string literals
func (l *Lexer) readString(delimiter rune, line, col int) token.Token {
	l.readChar() // Skip opening quote
	start := l.position

	for l.ch != delimiter && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // Skip escaped character
		}
		l.readChar()
	}

	if l.ch == 0 {
		return l.newToken(token.ILLEGAL, "unterminated string", line, col)
	}

	literal := l.input[start:l.position]
	// Process escape sequences
	literal = l.processEscapes(literal)

	// Advance past closing quote
	l.readChar()

	return l.newToken(token.STRING, literal, line, col)
}

// readFString reads f-string literals
func (l *Lexer) readFString(line, col int) token.Token {
	// Note: when called from readIdentifier, l.ch is '"'
	l.readChar() // Skip '"'

	start := l.position

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // Skip escaped character
		}
		l.readChar()
	}

	if l.ch == 0 {
		return l.newToken(token.ILLEGAL, "unterminated f-string", line, col)
	}

	literal := l.input[start:l.position]

	// Advance past closing quote
	l.readChar()

	return l.newToken(token.FSTRING, literal, line, col)
}

// readLineComment reads line comments starting with #
func (l *Lexer) readLineComment(line, col int) token.Token {
	start := l.position

	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	literal := l.input[start:l.position]
	return l.newToken(token.COMMENT, literal, line, col)
}

// readBlockComment reads block comments /* ... */
func (l *Lexer) readBlockComment(line, col int) token.Token {
	start := l.position
	l.readChar() // Skip '/'
	l.readChar() // Skip '*'

	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // Skip '*'
			l.readChar() // Skip '/'
			break
		}
		l.readChar()
	}

	literal := l.input[start:l.position]
	return l.newToken(token.COMMENT, literal, line, col)
}

// readTripleBacktickComment reads triple backtick comments ``` ... ```
func (l *Lexer) readTripleBacktickComment(line, col int) token.Token {
	start := l.position
	l.readChar() // Skip first `
	l.readChar() // Skip second `
	l.readChar() // Skip third `

	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '`' && l.peekChar() == '`' && l.peekCharN(2) == '`' {
			l.readChar() // Skip first `
			l.readChar() // Skip second `
			l.readChar() // Skip third `
			break
		}
		l.readChar()
	}

	literal := l.input[start:l.position]
	return l.newToken(token.COMMENT, literal, line, col)
}

// processEscapes processes escape sequences in strings
func (l *Lexer) processEscapes(s string) string {
	result := strings.Builder{}

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case '\'':
				result.WriteByte('\'')
			default:
				result.WriteByte(s[i])
				result.WriteByte(s[i+1])
			}
			i++ // Skip next character
		} else {
			result.WriteByte(s[i])
		}
	}

	return result.String()
}

// newToken creates a new token with position information
func (l *Lexer) newToken(tokenType token.TokenType, literal string, line, col int) token.Token {
	return token.Token{
		Type:     tokenType,
		Literal:  literal,
		Filename: l.sourceFile,
		Line:     line,
		Column:   col,
	}
}

// Helper functions

// isLetter checks if a character is a letter or underscore
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit checks if a character is a digit
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// isAlphaNumeric checks if a character is alphanumeric or underscore
func isAlphaNumeric(ch rune) bool {
	return isLetter(ch) || isDigit(ch)
}
