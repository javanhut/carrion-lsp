package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/javanhut/carrion-lsp/internal/carrion/ast"
)

func TestParseWithTripleBacktickComments(t *testing.T) {
	input := "```\nThis is a documentation comment\nusing triple backticks\n```\ngrim String:\n    ```\n    Initialize the string grimoire\n    ```\n    init(value):\n        self.value = value\n    \n    ```\n    Return the length of the string\n    ```\n    spell length():\n        return len(self.value)\n"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should have 1 statement (the grim declaration)
	require.Len(t, program.Statements, 1, "program should have 1 statement")

	// Check it's a class statement
	stmt, ok := program.Statements[0].(*ast.ClassStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ClassStatement")
	assert.Equal(t, "String", stmt.Name.Value)
}

func TestParseWithLineComments(t *testing.T) {
	input := "# This is a comment\nx = 5  # Another comment\n# More comments\ny = 10"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should have 2 statements (the assignments)
	require.Len(t, program.Statements, 2, "program should have 2 statements")

	// Check first assignment
	stmt1, ok := program.Statements[0].(*ast.AssignStatement)
	require.True(t, ok, "program.Statements[0] is not ast.AssignStatement")
	assert.Equal(t, "x", stmt1.Name.Value)

	// Check second assignment
	stmt2, ok := program.Statements[1].(*ast.AssignStatement)
	require.True(t, ok, "program.Statements[1] is not ast.AssignStatement")
	assert.Equal(t, "y", stmt2.Name.Value)
}

func TestParseWithBlockComments(t *testing.T) {
	input := "/* This is a block comment */\nx = 5\n/* Multi-line\n   block\n   comment */\ny = 10"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should have 2 statements (the assignments)
	require.Len(t, program.Statements, 2, "program should have 2 statements")

	// Check first assignment
	stmt1, ok := program.Statements[0].(*ast.AssignStatement)
	require.True(t, ok, "program.Statements[0] is not ast.AssignStatement")
	assert.Equal(t, "x", stmt1.Name.Value)

	// Check second assignment
	stmt2, ok := program.Statements[1].(*ast.AssignStatement)
	require.True(t, ok, "program.Statements[1] is not ast.AssignStatement")
	assert.Equal(t, "y", stmt2.Name.Value)
}

func TestParseMuninStyleClass(t *testing.T) {
	// Test actual Munin-style class with docstrings
	input := "grim String:\n" +
		"    ```\n" +
		"    Initializes a new String grimoire instance with the provided text value.\n" +
		"    \n" +
		"    Parameters:\n" +
		"        value: The string content to be managed by this grimoire instance\n" +
		"    ```\n" +
		"    init(value):\n" +
		"        self.value = value\n" +
		"    \n" +
		"    ```\n" +
		"    Returns the number of characters in the string.\n" +
		"    \n" +
		"    Returns:\n" +
		"        Integer representing the total number of characters in the string\n" +
		"    ```\n" +
		"    spell length():\n" +
		"        return len(self.value)\n" +
		"    \n" +
		"    ```\n" +
		"    Converts all uppercase letters in the string to lowercase.\n" +
		"    \n" +
		"    Returns:\n" +
		"        A new string with all uppercase letters converted to lowercase\n" +
		"    ```\n" +
		"    spell lower():\n" +
		"        result = \"\"\n" +
		"        for char in self.value:\n" +
		"            code = ord(char)\n" +
		"            if code >= 65 and code <= 90:\n" +
		"                result = result + chr(code + 32)\n" +
		"            else:\n" +
		"                result = result + char\n" +
		"        return result"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should have 1 statement (the grim declaration)
	require.Len(t, program.Statements, 1, "program should have 1 statement")

	// Check it's a class statement
	stmt, ok := program.Statements[0].(*ast.ClassStatement)
	require.True(t, ok, "program.Statements[0] is not ast.ClassStatement")
	assert.Equal(t, "String", stmt.Name.Value)
	
	// Should have methods defined in the body
	require.NotNil(t, stmt.Body, "Class body should not be nil")
	require.GreaterOrEqual(t, len(stmt.Body.Statements), 3, "Class should have at least 3 statements (init + 2 spells)")
}