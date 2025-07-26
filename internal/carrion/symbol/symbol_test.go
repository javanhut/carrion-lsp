package symbol

import (
	"testing"

	"github.com/javanhut/carrion-lsp/internal/carrion/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()

	assert.NotNil(t, st.GlobalScope)
	assert.Equal(t, st.GlobalScope, st.CurrentScope)
	assert.Equal(t, GlobalScope, st.GlobalScope.Type)

	// Check that built-ins are present
	assert.True(t, len(st.Builtins) > 0)

	// Check some specific built-ins
	_, exists := st.Lookup("print")
	assert.True(t, exists)

	_, exists = st.Lookup("len")
	assert.True(t, exists)

	_, exists = st.Lookup("True")
	assert.True(t, exists)
}

func TestScope_DefineAndLookup(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable
	symbol, err := st.Define("x", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "x", Line: 1, Column: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, "x", symbol.Name)
	assert.Equal(t, VariableSymbol, symbol.Type)

	// Look up the variable
	found, exists := st.Lookup("x")
	assert.True(t, exists)
	assert.Equal(t, symbol, found)

	// Try to define the same variable again
	_, err = st.Define("x", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "x", Line: 2, Column: 1,
	})
	assert.Error(t, err)
}

func TestScope_Hierarchy(t *testing.T) {
	st := NewSymbolTable()

	// Define variable in global scope
	globalVar, err := st.Define("global_var", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "global_var", Line: 1, Column: 1,
	})
	require.NoError(t, err)

	// Enter function scope
	funcScope := st.EnterScope(FunctionScope, "test_func", nil)
	assert.Equal(t, funcScope, st.CurrentScope)
	assert.Equal(t, st.GlobalScope, funcScope.Parent)

	// Define variable in function scope
	localVar, err := st.Define("local_var", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "local_var", Line: 2, Column: 1,
	})
	require.NoError(t, err)

	// Should be able to access both global and local variables
	found, exists := st.Lookup("global_var")
	assert.True(t, exists)
	assert.Equal(t, globalVar, found)

	found, exists = st.Lookup("local_var")
	assert.True(t, exists)
	assert.Equal(t, localVar, found)

	// Exit function scope
	st.ExitScope()
	assert.Equal(t, st.GlobalScope, st.CurrentScope)

	// Should still access global variable
	found, exists = st.Lookup("global_var")
	assert.True(t, exists)
	assert.Equal(t, globalVar, found)

	// Should not access local variable
	_, exists = st.Lookup("local_var")
	assert.False(t, exists)
}

func TestScope_GetAllSymbols(t *testing.T) {
	st := NewSymbolTable()

	// Define global variable
	globalVar, err := st.Define("global_var", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "global_var", Line: 1, Column: 1,
	})
	require.NoError(t, err)

	// Enter function scope
	st.EnterScope(FunctionScope, "test_func", nil)

	// Define local variable
	localVar, err := st.Define("local_var", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "local_var", Line: 2, Column: 1,
	})
	require.NoError(t, err)

	// Get all accessible symbols
	allSymbols := st.GetAllAccessibleSymbols()

	// Should contain both global and local variables, plus built-ins
	assert.Contains(t, allSymbols, "global_var")
	assert.Contains(t, allSymbols, "local_var")
	assert.Contains(t, allSymbols, "print") // built-in

	assert.Equal(t, globalVar, allSymbols["global_var"])
	assert.Equal(t, localVar, allSymbols["local_var"])

	// Get only local symbols
	localSymbols := st.GetCurrentScopeSymbols()
	assert.Contains(t, localSymbols, "local_var")
	assert.NotContains(t, localSymbols, "global_var")
	assert.NotContains(t, localSymbols, "print")
}

func TestScope_Shadowing(t *testing.T) {
	st := NewSymbolTable()

	// Define variable in global scope
	globalVar, err := st.Define("x", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "x", Line: 1, Column: 1,
	})
	require.NoError(t, err)

	// Enter function scope
	st.EnterScope(FunctionScope, "test_func", nil)

	// Define variable with same name in function scope (shadowing)
	localVar, err := st.Define("x", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "x", Line: 2, Column: 1,
	})
	require.NoError(t, err)

	// Should find the local variable (shadows global)
	found, exists := st.Lookup("x")
	assert.True(t, exists)
	assert.Equal(t, localVar, found)
	assert.NotEqual(t, globalVar, found)

	// Exit function scope
	st.ExitScope()

	// Should now find the global variable
	found, exists = st.Lookup("x")
	assert.True(t, exists)
	assert.Equal(t, globalVar, found)
}

func TestSymbolTable_GetSymbolsOfType(t *testing.T) {
	st := NewSymbolTable()

	// Define various symbols
	st.Define("var1", VariableSymbol, nil, token.Token{Type: token.IDENT, Literal: "var1", Line: 1, Column: 1})
	st.Define("var2", VariableSymbol, nil, token.Token{Type: token.IDENT, Literal: "var2", Line: 2, Column: 1})
	st.Define("func1", FunctionSymbol, nil, token.Token{Type: token.IDENT, Literal: "func1", Line: 3, Column: 1})
	st.Define("class1", ClassSymbol, nil, token.Token{Type: token.IDENT, Literal: "class1", Line: 4, Column: 1})

	// Get variables
	variables := st.GetSymbolsOfType(VariableSymbol)
	varNames := make([]string, len(variables))
	for i, v := range variables {
		varNames[i] = v.Name
	}
	assert.Contains(t, varNames, "var1")
	assert.Contains(t, varNames, "var2")

	// Get functions (should include built-ins)
	functions := st.GetSymbolsOfType(FunctionSymbol)
	funcNames := make([]string, len(functions))
	for i, f := range functions {
		funcNames[i] = f.Name
	}
	assert.Contains(t, funcNames, "func1")

	// Get classes
	classes := st.GetSymbolsOfType(ClassSymbol)
	assert.Len(t, classes, 1)
	assert.Equal(t, "class1", classes[0].Name)
}

func TestSymbolTable_GetSymbolsByPrefix(t *testing.T) {
	st := NewSymbolTable()

	// Define symbols with various prefixes
	st.Define("test_var", VariableSymbol, nil, token.Token{Type: token.IDENT, Literal: "test_var", Line: 1, Column: 1})
	st.Define("test_func", FunctionSymbol, nil, token.Token{Type: token.IDENT, Literal: "test_func", Line: 2, Column: 1})
	st.Define("other_var", VariableSymbol, nil, token.Token{Type: token.IDENT, Literal: "other_var", Line: 3, Column: 1})

	// Get symbols with "test" prefix
	testSymbols := st.GetSymbolsByPrefix("test")
	assert.Len(t, testSymbols, 2)

	names := make([]string, len(testSymbols))
	for i, s := range testSymbols {
		names[i] = s.Name
	}
	assert.Contains(t, names, "test_var")
	assert.Contains(t, names, "test_func")

	// Get symbols with "other" prefix
	otherSymbols := st.GetSymbolsByPrefix("other")
	assert.Len(t, otherSymbols, 1)
	assert.Equal(t, "other_var", otherSymbols[0].Name)

	// Get symbols with non-existent prefix
	noneSymbols := st.GetSymbolsByPrefix("nonexistent")
	assert.Len(t, noneSymbols, 0)
}

func TestSymbol_Position(t *testing.T) {
	st := NewSymbolTable()

	symbol, err := st.Define("test", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "test", Line: 10, Column: 5,
	})
	require.NoError(t, err)

	line, column := symbol.Position()
	assert.Equal(t, 10, line)
	assert.Equal(t, 5, column)
}

func TestScope_LookupLocal(t *testing.T) {
	st := NewSymbolTable()

	// Define global variable
	st.Define("global_var", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "global_var", Line: 1, Column: 1,
	})

	// Enter function scope
	st.EnterScope(FunctionScope, "test_func", nil)

	// Define local variable
	localVar, err := st.Define("local_var", VariableSymbol, nil, token.Token{
		Type: token.IDENT, Literal: "local_var", Line: 2, Column: 1,
	})
	require.NoError(t, err)

	// LookupLocal should find local variable
	found, exists := st.CurrentScope.LookupLocal("local_var")
	assert.True(t, exists)
	assert.Equal(t, localVar, found)

	// LookupLocal should NOT find global variable
	_, exists = st.CurrentScope.LookupLocal("global_var")
	assert.False(t, exists)

	// Regular Lookup should find both
	_, exists = st.Lookup("local_var")
	assert.True(t, exists)

	_, exists = st.Lookup("global_var")
	assert.True(t, exists)
}
