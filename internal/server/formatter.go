package server

import (
	"strings"

	"github.com/javanhut/carrion-lsp/internal/carrion/token"
	"github.com/javanhut/carrion-lsp/internal/protocol"
)

// CarrionFormatter handles code formatting for Carrion language
type CarrionFormatter struct {
	TabSize      int
	InsertSpaces bool
}

// NewCarrionFormatter creates a new formatter with given options
func NewCarrionFormatter(options protocol.FormattingOptions) *CarrionFormatter {
	return &CarrionFormatter{
		TabSize:      options.TabSize,
		InsertSpaces: options.InsertSpaces,
	}
}

// FormatDocument formats the entire document and returns text edits
func (f *CarrionFormatter) FormatDocument(text string) []protocol.TextEdit {
	lines := strings.Split(text, "\n")
	var edits []protocol.TextEdit

	indentLevel := 0
	var formattedLines []string

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			formattedLines = append(formattedLines, "")
			continue
		}

		// Handle dedents (lines that decrease indentation)
		if f.isDeindentLine(trimmedLine) {
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
			}
		}

		// Format the line with proper indentation
		indentStr := f.getIndentString(indentLevel)
		formattedLine := indentStr + f.formatLineContent(trimmedLine)
		formattedLines = append(formattedLines, formattedLine)

		// Handle indents (lines that increase indentation)
		if f.isIndentLine(trimmedLine) {
			indentLevel++
		}

		// Create edit if the line changed
		if line != formattedLine {
			edit := protocol.TextEdit{
				Range: protocol.Range{
					Start: protocol.Position{Line: i, Character: 0},
					End:   protocol.Position{Line: i, Character: len(line)},
				},
				NewText: formattedLine,
			}
			edits = append(edits, edit)
		}
	}

	return edits
}

// isIndentLine checks if a line should increase indentation for the next line
func (f *CarrionFormatter) isIndentLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Lines ending with ':' typically increase indentation
	if strings.HasSuffix(trimmed, ":") {
		return true
	}

	return false
}

// isDeindentLine checks if a line should decrease indentation
func (f *CarrionFormatter) isDeindentLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Keywords that typically decrease indentation
	deindentKeywords := []string{"else:", "except:", "finally:", ""}

	for _, keyword := range deindentKeywords {
		if keyword != "" && strings.HasPrefix(trimmed, keyword) {
			return true
		}
	}

	return false
}

// getIndentString returns the appropriate indentation string
func (f *CarrionFormatter) getIndentString(level int) string {
	if f.InsertSpaces {
		return strings.Repeat(" ", level*f.TabSize)
	}
	return strings.Repeat("\t", level)
}

// formatLineContent formats the content of a line (without indentation)
func (f *CarrionFormatter) formatLineContent(line string) string {
	// For now, just return the line as-is to focus on indentation
	// In the future, this could handle spacing around operators, etc.
	return line
}

// getSpacingBetweenTokens determines appropriate spacing between two tokens
func (f *CarrionFormatter) getSpacingBetweenTokens(current, next token.Token) string {
	// No space before punctuation (except for operators)
	if f.isPunctuation(next.Type) && !f.isOperator(next.Type) {
		return ""
	}

	// No space after opening brackets
	if current.Type == token.LPAREN || current.Type == token.LBRACKET || current.Type == token.LBRACE {
		return ""
	}

	// No space before closing brackets
	if next.Type == token.RPAREN || next.Type == token.RBRACKET || next.Type == token.RBRACE {
		return ""
	}

	// No space before opening parentheses (function calls)
	if next.Type == token.LPAREN {
		return ""
	}

	// No space after dots or before dots
	if current.Type == token.DOT || next.Type == token.DOT {
		return ""
	}

	// Space around operators (but not assignment which is handled separately)
	if f.isOperator(current.Type) && current.Type != token.ASSIGN {
		return " "
	}
	if f.isOperator(next.Type) && next.Type != token.ASSIGN {
		return " "
	}

	// Space around assignment
	if current.Type == token.ASSIGN || next.Type == token.ASSIGN {
		return " "
	}

	// Space after commas
	if current.Type == token.COMMA {
		return " "
	}

	// Space after certain keywords (but not before parentheses)
	if f.isKeyword(current.Type) && next.Type != token.LPAREN && next.Type != token.COLON {
		return " "
	}

	// No space between identifiers and literals that are already properly spaced
	if (current.Type == token.IDENT && (next.Type == token.STRING || next.Type == token.INT || next.Type == token.FLOAT)) ||
		(next.Type == token.IDENT && (current.Type == token.STRING || current.Type == token.INT || current.Type == token.FLOAT)) {
		return " "
	}

	// Default: no space (most tokens don't need spacing)
	return ""
}

// isPunctuation checks if a token type is punctuation
func (f *CarrionFormatter) isPunctuation(t token.TokenType) bool {
	punctuation := []token.TokenType{
		token.COMMA, token.DOT, token.SEMICOLON, token.COLON,
		token.RPAREN, token.RBRACKET, token.RBRACE,
	}

	for _, p := range punctuation {
		if t == p {
			return true
		}
	}
	return false
}

// isOperator checks if a token type is an operator
func (f *CarrionFormatter) isOperator(t token.TokenType) bool {
	operators := []token.TokenType{
		token.ASSIGN, token.PLUS, token.MINUS, token.ASTERISK, token.SLASH,
		token.MODULO, token.POWER, token.FLOOR_DIV,
		token.EQ, token.NOT_EQ, token.LT, token.GT, token.LTE, token.GTE,
		token.AND, token.OR, token.NOT, token.IN, token.NOT_IN, token.IS, token.IS_NOT,
	}

	for _, op := range operators {
		if t == op {
			return true
		}
	}
	return false
}

// isKeyword checks if a token type is a keyword
func (f *CarrionFormatter) isKeyword(t token.TokenType) bool {
	keywords := []token.TokenType{
		token.SPELL, token.GRIM, token.IF, token.ELSE, token.WHILE, token.FOR,
		token.RETURN, token.IMPORT, token.AS, token.IN, token.AND, token.OR, token.NOT,
		token.TRUE, token.FALSE, token.NONE, token.IGNORE,
	}

	for _, kw := range keywords {
		if t == kw {
			return true
		}
	}
	return false
}
