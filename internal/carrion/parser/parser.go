package parser

import (
	"fmt"
	"strconv"

	"github.com/javanhut/carrion-lsp/internal/carrion/ast"
	"github.com/javanhut/carrion-lsp/internal/carrion/lexer"
	"github.com/javanhut/carrion-lsp/internal/carrion/token"
)

// Precedence constants for operator precedence parsing
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

// precedences maps token types to their precedence
var precedences = map[token.TokenType]int{
	token.EQ:        EQUALS,
	token.NOT_EQ:    EQUALS,
	token.LT:        LESSGREATER,
	token.GT:        LESSGREATER,
	token.LTE:       LESSGREATER,
	token.GTE:       LESSGREATER,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.SLASH:     PRODUCT,
	token.ASTERISK:  PRODUCT,
	token.MODULO:    PRODUCT,
	token.POWER:     PRODUCT,
	token.FLOOR_DIV: PRODUCT,
	token.DOT:       CALL, // Same precedence as function calls
	token.LPAREN:    CALL,
	token.LBRACKET:  INDEX,
	token.AND:       EQUALS,
	token.OR:        EQUALS,
	token.IN:        EQUALS,
	token.NOT_IN:    EQUALS,
	token.IS:        EQUALS,
	token.IS_NOT:    EQUALS,
}

// Parser represents the parser
type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	// Pratt parsing function maps
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new parser instance
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	// Initialize prefix parse functions
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INIT, p.parseIdentifier) // Allow init as identifier
	p.registerPrefix(token.SELF, p.parseIdentifier) // Allow self as identifier
	p.registerPrefix(token.MAIN, p.parseIdentifier) // Allow main as identifier
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.FSTRING, p.parseFStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NONE, p.parseNoneLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS, p.parsePrefixExpression)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.BITWISE_NOT, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	// Initialize infix parse functions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.POWER, p.parseInfixExpression)
	p.registerInfix(token.FLOOR_DIV, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseInfixExpression)
	p.registerInfix(token.NOT_IN, p.parseInfixExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.IS_NOT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// nextToken advances the parser to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// Errors returns parsing errors
func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parses the entire program and returns the AST
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		// Skip newlines and indentation tokens at top level
		if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.INDENT) || p.curTokenIs(token.DEDENT) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	program.Errors = p.errors
	return program
}

// parseStatement parses a statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SPELL:
		return p.parseFunctionStatement()
	case token.GRIM:
		return p.parseClassStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.STOP:
		return p.parseStopStatement()
	case token.SKIP:
		return p.parseSkipStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.IGNORE:
		return p.parseIgnoreStatement()
	case token.IDENT, token.INIT, token.SELF, token.MAIN:
		// Check if this is an assignment or expression statement
		return p.parseAssignOrExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseAssignStatement parses assignment statements (x = 5)
func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curToken}

	if !p.curTokenIs(token.IDENT) {
		p.addError("expected identifier in assignment")
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// parseAssignOrExpressionStatement determines if this is assignment or expression
func (p *Parser) parseAssignOrExpressionStatement() ast.Statement {
	// Check if this is a bare init(): function definition (constructor)
	if (p.curTokenIs(token.INIT) || (p.curTokenIs(token.IDENT) && p.curToken.Literal == "init")) && p.peekTokenIs(token.LPAREN) {
		// This is init(): constructor syntax - parse as function
		return p.parseBareInitFunction()
	}
	
	// Check if this is a main: block definition (no parentheses)
	if (p.curTokenIs(token.MAIN) || (p.curTokenIs(token.IDENT) && p.curToken.Literal == "main")) && p.peekTokenIs(token.COLON) {
		// This is main: block syntax - parse as main block
		return p.parseMainBlockStatement()
	}
	
	// Look ahead to see if this is an assignment
	if p.curTokenIsIdent() && p.peekTokenIs(token.ASSIGN) {
		// Simple assignment: x = value
		return p.parseAssignStatement()
	}

	// Could be member assignment: obj.member = value
	// Parse the left side as expression first
	expr := p.parseExpression(LOWEST)

	// Check if it's a member expression followed by assignment
	if memberExpr, ok := expr.(*ast.MemberExpression); ok && p.peekTokenIs(token.ASSIGN) {
		return p.parseMemberAssignStatement(memberExpr)
	}

	// Check if it's a simple identifier followed by assignment
	if ident, ok := expr.(*ast.Identifier); ok && p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // consume ASSIGN
		p.nextToken() // move to value

		value := p.parseExpression(LOWEST)

		stmt := &ast.AssignStatement{
			Token: ident.Token,
			Name:  ident,
			Value: value,
		}

		// Skip optional newline
		if p.peekTokenIs(token.NEWLINE) {
			p.nextToken()
		}

		return stmt
	}

	// Otherwise it's an expression statement
	var stmtToken token.Token
	switch e := expr.(type) {
	case *ast.Identifier:
		stmtToken = e.Token
	case *ast.CallExpression:
		stmtToken = e.Token
	case *ast.MemberExpression:
		stmtToken = e.Token
	default:
		// Use current token as fallback
		stmtToken = p.curToken
	}
	stmt := &ast.ExpressionStatement{Token: stmtToken, Expression: expr}

	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// parseMemberAssignStatement parses member assignment (obj.member = value)
func (p *Parser) parseMemberAssignStatement(memberExpr *ast.MemberExpression) *ast.MemberAssignStatement {
	stmt := &ast.MemberAssignStatement{
		Token:  memberExpr.Token,
		Object: memberExpr.Object,
		Member: memberExpr.Member,
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parses return statements
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// Check if return has a value
	if !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// parseStopStatement parses stop statements (break)
func (p *Parser) parseStopStatement() *ast.StopStatement {
	stmt := &ast.StopStatement{Token: p.curToken}
	
	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}
	
	return stmt
}

// parseSkipStatement parses skip statements (continue)
func (p *Parser) parseSkipStatement() *ast.SkipStatement {
	stmt := &ast.SkipStatement{Token: p.curToken}
	
	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}
	
	return stmt
}

// parseExpressionStatement parses expression statements
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// parseBlockStatement parses block statements
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	// Skip any additional newlines before indent
	p.skipNewlines()

	// Expect INDENT to start block
	if !p.curTokenIs(token.INDENT) {
		p.addError(fmt.Sprintf("expected INDENT, got %s instead", p.curToken.Type))
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(token.DEDENT) && !p.curTokenIs(token.EOF) {
		// Skip newlines
		if p.curTokenIs(token.NEWLINE) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseIfStatement parses if statements
func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if !p.expectPeek(token.NEWLINE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	// Check for else clause
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.COLON) {
			return nil
		}

		if !p.expectPeek(token.NEWLINE) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

// parseWhileStatement parses while statements
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if !p.expectPeek(token.NEWLINE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForStatement parses for statements
func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if !p.expectPeek(token.NEWLINE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseFunctionStatement parses spell (function) definitions
func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeekIdent() {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if !p.expectPeek(token.NEWLINE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseFunctionParameters parses function parameters
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	if !p.expectPeekIdent() {
		return nil
	}

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if !p.expectPeekIdent() {
			return nil
		}
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// parseBareInitFunction parses init(): constructor functions without spell keyword
func (p *Parser) parseBareInitFunction() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}
	
	// Set the function name to "init"
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: "init"}
	
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	
	stmt.Parameters = p.parseFunctionParameters()
	
	if !p.expectPeek(token.COLON) {
		return nil
	}
	
	if !p.expectPeek(token.NEWLINE) {
		return nil
	}
	
	stmt.Body = p.parseBlockStatement()
	
	return stmt
}

// parseBareFunctionStatement parses bare function definitions like main(): without spell keyword
func (p *Parser) parseBareFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}
	
	// Set the function name from current token
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	
	stmt.Parameters = p.parseFunctionParameters()
	
	if !p.expectPeek(token.COLON) {
		return nil
	}
	
	if !p.expectPeek(token.NEWLINE) {
		return nil
	}
	
	stmt.Body = p.parseBlockStatement()
	
	return stmt
}

// parseMainBlockStatement parses main: block definitions (special Carrion syntax)
func (p *Parser) parseMainBlockStatement() *ast.BlockStatement {
	stmt := &ast.BlockStatement{Token: p.curToken}
	
	if !p.expectPeek(token.COLON) {
		return nil
	}
	
	if !p.expectPeek(token.NEWLINE) {
		return nil
	}
	
	// Parse the main block content
	if !p.expectPeek(token.INDENT) {
		return nil
	}
	
	p.nextToken()
	stmt.Statements = []ast.Statement{}
	
	for !p.curTokenIs(token.DEDENT) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.NEWLINE) {
			p.nextToken()
			continue
		}
		
		statement := p.parseStatement()
		if statement != nil {
			stmt.Statements = append(stmt.Statements, statement)
		}
		p.nextToken()
	}
	
	return stmt
}

// parseClassStatement parses grim (class) definitions
func (p *Parser) parseClassStatement() *ast.ClassStatement {
	stmt := &ast.ClassStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for inheritance
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.Parent = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if !p.expectPeek(token.NEWLINE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseImportStatement parses import statements
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Module = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for alias (import x as y)
	if p.peekTokenIs(token.AS) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	return stmt
}

// parseIgnoreStatement parses ignore statements
func (p *Parser) parseIgnoreStatement() *ast.IgnoreStatement {
	stmt := &ast.IgnoreStatement{Token: p.curToken}

	// Skip optional newline
	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// EXPRESSION PARSING

// parseExpression parses expressions using Pratt parsing
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.NEWLINE) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier parses identifiers
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral parses integer literals
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses float literals
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses string literals
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseFStringLiteral parses f-string literals
func (p *Parser) parseFStringLiteral() ast.Expression {
	return &ast.FStringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseBooleanLiteral parses boolean literals
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseNoneLiteral parses None literal
func (p *Parser) parseNoneLiteral() ast.Expression {
	return &ast.NoneLiteral{Token: p.curToken}
}

// parsePrefixExpression parses prefix expressions
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression parses infix expressions
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseGroupedExpression parses grouped expressions (parentheses)
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseArrayLiteral parses array literals
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseHashLiteral parses hash literals
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

// parseCallExpression parses function calls
func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: fn}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseIndexExpression parses index expressions
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseMemberExpression parses member access expressions (obj.member)
func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{Token: p.curToken, Object: left}

	if !p.expectPeekIdent() {
		return nil
	}

	exp.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

// parseExpressionList parses a list of expressions
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// HELPER METHODS

// curTokenIs checks if current token matches the given type
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if peek token matches the given type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks peek token and advances if it matches
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// expectPeekIdent checks for identifier or keywords that can be used as identifiers
func (p *Parser) expectPeekIdent() bool {
	if p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.INIT) || p.peekTokenIs(token.SELF) || p.peekTokenIs(token.MAIN) {
		p.nextToken()
		return true
	} else {
		p.peekError(token.IDENT)
		return false
	}
}

// skipNewlines skips multiple consecutive newline tokens
func (p *Parser) skipNewlines() {
	for p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}
}

// expectNewlineAndIndent expects at least one newline followed by indent, skipping extra newlines
func (p *Parser) expectNewlineAndIndent() bool {
	if !p.expectPeek(token.NEWLINE) {
		return false
	}
	
	// Skip additional newlines
	p.skipNewlines()
	
	if !p.curTokenIs(token.INDENT) {
		p.addError(fmt.Sprintf("expected INDENT, got %s instead", p.curToken.Type))
		return false
	}
	
	return true
}

// curTokenIsIdent checks if current token can be used as identifier
func (p *Parser) curTokenIsIdent() bool {
	return p.curTokenIs(token.IDENT) || p.curTokenIs(token.INIT) || p.curTokenIs(token.SELF) || p.curTokenIs(token.MAIN)
}

// peekPrecedence returns the precedence of the peek token
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence returns the precedence of the current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ERROR HANDLING

// addError adds an error message
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d, column %d: %s",
		p.curToken.Line, p.curToken.Column, msg))
}

// peekError adds a peek token error
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.addError(msg)
}

// noPrefixParseFnError adds a no prefix parse function error
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg)
}
