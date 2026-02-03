package parser

import (
	"github.com/grqphical/webc/internal/ast"
	"github.com/grqphical/webc/internal/lexer"
)

type Parser struct {
	l *lexer.Lexer

	curToken  lexer.Token
	peekToken lexer.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// read the first two characters so that curToken and peekToken are set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		return false
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.TK_INT, lexer.TK_FLOAT, lexer.TK_CHAR:
		return p.parseVariableDefineStatement()
	default:
		return nil
	}
}

func (p *Parser) parseVariableDefineStatement() ast.Statement {
	stmt := &ast.VariableDefineStatement{Token: p.curToken, Type: ast.ValueType(p.curToken.Literal)}

	if !p.expectPeek(lexer.TK_IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.expectPeek(lexer.TK_SEMICOLON) {
		// just defining the variable to be uninitialized
		return stmt
	}

	if !p.expectPeek(lexer.TK_EQUAL) {
		return nil
	}

	for !p.curTokenIs(lexer.TK_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = make([]ast.Statement, 0)

	for p.curToken.Type != lexer.TK_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}
