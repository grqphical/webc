package lexer

import (
	"fmt"
	"strings"
)

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNumber(c byte) bool {
	return c >= '0' && c <= '9'
}

type LexerError struct {
	Message string
	Line    int
}

func (l LexerError) Error() string {
	return fmt.Sprintf("lexical error on line %d: %s", l.Line, l.Message)
}

type TokenType string

const (
	TK_KEYWORD   TokenType = "KEYWORD"
	TK_IDENT     TokenType = "IDENTIFIER"
	TK_NUMBER    TokenType = "NUMBER"
	TK_LPAREN    TokenType = "("
	TK_RPAREN    TokenType = ")"
	TK_LBRACE    TokenType = "{"
	TK_RBRACE    TokenType = "}"
	TK_SEMICOLON TokenType = ";"
	TK_EQUAL     TokenType = "="
	TK_DASH      TokenType = "-"
	TK_EOF       TokenType = "EOF"
)

var keywords map[string]any = map[string]any{
	"int":    nil,
	"return": nil,
}

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	source    string
	lineCount int
	head      int
	tokens    []Token
}

func New(source string) *Lexer {
	return &Lexer{
		source:    source,
		lineCount: 1,
		head:      0,
		tokens:    make([]Token, 0),
	}
}

func (l *Lexer) getCurrentChar() byte {
	return l.source[l.head]
}

func (l *Lexer) makeLiteral() {
	literal := ""

	for isLetter(l.getCurrentChar()) {
		literal += string(l.getCurrentChar())
		l.head++
	}

	var tokenType TokenType
	if _, exists := keywords[literal]; exists {
		tokenType = TK_KEYWORD
	} else {
		tokenType = TK_IDENT
	}

	l.tokens = append(l.tokens, Token{
		Type:    tokenType,
		Literal: literal,
	})

}

func (l *Lexer) makeNumber() {
	var literal strings.Builder

	for isNumber(l.getCurrentChar()) {
		literal.WriteString(string(l.getCurrentChar()))
		l.head++
	}

	l.tokens = append(l.tokens, Token{
		Type:    TK_NUMBER,
		Literal: literal.String(),
	})
}

func (l *Lexer) ParseSource() ([]Token, error) {
	for l.head < len(l.source) {
		tok := l.getCurrentChar()

		switch tok {
		case ' ', '\t':
			l.head++

		case '\n':
			l.lineCount += 1
			l.head++
		case '{':
			l.tokens = append(l.tokens, Token{
				Type:    TK_LBRACE,
				Literal: "{",
			})
			l.head++
		case '}':
			l.tokens = append(l.tokens, Token{
				Type:    TK_RBRACE,
				Literal: "}",
			})
			l.head++
		case '(':
			l.tokens = append(l.tokens, Token{
				Type:    TK_LPAREN,
				Literal: "(",
			})
			l.head++
		case ')':
			l.tokens = append(l.tokens, Token{
				Type:    TK_RPAREN,
				Literal: ")",
			})
			l.head++
		case ';':
			l.tokens = append(l.tokens, Token{
				Type:    TK_SEMICOLON,
				Literal: ";",
			})
			l.head++
		case '=':
			l.tokens = append(l.tokens, Token{
				Type:    TK_EQUAL,
				Literal: "=",
			})
			l.head++
		case '-':
			l.tokens = append(l.tokens, Token{
				Type:    TK_DASH,
				Literal: "-",
			})
			l.head++

		default:
			if isLetter(tok) {
				l.makeLiteral()
			} else if isNumber(tok) {
				l.makeNumber()
			} else {
				return nil, LexerError{
					Line:    l.lineCount,
					Message: fmt.Sprintf("illegal token '%c'", tok),
				}
			}
		}

	}

	l.tokens = append(l.tokens, Token{
		Type:    TK_EOF,
		Literal: "EOF",
	})

	return l.tokens, nil
}
