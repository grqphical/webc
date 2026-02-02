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
