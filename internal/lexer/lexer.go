package lexer

import (
	"errors"
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
	TK_KEYWORD TokenType = "KEYWORD"
	TK_IDENT   TokenType = "IDENTIFIER"

	TK_INTEGER TokenType = "INTEGER"
	TK_FLOAT   TokenType = "FLOAT"

	TK_LPAREN TokenType = "("
	TK_RPAREN TokenType = ")"
	TK_LBRACE TokenType = "{"
	TK_RBRACE TokenType = "}"

	TK_SEMICOLON TokenType = ";"

	TK_EQUAL TokenType = "="
	TK_DASH  TokenType = "-"
	TK_PLUS  TokenType = "+"
	TK_STAR  TokenType = "*"
	TK_SLASH TokenType = "/"

	TK_EOF TokenType = "EOF"
)

var keywords map[string]any = map[string]any{
	"int":    nil,
	"float":  nil,
	"return": nil,
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
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
		Line:    l.lineCount,
	})

}

func (l *Lexer) makeNumber() error {
	var literal strings.Builder
	dotCount := 0

	for isNumber(l.getCurrentChar()) || l.getCurrentChar() == '.' {
		if l.getCurrentChar() == '.' {
			if dotCount != 0 {
				return errors.New("invalid number literal")
			}
			dotCount += 1
		}
		literal.WriteString(string(l.getCurrentChar()))
		l.head++
	}

	t := TK_INTEGER
	if dotCount == 1 {
		t = TK_FLOAT
	}

	l.tokens = append(l.tokens, Token{
		Type:    t,
		Literal: literal.String(),
		Line:    l.lineCount,
	})
	return nil
}

func (l *Lexer) peek() byte {
	return l.source[l.head+1]
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
				Line:    l.lineCount,
			})
			l.head++
		case '}':
			l.tokens = append(l.tokens, Token{
				Type:    TK_RBRACE,
				Literal: "}",
				Line:    l.lineCount,
			})
			l.head++
		case '(':
			l.tokens = append(l.tokens, Token{
				Type:    TK_LPAREN,
				Literal: "(",
				Line:    l.lineCount,
			})
			l.head++
		case ')':
			l.tokens = append(l.tokens, Token{
				Type:    TK_RPAREN,
				Literal: ")",
				Line:    l.lineCount,
			})
			l.head++
		case ';':
			l.tokens = append(l.tokens, Token{
				Type:    TK_SEMICOLON,
				Literal: ";",
				Line:    l.lineCount,
			})
			l.head++
		case '=':
			l.tokens = append(l.tokens, Token{
				Type:    TK_EQUAL,
				Literal: "=",
				Line:    l.lineCount,
			})
			l.head++
		case '+':
			l.tokens = append(l.tokens, Token{
				Type:    TK_PLUS,
				Literal: "+",
				Line:    l.lineCount,
			})
			l.head++
		case '-':
			l.tokens = append(l.tokens, Token{
				Type:    TK_DASH,
				Literal: "-",
				Line:    l.lineCount,
			})
			l.head++
		case '*':
			l.tokens = append(l.tokens, Token{
				Type:    TK_STAR,
				Literal: "*",
				Line:    l.lineCount,
			})
			l.head++
		case '/':
			if l.peek() == '/' {
				// skip lines with comments
				for l.getCurrentChar() != '\n' {
					l.head++
				}
			} else {
				l.tokens = append(l.tokens, Token{
					Type:    TK_SLASH,
					Literal: "/",
					Line:    l.lineCount,
				})
				l.head++
			}

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
		Line:    l.lineCount,
	})

	return l.tokens, nil
}
