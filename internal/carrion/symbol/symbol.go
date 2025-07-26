package symbol

import (
	"fmt"

	"github.com/javanhut/carrion-lsp/internal/carrion/ast"
	"github.com/javanhut/carrion-lsp/internal/carrion/token"
)

// SymbolType represents the type of symbol
type SymbolType string

const (
	VariableSymbol  SymbolType = "VARIABLE"
	FunctionSymbol  SymbolType = "FUNCTION"
	ClassSymbol     SymbolType = "CLASS"
	ParameterSymbol SymbolType = "PARAMETER"
	ModuleSymbol    SymbolType = "MODULE"
	BuiltinSymbol   SymbolType = "BUILTIN"
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name       string
	Type       SymbolType
	Scope      *Scope
	Node       ast.Node           // AST node where symbol is defined
	Token      token.Token        // Token for the symbol name
	DataType   string             // Inferred or declared type (e.g., "int", "str", "MyClass")
	Parameters []*Symbol          // For functions - their parameters
	ReturnType string             // For functions - return type
	Parent     *Symbol            // For classes - parent class
	Members    map[string]*Symbol // For classes - methods and attributes
}

// Position returns the line and column where this symbol is defined
func (s *Symbol) Position() (line, column int) {
	return s.Token.Line, s.Token.Column
}

// String returns a string representation of the symbol
func (s *Symbol) String() string {
	return fmt.Sprintf("Symbol{Name: %s, Type: %s, DataType: %s}", s.Name, s.Type, s.DataType)
}

// ScopeType represents the type of scope
type ScopeType string

const (
	GlobalScope   ScopeType = "GLOBAL"
	FunctionScope ScopeType = "FUNCTION"
	ClassScope    ScopeType = "CLASS"
	BlockScope    ScopeType = "BLOCK"
	ModuleScope   ScopeType = "MODULE"
)

// Scope represents a lexical scope
type Scope struct {
	Type     ScopeType
	Name     string             // Name of the scope (function name, class name, etc.)
	Parent   *Scope             // Parent scope
	Children []*Scope           // Child scopes
	Symbols  map[string]*Symbol // Symbols defined in this scope
	Node     ast.Node           // AST node this scope represents
}

// NewScope creates a new scope
func NewScope(scopeType ScopeType, name string, parent *Scope, node ast.Node) *Scope {
	scope := &Scope{
		Type:     scopeType,
		Name:     name,
		Parent:   parent,
		Children: []*Scope{},
		Symbols:  make(map[string]*Symbol),
		Node:     node,
	}

	if parent != nil {
		parent.Children = append(parent.Children, scope)
	}

	return scope
}

// Define adds a symbol to this scope
func (s *Scope) Define(symbol *Symbol) error {
	if existing, exists := s.Symbols[symbol.Name]; exists {
		return fmt.Errorf("symbol '%s' already defined at line %d", symbol.Name, existing.Token.Line)
	}

	symbol.Scope = s
	s.Symbols[symbol.Name] = symbol
	return nil
}

// Lookup finds a symbol in this scope or parent scopes
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	// Check current scope first
	if symbol, exists := s.Symbols[name]; exists {
		return symbol, true
	}

	// Check parent scopes
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}

	return nil, false
}

// LookupLocal finds a symbol only in this scope (not parent scopes)
func (s *Scope) LookupLocal(name string) (*Symbol, bool) {
	symbol, exists := s.Symbols[name]
	return symbol, exists
}

// GetAllSymbols returns all symbols accessible from this scope
func (s *Scope) GetAllSymbols() map[string]*Symbol {
	symbols := make(map[string]*Symbol)

	// Collect symbols from parent scopes first (can be overridden)
	if s.Parent != nil {
		parentSymbols := s.Parent.GetAllSymbols()
		for name, symbol := range parentSymbols {
			symbols[name] = symbol
		}
	}

	// Add symbols from current scope (override parent symbols)
	for name, symbol := range s.Symbols {
		symbols[name] = symbol
	}

	return symbols
}

// GetLocalSymbols returns only symbols defined in this scope
func (s *Scope) GetLocalSymbols() map[string]*Symbol {
	symbols := make(map[string]*Symbol)
	for name, symbol := range s.Symbols {
		symbols[name] = symbol
	}
	return symbols
}

// String returns a string representation of the scope
func (s *Scope) String() string {
	return fmt.Sprintf("Scope{Type: %s, Name: %s, Symbols: %d}", s.Type, s.Name, len(s.Symbols))
}

// SymbolTable represents the global symbol table for a program
type SymbolTable struct {
	GlobalScope  *Scope
	CurrentScope *Scope
	Builtins     map[string]*Symbol
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	globalScope := NewScope(GlobalScope, "global", nil, nil)

	st := &SymbolTable{
		GlobalScope:  globalScope,
		CurrentScope: globalScope,
		Builtins:     make(map[string]*Symbol),
	}

	// Add built-in symbols
	st.addBuiltins()

	return st
}

// addBuiltins adds built-in functions and types to the symbol table
func (st *SymbolTable) addBuiltins() {
	builtins := []struct {
		name       string
		symbolType SymbolType
		dataType   string
	}{
		// Built-in functions
		{"print", FunctionSymbol, "function"},
		{"len", FunctionSymbol, "function"},
		{"str", FunctionSymbol, "function"},
		{"int", FunctionSymbol, "function"},
		{"float", FunctionSymbol, "function"},
		{"bool", FunctionSymbol, "function"},
		{"list", FunctionSymbol, "function"},
		{"dict", FunctionSymbol, "function"},
		{"range", FunctionSymbol, "function"},
		{"enumerate", FunctionSymbol, "function"},
		{"zip", FunctionSymbol, "function"},
		{"map", FunctionSymbol, "function"},
		{"filter", FunctionSymbol, "function"},
		{"sorted", FunctionSymbol, "function"},
		{"sum", FunctionSymbol, "function"},
		{"min", FunctionSymbol, "function"},
		{"max", FunctionSymbol, "function"},
		{"abs", FunctionSymbol, "function"},
		{"round", FunctionSymbol, "function"},
		{"open", FunctionSymbol, "function"},

		// Built-in types/constants
		{"True", VariableSymbol, "bool"},
		{"False", VariableSymbol, "bool"},
		{"None", VariableSymbol, "NoneType"},
	}

	for _, builtin := range builtins {
		symbol := &Symbol{
			Name:     builtin.name,
			Type:     BuiltinSymbol,
			DataType: builtin.dataType,
			Token:    token.Token{Type: token.IDENT, Literal: builtin.name, Line: 0, Column: 0},
		}
		st.Builtins[builtin.name] = symbol
		st.GlobalScope.Symbols[builtin.name] = symbol
	}
}

// EnterScope creates and enters a new scope
func (st *SymbolTable) EnterScope(scopeType ScopeType, name string, node ast.Node) *Scope {
	newScope := NewScope(scopeType, name, st.CurrentScope, node)
	st.CurrentScope = newScope
	return newScope
}

// ExitScope returns to the parent scope
func (st *SymbolTable) ExitScope() *Scope {
	if st.CurrentScope.Parent != nil {
		st.CurrentScope = st.CurrentScope.Parent
	}
	return st.CurrentScope
}

// Define adds a symbol to the current scope
func (st *SymbolTable) Define(name string, symbolType SymbolType, node ast.Node, tok token.Token) (*Symbol, error) {
	symbol := &Symbol{
		Name:     name,
		Type:     symbolType,
		Node:     node,
		Token:    tok,
		DataType: st.inferDataType(node, symbolType),
		Members:  make(map[string]*Symbol),
	}

	err := st.CurrentScope.Define(symbol)
	if err != nil {
		return nil, err
	}

	return symbol, nil
}

// Lookup finds a symbol in the current scope or parent scopes
func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
	return st.CurrentScope.Lookup(name)
}

// LookupInScope finds a symbol in a specific scope
func (st *SymbolTable) LookupInScope(name string, scope *Scope) (*Symbol, bool) {
	return scope.Lookup(name)
}

// GetCurrentScopeSymbols returns all symbols in the current scope
func (st *SymbolTable) GetCurrentScopeSymbols() map[string]*Symbol {
	return st.CurrentScope.GetLocalSymbols()
}

// GetAllAccessibleSymbols returns all symbols accessible from current scope
func (st *SymbolTable) GetAllAccessibleSymbols() map[string]*Symbol {
	return st.CurrentScope.GetAllSymbols()
}

// GetAllSymbols returns all symbols accessible from the global scope
func (st *SymbolTable) GetAllSymbols() map[string]*Symbol {
	return st.GlobalScope.GetAllSymbols()
}

// FindScopeAtPosition finds the most specific scope that contains the given position
func (st *SymbolTable) FindScopeAtPosition(line, column int) *Scope {
	return st.findScopeAtPositionRecursive(st.GlobalScope, line, column)
}

// findScopeAtPositionRecursive recursively searches for the scope at position
func (st *SymbolTable) findScopeAtPositionRecursive(scope *Scope, line, column int) *Scope {
	// Check if this scope contains the position
	if scope.Node != nil {
		scopeLine, scopeColumn := scope.Node.Position()
		if line < scopeLine || (line == scopeLine && column < scopeColumn) {
			return scope.Parent // Position is before this scope
		}
	}

	// Check child scopes for a more specific match
	for _, child := range scope.Children {
		if childScope := st.findScopeAtPositionRecursive(child, line, column); childScope != nil {
			return childScope
		}
	}

	return scope
}

// inferDataType attempts to infer the data type of a symbol from its AST node
func (st *SymbolTable) inferDataType(node ast.Node, symbolType SymbolType) string {
	switch symbolType {
	case FunctionSymbol:
		return "function"
	case ClassSymbol:
		return "class"
	case ParameterSymbol:
		return "unknown" // Would need type annotations or analysis
	case ModuleSymbol:
		return "module"
	case VariableSymbol:
		// Try to infer from assignment value
		return st.inferTypeFromExpression(node)
	}
	return "unknown"
}

// inferTypeFromExpression attempts to infer type from an expression
func (st *SymbolTable) inferTypeFromExpression(node ast.Node) string {
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		return "int"
	case *ast.FloatLiteral:
		return "float"
	case *ast.StringLiteral:
		return "str"
	case *ast.FStringLiteral:
		return "str"
	case *ast.BooleanLiteral:
		return "bool"
	case *ast.NoneLiteral:
		return "NoneType"
	case *ast.ArrayLiteral:
		return "list"
	case *ast.HashLiteral:
		return "dict"
	case *ast.Identifier:
		// Look up the identifier's type
		if symbol, exists := st.Lookup(n.Value); exists {
			return symbol.DataType
		}
		return "unknown"
	case *ast.CallExpression:
		// Try to infer return type from function
		if ident, ok := n.Function.(*ast.Identifier); ok {
			if symbol, exists := st.Lookup(ident.Value); exists && symbol.ReturnType != "" {
				return symbol.ReturnType
			}
		}
		return "unknown"
	case *ast.InfixExpression:
		// Basic arithmetic/comparison inference
		switch n.Operator {
		case "+", "-", "*", "/", "//", "%", "**":
			leftType := st.inferTypeFromExpression(n.Left)
			rightType := st.inferTypeFromExpression(n.Right)
			if leftType == "int" && rightType == "int" {
				if n.Operator == "/" {
					return "float"
				}
				return "int"
			}
			if (leftType == "float" || rightType == "float") &&
				(n.Operator == "+" || n.Operator == "-" || n.Operator == "*" || n.Operator == "/") {
				return "float"
			}
			if (leftType == "str" || rightType == "str") && n.Operator == "+" {
				return "str"
			}
			return "unknown"
		case "==", "!=", "<", ">", "<=", ">=", "and", "or", "not", "in", "not in", "is", "is not":
			return "bool"
		}
		return "unknown"
	}
	return "unknown"
}

// GetSymbolsOfType returns all symbols of a specific type in the current scope
func (st *SymbolTable) GetSymbolsOfType(symbolType SymbolType) []*Symbol {
	var symbols []*Symbol
	allSymbols := st.GetAllAccessibleSymbols()

	for _, symbol := range allSymbols {
		if symbol.Type == symbolType {
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

// GetSymbolsByPrefix returns all symbols starting with the given prefix
func (st *SymbolTable) GetSymbolsByPrefix(prefix string) []*Symbol {
	var symbols []*Symbol
	allSymbols := st.GetAllAccessibleSymbols()

	for name, symbol := range allSymbols {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}
