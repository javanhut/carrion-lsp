package ast

import (
	"fmt"
	"strings"

	"github.com/javanhut/carrion-lsp/internal/carrion/token"
)

// Node represents any node in the AST
type Node interface {
	TokenLiteral() string
	String() string
	Position() (line, column int)
}

// Statement represents statement nodes in the AST
type Statement interface {
	Node
	statementNode()
}

// Expression represents expression nodes in the AST
type Expression interface {
	Node
	expressionNode()
}

// Program represents the root of every AST
type Program struct {
	Statements []Statement
	Errors     []string
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (p *Program) Position() (line, column int) {
	if len(p.Statements) > 0 {
		return p.Statements[0].Position()
	}
	return 0, 0
}

// Identifier represents identifier expressions
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()              {}
func (i *Identifier) TokenLiteral() string         { return i.Token.Literal }
func (i *Identifier) String() string               { return i.Value }
func (i *Identifier) Position() (line, column int) { return i.Token.Line, i.Token.Column }

// IntegerLiteral represents integer literals
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()              {}
func (il *IntegerLiteral) TokenLiteral() string         { return il.Token.Literal }
func (il *IntegerLiteral) String() string               { return il.Token.Literal }
func (il *IntegerLiteral) Position() (line, column int) { return il.Token.Line, il.Token.Column }

// FloatLiteral represents floating point literals
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()              {}
func (fl *FloatLiteral) TokenLiteral() string         { return fl.Token.Literal }
func (fl *FloatLiteral) String() string               { return fl.Token.Literal }
func (fl *FloatLiteral) Position() (line, column int) { return fl.Token.Line, fl.Token.Column }

// StringLiteral represents string literals
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()              {}
func (sl *StringLiteral) TokenLiteral() string         { return sl.Token.Literal }
func (sl *StringLiteral) String() string               { return fmt.Sprintf(`"%s"`, sl.Value) }
func (sl *StringLiteral) Position() (line, column int) { return sl.Token.Line, sl.Token.Column }

// FStringLiteral represents f-string literals
type FStringLiteral struct {
	Token token.Token
	Value string
}

func (fsl *FStringLiteral) expressionNode()              {}
func (fsl *FStringLiteral) TokenLiteral() string         { return fsl.Token.Literal }
func (fsl *FStringLiteral) String() string               { return fmt.Sprintf(`f"%s"`, fsl.Value) }
func (fsl *FStringLiteral) Position() (line, column int) { return fsl.Token.Line, fsl.Token.Column }

// BooleanLiteral represents boolean literals (True/False)
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()              {}
func (bl *BooleanLiteral) TokenLiteral() string         { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string               { return bl.Token.Literal }
func (bl *BooleanLiteral) Position() (line, column int) { return bl.Token.Line, bl.Token.Column }

// NoneLiteral represents None literal
type NoneLiteral struct {
	Token token.Token
}

func (nl *NoneLiteral) expressionNode()              {}
func (nl *NoneLiteral) TokenLiteral() string         { return nl.Token.Literal }
func (nl *NoneLiteral) String() string               { return "None" }
func (nl *NoneLiteral) Position() (line, column int) { return nl.Token.Line, nl.Token.Column }

// PrefixExpression represents prefix expressions (!-x, -x, +x, ~x)
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	if pe.Operator == "not" {
		return fmt.Sprintf("(%s %s)", pe.Operator, pe.Right.String())
	}
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}
func (pe *PrefixExpression) Position() (line, column int) { return pe.Token.Line, pe.Token.Column }

// InfixExpression represents infix expressions (x + y, x == y, etc.)
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}
func (ie *InfixExpression) Position() (line, column int) { return ie.Token.Line, ie.Token.Column }

// CallExpression represents function calls
type CallExpression struct {
	Token     token.Token
	Function  Expression // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}
func (ce *CallExpression) Position() (line, column int) { return ce.Token.Line, ce.Token.Column }

// IndexExpression represents array/dict indexing (arr[0], dict["key"])
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return fmt.Sprintf("(%s[%s])", ie.Left.String(), ie.Index.String())
}
func (ie *IndexExpression) Position() (line, column int) { return ie.Token.Line, ie.Token.Column }

// ArrayLiteral represents array literals [1, 2, 3]
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var elements []string
	for _, e := range al.Elements {
		elements = append(elements, e.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}
func (al *ArrayLiteral) Position() (line, column int) { return al.Token.Line, al.Token.Column }

// HashLiteral represents hash/dict literals {key: value}
type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var pairs []string
	for key, value := range hl.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key.String(), value.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}
func (hl *HashLiteral) Position() (line, column int) { return hl.Token.Line, hl.Token.Column }

// MemberExpression represents member access (obj.member)
type MemberExpression struct {
	Token  token.Token // the DOT token
	Object Expression
	Member *Identifier
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) String() string {
	return fmt.Sprintf("%s.%s", me.Object.String(), me.Member.String())
}
func (me *MemberExpression) Position() (line, column int) { return me.Token.Line, me.Token.Column }

// STATEMENTS

// ExpressionStatement represents expression statements
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}
func (es *ExpressionStatement) Position() (line, column int) { return es.Token.Line, es.Token.Column }

// AssignStatement represents assignment statements (x = 5)
type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	return fmt.Sprintf("%s = %s", as.Name.String(), as.Value.String())
}
func (as *AssignStatement) Position() (line, column int) { return as.Token.Line, as.Token.Column }

// MemberAssignStatement represents member assignment statements (obj.member = value)
type MemberAssignStatement struct {
	Token  token.Token
	Object Expression
	Member *Identifier
	Value  Expression
}

func (mas *MemberAssignStatement) statementNode()       {}
func (mas *MemberAssignStatement) TokenLiteral() string { return mas.Token.Literal }
func (mas *MemberAssignStatement) String() string {
	return fmt.Sprintf("%s.%s = %s", mas.Object.String(), mas.Member.String(), mas.Value.String())
}
func (mas *MemberAssignStatement) Position() (line, column int) {
	return mas.Token.Line, mas.Token.Column
}

// ReturnStatement represents return statements
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	if rs.ReturnValue != nil {
		return fmt.Sprintf("return %s", rs.ReturnValue.String())
	}
	return "return"
}
func (rs *ReturnStatement) Position() (line, column int) { return rs.Token.Line, rs.Token.Column }

// BlockStatement represents block statements (groups of statements)
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out strings.Builder
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
func (bs *BlockStatement) Position() (line, column int) { return bs.Token.Line, bs.Token.Column }

// IfStatement represents if statements
type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ifs *IfStatement) statementNode()       {}
func (ifs *IfStatement) TokenLiteral() string { return ifs.Token.Literal }
func (ifs *IfStatement) String() string {
	var out strings.Builder
	out.WriteString("if ")
	out.WriteString(ifs.Condition.String())
	out.WriteString(":\n")
	out.WriteString(ifs.Consequence.String())
	if ifs.Alternative != nil {
		out.WriteString("else:\n")
		out.WriteString(ifs.Alternative.String())
	}
	return out.String()
}
func (ifs *IfStatement) Position() (line, column int) { return ifs.Token.Line, ifs.Token.Column }

// WhileStatement represents while loops
type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	return fmt.Sprintf("while %s:\n%s", ws.Condition.String(), ws.Body.String())
}
func (ws *WhileStatement) Position() (line, column int) { return ws.Token.Line, ws.Token.Column }

// ForStatement represents for loops
type ForStatement struct {
	Token    token.Token
	Variable *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	return fmt.Sprintf("for %s in %s:\n%s", fs.Variable.String(), fs.Iterable.String(), fs.Body.String())
}
func (fs *ForStatement) Position() (line, column int) { return fs.Token.Line, fs.Token.Column }

// FunctionStatement represents spell (function) definitions
type FunctionStatement struct {
	Token      token.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var params []string
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("spell %s(%s):\n%s", fs.Name.String(), strings.Join(params, ", "), fs.Body.String())
}
func (fs *FunctionStatement) Position() (line, column int) { return fs.Token.Line, fs.Token.Column }

// ClassStatement represents grim (class) definitions
type ClassStatement struct {
	Token   token.Token
	Name    *Identifier
	Parent  *Identifier // Optional parent class
	Methods []*FunctionStatement
	Body    *BlockStatement
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStatement) String() string {
	var out strings.Builder
	out.WriteString("grim ")
	out.WriteString(cs.Name.String())
	if cs.Parent != nil {
		out.WriteString("(")
		out.WriteString(cs.Parent.String())
		out.WriteString(")")
	}
	out.WriteString(":\n")
	if cs.Body != nil {
		out.WriteString(cs.Body.String())
	}
	return out.String()
}
func (cs *ClassStatement) Position() (line, column int) { return cs.Token.Line, cs.Token.Column }

// ImportStatement represents import statements
type ImportStatement struct {
	Token  token.Token
	Module *Identifier
	Alias  *Identifier // Optional alias (import x as y)
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	if is.Alias != nil {
		return fmt.Sprintf("import %s as %s", is.Module.String(), is.Alias.String())
	}
	return fmt.Sprintf("import %s", is.Module.String())
}
func (is *ImportStatement) Position() (line, column int) { return is.Token.Line, is.Token.Column }

// IgnoreStatement represents ignore statements (no-op)
type IgnoreStatement struct {
	Token token.Token
}

func (igs *IgnoreStatement) statementNode()               {}
func (igs *IgnoreStatement) TokenLiteral() string         { return igs.Token.Literal }
func (igs *IgnoreStatement) String() string               { return "ignore" }
func (igs *IgnoreStatement) Position() (line, column int) { return igs.Token.Line, igs.Token.Column }
