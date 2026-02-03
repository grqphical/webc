package lexer_test

import (
	"testing"

	"github.com/grqphical/webc/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func TestMainFunction(t *testing.T) {
	exampleCode := `int main() {}`
	expectedOutput := []lexer.Token{
		{
			Type:    lexer.TK_KEYWORD,
			Literal: "int",
			Line:    1,
		},
		{
			Type:    lexer.TK_IDENT,
			Literal: "main",
			Line:    1,
		},
		{
			Type:    lexer.TK_LPAREN,
			Literal: "(",
			Line:    1,
		},
		{
			Type:    lexer.TK_RPAREN,
			Literal: ")",
			Line:    1,
		},
		{
			Type:    lexer.TK_LBRACE,
			Literal: "{",
			Line:    1,
		},
		{
			Type:    lexer.TK_RBRACE,
			Literal: "}",
			Line:    1,
		},
		{
			Type:    lexer.TK_EOF,
			Literal: "EOF",
			Line:    1,
		},
	}

	l := lexer.New(exampleCode)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)
	assert.ElementsMatch(t, tokens, expectedOutput)
}

func TestIntegerVariableDeclarations(t *testing.T) {
	exampleCode := "int x = 100;"
	expectedOutput := []lexer.Token{
		{
			Type:    lexer.TK_KEYWORD,
			Literal: "int",
			Line:    1,
		},
		{
			Type:    lexer.TK_IDENT,
			Literal: "x",
			Line:    1,
		},
		{
			Type:    lexer.TK_EQUAL,
			Literal: "=",
			Line:    1,
		},
		{
			Type:    lexer.TK_INTEGER,
			Literal: "100",
			Line:    1,
		},
		{
			Type:    lexer.TK_SEMICOLON,
			Literal: ";",
			Line:    1,
		},
		{
			Type:    lexer.TK_EOF,
			Literal: "EOF",
			Line:    1,
		},
	}

	l := lexer.New(exampleCode)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)
	assert.ElementsMatch(t, tokens, expectedOutput)
}

func TestFloatVariableDeclarations(t *testing.T) {
	exampleCode := "float pi = 3.14159;"
	expectedOutput := []lexer.Token{
		{
			Type:    lexer.TK_KEYWORD,
			Literal: "float",
			Line:    1,
		},
		{
			Type:    lexer.TK_IDENT,
			Literal: "pi",
			Line:    1,
		},
		{
			Type:    lexer.TK_EQUAL,
			Literal: "=",
			Line:    1,
		},
		{
			Type:    lexer.TK_FLOAT,
			Literal: "3.14159",
			Line:    1,
		},
		{
			Type:    lexer.TK_SEMICOLON,
			Literal: ";",
			Line:    1,
		},
		{
			Type:    lexer.TK_EOF,
			Literal: "EOF",
			Line:    1,
		},
	}

	l := lexer.New(exampleCode)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)
	assert.ElementsMatch(t, tokens, expectedOutput)
}

func TestCharVariableDeclarations(t *testing.T) {
	exampleCode := "char x = 'a';"
	expectedOutput := []lexer.Token{
		{
			Type:    lexer.TK_KEYWORD,
			Literal: "char",
			Line:    1,
		},
		{
			Type:    lexer.TK_IDENT,
			Literal: "x",
			Line:    1,
		},
		{
			Type:    lexer.TK_EQUAL,
			Literal: "=",
			Line:    1,
		},
		{
			Type:    lexer.TK_CHAR_LITERAL,
			Literal: "a",
			Line:    1,
		},
		{
			Type:    lexer.TK_SEMICOLON,
			Literal: ";",
			Line:    1,
		},
		{
			Type:    lexer.TK_EOF,
			Literal: "EOF",
			Line:    1,
		},
	}

	l := lexer.New(exampleCode)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)
	assert.ElementsMatch(t, tokens, expectedOutput)
}

func TestIllegalToken(t *testing.T) {
	illegalChar := "$ foo = 10;"

	l := lexer.New(illegalChar)
	_, err := l.ParseSource()
	assert.Error(t, err)
}

func TestUnterminatedChar(t *testing.T) {
	unterminatedCharLiteral := "'a"

	l := lexer.New(unterminatedCharLiteral)
	_, err := l.ParseSource()
	assert.Error(t, err)
}

func TestInvalidNumber(t *testing.T) {
	invalidNumber := "9.8.1"

	l := lexer.New(invalidNumber)
	_, err := l.ParseSource()
	assert.Error(t, err)
}

func TestComments(t *testing.T) {
	exampleCode := "// This should not be counted\nint foo = 3;"
	expectedOutput := []lexer.Token{
		{
			Type:    lexer.TK_KEYWORD,
			Literal: "int",
			Line:    2,
		},
		{
			Type:    lexer.TK_IDENT,
			Literal: "foo",
			Line:    2,
		},
		{
			Type:    lexer.TK_EQUAL,
			Literal: "=",
			Line:    2,
		},
		{
			Type:    lexer.TK_INTEGER,
			Literal: "3",
			Line:    2,
		},
		{
			Type:    lexer.TK_SEMICOLON,
			Literal: ";",
			Line:    2,
		},
		{
			Type:    lexer.TK_EOF,
			Literal: "EOF",
			Line:    2,
		},
	}

	l := lexer.New(exampleCode)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)
	assert.ElementsMatch(t, tokens, expectedOutput)
}

func TestTwoCharTokens(t *testing.T) {
	exampleCode := "+ += - -= * *= / /="
	expectedOutput := []lexer.Token{
		{
			Type:    lexer.TK_PLUS,
			Literal: "+",
			Line:    1,
		},
		{
			Type:    lexer.TK_PLUS_EQUAL,
			Literal: "+=",
			Line:    1,
		},
		{
			Type:    lexer.TK_DASH,
			Literal: "-",
			Line:    1,
		},
		{
			Type:    lexer.TK_MINUS_EQUAL,
			Literal: "-=",
			Line:    1,
		},
		{
			Type:    lexer.TK_STAR,
			Literal: "*",
			Line:    1,
		},
		{
			Type:    lexer.TK_TIMES_EQUAL,
			Literal: "*=",
			Line:    1,
		},
		{
			Type:    lexer.TK_SLASH,
			Literal: "/",
			Line:    1,
		},
		{
			Type:    lexer.TK_DIVIDE_EQUAL,
			Literal: "/=",
			Line:    1,
		},
		{
			Type:    lexer.TK_EOF,
			Literal: "EOF",
			Line:    1,
		},
	}

	l := lexer.New(exampleCode)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)
	assert.ElementsMatch(t, tokens, expectedOutput)
}
