package analyzer

import (
	"testing"

	"github.com/javanhut/carrion-lsp/internal/carrion/lexer"
	"github.com/javanhut/carrion-lsp/internal/carrion/parser"
	"github.com/javanhut/carrion-lsp/internal/carrion/symbol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create analyzer from Carrion code
func createAnalyzer(input string) (*Analyzer, error) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	analyzer := New()
	err := analyzer.Analyze(program)
	return analyzer, err
}

func TestAnalyzer_VariableAssignment(t *testing.T) {
	input := `
x = 42
y = "hello"
z = True
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Check that variables are defined
	xSymbol, exists := analyzer.SymbolTable.Lookup("x")
	assert.True(t, exists)
	assert.Equal(t, symbol.VariableSymbol, xSymbol.Type)
	assert.Equal(t, "int", xSymbol.DataType)

	ySymbol, exists := analyzer.SymbolTable.Lookup("y")
	assert.True(t, exists)
	assert.Equal(t, symbol.VariableSymbol, ySymbol.Type)
	assert.Equal(t, "str", ySymbol.DataType)

	zSymbol, exists := analyzer.SymbolTable.Lookup("z")
	assert.True(t, exists)
	assert.Equal(t, symbol.VariableSymbol, zSymbol.Type)
	assert.Equal(t, "bool", zSymbol.DataType)
}

func TestAnalyzer_FunctionDefinition(t *testing.T) {
	input := `
spell add(x, y):
    return x + y
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Check that function is defined
	funcSymbol, exists := analyzer.SymbolTable.Lookup("add")
	assert.True(t, exists)
	assert.Equal(t, symbol.FunctionSymbol, funcSymbol.Type)
	assert.Equal(t, "function", funcSymbol.DataType)

	// Check parameters
	assert.Len(t, funcSymbol.Parameters, 2)
	assert.Equal(t, "x", funcSymbol.Parameters[0].Name)
	assert.Equal(t, "y", funcSymbol.Parameters[1].Name)
	assert.Equal(t, symbol.ParameterSymbol, funcSymbol.Parameters[0].Type)
	assert.Equal(t, symbol.ParameterSymbol, funcSymbol.Parameters[1].Type)
}

func TestAnalyzer_ClassDefinition(t *testing.T) {
	input := `
grim Person:
    spell init(self, name):
        self.name = name
    
    spell greet(self):
        return "Hello, " + self.name
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Check that class is defined
	classSymbol, exists := analyzer.SymbolTable.Lookup("Person")
	assert.True(t, exists)
	assert.Equal(t, symbol.ClassSymbol, classSymbol.Type)
	assert.Equal(t, "class", classSymbol.DataType)

	// Check that methods are members of the class
	assert.Contains(t, classSymbol.Members, "init")
	assert.Contains(t, classSymbol.Members, "greet")

	initMethod := classSymbol.Members["init"]
	assert.Equal(t, symbol.FunctionSymbol, initMethod.Type)
	assert.Len(t, initMethod.Parameters, 2) // self, name
}

func TestAnalyzer_ClassInheritance(t *testing.T) {
	input := `
grim Animal:
    ignore

grim Dog(Animal):
    spell bark(self):
        return "Woof!"
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Check parent class
	animalSymbol, exists := analyzer.SymbolTable.Lookup("Animal")
	assert.True(t, exists)
	assert.Equal(t, symbol.ClassSymbol, animalSymbol.Type)

	// Check child class
	dogSymbol, exists := analyzer.SymbolTable.Lookup("Dog")
	assert.True(t, exists)
	assert.Equal(t, symbol.ClassSymbol, dogSymbol.Type)
	assert.Equal(t, animalSymbol, dogSymbol.Parent)
}

func TestAnalyzer_UndefinedVariable(t *testing.T) {
	input := `
x = undefined_var + 5
`

	analyzer, err := createAnalyzer(input)
	assert.Error(t, err)
	assert.True(t, len(analyzer.Errors) > 0)
	assert.Contains(t, analyzer.Errors[0], "undefined variable 'undefined_var'")
}

func TestAnalyzer_DuplicateDefinition(t *testing.T) {
	input := `
x = 5
x = 10
`

	analyzer, err := createAnalyzer(input)
	assert.Error(t, err)
	assert.Contains(t, analyzer.Errors[0], "symbol 'x' already defined")
}

func TestAnalyzer_FunctionScope(t *testing.T) {
	input := `
x = "global"

spell test():
    x = "local"
    y = 42
    return x + str(y)
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Global x should exist
	globalX, exists := analyzer.SymbolTable.GlobalScope.LookupLocal("x")
	assert.True(t, exists)
	assert.Equal(t, "str", globalX.DataType)

	// Function should exist
	funcSymbol, exists := analyzer.SymbolTable.Lookup("test")
	assert.True(t, exists)
	assert.Equal(t, symbol.FunctionSymbol, funcSymbol.Type)
}

func TestAnalyzer_ReturnOutsideFunction(t *testing.T) {
	input := `
x = 5
return x
`

	analyzer, err := createAnalyzer(input)
	assert.Error(t, err)
	assert.True(t, len(analyzer.Errors) > 0)
	assert.Contains(t, analyzer.Errors[0], "return statement outside function")
}

func TestAnalyzer_ForLoop(t *testing.T) {
	input := `
numbers = [1, 2, 3]
for num in numbers:
    print(num)
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// numbers variable should exist
	numbersSymbol, exists := analyzer.SymbolTable.Lookup("numbers")
	assert.True(t, exists)
	assert.Equal(t, "list", numbersSymbol.DataType)
}

func TestAnalyzer_ImportStatement(t *testing.T) {
	input := `
import os
import sys as system
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// os module should exist
	osSymbol, exists := analyzer.SymbolTable.Lookup("os")
	assert.True(t, exists)
	assert.Equal(t, symbol.ModuleSymbol, osSymbol.Type)

	// sys should be imported as system
	systemSymbol, exists := analyzer.SymbolTable.Lookup("system")
	assert.True(t, exists)
	assert.Equal(t, symbol.ModuleSymbol, systemSymbol.Type)

	// sys should not exist directly
	_, exists = analyzer.SymbolTable.Lookup("sys")
	assert.False(t, exists)
}

func TestAnalyzer_CallExpression(t *testing.T) {
	input := `
spell greet(name):
    return "Hello, " + name

result = greet("World")
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Function should exist
	greetSymbol, exists := analyzer.SymbolTable.Lookup("greet")
	assert.True(t, exists)
	assert.Equal(t, symbol.FunctionSymbol, greetSymbol.Type)

	// Result variable should exist
	resultSymbol, exists := analyzer.SymbolTable.Lookup("result")
	assert.True(t, exists)
	assert.Equal(t, symbol.VariableSymbol, resultSymbol.Type)
}

func TestAnalyzer_CallNonFunction(t *testing.T) {
	input := `
x = 42
result = x()
`

	analyzer, err := createAnalyzer(input)
	assert.Error(t, err)
	assert.True(t, len(analyzer.Errors) > 0)
	assert.Contains(t, analyzer.Errors[0], "'x' is not callable")
}

func TestAnalyzer_BuiltinFunctions(t *testing.T) {
	input := `
length = len("hello")
text = str(42)
number = int("123")
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Variables should be defined
	_, exists := analyzer.SymbolTable.Lookup("length")
	assert.True(t, exists)

	_, exists = analyzer.SymbolTable.Lookup("text")
	assert.True(t, exists)

	_, exists = analyzer.SymbolTable.Lookup("number")
	assert.True(t, exists)

	// Built-in functions should be accessible
	lenSymbol, exists := analyzer.SymbolTable.Lookup("len")
	assert.True(t, exists)
	assert.Equal(t, symbol.BuiltinSymbol, lenSymbol.Type)
}

func TestAnalyzer_TypeInference(t *testing.T) {
	tests := []struct {
		input    string
		varName  string
		expected string
	}{
		{"x = 42", "x", "int"},
		{"x = 3.14", "x", "float"},
		{"x = 'hello'", "x", "str"},
		{"x = True", "x", "bool"},
		{"x = False", "x", "bool"},
		{"x = None", "x", "NoneType"},
		{"x = [1, 2, 3]", "x", "list"},
		{"x = {'a': 1}", "x", "dict"},
		{"x = 5 + 3", "x", "int"},
		{"x = 5.0 + 3", "x", "float"},
		{"x = 'hello' + ' world'", "x", "str"},
		{"x = 5 > 3", "x", "bool"},
		{"x = 5 == 3", "x", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			analyzer, err := createAnalyzer(tt.input)
			require.NoError(t, err)

			symbol, exists := analyzer.SymbolTable.Lookup(tt.varName)
			assert.True(t, exists)
			assert.Equal(t, tt.expected, symbol.DataType, "Expected type %s for %s, got %s", tt.expected, tt.input, symbol.DataType)
		})
	}
}

func TestAnalyzer_ComplexProgram(t *testing.T) {
	input := `
counter = 0
name = "CarrionLSP"

grim Calculator:
    spell init(self):
        self.value = 0
    
    spell add(self, x):
        self.value = self.value + x
        return self.value
    
    spell multiply(self, x):
        self.value = self.value * x
        return self.value

spell fibonacci(n):
    if n <= 1:
        return n
    else:
        return fibonacci(n - 1) + fibonacci(n - 2)

calc = Calculator()
result = calc.add(10)
fib_result = fibonacci(5)

numbers = [1, 2, 3, 4, 5]
for num in numbers:
    counter = counter + num

if counter > 10:
    print("Counter is large: " + str(counter))
else:
    print("Counter is small")
`

	analyzer, err := createAnalyzer(input)
	if err != nil {
		t.Logf("Analysis errors: %v", analyzer.Errors)
	}
	require.NoError(t, err)

	// Check global variables
	counterSymbol, exists := analyzer.SymbolTable.Lookup("counter")
	assert.True(t, exists)
	assert.Equal(t, "int", counterSymbol.DataType)

	nameSymbol, exists := analyzer.SymbolTable.Lookup("name")
	assert.True(t, exists)
	assert.Equal(t, "str", nameSymbol.DataType)

	// Check class
	calcClassSymbol, exists := analyzer.SymbolTable.Lookup("Calculator")
	assert.True(t, exists)
	assert.Equal(t, symbol.ClassSymbol, calcClassSymbol.Type)
	assert.Contains(t, calcClassSymbol.Members, "init")
	assert.Contains(t, calcClassSymbol.Members, "add")
	assert.Contains(t, calcClassSymbol.Members, "multiply")

	// Check function
	fibSymbol, exists := analyzer.SymbolTable.Lookup("fibonacci")
	assert.True(t, exists)
	assert.Equal(t, symbol.FunctionSymbol, fibSymbol.Type)
	assert.Len(t, fibSymbol.Parameters, 1)
	assert.Equal(t, "n", fibSymbol.Parameters[0].Name)

	// Check variables created from function calls
	_, exists = analyzer.SymbolTable.Lookup("calc")
	assert.True(t, exists)

	_, exists = analyzer.SymbolTable.Lookup("result")
	assert.True(t, exists)

	_, exists = analyzer.SymbolTable.Lookup("fib_result")
	assert.True(t, exists)

	numbersSymbol, exists := analyzer.SymbolTable.Lookup("numbers")
	assert.True(t, exists)
	assert.Equal(t, "list", numbersSymbol.DataType)
}

func TestAnalyzer_GetCompletionItems(t *testing.T) {
	input := `
x = 42
y = "hello"

spell test_function():
    z = 3.14
    return z
`

	analyzer, err := createAnalyzer(input)
	require.NoError(t, err)

	// Get completion items with empty prefix (should return all symbols)
	items := analyzer.GetCompletionItems(1, 1, "")

	// Should contain global variables, functions, and built-ins
	itemNames := make([]string, len(items))
	for i, item := range items {
		itemNames[i] = item.Name
	}

	assert.Contains(t, itemNames, "x")
	assert.Contains(t, itemNames, "y")
	assert.Contains(t, itemNames, "test_function")
	assert.Contains(t, itemNames, "print") // built-in

	// Get completion items with prefix
	testItems := analyzer.GetCompletionItems(1, 1, "test")
	assert.Len(t, testItems, 1)
	assert.Equal(t, "test_function", testItems[0].Name)
}

func TestAnalyzer_GetDiagnostics(t *testing.T) {
	input := `
x = undefined_var
y = 42
y = "redefined"
`

	analyzer, _ := createAnalyzer(input) // Expect error

	diagnostics := analyzer.GetDiagnostics()
	assert.True(t, len(diagnostics) >= 2) // At least undefined var and redefinition

	// Check that we have error diagnostics
	for _, diag := range diagnostics {
		assert.Equal(t, DiagnosticError, diag.Severity)
		assert.Equal(t, "carrion-analyzer", diag.Source)
	}
}
