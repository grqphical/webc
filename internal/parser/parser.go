package parser

import (
	"fmt"

	"github.com/grqphical/webc/internal/lexer"
)

func isBinaryOperator(token lexer.Token) bool {
	return token.Type == lexer.TK_DASH || token.Type == lexer.TK_PLUS || token.Type == lexer.TK_SLASH || token.Type == lexer.TK_STAR
}

// Higher number = higher precedence (multiplication before addition)
func getPrecedence(token lexer.Token) int {
	switch token.Type {
	case lexer.TK_STAR, lexer.TK_SLASH:
		return 2
	case lexer.TK_PLUS, lexer.TK_DASH:
		return 1
	default:
		return 0
	}
}

type Symbol struct {
	Index int
	Type  string
}

type Node any

type Program struct {
	Functions []FunctionDecl
}

type FunctionDecl struct {
	Type            string
	Name            string
	Body            Block
	SymbolTable     map[string]Symbol
	NextSymbolIndex int
}

func (f *FunctionDecl) DefineSymbol(name string, typeName string) Symbol {
	sym := Symbol{
		Index: f.NextSymbolIndex,
		Type:  typeName,
	}
	f.SymbolTable[name] = sym
	f.NextSymbolIndex++
	return sym
}

func (f *FunctionDecl) GetSymbol(name string) (Symbol, bool) {
	sym, exists := f.SymbolTable[name]
	return sym, exists
}

type Block struct {
	Statements []Node
}

type ReturnStmt struct {
	Value Node
}

type VariableDefineStmt struct {
	Symbol Symbol
	Value  Node
}

type Constant struct {
	Value string
}

type VariableAccess struct {
	Index int
}

type UnaryExpression struct {
	Value     Node
	Operation string
}

type BinaryExpression struct {
	A         Node
	Operation string
	B         Node
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
		Type:            p.getCurrentToken().Literal,
		SymbolTable:     make(map[string]Symbol),
		NextSymbolIndex: 0,
	}
	p.head++

	funcDecl.Name = p.getCurrentToken().Literal
	p.head++

	if p.getCurrentToken().Type != lexer.TK_LPAREN {
		return FunctionDecl{}, fmt.Errorf("invalid token after function declaration, expected '(' got '%s' on line %d", p.getCurrentToken().Literal, p.getCurrentToken().Line)
	}
	p.head++
	if p.getCurrentToken().Type != lexer.TK_RPAREN {
		return FunctionDecl{}, fmt.Errorf("invalid token after '(' declaration, expected ')' on line %d", p.getCurrentToken().Line)
	}
	p.head++

	body, err := p.parseBlock(&funcDecl)
	if err != nil {
		return FunctionDecl{}, err
	}
	funcDecl.Body = body

	return funcDecl, nil

}

func (p *Parser) parseBlock(f *FunctionDecl) (Block, error) {
	block := Block{}

	// consume {
	p.head++

	for p.getCurrentToken().Type != lexer.TK_RBRACE && p.getCurrentToken().Type != lexer.TK_EOF {
		if p.getCurrentToken().Type == lexer.TK_SEMICOLON {
			p.head++
			continue
		}
		stmt, err := p.parseStatement(f)
		if err != nil {
			return Block{}, err
		}

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			return Block{}, fmt.Errorf("Unexpected token: %s on line %d\n", p.getCurrentToken().Literal, p.getCurrentToken().Line)
		}
	}

	return block, nil
}

func (p *Parser) parseStatement(f *FunctionDecl) (Node, error) {
	switch p.getCurrentToken().Literal {
	case "return":
		return p.parseReturn(f)
	case "int":
		return p.parseVarAssign(f)
	default:
		return nil, fmt.Errorf("invalid statement on line %d")
	}
}

func (p *Parser) parseReturn(f *FunctionDecl) (Node, error) {
	p.head++ // consume 'return'

	parsedExpr, err := p.parseExpression(f, 0)
	if err != nil {
		return nil, err
	}
	stmt := ReturnStmt{Value: parsedExpr}

	if p.getCurrentToken().Type != lexer.TK_SEMICOLON {
		return nil, fmt.Errorf("missing semicolon on line %d", p.getCurrentToken().Line)
	}
	p.head++

	return stmt, nil
}

func (p *Parser) parseVarAssign(f *FunctionDecl) (Node, error) {
	p.head++ // consume int (for now)

	name := p.getCurrentToken().Literal
	sym := f.DefineSymbol(name, "int")
	p.head++

	if p.getCurrentToken().Literal == "=" {
		p.head++ // consume '='
	}

	parsedExpr, err := p.parseExpression(f, 0)
	if err != nil {
		return nil, err
	}
	stmt := VariableDefineStmt{Value: parsedExpr, Symbol: sym}

	if p.getCurrentToken().Type != lexer.TK_SEMICOLON {
		return nil, fmt.Errorf("missing semicolon on line %d", p.getCurrentToken().Line)
	}
	p.head++
	return stmt, nil
}

func (p *Parser) parseExpression(f *FunctionDecl, minPrecedence int) (Node, error) {
	lhs, err := p.parsePrimary(f)
	if err != nil {
		return nil, err
	}

	for {
		opToken := p.getCurrentToken()

		if opToken.Type == lexer.TK_SEMICOLON || opToken.Type == lexer.TK_EOF || opToken.Type == lexer.TK_RPAREN {
			break
		}

		if !isBinaryOperator(opToken) {
			break
		}

		precedence := getPrecedence(opToken)

		// If the next operator has lower precedence than what we are currently
		// working on, return what we have so far.
		if precedence < minPrecedence {
			break
		}

		op := opToken.Literal
		p.head++

		// We pass precedence + 1 to ensure left-associativity
		// (so 1-2-3 parses as (1-2)-3, not 1-(2-3))
		rhs, err := p.parseExpression(f, precedence+1)
		if err != nil {
			return nil, err
		}

		lhs = BinaryExpression{
			A:         lhs,
			Operation: op,
			B:         rhs,
		}
	}

	return lhs, nil
}

func (p *Parser) parsePrimary(f *FunctionDecl) (Node, error) {
	token := p.getCurrentToken()

	switch token.Type {
	case lexer.TK_NUMBER:
		p.head++ // Consume the number
		return Constant{Value: token.Literal}, nil

	case lexer.TK_IDENT:
		p.head++ // Consume the identifier
		sym, exists := f.GetSymbol(token.Literal)
		if !exists {
			return nil, fmt.Errorf("undeclared variable '%s' on line %d", token.Literal, token.Line)
		}
		return VariableAccess{Index: sym.Index}, nil

	case lexer.TK_DASH:
		// Handle Unary Minus (e.g. -5)
		p.head++ // Consume '-'
		expr := UnaryExpression{Operation: "-"}

		// Parse the value being negated with high precedence
		// to ensure -5+2 parses as (-5)+2
		value, err := p.parseExpression(f, getPrecedence(token))
		if err != nil {
			return nil, err
		}
		expr.Value = value
		return expr, nil

	case lexer.TK_LPAREN:
		// Handle (expression)
		p.head++                             // consume '('
		expr, err := p.parseExpression(f, 0) // reset precedence inside parens
		if err != nil {
			return nil, err
		}
		if p.getCurrentToken().Type != lexer.TK_RPAREN {
			return nil, fmt.Errorf("expected ')', got '%s' on line %d", p.getCurrentToken().Literal, p.getCurrentToken().Line)
		}
		p.head++ // consume ')'
		return expr, nil

	default:
		return nil, fmt.Errorf("Unexpected token in expression: '%s' on line %d\n", token.Literal, token.Line)
	}
}

func (p *Parser) Parse() (Program, error) {
	prog := Program{
		Functions: make([]FunctionDecl, 0),
	}

	for p.head < len(p.tokens) {
		tok := p.getCurrentToken()
		if tok.Type == lexer.TK_KEYWORD || tok.Type == lexer.TK_IDENT {
			f, err := p.parseFunction()
			if err != nil {
				return Program{}, err
			}

			prog.Functions = append(prog.Functions, f)
		} else {
			p.head++
		}
	}

	return prog, nil
}
