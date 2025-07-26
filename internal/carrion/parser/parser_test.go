package parser

import (
	"fmt"
	"testing"

	"github.com/javanhut/carrion-lsp/internal/carrion/ast"
	"github.com/javanhut/carrion-lsp/internal/carrion/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create parser from input
func createParser(input string) *Parser {
	l := lexer.New(input)
	p := New(l)
	return p
}

// Helper function to check parser errors
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestAssignStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"x = 5", "x", 5},
		{"y = True", "y", true},
		{"foobar = y", "foobar", "y"},
		{"z = 3.14", "z", 3.14},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt := program.Statements[0]
		testAssignStatement(t, stmt, tt.expectedIdentifier)

		assignStmt := stmt.(*ast.AssignStatement)
		testLiteralExpression(t, assignStmt.Value, tt.expectedValue)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5", 5},
		{"return True", true},
		{"return foobar", "foobar"},
		{"return", nil},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		require.True(t, ok, "stmt not *ast.ReturnStatement")
		assert.Equal(t, "return", returnStmt.TokenLiteral())

		if tt.expectedValue != nil {
			testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue)
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

	ident, ok := stmt.Expression.(*ast.Identifier)
	require.True(t, ok, "exp not *ast.Identifier")
	assert.Equal(t, "foobar", ident.Value)
	assert.Equal(t, "foobar", ident.TokenLiteral())
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	require.True(t, ok, "exp not *ast.IntegerLiteral")
	assert.Equal(t, int64(5), literal.Value)
	assert.Equal(t, "5", literal.TokenLiteral())
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "3.14"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

	literal, ok := stmt.Expression.(*ast.FloatLiteral)
	require.True(t, ok, "exp not *ast.FloatLiteral")
	assert.Equal(t, 3.14, literal.Value)
	assert.Equal(t, "3.14", literal.TokenLiteral())
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world"`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	require.True(t, ok, "exp not *ast.StringLiteral")
	assert.Equal(t, "hello world", literal.Value)
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"True", true},
		{"False", false},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

		boolean, ok := stmt.Expression.(*ast.BooleanLiteral)
		require.True(t, ok, "exp not *ast.BooleanLiteral")
		assert.Equal(t, tt.expectedBoolean, boolean.Value)
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"-15", "-", 15},
		{"+10", "+", 10},
		{"not True", "not", true},
		{"not False", "not", false},
		{"~42", "~", 42},
	}

	for _, tt := range prefixTests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		require.True(t, ok, "stmt is not ast.PrefixExpression")
		assert.Equal(t, tt.operator, exp.Operator)
		testLiteralExpression(t, exp.Right, tt.value)
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"True == True", true, "==", true},
		{"True != False", true, "!=", false},
		{"False == False", false, "==", false},
	}

	for _, tt := range infixTests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")

		testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"not -a", "(not (-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"True", "True"},
		{"False", "False"},
		{"3 > 5 == False", "((3 > 5) == False)"},
		{"3 < 5 == True", "((3 < 5) == True)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"not (True == True)", "(not (True == True))"},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		assert.Equal(t, tt.expected, actual)
	}
}

func TestIfExpression(t *testing.T) {
	input := `if x < y:
    x`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	require.True(t, ok, "program.Statements[0] is not ast.IfStatement")

	testInfixExpression(t, stmt.Condition, "x", "<", "y")

	require.Len(t, stmt.Consequence.Statements, 1, "consequence should have 1 statement")

	consequence, ok := stmt.Consequence.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "Statements[0] is not ast.ExpressionStatement")

	testIdentifier(t, consequence.Expression, "x")

	assert.Nil(t, stmt.Alternative, "stmt.Alternative should be nil")
}

func TestIfElseExpression(t *testing.T) {
	input := `if x < y:
    x
else:
    y`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	require.True(t, ok, "program.Statements[0] is not ast.IfStatement")

	testInfixExpression(t, stmt.Condition, "x", "<", "y")

	require.Len(t, stmt.Consequence.Statements, 1, "consequence should have 1 statement")

	consequence, ok := stmt.Consequence.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "Statements[0] is not ast.ExpressionStatement")

	testIdentifier(t, consequence.Expression, "x")

	require.NotNil(t, stmt.Alternative, "stmt.Alternative should not be nil")
	require.Len(t, stmt.Alternative.Statements, 1, "alternative should have 1 statement")

	alternative, ok := stmt.Alternative.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "Alternative.Statements[0] is not ast.ExpressionStatement")

	testIdentifier(t, alternative.Expression, "y")
}

func TestFunctionStatement(t *testing.T) {
	input := `spell add(x, y):
    x + y`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.FunctionStatement)
	require.True(t, ok, "program.Statements[0] is not ast.FunctionStatement")

	assert.Equal(t, "add", stmt.Name.Value)
	require.Len(t, stmt.Parameters, 2, "function should have 2 parameters")

	testLiteralExpression(t, stmt.Parameters[0], "x")
	testLiteralExpression(t, stmt.Parameters[1], "y")

	require.Len(t, stmt.Body.Statements, 1, "body should have 1 statement")

	bodyStmt, ok := stmt.Body.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "function body stmt is not ast.ExpressionStatement")

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5)"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "stmt is not ast.ExpressionStatement")

	exp, ok := stmt.Expression.(*ast.CallExpression)
	require.True(t, ok, "stmt.Expression is not ast.CallExpression")

	testIdentifier(t, exp.Function, "add")
	require.Len(t, exp.Arguments, 3, "wrong length of arguments")

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestArrayLiteralParsing(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "statement is not ast.ExpressionStatement")

	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	require.True(t, ok, "exp not ast.ArrayLiteral")
	require.Len(t, array.Elements, 3, "len(array.Elements) not 3")

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "statement is not ast.ExpressionStatement")

	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	require.True(t, ok, "exp not *ast.IndexExpression")

	testIdentifier(t, indexExp.Left, "myArray")
	testInfixExpression(t, indexExp.Index, 1, "+", 1)
}

func TestHashLiteralParsing(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, "statement is not ast.ExpressionStatement")

	hash, ok := stmt.Expression.(*ast.HashLiteral)
	require.True(t, ok, "exp is not ast.HashLiteral")

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	require.Len(t, hash.Pairs, 3, "hash.Pairs has wrong length")

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		require.True(t, ok, "key is not ast.StringLiteral")

		expectedValue := expected[literal.Value]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestClassStatement(t *testing.T) {
	input := `grim Person:
    pass`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 1, "program should have 1 statement")

	stmt, ok := program.Statements[0].(*ast.ClassStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ClassStatement")

	assert.Equal(t, "Person", stmt.Name.Value)
	assert.Nil(t, stmt.Parent, "Parent should be nil")
}

func TestImportStatement(t *testing.T) {
	tests := []struct {
		input          string
		expectedModule string
		expectedAlias  string
	}{
		{"import os", "os", ""},
		{"import sys as system", "sys", "system"},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt, ok := program.Statements[0].(*ast.ImportStatement)
		require.True(t, ok, "program.Statements[0] is not ast.ImportStatement")

		assert.Equal(t, tt.expectedModule, stmt.Module.Value)
		if tt.expectedAlias != "" {
			require.NotNil(t, stmt.Alias, "Alias should not be nil")
			assert.Equal(t, tt.expectedAlias, stmt.Alias.Value)
		} else {
			assert.Nil(t, stmt.Alias, "Alias should be nil")
		}
	}
}

func TestMemberExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"obj.member", "obj.member"},
		{"obj.method()", "obj.method()"},
		{"ex.example_print()", "ex.example_print()"},
		{"self.value", "self.value"},
		{"a.b.c", "a.b.c"},
		{"a.b.c()", "a.b.c()"},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Len(t, program.Statements, 1, "program should have 1 statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		require.True(t, ok, "program.Statements[0] is not ast.ExpressionStatement")
		
		assert.Equal(t, tt.expected, stmt.Expression.String())
	}
}

func TestMemberExpressionWithNewline(t *testing.T) {
	// Test that member expressions stop at newlines
	input := `ex = Example()
ex.example_print()`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Len(t, program.Statements, 2, "program should have 2 statements")

	// First statement should be assignment
	stmt1, ok := program.Statements[0].(*ast.AssignStatement)
	require.True(t, ok, "program.Statements[0] is not ast.AssignStatement")
	assert.Equal(t, "ex", stmt1.Name.Value)

	// Second statement should be method call
	stmt2, ok := program.Statements[1].(*ast.ExpressionStatement)
	require.True(t, ok, "program.Statements[1] is not ast.ExpressionStatement")
	assert.Equal(t, "ex.example_print()", stmt2.Expression.String())
}

// HELPER FUNCTIONS

func testAssignStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != name {
		t.Errorf("s.TokenLiteral not '%s'. got=%s", name, s.TokenLiteral())
		return false
	}

	assignStmt, ok := s.(*ast.AssignStatement)
	if !ok {
		t.Errorf("s not *ast.AssignStatement. got=%T", s)
		return false
	}

	if assignStmt.Name.Value != name {
		t.Errorf("assignStmt.Name.Value not '%s'. got=%s", name, assignStmt.Name.Value)
		return false
	}

	if assignStmt.Name.TokenLiteral() != name {
		t.Errorf("assignStmt.Name.TokenLiteral() not '%s'. got=%s",
			name, assignStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	case float64:
		return testFloatLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp not *ast.BooleanLiteral. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	expectedLiteral := "True"
	if !value {
		expectedLiteral = "False"
	}
	if bo.TokenLiteral() != expectedLiteral {
		t.Errorf("bo.TokenLiteral not %s. got=%s",
			expectedLiteral, bo.TokenLiteral())
		return false
	}

	return true
}

func testFloatLiteral(t *testing.T, exp ast.Expression, value float64) bool {
	fl, ok := exp.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("exp not *ast.FloatLiteral. got=%T", exp)
		return false
	}

	if fl.Value != value {
		t.Errorf("fl.Value not %f. got=%f", value, fl.Value)
		return false
	}

	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}
