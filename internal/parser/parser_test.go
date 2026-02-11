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
	char z = 'a';`

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.Empty(t, p.Errors())
	assert.NotNil(t, program)
	assert.Equal(t, len(program.Statements), 3, "should have three statements")

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"z"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]

		assert.Contains(t, []string{"float", "int", "char"}, stmt.TokenLiteral(), "statement isn't type of int, float, or char")

		defineStmt, ok := stmt.(*ast.VariableDefineStatement)
		assert.Truef(t, ok, "could not cast statement to *ast.VariableDefineStatement, got %T instead", stmt)

		assert.Equal(t, tt.expectedIdentifier, defineStmt.Name.Value, "statement names are not equal")
		assert.Equal(t, tt.expectedIdentifier, defineStmt.Name.TokenLiteral(), "statement name token literal not equal")
	}

}

func TestReturnStatement(t *testing.T) {
	input := `return 5;
	return 10;
	return 10.58;
	return 'b';`

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, len(program.Statements), 4, "should have four statements")

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
	assert.Truef(t, ok, "could not cast expression to Identifier got %T instead", stmt.Expression)
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

func TestFloatLiteralExpression(t *testing.T) {
	input := "5.4;"

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, 1, len(program.Statements), "should have one statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "could not cast statement to ExpressionStatement")

	literal, ok := stmt.Expression.(*ast.FloatLiteral)
	assert.True(t, ok, "could not cast expression to FloatLiteral")
	assert.Equal(t, "5.4", literal.TokenLiteral())
}

func TestCharLiteralExpression(t *testing.T) {
	input := "'z';"

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Equal(t, 1, len(program.Statements), "should have one statement")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "could not cast statement to ExpressionStatement")

	literal, ok := stmt.Expression.(*ast.CharLiteral)
	assert.True(t, ok, "could not cast expression to CharLiteral")
	assert.Equal(t, "z", literal.TokenLiteral())
}

func TestIntegerPrefixOperators(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"-5", "-", 5},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		assert.NotNil(t, program)

		assert.Equal(t, 1, len(program.Statements), "should have one statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "could not cast statement to ExpressionStatement")

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(t, ok, "could not cast expression to PrefixExpression")
		assert.Equal(t, tt.operator, exp.Operator)

		value, ok := exp.Right.(*ast.IntegerLiteral)
		assert.True(t, ok, "could not cast value to IntegerLiteral")
		assert.Equal(t, tt.integerValue, value.Value)
	}
}

func TestFloatPrefixOperators(t *testing.T) {
	prefixTests := []struct {
		input      string
		operator   string
		floatValue float64
	}{
		{"-5.1", "-", 5.1},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		assert.NotNil(t, program)

		assert.Equal(t, 1, len(program.Statements), "should have one statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "could not cast statement to ExpressionStatement")

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(t, ok, "could not cast expression to PrefixExpression")
		assert.Equal(t, tt.operator, exp.Operator)

		value, ok := exp.Right.(*ast.FloatLiteral)
		assert.True(t, ok, "could not cast value to FloatLiteral")
		assert.Equal(t, tt.floatValue, value.Value)
	}
}

func TestIntegerInfixOperators(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"8 + 2;", 8, "+", 2},
		{"8 - 2;", 8, "-", 2},
		{"8 * 2;", 8, "*", 2},
		{"8 / 2;", 8, "/", 2},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		assert.NotNil(t, program)

		assert.Equal(t, 1, len(program.Statements), "should have one statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "could not cast statement to ExpressionStatement")

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		assert.True(t, ok, "could not cast expression to InfixExpression")
		assert.Equal(t, tt.operator, exp.Operator)

		leftValue, ok := exp.Left.(*ast.IntegerLiteral)
		assert.True(t, ok, "could not cast value to IntegerLiteral")
		assert.Equal(t, tt.leftValue, leftValue.Value)

		rightValue, ok := exp.Right.(*ast.IntegerLiteral)
		assert.True(t, ok, "could not cast value to IntegerLiteral")
		assert.Equal(t, tt.rightValue, rightValue.Value)
	}
}

func TestFloatInfixOperators(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  float64
		operator   string
		rightValue float64
	}{
		{"8.1 + 2.1;", 8.1, "+", 2.1},
		{"8.1 - 2.1;", 8.1, "-", 2.1},
		{"8.1 * 2.1;", 8.1, "*", 2.1},
		{"8.1 / 2.1;", 8.1, "/", 2.1},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		assert.NotNil(t, program)

		assert.Equal(t, 1, len(program.Statements), "should have one statement")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "could not cast statement to ExpressionStatement")

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		assert.True(t, ok, "could not cast expression to InfixExpression")
		assert.Equal(t, tt.operator, exp.Operator)

		leftValue, ok := exp.Left.(*ast.FloatLiteral)
		assert.True(t, ok, "could not cast value to IntegerLiteral")
		assert.Equal(t, tt.leftValue, leftValue.Value)

		rightValue, ok := exp.Right.(*ast.FloatLiteral)
		assert.True(t, ok, "could not cast value to IntegerLiteral")
		assert.Equal(t, tt.rightValue, rightValue.Value)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		}, {
			"a + b - c",
			"((a + b) - c)",
		}, {
			"a * b * c",
			"((a * b) * c)",
		}, {
			"a * b / c",
			"((a * b) / c)",
		}, {
			"a + b / c",
			"(a + (b / c))",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		assert.NotNil(t, program)

		actual := program.String()
		assert.Equal(t, tt.expected, actual)
	}

}

func TestFunctionDeclarations(t *testing.T) {
	input := `int main(){int x = 5;}`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	assert.NotNil(t, program)

	assert.Equal(t, 1, len(program.Functions), "didnt get one function declaration")
	assert.Equal(t, "main", program.Functions[0].Name)
	assert.Equal(t, 1, len(program.Functions[0].Statements), "did not get one statement")

	_, ok := program.Functions[0].Statements[0].(*ast.VariableDefineStatement)
	assert.True(t, ok, "cannot cast statement to VariableDefineStatement")
}

func TestVariableUpdate(t *testing.T) {
	input := `int x;
	x = 5;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	assert.NotNil(t, program)

	assert.Equal(t, 2, len(program.Statements), "expected two statements")

}

func TestExternFunction(t *testing.T) {
	input := `extern void foo();`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Empty(t, p.Errors())

	assert.Equal(t, 1, len(program.ExternalFunctions), "expected one external function")
	assert.Equal(t, 0, len(program.Functions), "expected zero functions")
}

func TestFunctionArguments(t *testing.T) {
	input := `void foo(int a, float b);`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	assert.NotNil(t, program)
	assert.Empty(t, p.Errors())

	assert.Equal(t, 1, len(program.Functions), "expected one function")
}
