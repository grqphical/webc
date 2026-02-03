package ast

import "github.com/grqphical/webc/internal/lexer"

type Node interface {
	TokenLiteral() string
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

type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

type VariableDefineStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (vds *VariableDefineStatement) statementNode() {}
func (vds *VariableDefineStatement) TokenLiteral() string {
	return vds.Token.Literal
}
