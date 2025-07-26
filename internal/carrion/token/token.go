package token

import "fmt"

// TokenType represents the type of a token
type TokenType string

// Token represents a single token with position information
type Token struct {
	Type     TokenType
	Literal  string
	Filename string
	Line     int // 1-based line number
	Column   int // 1-based column number
}

// Token type constants - based on Carrion language specification
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Line structure
	NEWLINE TokenType = "NEWLINE"
	INDENT  TokenType = "INDENT"
	DEDENT  TokenType = "DEDENT"

	// Comments
	COMMENT TokenType = "COMMENT"

	// Identifiers and literals
	IDENT     TokenType = "IDENT"     // variable names, function names
	INT       TokenType = "INT"       // integers
	FLOAT     TokenType = "FLOAT"     // floating point numbers
	STRING    TokenType = "STRING"    // string literals
	FSTRING   TokenType = "FSTRING"   // f-string literals
	DOCSTRING TokenType = "DOCSTRING" // documentation strings
	INTERP    TokenType = "INTERP"    // interpolated expressions

	// Operators - Arithmetic
	PLUS      TokenType = "PLUS"      // +
	MINUS     TokenType = "MINUS"     // -
	ASTERISK  TokenType = "ASTERISK"  // *
	SLASH     TokenType = "SLASH"     // /
	MODULO    TokenType = "MODULO"    // %
	POWER     TokenType = "POWER"     // **
	FLOOR_DIV TokenType = "FLOOR_DIV" // //

	// Operators - Assignment
	ASSIGN          TokenType = "ASSIGN"          // =
	PLUS_ASSIGN     TokenType = "PLUS_ASSIGN"     // +=
	MINUS_ASSIGN    TokenType = "MINUS_ASSIGN"    // -=
	MULTIPLY_ASSIGN TokenType = "MULTIPLY_ASSIGN" // *=
	DIVIDE_ASSIGN   TokenType = "DIVIDE_ASSIGN"   // /=
	MODULO_ASSIGN   TokenType = "MODULO_ASSIGN"   // %=
	POWER_ASSIGN    TokenType = "POWER_ASSIGN"    // **=
	FLOOR_ASSIGN    TokenType = "FLOOR_ASSIGN"    // //=

	// Operators - Increment/Decrement
	INCREMENT TokenType = "INCREMENT" // ++
	DECREMENT TokenType = "DECREMENT" // --

	// Operators - Comparison
	EQ     TokenType = "EQ"     // ==
	NOT_EQ TokenType = "NOT_EQ" // !=
	LT     TokenType = "LT"     // <
	GT     TokenType = "GT"     // >
	LTE    TokenType = "LTE"    // <=
	GTE    TokenType = "GTE"    // >=

	// Operators - Logical
	AND TokenType = "AND" // and
	OR  TokenType = "OR"  // or
	NOT TokenType = "NOT" // not

	// Operators - Bitwise
	BITWISE_AND TokenType = "BITWISE_AND" // &
	BITWISE_OR  TokenType = "BITWISE_OR"  // |
	BITWISE_XOR TokenType = "BITWISE_XOR" // ^
	BITWISE_NOT TokenType = "BITWISE_NOT" // ~
	LEFT_SHIFT  TokenType = "LEFT_SHIFT"  // <<
	RIGHT_SHIFT TokenType = "RIGHT_SHIFT" // >>

	// Operators - Other
	IN     TokenType = "IN"     // in
	NOT_IN TokenType = "NOT_IN" // not in
	IS     TokenType = "IS"     // is
	IS_NOT TokenType = "IS_NOT" // is not

	// Delimiters
	COMMA     TokenType = "COMMA"     // ,
	SEMICOLON TokenType = "SEMICOLON" // ;
	COLON     TokenType = "COLON"     // :
	DOT       TokenType = "DOT"       // .
	ARROW     TokenType = "ARROW"     // ->
	UNPACK    TokenType = "UNPACK"    // <-
	HASH      TokenType = "HASH"      // #
	AT        TokenType = "AT"        // @

	// Brackets
	LPAREN   TokenType = "LPAREN"   // (
	RPAREN   TokenType = "RPAREN"   // )
	LBRACE   TokenType = "LBRACE"   // {
	RBRACE   TokenType = "RBRACE"   // }
	LBRACKET TokenType = "LBRACKET" // [
	RBRACKET TokenType = "RBRACKET" // ]

	// Keywords - Function/Class definition
	SPELL TokenType = "SPELL" // spell (function)
	GRIM  TokenType = "GRIM"  // grim (class)

	// Keywords - OOP
	ARCANE TokenType = "ARCANE" // arcane (abstract)
	INIT   TokenType = "INIT"   // init (constructor)
	SELF   TokenType = "SELF"   // self
	SUPER  TokenType = "SUPER"  // super

	// Keywords - Control flow
	IF        TokenType = "IF"        // if
	OTHERWISE TokenType = "OTHERWISE" // otherwise
	ELSE      TokenType = "ELSE"      // else
	FOR       TokenType = "FOR"       // for
	WHILE     TokenType = "WHILE"     // while
	SKIP      TokenType = "SKIP"      // skip (continue)
	STOP      TokenType = "STOP"      // stop (break)
	RETURN    TokenType = "RETURN"    // return
	MATCH     TokenType = "MATCH"     // match
	CASE      TokenType = "CASE"      // case

	// Keywords - Error handling
	ATTEMPT TokenType = "ATTEMPT" // attempt (try)
	ENSNARE TokenType = "ENSNARE" // ensnare (catch)
	RESOLVE TokenType = "RESOLVE" // resolve (finally)
	RAISE   TokenType = "RAISE"   // raise (throw)
	CHECK   TokenType = "CHECK"   // check (assert)

	// Keywords - Boolean/Null
	TRUE  TokenType = "TRUE"  // True
	FALSE TokenType = "FALSE" // False
	NONE  TokenType = "NONE"  // None

	// Keywords - Import/Module
	IMPORT TokenType = "IMPORT" // import
	AS     TokenType = "AS"     // as

	// Keywords - Other
	GLOBAL    TokenType = "GLOBAL"    // global
	IGNORE    TokenType = "IGNORE"    // ignore
	MAIN      TokenType = "MAIN"      // main
	AUTOCLOSE TokenType = "AUTOCLOSE" // autoclose
	DIVERGE   TokenType = "DIVERGE"   // diverge
	CONVERGE  TokenType = "CONVERGE"  // converge
)

// keywords maps string literals to their token types
var keywords = map[string]TokenType{
	// Function/Class definition
	"spell": SPELL,
	"grim":  GRIM,

	// OOP
	"arcane": ARCANE,
	"init":   INIT,
	"self":   SELF,
	"super":  SUPER,

	// Control flow
	"if":        IF,
	"otherwise": OTHERWISE,
	"else":      ELSE,
	"for":       FOR,
	"while":     WHILE,
	"in":        IN,
	"skip":      SKIP,
	"stop":      STOP,
	"return":    RETURN,
	"match":     MATCH,
	"case":      CASE,

	// Error handling
	"attempt": ATTEMPT,
	"ensnare": ENSNARE,
	"resolve": RESOLVE,
	"raise":   RAISE,
	"check":   CHECK,

	// Logical operators
	"and": AND,
	"or":  OR,
	"not": NOT,

	// Boolean/Null
	"True":  TRUE,
	"False": FALSE,
	"None":  NONE,

	// Import/Module
	"import": IMPORT,
	"as":     AS,

	// Other
	"global":    GLOBAL,
	"ignore":    IGNORE,
	"main":      MAIN,
	"autoclose": AUTOCLOSE,
	"diverge":   DIVERGE,
	"converge":  CONVERGE,
}

// NewToken creates a new token with the given parameters
func NewToken(tokenType TokenType, literal, filename string, line, column int) Token {
	return Token{
		Type:     tokenType,
		Literal:  literal,
		Filename: filename,
		Line:     line,
		Column:   column,
	}
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// String returns a string representation of the TokenType
func (tt TokenType) String() string {
	return string(tt)
}

// String returns a string representation of the Token
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: '%s', Position: %d:%d}",
		t.Type, t.Literal, t.Line, t.Column)
}

// LSPPosition converts 1-based Carrion positions to 0-based LSP positions
func (t Token) LSPPosition() (line, character int) {
	return t.Line - 1, t.Column - 1
}

// Range returns the LSP range for this token (start and end positions)
func (t Token) Range() (startLine, startChar, endLine, endChar int) {
	startLine, startChar = t.LSPPosition()
	endLine = startLine
	endChar = startChar + len(t.Literal)
	return
}

// IsKeyword returns true if this token is a keyword
func (t Token) IsKeyword() bool {
	switch t.Type {
	case SPELL, GRIM, ARCANE, INIT, SELF, SUPER,
		IF, OTHERWISE, ELSE, FOR, WHILE, IN, SKIP, STOP, RETURN, MATCH, CASE,
		ATTEMPT, ENSNARE, RESOLVE, RAISE, CHECK,
		AND, OR, NOT, TRUE, FALSE, NONE,
		IMPORT, AS, GLOBAL, IGNORE, MAIN, AUTOCLOSE, DIVERGE, CONVERGE:
		return true
	default:
		return false
	}
}

// IsOperator returns true if this token is an operator
func (t Token) IsOperator() bool {
	switch t.Type {
	case PLUS, MINUS, ASTERISK, SLASH, MODULO, POWER, FLOOR_DIV,
		ASSIGN, PLUS_ASSIGN, MINUS_ASSIGN, MULTIPLY_ASSIGN, DIVIDE_ASSIGN,
		MODULO_ASSIGN, POWER_ASSIGN, FLOOR_ASSIGN, INCREMENT, DECREMENT,
		EQ, NOT_EQ, LT, GT, LTE, GTE, AND, OR, NOT,
		BITWISE_AND, BITWISE_OR, BITWISE_XOR, BITWISE_NOT, LEFT_SHIFT, RIGHT_SHIFT,
		IN, NOT_IN, IS, IS_NOT:
		return true
	default:
		return false
	}
}

// IsLiteral returns true if this token is a literal value
func (t Token) IsLiteral() bool {
	switch t.Type {
	case INT, FLOAT, STRING, FSTRING, TRUE, FALSE, NONE:
		return true
	default:
		return false
	}
}

// IsError returns true if this token represents an error
func (t Token) IsError() bool {
	return t.Type == ILLEGAL
}

// IsEOF returns true if this token represents end of file
func (t Token) IsEOF() bool {
	return t.Type == EOF
}
