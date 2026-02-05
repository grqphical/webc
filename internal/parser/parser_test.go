package parser_test

import (
	"testing"

	"github.com/grqphical/webc/internal/ast"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestVariableDefineStatement(t *testing.T) {
	input := `int x = 5;
	float y = 5.6;
	char foobar = 'b';`

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, len(program.Statements), 3, "should have three statements")

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]

		assert.Contains(t, []string{"float", "int", "char"}, stmt.TokenLiteral(), "statement isn't type of int, float, or char")

		defineStmt, ok := stmt.(*ast.VariableDefineStatement)
		assert.True(t, ok, "could not cast statement to *ast.VariableDefineStatement")

		assert.Equal(t, tt.expectedIdentifier, defineStmt.Name.Value, "statement names are not equal")
		assert.Equal(t, tt.expectedIdentifier, defineStmt.Name.TokenLiteral(), "statement name token literal not equal")
	}

}

func TestReturnStatement(t *testing.T) {
	input := `return 5;
	return 10;
	return 10.58;`

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, len(program.Statements), 3, "should have three statements")

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		assert.True(t, ok, "could not cast statement to ReturnStatement")
		assert.Equal(t, "return", returnStmt.TokenLiteral(), "token literal not equal to 'return'")
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, 1, len(program.Statements), "should have one statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "could not cast statement to ExpressionStatement")

	ident, ok := stmt.Expression.(*ast.Identifier)
	assert.True(t, ok, "could not cast expression to Identifier")
	assert.Equal(t, "foobar", ident.Value)
	assert.Equal(t, "foobar", ident.TokenLiteral())
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "12345;"

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, 1, len(program.Statements), "should have one statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "could not cast statement to ExpressionStatement")

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	assert.True(t, ok, "could not cast expression to IntegerLiteral")
	assert.Equal(t, "12345", literal.TokenLiteral())
}
