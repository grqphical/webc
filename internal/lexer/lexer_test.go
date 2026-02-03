package lexer_test

import (
	"testing"

	"github.com/grqphical/webc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func TestSingleTokens(t *testing.T) {
	input := `=+(){};`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.TK_EQUAL, "="},
		{lexer.TK_PLUS, "+"},
		{lexer.TK_LPAREN, "("},
		{lexer.TK_RPAREN, ")"},
		{lexer.TK_LBRACE, "{"},
		{lexer.TK_RBRACE, "}"},
		{lexer.TK_SEMICOLON, ";"},
		{lexer.TK_EOF, ""},
	}
	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] failed: token type wrong", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] failed: token literal wrong", i)
	}
}

func TestFunctionTokenization(t *testing.T) {
	input := `int main() {
		float x = 0.1;
		int y = 5;

		return y;
	}`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.TK_INT, "int"},
		{lexer.TK_IDENT, "main"},
		{lexer.TK_LPAREN, "("},
		{lexer.TK_RPAREN, ")"},
		{lexer.TK_LBRACE, "{"},
		{lexer.TK_FLOAT, "float"},
		{lexer.TK_IDENT, "x"},
		{lexer.TK_EQUAL, "="},
		{lexer.TK_FLOAT_LITERAL, "0.1"},
		{lexer.TK_SEMICOLON, ";"},
		{lexer.TK_INT, "int"},
		{lexer.TK_IDENT, "y"},
		{lexer.TK_EQUAL, "="},
		{lexer.TK_INTEGER_LITERAL, "5"},
		{lexer.TK_SEMICOLON, ";"},
		{lexer.TK_RETURN, "return"},
		{lexer.TK_IDENT, "y"},
		{lexer.TK_SEMICOLON, ";"},
		{lexer.TK_RBRACE, "}"},
		{lexer.TK_EOF, ""},
	}
	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] failed: token type wrong", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] failed: token literal wrong", i)
	}
}
