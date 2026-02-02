package parser

import (
	"fmt"

	"github.com/grqphical/webc/internal/lexer"
)

type ValueType = string

const (
	TypeInt   ValueType = "int"
	TypeFloat ValueType = "float"
	TypeChar  ValueType = "char"
)

func isCompatibleOperationType(a ValueType, b ValueType) bool {
	if a == b {
		return true
	}

	switch a {
	case TypeInt:
		return b == TypeChar
	case TypeChar:
		return b == TypeInt
	default:
		return false
	}
}

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
	Index    int
	Type     ValueType
	Constant bool
}

type Node interface {
	GetType() ValueType
}

type Program struct {
	Functions []FunctionDecl
}

type FunctionDecl struct {
	Type            ValueType
	Name            string
	Body            Block
	SymbolTable     map[string]Symbol
	NextSymbolIndex int
}

func (f *FunctionDecl) DefineSymbol(name string, typeName ValueType, constant bool) Symbol {
	sym := Symbol{
		Index:    f.NextSymbolIndex,
		Type:     typeName,
		Constant: constant,
	}
	f.SymbolTable[name] = sym
	f.NextSymbolIndex++
	return sym
}

func (f *FunctionDecl) GetSymbol(name string) (Symbol, bool) {
	sym, exists := f.SymbolTable[name]
	return sym, exists
}

func (f *FunctionDecl) GetVariableCounts() (intCount int, floatCount int) {
	floatCount = 0
	intCount = 0

	for _, sym := range f.SymbolTable {
		switch sym.Type {
		case TypeInt:
			intCount++
		case TypeFloat:
			floatCount++
		case TypeChar:
			intCount++
		}
	}

	return
}

type Block struct {
	Statements []Node
}

type ReturnStmt struct {
	Value Node
}

func (r ReturnStmt) GetType() ValueType {
	return r.Value.GetType()
}

type VariableDefineStmt struct {
	Symbol   Symbol
	Value    Node
	Type     ValueType
	Constant bool
}

func (v VariableDefineStmt) GetType() ValueType {
	return v.GetType()
}

type Constant struct {
	Value string
	Type  ValueType
}

func (c Constant) GetType() ValueType {
	return c.Type
}

type VariableAccess struct {
	Index int
	Type  ValueType
}

func (v VariableAccess) GetType() ValueType {
	return v.Type
}

type VariableUpdateStmt struct {
	Index int
	Value Node
	Type  ValueType
}

func (v VariableUpdateStmt) GetType() ValueType {
	return v.Type
}

type UnaryExpression struct {
	Value     Node
	Operation string
}

func (u UnaryExpression) GetType() ValueType {
	return u.Value.GetType()
}

type BinaryExpression struct {
	A         Node
	Operation string
	B         Node
}

func (b BinaryExpression) GetType() ValueType {
	typeA := b.A.GetType()
	typeB := b.B.GetType()

	if typeA == TypeFloat || typeB == TypeFloat {
		return TypeFloat
	}
	return TypeInt
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
	if p.getCurrentToken().Type == lexer.TK_IDENT {
		// variable update statement
		return p.parseVariableUpdate(f)
	}

	switch p.getCurrentToken().Literal {
	case "return":
		return p.parseReturn(f)
	case "int", "float", "char", "const":
		return p.parseVarAssign(f)
	default:
		return nil, fmt.Errorf("invalid statement on line %d", p.getCurrentToken().Line)
	}
}

func (p *Parser) parseVariableUpdate(f *FunctionDecl) (Node, error) {
	name := p.getCurrentToken().Literal

	sym, exists := f.GetSymbol(name)
	if !exists {
		return nil, fmt.Errorf("no variable defined named '%s' on line %d", name, p.getCurrentToken().Line)
	}

	if sym.Constant {
		return nil, fmt.Errorf("cannot modify constant '%s' on line %d", name, p.getCurrentToken().Line)
	}

	p.head++ // consume variable name

	operatorToken := p.getCurrentToken()

	stmt := VariableUpdateStmt{
		Index: sym.Index,
		Type:  sym.Type,
	}
	p.head++ // consume assignment operator

	value, err := p.parseExpression(f, 0)
	if err != nil {
		return nil, err
	}

	switch operatorToken.Type {
	case lexer.TK_EQUAL:
		stmt.Value = value
	case lexer.TK_PLUS_EQUAL:
		stmt.Value = BinaryExpression{
			A: VariableAccess{
				Index: sym.Index,
				Type:  sym.Type,
			},
			Operation: "+",

			B: value,
		}
	case lexer.TK_MINUS_EQUAL:
		stmt.Value = BinaryExpression{
			A: VariableAccess{
				Index: sym.Index,
				Type:  sym.Type,
			},
			Operation: "-",

			B: value,
		}
	case lexer.TK_TIMES_EQUAL:
		stmt.Value = BinaryExpression{
			A: VariableAccess{
				Index: sym.Index,
				Type:  sym.Type,
			},
			Operation: "*",

			B: value,
		}
	case lexer.TK_DIVIDE_EQUAL:
		stmt.Value = BinaryExpression{
			A: VariableAccess{
				Index: sym.Index,
				Type:  sym.Type,
			},
			Operation: "/",

			B: value,
		}
	default:
		return nil, fmt.Errorf("invalid assignment operator on line %d", p.getCurrentToken().Line)
	}

	return stmt, nil
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
	constant := false

	if p.getCurrentToken().Literal == "const" {
		p.head++
		constant = true
	}

	valueType := p.getCurrentToken().Literal
	p.head++

	name := p.getCurrentToken().Literal
	sym := f.DefineSymbol(name, valueType, constant)
	p.head++

	if p.getCurrentToken().Literal == "=" {
		p.head++ // consume '='
	}

	parsedExpr, err := p.parseExpression(f, 0)
	if err != nil {
		return nil, err
	}
	if !isCompatibleOperationType(parsedExpr.GetType(), valueType) {
		return nil, fmt.Errorf("cannot assign %s to %s on line %d", parsedExpr.GetType(), valueType, p.getCurrentToken().Line)
	}

	stmt := VariableDefineStmt{Value: parsedExpr, Symbol: sym, Type: valueType, Constant: constant}

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

		if !isCompatibleOperationType(lhs.GetType(), rhs.GetType()) {
			return nil, fmt.Errorf("cannot do binary expression between %s and %s on line %d", lhs.GetType(), rhs.GetType(), p.getCurrentToken().Line)
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
	case lexer.TK_INTEGER:
		p.head++
		return Constant{Value: token.Literal, Type: TypeInt}, nil
	case lexer.TK_FLOAT:
		p.head++
		return Constant{Value: token.Literal, Type: TypeFloat}, nil
	case lexer.TK_CHAR_LITERAL:
		p.head++
		return Constant{Value: token.Literal, Type: TypeChar}, nil
	case lexer.TK_IDENT:
		p.head++ // Consume the identifier
		sym, exists := f.GetSymbol(token.Literal)
		if !exists {
			return nil, fmt.Errorf("undeclared variable '%s' on line %d", token.Literal, token.Line)
		}
		return VariableAccess{Index: sym.Index, Type: sym.Type}, nil

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
