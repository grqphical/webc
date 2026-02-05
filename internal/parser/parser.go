package parser

import (
	"fmt"
	"strconv"

	"github.com/grqphical/webc/internal/ast"
	"github.com/grqphical/webc/internal/lexer"
)

// ordering for expression evaluation
const (
	_ int = iota
	PrecendenceLowest
	PrecedenceEquals
	PrecedenceLessGreater
	PrecedenceSum
	PrecedenceProduct
	PrecedencePrefix
	PrecedenceCall
)

var precedenceLookup = map[lexer.TokenType]int{
	lexer.TokenPlus:  PrecedenceSum,
	lexer.TokenDash:  PrecedenceSum,
	lexer.TokenStar:  PrecedenceProduct,
	lexer.TokenSlash: PrecedenceProduct,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
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

	curToken   lexer.Token
	peekToken  lexer.Token
	peekToken2 lexer.Token

	errors []ParseError

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn

	curFunction *ast.Function
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: make([]ParseError, 0),
	}

	// read the first three characters so that curToken, peekToken and peekToken2 are set
	p.nextToken()
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.TokenIdent, p.parseIdentifier)
	p.registerPrefix(lexer.TokenIntLiteral, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TokenFloatLiteral, p.parseFloatLiteral)
	p.registerPrefix(lexer.TokenDash, p.parsePrefixExpression)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.TokenPlus, p.parseInfixExpression)
	p.registerInfix(lexer.TokenDash, p.parseInfixExpression)
	p.registerInfix(lexer.TokenStar, p.parseInfixExpression)
	p.registerInfix(lexer.TokenSlash, p.parseInfixExpression)

	return p
}

func (p *Parser) Errors() []ParseError {
	return p.errors
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType, line int) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, ParseError{
		message: msg,
		line:    line,
	})
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, ParseError{
		message: msg,
		line:    p.peekToken.Line,
	})
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedenceLookup[p.peekToken.Type]; ok {
		return p
	}

	return PrecendenceLowest
}

func (p *Parser) currentPrecedence() int {
	if p, ok := precedenceLookup[p.curToken.Type]; ok {
		return p
	}

	return PrecendenceLowest
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.peekToken2
	p.peekToken2 = p.l.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// peeks ahead by two
func (p *Parser) doublePeekTokenIs(t lexer.TokenType) bool {
	return p.peekToken2.Type == t
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
	case lexer.TokenReturn:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseVariableDefineStatement() ast.Statement {
	t := ast.ValueType(p.curToken.Literal)
	stmt := &ast.VariableDefineStatement{Token: p.curToken, Type: t}

	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal, Symbol: p.curFunction.SetSymbol(p.curToken.Literal, t)}

	if p.peekTokenIs(lexer.TokenSemicolon) {
		// just defining the variable to be uninitialized
		p.nextToken()
		return stmt
	}

	if !p.expectPeek(lexer.TokenEqual) {
		return nil
	}
	p.nextToken()

	stmt.Value = p.parseExpression(PrecendenceLowest)

	if !p.expectPeek(lexer.TokenSemicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() // consume 'return'

	stmt.ReturnValue = p.parseExpression(PrecendenceLowest)

	if !p.expectPeek(lexer.TokenSemicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type, p.curToken.Line)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TokenSemicolon) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(PrecendenceLowest)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken, Type: ast.ValueTypeInt}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as an integer", p.curToken.Literal)
		p.errors = append(p.errors, ParseError{
			message: msg,
			line:    p.curToken.Line,
		})
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken, Type: ast.ValueTypeFloat}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as a float", p.curToken.Literal)
		p.errors = append(p.errors, ParseError{
			message: msg,
			line:    p.curToken.Line,
		})
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PrecedencePrefix)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseFunction() *ast.Function {
	t := p.curToken.Literal
	p.nextToken()
	name := p.curToken.Literal
	function := ast.NewFunction(name, ast.ValueType(t))
	function.Statements = make([]ast.Statement, 0)
	p.curFunction = function

	// just skip '()' for now, we will deal with arguments later
	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	// skip '{'
	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}
	p.nextToken()

	for p.curToken.Type != lexer.TokenRBrace {
		if p.curToken.Type == lexer.TokenEndOfFile {
			p.errors = append(p.errors, ParseError{
				message: "expected }, got EOF instead",
				line:    p.curToken.Line,
			})
			return nil
		}

		stmt := p.parseStatement()
		if stmt != nil {
			function.Statements = append(function.Statements, stmt)
		}
		p.nextToken()
	}

	return function
}

func (p *Parser) isTypeKeyword(t lexer.TokenType) bool {
	return t == lexer.TokenIntKeyword ||
		t == lexer.TokenFloatKeyword ||
		t == lexer.TokenCharKeyword
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Functions = make([]*ast.Function, 0)

	for p.curToken.Type != lexer.TokenEndOfFile {
		// functions are form [type] [identifier]()
		isFunc := p.isTypeKeyword(p.curToken.Type) &&
			p.peekTokenIs(lexer.TokenIdent) &&
			p.doublePeekTokenIs(lexer.TokenLParen)

		if isFunc {
			function := p.parseFunction()
			if function != nil {
				program.Functions = append(program.Functions, function)
			}
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				program.Statements = append(program.Statements, stmt)
			}
		}

		p.nextToken()
	}

	return program
}
