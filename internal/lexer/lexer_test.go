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
		{lexer.TokenEqual, "="},
		{lexer.TokenPlus, "+"},
		{lexer.TokenLParen, "("},
		{lexer.TokenRParen, ")"},
		{lexer.TokenLBrace, "{"},
		{lexer.TokenRBrace, "}"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenEndOfFile, ""},
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
		float x = 0.1 + 0.2;
		int y = 5;
		char z = 'a';

		return y;
	}`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.TokenIntKeyword, "int"},
		{lexer.TokenIdent, "main"},
		{lexer.TokenLParen, "("},
		{lexer.TokenRParen, ")"},
		{lexer.TokenLBrace, "{"},
		{lexer.TokenFloatKeyword, "float"},
		{lexer.TokenIdent, "x"},
		{lexer.TokenEqual, "="},
		{lexer.TokenFloatLiteral, "0.1"},
		{lexer.TokenPlus, "+"},
		{lexer.TokenFloatLiteral, "0.2"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenIntKeyword, "int"},
		{lexer.TokenIdent, "y"},
		{lexer.TokenEqual, "="},
		{lexer.TokenIntLiteral, "5"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenCharKeyword, "char"},
		{lexer.TokenIdent, "z"},
		{lexer.TokenEqual, "="},
		{lexer.TokenCharLiteral, "a"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenReturn, "return"},
		{lexer.TokenIdent, "y"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenRBrace, "}"},
		{lexer.TokenEndOfFile, ""},
	}
	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] failed: token type wrong", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] failed: token literal wrong", i)
	}
}

func TestTwoCharTokens(t *testing.T) {
	input := `int main() {
		int x = 10;
		x += 10;

		return x;
	}`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.TokenIntKeyword, "int"},
		{lexer.TokenIdent, "main"},
		{lexer.TokenLParen, "("},
		{lexer.TokenRParen, ")"},
		{lexer.TokenLBrace, "{"},
		{lexer.TokenIntKeyword, "int"},
		{lexer.TokenIdent, "x"},
		{lexer.TokenEqual, "="},
		{lexer.TokenIntLiteral, "10"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenIdent, "x"},
		{lexer.TokenPlusEqual, "+="},
		{lexer.TokenIntLiteral, "10"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenReturn, "return"},
		{lexer.TokenIdent, "x"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenRBrace, "}"},
		{lexer.TokenEndOfFile, ""},
	}
	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] failed: token type wrong", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] failed: token literal wrong", i)
	}
}

func TestComments(t *testing.T) {
	input := `// int main()
	int x = 5;`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.TokenIntKeyword, "int"},
		{lexer.TokenIdent, "x"},
		{lexer.TokenEqual, "="},
		{lexer.TokenIntLiteral, "5"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenEndOfFile, ""},
	}
	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] failed: token type wrong", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] failed: token literal wrong", i)
	}
}

func TestMultilineComments(t *testing.T) {
	input := `/*
	foobar, I am a comment
	*/
	int x = 5;`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.TokenIntKeyword, "int"},
		{lexer.TokenIdent, "x"},
		{lexer.TokenEqual, "="},
		{lexer.TokenIntLiteral, "5"},
		{lexer.TokenSemicolon, ";"},
		{lexer.TokenEndOfFile, ""},
	}
	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] failed: token type wrong", i)
		assert.Equal(t, tt.expectedLiteral, tok.Literal, "test[%d] failed: token literal wrong", i)
	}
}
