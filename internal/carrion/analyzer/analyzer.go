package analyzer

import (
	"fmt"
	"strings"

	"github.com/javanhut/carrion-lsp/internal/carrion/ast"
	"github.com/javanhut/carrion-lsp/internal/carrion/symbol"
	"github.com/javanhut/carrion-lsp/internal/carrion/token"
)

// Analyzer performs semantic analysis and builds symbol tables
type Analyzer struct {
	SymbolTable *symbol.SymbolTable
	Errors      []string
	Diagnostics []Diagnostic
	References  map[string][]ReferenceLocation // Maps symbol names to their reference locations
}

// New creates a new analyzer
func New() *Analyzer {
	analyzer := &Analyzer{
		SymbolTable: symbol.NewSymbolTable(),
		Errors:      []string{},
		Diagnostics: []Diagnostic{},
		References:  make(map[string][]ReferenceLocation),
	}
	
	// Initialize built-in symbols
	analyzer.initializeBuiltins()
	
	return analyzer
}

// initializeBuiltins defines built-in functions and modules
func (a *Analyzer) initializeBuiltins() {
	// Built-in functions
	builtinFunctions := []string{
		"print", "len", "type", "str", "int", "float", "bool", "list", "dict",
		"range", "enumerate", "zip", "map", "filter", "sorted", "reversed",
		"min", "max", "sum", "any", "all", "abs", "round", "pow", "ord", "chr",
		"input", "open", "exit", "help",
	}
	
	for _, name := range builtinFunctions {
		a.SymbolTable.Define(
			name,
			symbol.BuiltinSymbol,
			nil, // No AST node for built-ins
			token.Token{Type: token.IDENT, Literal: name, Line: 0, Column: 0},
		)
	}
	
	// Built-in modules/classes with their common methods
	builtinModules := map[string][]string{
		"os": {"cwd", "listdir", "mkdir", "rmdir", "remove", "rename", "getcwd", "chdir", "getenv", "setenv"},
		"sys": {"argv", "exit", "version", "platform", "path"},
		"time": {"time", "sleep", "strftime", "strptime", "clock"},
		"math": {"sin", "cos", "tan", "sqrt", "pow", "floor", "ceil", "abs"},
		"random": {"random", "randint", "choice", "shuffle", "seed"},
		"json": {"loads", "dumps", "load", "dump"},
		"re": {"match", "search", "findall", "sub", "split"},
		"http": {"get", "post", "put", "delete", "request"},
		"file": {"open", "read", "write", "close", "exists"},
		"socket": {"socket", "bind", "listen", "accept", "connect", "send", "recv"},
	}
	
	for moduleName, methods := range builtinModules {
		moduleSymbol, _ := a.SymbolTable.Define(
			moduleName,
			symbol.ModuleSymbol,
			nil, // No AST node for built-ins
			token.Token{Type: token.IDENT, Literal: moduleName, Line: 0, Column: 0},
		)
		
		// Add methods to the module
		for _, methodName := range methods {
			methodSymbol := &symbol.Symbol{
				Name:     methodName,
				Type:     symbol.FunctionSymbol,
				Node:     nil,
				Token:    token.Token{Type: token.IDENT, Literal: methodName, Line: 0, Column: 0},
				DataType: "function",
				Members:  make(map[string]*symbol.Symbol),
			}
			moduleSymbol.Members[methodName] = methodSymbol
		}
	}
}

// Analyze performs semantic analysis on an AST program
func (a *Analyzer) Analyze(program *ast.Program) error {
	// Reset state
	a.Errors = []string{}
	a.Diagnostics = []Diagnostic{}
	a.References = make(map[string][]ReferenceLocation)

	// Analyze all statements
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}

	// Add parser errors to analyzer errors
	for _, err := range program.Errors {
		a.addError(err)
	}

	if len(a.Errors) > 0 {
		return fmt.Errorf("analysis failed with %d errors", len(a.Errors))
	}

	return nil
}

// analyzeStatement analyzes a statement and updates the symbol table
func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch node := stmt.(type) {
	case *ast.AssignStatement:
		a.analyzeAssignStatement(node)
	case *ast.MemberAssignStatement:
		a.analyzeMemberAssignStatement(node)
	case *ast.FunctionStatement:
		a.analyzeFunctionStatement(node)
	case *ast.ClassStatement:
		a.analyzeClassStatement(node)
	case *ast.ImportStatement:
		a.analyzeImportStatement(node)
	case *ast.ReturnStatement:
		a.analyzeReturnStatement(node)
	case *ast.IfStatement:
		a.analyzeIfStatement(node)
	case *ast.WhileStatement:
		a.analyzeWhileStatement(node)
	case *ast.ForStatement:
		a.analyzeForStatement(node)
	case *ast.ExpressionStatement:
		a.analyzeExpression(node.Expression)
	case *ast.BlockStatement:
		a.analyzeBlockStatement(node)
	case *ast.IgnoreStatement:
		// No analysis needed for ignore statements
	}
}

// analyzeAssignStatement analyzes variable assignments
func (a *Analyzer) analyzeAssignStatement(node *ast.AssignStatement) {
	// Analyze the value expression first
	a.analyzeExpression(node.Value)

	// Infer the type from the assignment value
	varType := a.inferTypeFromAssignment(node.Value)

	// Define the variable in current scope
	varSymbol, err := a.SymbolTable.Define(
		node.Name.Value,
		symbol.VariableSymbol,
		node.Value, // Use the value node for type inference
		node.Name.Token,
	)

	if err != nil {
		// Check if this is trying to shadow a built-in - that's okay
		if existingSym, exists := a.SymbolTable.Lookup(node.Name.Value); exists && 
		   (existingSym.Type == symbol.BuiltinSymbol || existingSym.Type == symbol.ModuleSymbol) &&
		   existingSym.Token.Line == 0 { // Built-ins have line 0
			// Allow shadowing built-ins - force define in current scope
			scope := a.SymbolTable.CurrentScope
			varSymbol = &symbol.Symbol{
				Name:     node.Name.Value,
				Type:     symbol.VariableSymbol,
				Node:     node.Value,
				Token:    node.Name.Token,
				DataType: varType,
				Members:  make(map[string]*symbol.Symbol),
			}
			scope.Symbols[node.Name.Value] = varSymbol
		} else {
			a.addError(fmt.Sprintf("line %d: %s", node.Token.Line, err.Error()))
			a.addDiagnostic(node.Name.Token, err.Error(), DiagnosticError)
		}
	} else if varSymbol != nil {
		// Set the inferred type
		varSymbol.DataType = varType
	}
}

// analyzeMemberAssignStatement analyzes member assignment statements (obj.member = value)
func (a *Analyzer) analyzeMemberAssignStatement(node *ast.MemberAssignStatement) {
	// Analyze the object and value expressions
	a.analyzeExpression(node.Object)
	a.analyzeExpression(node.Value)

	// Note: We don't track object member assignments in the symbol table currently
	// This would require more sophisticated object tracking
}

// analyzeFunctionStatement analyzes function definitions
func (a *Analyzer) analyzeFunctionStatement(node *ast.FunctionStatement) {
	if node == nil {
		return
	}

	// Define function in current scope
	funcSymbol, err := a.SymbolTable.Define(
		node.Name.Value,
		symbol.FunctionSymbol,
		node,
		node.Name.Token,
	)

	if err != nil {
		a.addError(fmt.Sprintf("line %d: %s", node.Token.Line, err.Error()))
		a.addDiagnostic(node.Name.Token, err.Error(), DiagnosticError)
		return
	}

	// Enter function scope
	funcScope := a.SymbolTable.EnterScope(symbol.FunctionScope, node.Name.Value, node)

	// Add parameters to function scope
	var paramSymbols []*symbol.Symbol
	for _, param := range node.Parameters {
		paramSymbol, err := a.SymbolTable.Define(
			param.Value,
			symbol.ParameterSymbol,
			param,
			param.Token,
		)

		if err != nil {
			a.addError(fmt.Sprintf("line %d: %s", param.Token.Line, err.Error()))
			a.addDiagnostic(param.Token, err.Error(), DiagnosticError)
		} else {
			paramSymbols = append(paramSymbols, paramSymbol)
		}
	}

	// Store parameters in function symbol
	funcSymbol.Parameters = paramSymbols

	// Analyze function body
	a.analyzeBlockStatement(node.Body)

	// Infer return type from return statements
	a.inferFunctionReturnType(funcSymbol, funcScope)

	// Exit function scope
	a.SymbolTable.ExitScope()
}

// analyzeClassStatement analyzes class definitions
func (a *Analyzer) analyzeClassStatement(node *ast.ClassStatement) {
	// Define class in current scope
	classSymbol, err := a.SymbolTable.Define(
		node.Name.Value,
		symbol.ClassSymbol,
		node,
		node.Name.Token,
	)

	if err != nil {
		a.addError(fmt.Sprintf("line %d: %s", node.Token.Line, err.Error()))
		a.addDiagnostic(node.Name.Token, err.Error(), DiagnosticError)
		return
	}

	// Handle inheritance
	if node.Parent != nil {
		if parentSymbol, exists := a.SymbolTable.Lookup(node.Parent.Value); exists {
			if parentSymbol.Type != symbol.ClassSymbol {
				a.addError(fmt.Sprintf("line %d: '%s' is not a class", node.Parent.Token.Line, node.Parent.Value))
				a.addDiagnostic(node.Parent.Token, fmt.Sprintf("'%s' is not a class", node.Parent.Value), DiagnosticError)
			} else {
				classSymbol.Parent = parentSymbol
			}
		} else {
			a.addError(fmt.Sprintf("line %d: undefined class '%s'", node.Parent.Token.Line, node.Parent.Value))
			a.addDiagnostic(node.Parent.Token, fmt.Sprintf("undefined class '%s'", node.Parent.Value), DiagnosticError)
		}
	}

	// Enter class scope
	a.SymbolTable.EnterScope(symbol.ClassScope, node.Name.Value, node)

	// Add 'self' parameter to class scope
	selfSymbol, _ := a.SymbolTable.Define(
		"self",
		symbol.ParameterSymbol,
		node,
		token.Token{Type: token.SELF, Literal: "self", Line: node.Token.Line, Column: node.Token.Column},
	)
	selfSymbol.DataType = node.Name.Value

	// Analyze class body
	if node.Body != nil {
		a.analyzeBlockStatement(node.Body)
	}

	// Collect methods for the class symbol
	for name, sym := range a.SymbolTable.CurrentScope.Symbols {
		if sym.Type == symbol.FunctionSymbol {
			classSymbol.Members[name] = sym
		}
	}

	// Exit class scope
	a.SymbolTable.ExitScope()
}

// analyzeImportStatement analyzes import statements
func (a *Analyzer) analyzeImportStatement(node *ast.ImportStatement) {
	moduleName := node.Module.Value
	if node.Alias != nil {
		moduleName = node.Alias.Value
	}

	// Define module in current scope
	_, err := a.SymbolTable.Define(
		moduleName,
		symbol.ModuleSymbol,
		node,
		node.Module.Token,
	)

	if err != nil {
		a.addError(fmt.Sprintf("line %d: %s", node.Token.Line, err.Error()))
		a.addDiagnostic(node.Module.Token, err.Error(), DiagnosticError)
	}
}

// analyzeReturnStatement analyzes return statements
func (a *Analyzer) analyzeReturnStatement(node *ast.ReturnStatement) {
	if node.ReturnValue != nil {
		a.analyzeExpression(node.ReturnValue)
	}

	// Check if we're in a function scope
	scope := a.SymbolTable.CurrentScope
	for scope != nil && scope.Type != symbol.FunctionScope {
		scope = scope.Parent
	}

	if scope == nil {
		a.addError(fmt.Sprintf("line %d: return statement outside function", node.Token.Line))
		a.addDiagnostic(node.Token, "return statement outside function", DiagnosticError)
	}
}

// analyzeIfStatement analyzes if statements
func (a *Analyzer) analyzeIfStatement(node *ast.IfStatement) {
	// Analyze condition
	a.analyzeExpression(node.Condition)

	// Analyze consequence block
	a.analyzeBlockStatement(node.Consequence)

	// Analyze alternative block if present
	if node.Alternative != nil {
		a.analyzeBlockStatement(node.Alternative)
	}
}

// analyzeWhileStatement analyzes while statements
func (a *Analyzer) analyzeWhileStatement(node *ast.WhileStatement) {
	// Analyze condition
	a.analyzeExpression(node.Condition)

	// Analyze body
	a.analyzeBlockStatement(node.Body)
}

// analyzeForStatement analyzes for statements
func (a *Analyzer) analyzeForStatement(node *ast.ForStatement) {
	// Enter block scope for the loop
	a.SymbolTable.EnterScope(symbol.BlockScope, "for-loop", node)

	// Define loop variable
	_, err := a.SymbolTable.Define(
		node.Variable.Value,
		symbol.VariableSymbol,
		node.Variable,
		node.Variable.Token,
	)

	if err != nil {
		a.addError(fmt.Sprintf("line %d: %s", node.Variable.Token.Line, err.Error()))
		a.addDiagnostic(node.Variable.Token, err.Error(), DiagnosticError)
	}

	// Analyze iterable expression
	a.analyzeExpression(node.Iterable)

	// Analyze loop body
	a.analyzeBlockStatement(node.Body)

	// Exit block scope
	a.SymbolTable.ExitScope()
}

// analyzeBlockStatement analyzes block statements
func (a *Analyzer) analyzeBlockStatement(node *ast.BlockStatement) {
	for _, stmt := range node.Statements {
		a.analyzeStatement(stmt)
	}
}

// analyzeExpression analyzes expressions and checks for undefined variables
func (a *Analyzer) analyzeExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch node := expr.(type) {
	case *ast.Identifier:
		a.analyzeIdentifier(node)
	case *ast.CallExpression:
		a.analyzeCallExpression(node)
	case *ast.IndexExpression:
		a.analyzeIndexExpression(node)
	case *ast.MemberExpression:
		a.analyzeMemberExpression(node)
	case *ast.InfixExpression:
		a.analyzeExpression(node.Left)
		a.analyzeExpression(node.Right)
	case *ast.PrefixExpression:
		a.analyzeExpression(node.Right)
	case *ast.ArrayLiteral:
		for _, elem := range node.Elements {
			a.analyzeExpression(elem)
		}
	case *ast.HashLiteral:
		for key, value := range node.Pairs {
			a.analyzeExpression(key)
			a.analyzeExpression(value)
		}
	// Literals don't need analysis
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.FStringLiteral, *ast.BooleanLiteral, *ast.NoneLiteral:
		// No analysis needed for literals
	}
}

// analyzeIdentifier checks if an identifier is defined
func (a *Analyzer) analyzeIdentifier(node *ast.Identifier) {
	if _, exists := a.SymbolTable.Lookup(node.Value); !exists {
		a.addError(fmt.Sprintf("line %d: undefined variable '%s'", node.Token.Line, node.Value))
		a.addDiagnostic(node.Token, fmt.Sprintf("undefined variable '%s'", node.Value), DiagnosticError)
	} else {
		// Record this as a reference to the symbol
		a.addReference(node.Value, node.Token)
	}
}

// analyzeCallExpression analyzes function calls
func (a *Analyzer) analyzeCallExpression(node *ast.CallExpression) {
	// Analyze function expression
	a.analyzeExpression(node.Function)

	// Analyze arguments
	for _, arg := range node.Arguments {
		a.analyzeExpression(arg)
	}

	// Check if function exists and is callable
	if ident, ok := node.Function.(*ast.Identifier); ok {
		if sym, exists := a.SymbolTable.Lookup(ident.Value); exists {
			if sym.Type != symbol.FunctionSymbol && sym.Type != symbol.BuiltinSymbol && sym.Type != symbol.ClassSymbol && sym.Type != symbol.ModuleSymbol {
				a.addError(fmt.Sprintf("line %d: '%s' is not callable", node.Token.Line, ident.Value))
				a.addDiagnostic(node.Token, fmt.Sprintf("'%s' is not callable", ident.Value), DiagnosticError)
			}
		}
	}
}

// analyzeIndexExpression analyzes array/dict indexing
func (a *Analyzer) analyzeIndexExpression(node *ast.IndexExpression) {
	a.analyzeExpression(node.Left)
	a.analyzeExpression(node.Index)
}

// analyzeMemberExpression analyzes member access (obj.member)
func (a *Analyzer) analyzeMemberExpression(node *ast.MemberExpression) {
	// Analyze the object being accessed
	a.analyzeExpression(node.Object)
	
	// Check if the object exists and has the requested member
	if ident, ok := node.Object.(*ast.Identifier); ok {
		if sym, exists := a.SymbolTable.Lookup(ident.Value); exists {
			// Check what type of object this is
			switch sym.Type {
			case symbol.ClassSymbol:
				// For class symbols, check if the member exists in the class
				if _, hasMember := sym.Members[node.Member.Value]; !hasMember {
					a.addError(fmt.Sprintf("line %d: class '%s' has no member '%s'", 
						node.Member.Token.Line, sym.Name, node.Member.Value))
					a.addDiagnostic(node.Member.Token, 
						fmt.Sprintf("class '%s' has no member '%s'", sym.Name, node.Member.Value), 
						DiagnosticError)
				}
			case symbol.VariableSymbol:
				// For variables, check if the variable's type has the member
				if sym.DataType != "" {
					// Look up the type (class or module) of this variable
					if typeSym, typeExists := a.SymbolTable.Lookup(sym.DataType); typeExists {
						if typeSym.Type == symbol.ClassSymbol || typeSym.Type == symbol.ModuleSymbol {
							if _, hasMember := typeSym.Members[node.Member.Value]; !hasMember {
								objectType := "object"
								if typeSym.Type == symbol.ModuleSymbol {
									objectType = "module instance"
								}
								a.addError(fmt.Sprintf("line %d: %s of type '%s' has no member '%s'", 
									node.Member.Token.Line, objectType, sym.DataType, node.Member.Value))
								a.addDiagnostic(node.Member.Token, 
									fmt.Sprintf("%s of type '%s' has no member '%s'", objectType, sym.DataType, node.Member.Value), 
									DiagnosticError)
							}
						}
					}
				}
			case symbol.ModuleSymbol:
				// For module symbols (static access), check module members
				if _, hasMember := sym.Members[node.Member.Value]; !hasMember {
					a.addError(fmt.Sprintf("line %d: module '%s' has no member '%s'", 
						node.Member.Token.Line, sym.Name, node.Member.Value))
					a.addDiagnostic(node.Member.Token, 
						fmt.Sprintf("module '%s' has no member '%s'", sym.Name, node.Member.Value), 
						DiagnosticError)
				}
			}
		}
	}
}

// inferFunctionReturnType infers the return type of a function from its return statements
func (a *Analyzer) inferFunctionReturnType(funcSymbol *symbol.Symbol, funcScope *symbol.Scope) {
	// This is a simplified implementation
	// In a more complete implementation, we would traverse the function body
	// looking for return statements and infer the common return type
	funcSymbol.ReturnType = "unknown"
}

// addError adds an error to the analyzer
func (a *Analyzer) addError(msg string) {
	a.Errors = append(a.Errors, msg)
}

// addDiagnostic adds a diagnostic with position information
func (a *Analyzer) addDiagnostic(tok token.Token, message string, severity DiagnosticSeverity) {
	diagnostic := Diagnostic{
		Range: Range{
			Start: Position{
				Line:      tok.Line - 1, // Convert 1-based to 0-based
				Character: tok.Column - 1,
			},
			End: Position{
				Line:      tok.Line - 1,
				Character: tok.Column - 1 + len(tok.Literal),
			},
		},
		Message:  message,
		Severity: severity,
		Source:   "carrion-analyzer",
	}
	a.Diagnostics = append(a.Diagnostics, diagnostic)
}

// addReference records a reference to a symbol
func (a *Analyzer) addReference(symbolName string, tok token.Token) {
	ref := ReferenceLocation{
		Line:   tok.Line,
		Column: tok.Column,
		Length: len(symbolName),
	}
	a.References[symbolName] = append(a.References[symbolName], ref)
}

// GetErrors returns all analysis errors
func (a *Analyzer) GetErrors() []string {
	return a.Errors
}

// GetSymbolTable returns the symbol table
func (a *Analyzer) GetSymbolTable() *symbol.SymbolTable {
	return a.SymbolTable
}

// GetSymbolAtPosition finds the symbol at a specific position
func (a *Analyzer) GetSymbolAtPosition(line, column int) *symbol.Symbol {
	scope := a.SymbolTable.FindScopeAtPosition(line, column)
	if scope == nil {
		return nil
	}

	// This is a simplified implementation
	// In practice, we'd need to track which identifiers are at which positions
	// For now, return nil as we'd need additional position tracking
	return nil
}

// GetCompletionItems returns symbols available for code completion at a position
func (a *Analyzer) GetCompletionItems(line, column int, prefix string) []*symbol.Symbol {
	scope := a.SymbolTable.FindScopeAtPosition(line, column)
	if scope == nil {
		scope = a.SymbolTable.GlobalScope
	}

	// Get all symbols accessible from this scope
	allSymbols := scope.GetAllSymbols()
	var completionItems []*symbol.Symbol

	for name, sym := range allSymbols {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			completionItems = append(completionItems, sym)
		}
	}

	// Sort completion items by relevance (built-ins last, local symbols first)
	return a.sortCompletionItems(completionItems)
}

// sortCompletionItems sorts completion items by relevance
func (a *Analyzer) sortCompletionItems(items []*symbol.Symbol) []*symbol.Symbol {
	// Simple sort: put user-defined symbols first, then built-ins
	var userDefined []*symbol.Symbol
	var builtins []*symbol.Symbol
	
	for _, item := range items {
		if item.Token.Line == 0 { // Built-ins have line 0
			builtins = append(builtins, item)
		} else {
			userDefined = append(userDefined, item)
		}
	}
	
	// Combine: user-defined first, then built-ins
	result := append(userDefined, builtins...)
	return result
}

// GetMemberCompletionItems returns completion items for member access (obj.member)
func (a *Analyzer) GetMemberCompletionItems(objectName, memberPrefix string, line, column int) []*symbol.Symbol {
	scope := a.SymbolTable.FindScopeAtPosition(line, column)
	if scope == nil {
		scope = a.SymbolTable.GlobalScope
	}

	// Find the object symbol
	objectSymbol, exists := scope.Lookup(objectName)
	if !exists {
		return []*symbol.Symbol{}
	}

	var completionItems []*symbol.Symbol

	// Handle different types of objects
	switch objectSymbol.Type {
	case symbol.VariableSymbol:
		// Check if this variable is a module instance (e.g., sys = os())
		if objectSymbol.DataType != "" && objectSymbol.DataType != "unknown" {
			// First check if it's a built-in module instance
			if moduleMembers := a.getBuiltinModuleMembers(objectSymbol.DataType); len(moduleMembers) > 0 {
				for _, member := range moduleMembers {
					if memberPrefix == "" || strings.HasPrefix(member.Name, memberPrefix) {
						completionItems = append(completionItems, member)
					}
				}
				return completionItems
			}
			
			// Then check if it's a class instance
			if classSymbol, exists := scope.Lookup(objectSymbol.DataType); exists && classSymbol.Type == symbol.ClassSymbol {
				// Add class members (methods and attributes)
				for memberName, member := range classSymbol.Members {
					if memberPrefix == "" || strings.HasPrefix(memberName, memberPrefix) {
						completionItems = append(completionItems, member)
					}
				}
			}
		}

	case symbol.ClassSymbol:
		// For class symbols (static access), return class members
		for memberName, member := range objectSymbol.Members {
			if memberPrefix == "" || strings.HasPrefix(memberName, memberPrefix) {
				completionItems = append(completionItems, member)
			}
		}

	case symbol.ModuleSymbol:
		// For modules, return exported symbols
		for memberName, member := range objectSymbol.Members {
			if memberPrefix == "" || strings.HasPrefix(memberName, memberPrefix) {
				completionItems = append(completionItems, member)
			}
		}
	}

	return completionItems
}

// getBuiltinModuleMembers returns the members for built-in module instances
func (a *Analyzer) getBuiltinModuleMembers(moduleName string) []*symbol.Symbol {
	var members []*symbol.Symbol
	
	switch moduleName {
	case "os":
		members = append(members, &symbol.Symbol{
			Name: "cwd", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Get current working directory",
		})
		members = append(members, &symbol.Symbol{
			Name: "listdir", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "List directory contents",
		})
		members = append(members, &symbol.Symbol{
			Name: "mkdir", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Create a directory",
		})
		members = append(members, &symbol.Symbol{
			Name: "rmdir", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Remove a directory",
		})
		members = append(members, &symbol.Symbol{
			Name: "remove", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Remove a file",
		})
		members = append(members, &symbol.Symbol{
			Name: "rename", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Rename a file or directory",
		})
		members = append(members, &symbol.Symbol{
			Name: "getcwd", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Get current working directory (alias for cwd)",
		})
		members = append(members, &symbol.Symbol{
			Name: "chdir", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Change current directory",
		})
		members = append(members, &symbol.Symbol{
			Name: "getenv", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Get environment variable",
		})
		members = append(members, &symbol.Symbol{
			Name: "setenv", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Set environment variable",
		})
	case "file":
		members = append(members, &symbol.Symbol{
			Name: "open", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Open a file",
		})
		members = append(members, &symbol.Symbol{
			Name: "read", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Read from a file",
		})
		members = append(members, &symbol.Symbol{
			Name: "write", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Write to a file",
		})
		members = append(members, &symbol.Symbol{
			Name: "close", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Close a file",
		})
	case "http":
		members = append(members, &symbol.Symbol{
			Name: "get", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Make HTTP GET request",
		})
		members = append(members, &symbol.Symbol{
			Name: "post", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Make HTTP POST request",
		})
		members = append(members, &symbol.Symbol{
			Name: "put", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Make HTTP PUT request",
		})
		members = append(members, &symbol.Symbol{
			Name: "delete", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Make HTTP DELETE request",
		})
	case "time":
		members = append(members, &symbol.Symbol{
			Name: "now", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Get current timestamp",
		})
		members = append(members, &symbol.Symbol{
			Name: "sleep", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Sleep for specified seconds",
		})
		members = append(members, &symbol.Symbol{
			Name: "format", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Format timestamp",
		})
	case "math":
		members = append(members, &symbol.Symbol{
			Name: "abs", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Absolute value",
		})
		members = append(members, &symbol.Symbol{
			Name: "sqrt", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Square root",
		})
		members = append(members, &symbol.Symbol{
			Name: "pow", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Power function",
		})
		members = append(members, &symbol.Symbol{
			Name: "floor", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Floor function",
		})
		members = append(members, &symbol.Symbol{
			Name: "ceil", Type: symbol.FunctionSymbol, DataType: "function",
			Description: "Ceiling function",
		})
	}
	
	return members
}

// inferTypeFromAssignment infers the type of a variable from its assignment value
func (a *Analyzer) inferTypeFromAssignment(valueNode ast.Expression) string {
	switch node := valueNode.(type) {
	case *ast.CallExpression:
		// Check if this is a class constructor call
		if ident, ok := node.Function.(*ast.Identifier); ok {
			// Look up the function/class being called
			if sym, exists := a.SymbolTable.Lookup(ident.Value); exists {
				if sym.Type == symbol.ClassSymbol {
					// This is a class constructor, the variable type is the class name
					return sym.Name
				} else if sym.Type == symbol.ModuleSymbol {
					// This is a module constructor call, the variable type is the module name
					return sym.Name
				} else if sym.Type == symbol.FunctionSymbol && sym.ReturnType != "" {
					// This is a function call, use the return type
					return sym.ReturnType
				}
			}
		}
		return "unknown"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.FloatLiteral:
		return "float"
	case *ast.StringLiteral, *ast.FStringLiteral:
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
		if symbol, exists := a.SymbolTable.Lookup(node.Value); exists {
			return symbol.DataType
		}
		return "unknown"
	case *ast.InfixExpression:
		// Handle binary operations
		leftType := a.inferTypeFromAssignment(node.Left)
		rightType := a.inferTypeFromAssignment(node.Right)

		switch node.Operator {
		case "+", "-", "*", "/", "%", "**":
			// Arithmetic operations
			if leftType == "float" || rightType == "float" {
				return "float"
			} else if leftType == "int" && rightType == "int" {
				return "int"
			} else if leftType == "str" && rightType == "str" && node.Operator == "+" {
				return "str"
			}
			return "unknown"
		case "==", "!=", "<", ">", "<=", ">=":
			// Comparison operations always return bool
			return "bool"
		case "and", "or":
			// Logical operations return bool
			return "bool"
		default:
			return "unknown"
		}
	default:
		return "unknown"
	}
}

// GetDiagnostics returns diagnostic information for LSP
func (a *Analyzer) GetDiagnostics() []Diagnostic {
	// Return the diagnostics array which now has proper position information
	return a.Diagnostics
}

// FindReferences finds all references to a symbol at the given position
func (a *Analyzer) FindReferences(line, column int, includeDeclaration bool) []ReferenceLocation {
	var references []ReferenceLocation

	// For now, we'll use a simple approach: find the identifier at the position
	// by looking through all known symbols and their references
	var symbolName string

	// Check all references to find which symbol is at this position
	for name, refs := range a.References {
		for _, ref := range refs {
			if ref.Line == line && ref.Column <= column && column < ref.Column+ref.Length {
				symbolName = name
				break
			}
		}
		if symbolName != "" {
			break
		}
	}

	// If we didn't find a reference at this position, check symbol definitions
	if symbolName == "" {
		for name, sym := range a.SymbolTable.GetAllSymbols() {
			if sym.Token.Line == line && sym.Token.Column <= column && column < sym.Token.Column+len(name) {
				symbolName = name
				break
			}
		}
	}

	if symbolName == "" {
		return references
	}

	// Include declaration if requested
	if includeDeclaration {
		if sym, exists := a.SymbolTable.Lookup(symbolName); exists && sym.Token.Line > 0 {
			references = append(references, ReferenceLocation{
				Line:   sym.Token.Line,
				Column: sym.Token.Column,
				Length: len(symbolName),
			})
		}
	}

	// Add all references to this symbol
	if refs, exists := a.References[symbolName]; exists {
		references = append(references, refs...)
	}

	return references
}

// ReferenceLocation represents a location where a symbol is referenced
type ReferenceLocation struct {
	Line   int
	Column int
	Length int
}

// Diagnostic represents a diagnostic message (error, warning, info)
type Diagnostic struct {
	Range    Range
	Message  string
	Severity DiagnosticSeverity
	Source   string
}

// Range represents a text range
type Range struct {
	Start Position
	End   Position
}

// Position represents a text position
type Position struct {
	Line      int
	Character int
}

// DiagnosticSeverity represents the severity of a diagnostic
type DiagnosticSeverity int

const (
	DiagnosticError       DiagnosticSeverity = 1
	DiagnosticWarning     DiagnosticSeverity = 2
	DiagnosticInformation DiagnosticSeverity = 3
	DiagnosticHint        DiagnosticSeverity = 4
)
