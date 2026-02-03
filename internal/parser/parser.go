package parser

import (
	"fmt"

	"github.com/grqphical/webc/internal/ast"
	"github.com/grqphical/webc/internal/lexer"
)

type ParseError struct {
	message string
	line    int
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("SyntaxError: %s, line: %d", pe.message, pe.line)
}

type Parser struct {
	l *lexer.Lexer

	curToken  lexer.Token
	peekToken lexer.Token

	errors []ParseError
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: make([]ParseError, 0),
	}

	// read the first two characters so that curToken and peekToken are set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []ParseError {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, ParseError{
		message: msg,
		line:    p.peekToken.Line,
	})
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
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.TokenIntKeyword, lexer.TokenFloatKeyword, lexer.TokenCharKeyword:
		return p.parseVariableDefineStatement()
	default:
		return nil
	}
}

func (p *Parser) parseVariableDefineStatement() ast.Statement {
	stmt := &ast.VariableDefineStatement{Token: p.curToken, Type: ast.ValueType(p.curToken.Literal)}

	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.expectPeek(lexer.TokenSemicolon) {
		// just defining the variable to be uninitialized
		return stmt
	}

	if !p.expectPeek(lexer.TokenEqual) {
		return nil
	}

	for !p.curTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = make([]ast.Statement, 0)

	for p.curToken.Type != lexer.TokenEndOfFile {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}
