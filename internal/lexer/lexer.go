package lexer

import (
	"fmt"
)

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

type LexerError struct {
	Message string
	Line    int
}

func (l LexerError) Error() string {
	return fmt.Sprintf("lexical error on line %d: %s", l.Line, l.Message)
}

type Lexer struct {
	source       string
	lineCount    int
	position     int
	readPosition int
	ch           byte
	tokens       []Token
}

func New(source string) *Lexer {
	l := &Lexer{
		source:    source,
		lineCount: 1,
		tokens:    make([]Token, 0),
	}
	l.readChar()
	return l
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.lineCount++
		}
		l.readChar()
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.source) {
		l.ch = 0
	} else {
		l.ch = l.source[l.readPosition]
	}
	l.position = l.readPosition

	l.readPosition++
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.source[position:l.position]
}

func (l *Lexer) readNumber() Token {
	var tok Token
	tok.Type = TK_INTEGER_LITERAL
	position := l.position

	for isDigit(l.ch) || l.ch == '.' {
		// ensure only one dot is part of the number literal
		if l.ch == '.' && tok.Type != TK_FLOAT_LITERAL {
			tok.Type = TK_FLOAT_LITERAL
		}
		l.readChar()
	}
	tok.Literal = l.source[position:l.position]
	return tok
}

func (l *Lexer) readCharLiteral() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}

	return l.source[position:l.position]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		tok = newToken(TK_EQUAL, string(l.ch), l.lineCount)
	case ';':
		tok = newToken(TK_SEMICOLON, string(l.ch), l.lineCount)
	case '(':
		tok = newToken(TK_LPAREN, string(l.ch), l.lineCount)
	case ')':
		tok = newToken(TK_RPAREN, string(l.ch), l.lineCount)
	case '{':
		tok = newToken(TK_LBRACE, string(l.ch), l.lineCount)
	case '}':
		tok = newToken(TK_RBRACE, string(l.ch), l.lineCount)
	case '+':
		tok = newToken(TK_PLUS, string(l.ch), l.lineCount)
	case '-':
		tok = newToken(TK_DASH, string(l.ch), l.lineCount)
	case '*':
		tok = newToken(TK_STAR, string(l.ch), l.lineCount)
	case '/':
		tok = newToken(TK_SLASH, string(l.ch), l.lineCount)
	case '\'':
		tok.Type = TK_CHAR_LITERAL
		tok.Literal = l.readCharLiteral()
	case 0:
		tok.Literal = ""
		tok.Type = TK_EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok = l.readNumber()
			return tok
		} else {
			tok = newToken(TK_ILLEGAL, string(l.ch), l.lineCount)
		}
	}

	l.readChar()
	return tok
}
