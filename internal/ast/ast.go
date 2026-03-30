package ast

import (
	"bytes"
	"fmt"

	"github.com/grqphical/webc/internal/lexer"
)

// Represents a type within the C programming language
// Used mainly for storing return types of functions and statements within the AST
type ValueType string

const (
	ValueTypeInt   ValueType = "int"
	ValueTypeFloat ValueType = "float"
	ValueTypeChar  ValueType = "char"
	ValueTypeVoid  ValueType = "void"
)

// Represents a Node in the Abstract Syntax Tree
type Node interface {
	// Returns the TokenLiteral of the AST Node
	TokenLiteral() string
	// Returns a string representation of the AST Node
	String() string
	// Returns the type of the AST Node
	ValueType() ValueType
}

// Represents a statement node (e.g. variable declaration, if statement etc.) in the AST
type Statement interface {
	Node
	statementNode()
}

// Represents an expression (e.g. 5 < 2, 6 + 7, 5 * (3 + 5))
type Expression interface {
	Node
	expressionNode()
}

// Stores the entire program and it's AST
type Program struct {
	// Every function declared
	Functions []*Function

	// Functions to be imported from JS
	ExternalFunctions []*Function

	// global statements
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Functions) > 0 {
		return p.Functions[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	for _, f := range p.Functions {
		out.WriteString(f.String())
	}
	return out.String()
}

func (p *Program) Type() ValueType {
	return ValueTypeVoid
}

// checks if a function exists, if so returns it's index.
// otherwise the function returns -1
func (p *Program) FunctionExists(name string) int {
	for i, f := range p.Functions {
		if f.Name == name {
			return i
		}
	}

	return -1
}

// Represents an argument for a function
type Argument struct {
	Name string
	Type ValueType
}

// Represents a function in the program
type Function struct {
	Name            string
	ReturnType      ValueType
	Statement       Statement
	SymbolIndex     map[string]int
	Symbols         []*Symbol
	NextSymbolIndex int
	Arguments       []Argument
}

func NewFunction(name string, returnType ValueType) *Function {
	return &Function{
		Name:            name,
		ReturnType:      returnType,
		Statement:       nil,
		Symbols:         make([]*Symbol, 0),
		SymbolIndex:     make(map[string]int),
		NextSymbolIndex: 0,
	}
}

func (f *Function) TokenLiteral() string {
	if f.Statement != nil {
		return f.Statement.TokenLiteral()
	}
	return ""
}

func (f *Function) String() string {
	if f.Statement != nil {
		return f.Statement.String()
	}
	return ""
}

func (f *Function) ValueType() ValueType {
	return f.ReturnType
}

// Gets the number of variables based on each main type, used to generate WASM locals to store the variables in
func (f *Function) GetVariableCounts() (integerCount, floatCount int) {
	for _, s := range f.Symbols {
		switch s.Type {
		case ValueTypeFloat:
			floatCount++
		case ValueTypeInt:
			integerCount++
		case ValueTypeChar:
			integerCount++
		}
	}
	return
}

// Registers a variable within the program
func (f *Function) SetSymbol(name string, t ValueType, constant bool) *Symbol {
	idx := len(f.Symbols)
	s := &Symbol{
		Index:    idx,
		Type:     t,
		Constant: constant,
	}
	f.SymbolIndex[name] = idx
	f.Symbols = append(f.Symbols, s)
	return s
}

// Retrieves the symbol for the given variable name
func (f *Function) GetSymbol(name string) *Symbol {
	idx, ok := f.SymbolIndex[name]
	if !ok {
		return nil
	}
	return f.Symbols[idx]
}

// Represents a defiend symbol such as a variable name
type Symbol struct {
	Index    int
	Type     ValueType
	Constant bool
}

// Represents an identifier expression (e.g. using a variable in an expression)
type Identifier struct {
	Token  lexer.Token
	Value  string
	Symbol *Symbol
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

func (i *Identifier) ValueType() ValueType {
	return i.Symbol.Type
}

// Represents a block of code contained within {}
type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
func (bs *BlockStatement) ValueType() ValueType { return ValueTypeVoid }

// Represents a statement that defines a variable (e.g. int x = 0;)
type VariableDefineStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
	Type  ValueType
}

func (vds *VariableDefineStatement) statementNode() {}
func (vds *VariableDefineStatement) TokenLiteral() string {
	return vds.Token.Literal
}

func (vds *VariableDefineStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vds.Token.Literal + " ")
	out.WriteString(vds.Name.String())
	out.WriteString(" = ")

	if vds.Value != nil {
		out.WriteString(vds.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (vds *VariableDefineStatement) ValueType() ValueType {
	return vds.Type
}

// Represents a statement that defines a variable (e.g. x = 5; x += 6;)
type VariableUpdateStatement struct {
	Name      *Identifier
	Token     lexer.Token
	NewValue  Expression
	Operation string
}

func (vus *VariableUpdateStatement) statementNode() {}
func (vus *VariableUpdateStatement) TokenLiteral() string {
	return vus.Token.Literal
}

func (vus *VariableUpdateStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vus.Name.String())
	out.WriteString(" = ")
	out.WriteString(vus.NewValue.String())

	out.WriteString(";")

	return out.String()
}

func (vus *VariableUpdateStatement) ValueType() ValueType {
	return vus.Name.Symbol.Type
}

// Represents a statement that returns a value from a function
type ReturnStatement struct {
	Token       lexer.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

func (rs *ReturnStatement) ValueType() ValueType {
	return rs.ReturnValue.ValueType()
}

// Represents an if statement in the AST
type IfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence Statement
	Alternative Statement
}

func (i *IfStatement) statementNode() {}
func (i *IfStatement) TokenLiteral() string {
	return i.Token.Literal
}

func (i *IfStatement) String() string {
	if i.Condition != nil {
		return "if (" + i.Condition.String() + ");"
	}
	return ""
}

func (i *IfStatement) ValueType() ValueType {
	return i.Condition.ValueType()
}

// Represents a standalone expression statement
type ExpressionStatement struct {
	Token      lexer.Token
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

func (es *ExpressionStatement) ValueType() ValueType {
	return es.Expression.ValueType()
}

// Represents a literal integer (e.g. 5, 682329, -2198)
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
	Type  ValueType
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) ValueType() ValueType { return il.Type }

// Represents a literal float (e.g. 5.0, 3.14159, -0.56791)
type FloatLiteral struct {
	Token lexer.Token
	Value float64
	Type  ValueType
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }
func (fl *FloatLiteral) ValueType() ValueType { return fl.Type }

// Represents a character literal (e.g. 'a', 'b', '\n')
type CharLiteral struct {
	Token lexer.Token
	Value byte
}

func (cl *CharLiteral) expressionNode()      {}
func (cl *CharLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *CharLiteral) String() string       { return cl.Token.Literal }
func (fl *CharLiteral) ValueType() ValueType { return ValueTypeChar }

// Represents an expression with a prefix operator (e.g. -5, +6)
type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

func (pe *PrefixExpression) ValueType() ValueType {
	return pe.Right.ValueType()
}

// Represents an expression with an operator and two values (e.g. 5 + 6)
type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

func (ie *InfixExpression) ValueType() ValueType {
	return ie.Left.ValueType()
}

// Represents a call to a function in an expression
type FunctionCallExpression struct {
	Token      lexer.Token
	Name       string
	Args       []Expression
	ReturnType ValueType
	Index      int
}

func (fce *FunctionCallExpression) expressionNode()      {}
func (fce *FunctionCallExpression) TokenLiteral() string { return fce.Token.Literal }
func (fce *FunctionCallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(fce.Name)
	out.WriteString("(")
	for _, arg := range fce.Args {
		out.WriteString(arg.String())
	}
	out.WriteString(")")

	return out.String()
}

func (fce *FunctionCallExpression) ValueType() ValueType {
	return fce.ReturnType
}

// Represents an expression with it's operator on the right side (e.g. 5++, i--)
type PostfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PostfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Operator)
	out.WriteString(")")
	return out.String()
}
func (pe *PostfixExpression) ValueType() ValueType {
	return pe.Left.ValueType()
}

// Represents a while loop
type WhileLoopStatement struct {
	Token     lexer.Token
	Condition Expression
	Statement Statement
}

func (ws *WhileLoopStatement) statementNode()       {}
func (ws *WhileLoopStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileLoopStatement) String() string {
	return fmt.Sprintf("while (%s) %s", ws.Condition.String(), ws.Statement.String())
}
func (ws *WhileLoopStatement) ValueType() ValueType {
	return ws.Condition.ValueType()
}

// Represents a for loop
type ForLoopStatement struct {
	Token     lexer.Token
	Initial   Statement  // initial statement (e.g. int i = 0;)
	Condition Expression // stopping condition (e.g. i < 10;)
	Increment Expression // code to run every iteration (e.g. i++)
	Statement Statement  // actual code to run throughout the loop
}

func (fs *ForLoopStatement) statementNode()       {}
func (fs *ForLoopStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForLoopStatement) String() string {
	return fmt.Sprintf("for (%s;%s;%s) %s", fs.Initial.String(), fs.Condition.String(), fs.Increment.String(), fs.Statement.String())
}
func (fs *ForLoopStatement) ValueType() ValueType {
	return fs.Condition.ValueType()
}
