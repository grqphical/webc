package ast

import (
	"bytes"

	"github.com/grqphical/webc/internal/lexer"
)

type ValueType string

const (
	ValueTypeInt   ValueType = "int"
	ValueTypeFloat ValueType = "float"
	ValueTypeChar  ValueType = "char"
	ValueTypeVoid  ValueType = "void"
)

type Node interface {
	TokenLiteral() string
	String() string
	ValueType() ValueType
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
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

type Function struct {
	Name            string
	ReturnType      ValueType
	Statements      []Statement
	SymbolIndex     map[string]int
	Symbols         []*Symbol
	NextSymbolIndex int
}

func NewFunction(name string, returnType ValueType) *Function {
	return &Function{
		Name:            name,
		ReturnType:      returnType,
		Statements:      make([]Statement, 0),
		Symbols:         make([]*Symbol, 0),
		SymbolIndex:     make(map[string]int),
		NextSymbolIndex: 0,
	}
}

func (f *Function) TokenLiteral() string {
	if len(f.Statements) > 0 {
		return f.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (f *Function) String() string {
	var out bytes.Buffer
	for _, s := range f.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (f *Function) ValueType() ValueType {
	return f.ReturnType
}

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

func (f *Function) GetSymbol(name string) *Symbol {
	idx, ok := f.SymbolIndex[name]
	if !ok {
		return nil
	}
	return f.Symbols[idx]
}

type Symbol struct {
	Index    int
	Type     ValueType
	Constant bool
}

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

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
	Type  ValueType
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) ValueType() ValueType { return il.Type }

type FloatLiteral struct {
	Token lexer.Token
	Value float64
	Type  ValueType
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }
func (fl *FloatLiteral) ValueType() ValueType { return fl.Type }

type CharLiteral struct {
	Token lexer.Token
	Value byte
}

func (cl *CharLiteral) expressionNode()      {}
func (cl *CharLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *CharLiteral) String() string       { return cl.Token.Literal }
func (fl *CharLiteral) ValueType() ValueType { return ValueTypeChar }

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
