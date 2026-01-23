package parser

import (
	"errors"
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
		return FunctionDecl{}, fmt.Errorf("invalid token after function declaration, expected '(' got '%s'", p.getCurrentToken().Literal)
	}
	p.head++
	if p.getCurrentToken().Type != lexer.TK_RPAREN {
		return FunctionDecl{}, errors.New("invalid token after '(' declaration, expected ')'")
	}
	p.head++

	funcDecl.Body = p.parseBlock(&funcDecl)

	return funcDecl, nil

}

func (p *Parser) parseBlock(f *FunctionDecl) Block {
	block := Block{}

	// consume {
	p.head++

	for p.getCurrentToken().Type != lexer.TK_RBRACE && p.getCurrentToken().Type != lexer.TK_EOF {
		if p.getCurrentToken().Type == lexer.TK_SEMICOLON {
			p.head++
			continue
		}
		stmt := p.parseStatement(f)

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			fmt.Printf("Unexpected token: %s\n", p.getCurrentToken().Literal)
			p.head++
		}
	}

	return block
}

func (p *Parser) parseStatement(f *FunctionDecl) Node {
	switch p.getCurrentToken().Literal {
	case "return":
		return p.parseReturn(f)
	case "int":
		return p.parseVarAssign(f)
	default:
		return nil
	}
}

func (p *Parser) parseReturn(f *FunctionDecl) Node {
	p.head++ // consume 'return'

	stmt := ReturnStmt{Value: p.parseExpression(f, 0)}

	if p.getCurrentToken().Type == lexer.TK_SEMICOLON {
		p.head++
	}

	return stmt
}

func (p *Parser) parseVarAssign(f *FunctionDecl) Node {
	p.head++ // consume int (for now)

	name := p.getCurrentToken().Literal
	sym := f.DefineSymbol(name, "int")
	p.head++

	if p.getCurrentToken().Literal == "=" {
		p.head++ // consume '='
	}

	stmt := VariableDefineStmt{Value: p.parseExpression(f, 0), Symbol: sym}

	if p.getCurrentToken().Type == lexer.TK_SEMICOLON {
		p.head++
	}
	return stmt
}

func (p *Parser) parseExpression(f *FunctionDecl, minPrecedence int) Node {
	lhs := p.parsePrimary(f)

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
		rhs := p.parseExpression(f, precedence+1)

		lhs = BinaryExpression{
			A:         lhs,
			Operation: op,
			B:         rhs,
		}
	}

	return lhs
}

func (p *Parser) parsePrimary(f *FunctionDecl) Node {
	token := p.getCurrentToken()

	switch token.Type {
	case lexer.TK_NUMBER:
		p.head++ // Consume the number
		return Constant{Value: token.Literal}

	case lexer.TK_IDENT:
		p.head++ // Consume the identifier
		sym, exists := f.GetSymbol(token.Literal)
		if !exists {
			fmt.Printf("error: variable '%s' is not defined\n", token.Literal)
			return nil
		}
		return VariableAccess{Index: sym.Index}

	case lexer.TK_DASH:
		// Handle Unary Minus (e.g. -5)
		p.head++ // Consume '-'
		expr := UnaryExpression{Operation: "-"}

		// Parse the value being negated with high precedence
		// to ensure -5+2 parses as (-5)+2
		expr.Value = p.parseExpression(f, getPrecedence(token))
		return expr

	case lexer.TK_LPAREN:
		// Handle (expression)
		p.head++                        // consume '('
		expr := p.parseExpression(f, 0) // reset precedence inside parens
		if p.getCurrentToken().Type != lexer.TK_RPAREN {
			fmt.Printf("expected ')'\n")
			return nil
		}
		p.head++ // consume ')'
		return expr

	default:
		fmt.Printf("Unexpected token in expression: %s\n", token.Literal)
		return nil
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
