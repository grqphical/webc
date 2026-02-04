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
)

type Node interface {
	TokenLiteral() string
	String() string
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
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}
func (i *Identifier) String() string {
	return i.Value
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
