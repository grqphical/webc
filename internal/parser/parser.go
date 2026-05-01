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
	PrecedencePostFix
)

// table to lookup what each token's precedence (order of operations) is
var precedenceLookup = map[lexer.TokenType]int{
	lexer.TokenPlus:           PrecedenceSum,
	lexer.TokenDash:           PrecedenceSum,
	lexer.TokenStar:           PrecedenceProduct,
	lexer.TokenSlash:          PrecedenceProduct,
	lexer.TokenBang:           PrecedenceLessGreater,
	lexer.TokenGreaterOrEqual: PrecedenceLessGreater,
	lexer.TokenLessOrEqual:    PrecedenceLessGreater,
	lexer.TokenGreaterThan:    PrecedenceLessGreater,
	lexer.TokenLessThan:       PrecedenceLessGreater,
	lexer.TokenEqualEqual:     PrecedenceLessGreater,
	lexer.TokenNotEqual:       PrecedenceLessGreater,
	lexer.TokenIncrement:      PrecedencePostFix,
	lexer.TokenDecrement:      PrecedencePostFix,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Represents an error that occured while parsing
type ParseError struct {
	message string
	line    int
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("SyntaxError: %s, line: %d", pe.message, pe.line)
}

// The parser itself. Stores the current state of the parser
type Parser struct {
	l *lexer.Lexer

	curToken   lexer.Token
	peekToken  lexer.Token
	peekToken2 lexer.Token

	errors []ParseError

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn

	curFunction *ast.Function

	program *ast.Program
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: make([]ParseError, 0),
		// create a global "function" to store global variables and stuff
		curFunction: ast.NewFunction("_global", ast.ValueTypeVoid),
	}

	// read the first three characters so that curToken, peekToken and peekToken2 are set
	p.nextToken()
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.TokenIdent, p.parseIdentifier)
	p.registerPrefix(lexer.TokenIntLiteral, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TokenFloatLiteral, p.parseFloatLiteral)
	p.registerPrefix(lexer.TokenCharLiteral, p.parseCharLiteral)
	p.registerPrefix(lexer.TokenDash, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenBang, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenLParen, p.parseGroupedExpression)
	p.registerPrefix(lexer.TokenIncrement, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenDecrement, p.parsePrefixExpression)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.TokenPlus, p.parseInfixExpression)
	p.registerInfix(lexer.TokenDash, p.parseInfixExpression)
	p.registerInfix(lexer.TokenStar, p.parseInfixExpression)
	p.registerInfix(lexer.TokenSlash, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLessThan, p.parseInfixExpression)
	p.registerInfix(lexer.TokenGreaterThan, p.parseInfixExpression)
	p.registerInfix(lexer.TokenGreaterOrEqual, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLessOrEqual, p.parseInfixExpression)
	p.registerInfix(lexer.TokenEqualEqual, p.parseInfixExpression)
	p.registerInfix(lexer.TokenIncrement, p.parsePostfixExpression)
	p.registerInfix(lexer.TokenDecrement, p.parsePostfixExpression)

	return p
}

// Returns any errors that occured during parsing
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

// Registers a prefix parser function, which runs when the given token is a prefix expression (e.g. -, !)
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// Registers an infix parser function, which runs when the given token is an infix expression (e.g. ==, +, >)
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Adds an error to the parser if the next token is not the expected type
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, ParseError{
		message: msg,
		line:    p.peekToken.Line,
	})
}

// Finds the precedence of the next token
func (p *Parser) peekPrecedence() int {
	if p, ok := precedenceLookup[p.peekToken.Type]; ok {
		return p
	}

	return PrecendenceLowest
}

// Finds the precedence of the current token
func (p *Parser) currentPrecedence() int {
	if p, ok := precedenceLookup[p.curToken.Type]; ok {
		return p
	}

	return PrecendenceLowest
}

// Advances the parser
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.peekToken2
	p.peekToken2 = p.l.NextToken()
}

// Checks if the current token is equal to the given token
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// Checks if the next token is equal to the given token
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// Checks if the next, next token is equal to the given token
func (p *Parser) doublePeekTokenIs(t lexer.TokenType) bool {
	return p.peekToken2.Type == t
}

// Similar to curTokenIs but if it's false, an error is added to the parser
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// Parses an entire statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.TokenIntKeyword, lexer.TokenFloatKeyword, lexer.TokenCharKeyword, lexer.TokenConst:
		return p.parseVariableDefineStatement()
	case lexer.TokenLBrace:
		return p.parseBlock()
	case lexer.TokenIdent:
		return p.parseVariableUpdateStatement()
	case lexer.TokenReturn:
		return p.parseReturnStatement()
	case lexer.TokenIf:
		return p.parseIfStatement()
	case lexer.TokenWhile:
		return p.parseWhileLoop()
	case lexer.TokenFor:
		return p.parseForLoop()
	default:
		return p.parseExpressionStatement()
	}
}

// Parses a block of code, contained in {}. Used in function bodies, for loops, etc.
func (p *Parser) parseBlock() ast.Statement {
	block := &ast.BlockStatement{
		Token: p.curToken,
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
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	return block
}

// Parses a C-style while loop
func (p *Parser) parseWhileLoop() ast.Statement {
	stmt := &ast.WhileLoopStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}
	p.nextToken()

	stmt.Condition = p.parseExpression(PrecendenceLowest)
	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}
	p.nextToken()

	stmt.Statement = p.parseStatement()
	return stmt
}

// Parses a C-style for loop
func (p *Parser) parseForLoop() ast.Statement {
	stmt := &ast.ForLoopStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}
	p.nextToken()

	if p.curTokenIs(lexer.TokenSemicolon) {
		stmt.Initial = nil
	} else {
		stmt.Initial = p.parseStatement()
		if !p.curTokenIs(lexer.TokenSemicolon) {
			p.errors = append(p.errors, ParseError{
				message: fmt.Sprintf("expected semicolon, got %s", p.curToken.Literal),
				line:    p.curToken.Line,
			})
			return nil
		}
	}
	p.nextToken()

	if p.curTokenIs(lexer.TokenSemicolon) {
		stmt.Condition = nil
	} else {
		stmt.Condition = p.parseExpression(PrecendenceLowest)
		if !p.expectPeek(lexer.TokenSemicolon) {
			return nil
		}
	}
	p.nextToken() // Move past the semicolon to the increment

	if p.curTokenIs(lexer.TokenRParen) {
		stmt.Increment = nil
	} else {
		stmt.Increment = p.parseExpression(PrecendenceLowest)
		if !p.expectPeek(lexer.TokenRParen) {
			return nil
		}
	}
	p.nextToken() // Move past the ')' to the body block

	stmt.Statement = p.parseStatement()

	return stmt
}

// Parses a variable update statement such as x = 5; or foo += 10;
func (p *Parser) parseVariableUpdateStatement() ast.Statement {
	if !p.peekTokenIs(lexer.TokenEqual) && !p.peekTokenIs(lexer.TokenPlusEqual) && !p.peekTokenIs(lexer.TokenMinusEqual) && !p.peekTokenIs(lexer.TokenTimesEqual) && !p.peekTokenIs(lexer.TokenDivideEqual) {
		return p.parseExpressionStatement()
	}

	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	symbol := p.curFunction.GetSymbol(p.curToken.Literal)
	if symbol == nil {
		p.errors = append(p.errors, ParseError{
			message: fmt.Sprintf("undeclared variable %s", p.curToken.Literal),
			line:    p.curToken.Line,
		})
		return nil
	}
	if symbol.Constant {
		p.errors = append(p.errors, ParseError{
			message: fmt.Sprintf("cannot modify constant variable %s", p.curToken.Literal),
			line:    p.curToken.Line,
		})

		// ERROR RECOVERY: We must skip the rest of this invalid statement
		// until we find a semicolon, otherwise the parser will crash on the next token.
		for !p.curTokenIs(lexer.TokenSemicolon) && !p.curTokenIs(lexer.TokenEndOfFile) {
			p.nextToken()
		}
		if p.peekTokenIs(lexer.TokenSemicolon) {
			p.nextToken()
		}

		return nil
	}

	name.Symbol = symbol

	stmt := &ast.VariableUpdateStatement{Name: name, Token: p.curToken}

	p.nextToken()
	stmt.Operation = p.curToken.Literal
	p.nextToken()

	stmt.NewValue = p.parseExpression(PrecendenceLowest)

	if !p.expectPeek(lexer.TokenSemicolon) {
		return nil
	}

	return stmt
}

// Parses a variable definition statment such as int x = 0;
func (p *Parser) parseVariableDefineStatement() ast.Statement {
	constant := false
	if p.curTokenIs(lexer.TokenConst) {
		constant = true
		p.nextToken()
	}

	t := ast.ValueType(p.curToken.Literal)
	stmt := &ast.VariableDefineStatement{Token: p.curToken, Type: t}

	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal, Symbol: p.curFunction.SetSymbol(p.curToken.Literal, t, constant)}

	if p.peekTokenIs(lexer.TokenSemicolon) {
		// just defining the variable to be uninitialized if there is a semicolon right after the variable name
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

// Parses a return statement
func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() // consume 'return'

	stmt.ReturnValue = p.parseExpression(PrecendenceLowest)

	if !p.expectPeek(lexer.TokenSemicolon) {
		return nil
	}

	return stmt
}

// Parses a C-Style if statement
func (p *Parser) parseIfStatement() ast.Statement {
	ifStmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}
	p.nextToken()

	exp := p.parseExpression(PrecendenceLowest)
	ifStmt.Condition = exp

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	p.nextToken()
	ifStmt.Consequence = p.parseStatement()

	if p.peekTokenIs(lexer.TokenElse) {
		p.nextToken()

		if p.peekTokenIs(lexer.TokenIf) {
			p.nextToken()
			ifStmt.Alternative = p.parseIfStatement()

		} else {
			p.nextToken()
			ifStmt.Alternative = p.parseStatement()
		}
	}
	return ifStmt
}

// Parses an identifier such as a variable access or function call
func (p *Parser) parseIdentifier() ast.Expression {
	if !p.peekTokenIs(lexer.TokenLParen) {
		// not a function call so return the identifier
		return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal, Symbol: p.curFunction.GetSymbol(p.curToken.Literal)}
	}
	f := &ast.FunctionCallExpression{Token: p.curToken, Name: p.curToken.Literal}

	funcExists := false
	for _, definedFunc := range p.program.Functions {
		if definedFunc.Name == f.Name {
			f.ReturnType = definedFunc.ReturnType
			funcExists = true
		}
	}
	for i, definedFunc := range p.program.ExternalFunctions {
		if definedFunc.Name == f.Name {
			f.ReturnType = definedFunc.ReturnType
			f.Index = i
			funcExists = true
		}
	}

	if !funcExists {
		p.errors = append(p.errors, ParseError{
			message: fmt.Sprintf("unknown function '%s'", f.Name),
			line:    f.Token.Line,
		})
		return nil
	}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	if p.peekTokenIs(lexer.TokenRParen) {
		p.nextToken() // consume ')'
		return f
	}

	p.nextToken() // Move p.curToken to the first argument
	f.Args = append(f.Args, p.parseExpression(PrecendenceLowest))

	for p.peekTokenIs(lexer.TokenComma) {
		p.nextToken()
		p.nextToken()
		f.Args = append(f.Args, p.parseExpression(PrecendenceLowest))
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	return f
}

// Parses an expression such as 5 + 5, or foo()
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type, p.curToken.Line)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TokenSemicolon) && !p.peekTokenIs(lexer.TokenComma) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	_, ok := left.(*ast.Identifier)
	if !ok {
		p.errors = append(p.errors, ParseError{
			line:    p.curToken.Line,
			message: "cannot apply postfix operator on non variable",
		})
		return nil
	}

	expression := &ast.PostfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	return expression
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

func (p *Parser) parseCharLiteral() ast.Expression {
	lit := &ast.CharLiteral{Token: p.curToken}
	lit.Value = p.curToken.Literal[0]

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PrecedencePrefix)

	if p.curToken.Literal == "++" || p.curToken.Literal == "--" {
		_, ok := expression.Right.(*ast.Identifier)
		if !ok {
			p.errors = append(p.errors, ParseError{
				line:    p.curToken.Line,
				message: "cannot apply postfix operator on non variable",
			})
			return nil
		}
	}

	return expression
}

// Parses an expression contained within parenthesis ()
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // consume the '('

	exp := p.parseExpression(PrecendenceLowest)

	// Ensure the next token is ')'
	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	return exp
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

func (p *Parser) parseFunction(extern bool) *ast.Function {
	t := p.curToken.Literal
	p.nextToken()
	name := p.curToken.Literal
	function := ast.NewFunction(name, ast.ValueType(t))
	p.curFunction = function

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	// handle argument parsing
	if !p.peekTokenIs(lexer.TokenRParen) {
		p.nextToken()

		for {
			if !p.isTypeKeyword(p.curToken.Type) {
				p.errors = append(p.errors, ParseError{
					message: fmt.Sprintf("expected type, got %s instead", p.curToken.Literal),
					line:    p.curToken.Line,
				})
				return nil
			}
			argType := p.curToken.Literal
			if !p.expectPeek(lexer.TokenIdent) {
				return nil
			}
			name := p.curToken.Literal

			function.Arguments = append(function.Arguments, ast.Argument{
				Name: name,
				Type: ast.ValueType(argType),
			})
			function.SetSymbol(name, ast.ValueType(argType), false)

			if p.peekTokenIs(lexer.TokenComma) {
				// consume identifier and comma
				p.nextToken()
				p.nextToken()
			} else {
				break
			}
		}
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	// if the line ends with a semicolon, its just a function definition
	if !extern {
		if p.peekTokenIs(lexer.TokenSemicolon) {
			p.nextToken()
			return function
		}
	} else {
		if !p.expectPeek(lexer.TokenSemicolon) {
			return nil
		}
	}

	if !extern {
		p.nextToken()
		function.Statement = p.parseStatement()
	}

	return function
}

// Check if the token is the keyword for a type. Used for variable and function declarations
func (p *Parser) isTypeKeyword(t lexer.TokenType) bool {
	return t == lexer.TokenIntKeyword ||
		t == lexer.TokenFloatKeyword ||
		t == lexer.TokenCharKeyword || t == lexer.TokenVoid
}

// Parses the entire program, returning the abstract syntax tree
func (p *Parser) ParseProgram() *ast.Program {
	p.program = &ast.Program{}
	p.program.Functions = make([]*ast.Function, 0)

	for p.curToken.Type != lexer.TokenEndOfFile {
		extern := p.curTokenIs(lexer.TokenExtern)
		if extern {
			p.nextToken()
		}

		// functions are form [type] [identifier]()
		isFunc := p.isTypeKeyword(p.curToken.Type) &&
			p.peekTokenIs(lexer.TokenIdent) &&
			p.doublePeekTokenIs(lexer.TokenLParen)

		if isFunc {
			var function *ast.Function
			if extern {
				function = p.parseFunction(true)
				if function != nil {
					if idx := p.program.FunctionExists(function.Name); idx == -1 {
						p.program.ExternalFunctions = append(p.program.ExternalFunctions, function)
					} else {
						p.errors = append(p.errors, ParseError{
							message: "function overrides not supported",
							line:    p.curToken.Line,
						})
					}
				}
			} else {
				function = p.parseFunction(false)
				if function != nil {
					if idx := p.program.FunctionExists(function.Name); idx == -1 {
						p.program.Functions = append(p.program.Functions, function)
					} else {
						// handle functions that have been defined without bodies (in header files for example)
						p.program.Functions[idx].Statement = function.Statement
					}
				}
			}

		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				p.program.Statements = append(p.program.Statements, stmt)
			}
		}

		p.nextToken()
	}

	return p.program
}
