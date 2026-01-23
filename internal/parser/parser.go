package parser

import (
	"errors"
	"fmt"

	"github.com/grqphical/webc/internal/lexer"
)

type Node any

type Program struct {
	Functions []FunctionDecl
}

type FunctionDecl struct {
	Type string
	Name string
	Body Block
}

type Block struct {
	Statements []Node
}

type ReturnStmt struct {
	Value Node
}

type Constant struct {
	Value string
}

type Parser struct {
	tokens []lexer.Token
	head   int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		head:   0,
	}
}

func (p *Parser) getCurrentToken() lexer.Token {
	return p.tokens[p.head]
}

func (p *Parser) parseFunction() (FunctionDecl, error) {
	funcDecl := FunctionDecl{
		Type: p.getCurrentToken().Literal,
	}
	p.head++

	funcDecl.Name = p.getCurrentToken().Literal

	if p.getCurrentToken().Type != lexer.TK_LPAREN {
		return FunctionDecl{}, fmt.Errorf("invalid token after function declaration, expected '(' got '%s'", p.getCurrentToken().Literal)
	}
	p.head++
	if p.getCurrentToken().Type != lexer.TK_RPAREN {
		return FunctionDecl{}, errors.New("invalid token after '(' declaration, expected ')'")
	}
	p.head++

	funcDecl.Body = p.parseBlock()

	return funcDecl, nil

}

func (p *Parser) parseBlock() Block {
	block := Block{}

	// consume {
	p.head++

	for p.getCurrentToken().Type != lexer.TK_RBRACE && p.getCurrentToken().Type != lexer.TK_EOF {
		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)
		p.head++
	}

	return block
}

func (p *Parser) parseStatement() Node {
	switch p.getCurrentToken().Literal {
	case "return":
		return p.parseReturn()
	default:
		return nil
	}
}

func (p *Parser) parseReturn() Node {
	p.head++ // consume 'return'

	stmt := ReturnStmt{Value: Constant{Value: p.getCurrentToken().Literal}}
	p.head++ // consume return value

	if p.getCurrentToken().Type == lexer.TK_SEMICOLON {
		p.head++
	}

	return stmt
}

func (p *Parser) Parse() (Program, error) {
	prog := Program{
		Functions: make([]FunctionDecl, 0),
	}

	for p.head < len(p.tokens) {
		tok := p.getCurrentToken()
		switch tok.Type {
		case lexer.TK_EOF:

		case lexer.TK_IDENT:
			f, err := p.parseFunction()
			if err != nil {
				return Program{}, err
			}

			prog.Functions = append(prog.Functions, f)
		}
		p.head++
	}

	return prog, nil
}
