package parser

import (
	"errors"
	"fmt"

	"github.com/grqphical/webc/internal/lexer"
)

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

func (f *FunctionDecl) GetSymbol(name string) Symbol {
	return f.SymbolTable[name]
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

func (p *Parser) peekToken() lexer.Token {
	if p.head+1 == len(p.tokens) {
		return lexer.Token{}
	}
	return p.tokens[p.head+1]
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
		block.Statements = append(block.Statements, stmt)
		p.head++
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

	stmt := ReturnStmt{Value: p.parseExpression(f)}
	p.head++ // consume return value

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

	stmt := VariableDefineStmt{Value: p.parseExpression(f), Symbol: sym}

	if p.getCurrentToken().Type == lexer.TK_SEMICOLON {
		p.head++
	}
	return stmt
}

func (p *Parser) parseExpression(f *FunctionDecl) Node {
	currentToken := p.getCurrentToken()

	switch currentToken.Type {
	case lexer.TK_NUMBER:
		if p.peekToken().Type == lexer.TK_SEMICOLON {
			return Constant{Value: currentToken.Literal}
		} else if p.peekToken().Type == lexer.TK_DASH {
			a := Constant{Value: currentToken.Literal}
			p.head++

			operation := p.getCurrentToken().Literal
			if p.peekToken().Type != lexer.TK_IDENT && p.peekToken().Type != lexer.TK_NUMBER {
				fmt.Printf("error parsing expression, expected number or identifier\n")
				return nil
			}
			p.head++

			b := p.parseExpression(f)
			return BinaryExpression{
				a,
				operation,
				b,
			}
		}
	case lexer.TK_IDENT:
		if p.peekToken().Type == lexer.TK_SEMICOLON {
			return VariableAccess{Index: f.GetSymbol(currentToken.Literal).Index}
		} else if p.peekToken().Type == lexer.TK_DASH {
			a := VariableAccess{Index: f.GetSymbol(currentToken.Literal).Index}
			p.head++

			operation := p.getCurrentToken().Literal
			if p.peekToken().Type != lexer.TK_IDENT && p.peekToken().Type != lexer.TK_NUMBER {
				fmt.Printf("error parsing expression, expected number or identifier\n")
				return nil
			}
			p.head++

			b := p.parseExpression(f)
			return BinaryExpression{
				a,
				operation,
				b,
			}
		}
	case lexer.TK_DASH:
		expr := UnaryExpression{Operation: "-"}
		p.head++
		expr.Value = p.parseExpression(f)
		return expr
	}

	return nil
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
