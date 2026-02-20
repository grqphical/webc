package lexer

// checks if the given byte is a letter (lower/upper case)
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// checks if the given byte is a digit
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// Lexer stores the current state of the lexer
type Lexer struct {
	source       string
	lineCount    int
	position     int
	readPosition int
	ch           byte
	tokens       []Token
}

// Creates a new lexer
func New(source string) *Lexer {
	l := &Lexer{
		source:    source,
		lineCount: 1,
		tokens:    make([]Token, 0),
	}
	l.readChar()
	return l
}

// Skips all whitespace until it hits a non whitespace character
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.lineCount++
		}
		l.readChar()
	}
}

// Advances the lexer
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.source) {
		l.ch = 0
	} else {
		l.ch = l.source[l.readPosition]
	}
	l.position = l.readPosition

	l.readPosition++
}

// Returns the next character in the source code. If the lexer is at the end of the source code, it returns `\0`
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.source) {
		return 0
	} else {
		return l.source[l.readPosition]
	}
}

// Tokenizes an identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.source[position:l.position]
}

// Tokenizes a number, determines if it's an integer or float literal and returns the token
func (l *Lexer) readNumber() Token {
	var tok Token
	tok.Type = TokenIntLiteral
	position := l.position

	for isDigit(l.ch) || l.ch == '.' {
		// ensure only one dot is part of the number literal
		if l.ch == '.' && tok.Type != TokenFloatLiteral {
			tok.Type = TokenFloatLiteral
		}
		l.readChar()
	}
	tok.Literal = l.source[position:l.position]
	return tok
}

// Reads a C character literal, e.g. 'a'
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

// Consumes the entire line, used when a comment is encountered
func (l *Lexer) readComment() {
	for l.ch != '\n' {
		l.readChar()
	}
}

// Consumes a multiline comment
func (l *Lexer) readMultiLineComment() {
	for true {
		l.readChar()
		if l.ch == '*' && l.peekChar() == '/' {
			break
		}
	}

	for l.ch != '\n' {
		l.readChar()
	}
}

// Creates a token based on the current character and then advances the lexer
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenEqualEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenEqual, string(l.ch), l.lineCount)
		}
	case ';':
		tok = newToken(TokenSemicolon, string(l.ch), l.lineCount)
	case '(':
		tok = newToken(TokenLParen, string(l.ch), l.lineCount)
	case ')':
		tok = newToken(TokenRParen, string(l.ch), l.lineCount)
	case '{':
		tok = newToken(TokenLBrace, string(l.ch), l.lineCount)
	case '}':
		tok = newToken(TokenRBrace, string(l.ch), l.lineCount)
	case ',':
		tok = newToken(TokenComma, string(l.ch), l.lineCount)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenNotEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenBang, string(l.ch), l.lineCount)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenLessOrEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenLessThan, string(l.ch), l.lineCount)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenGreaterOrEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenGreaterThan, string(l.ch), l.lineCount)
		}
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenPlusEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenPlus, string(l.ch), l.lineCount)
		}
	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenMinusEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenDash, string(l.ch), l.lineCount)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenTimesEqual, string(ch)+string(l.ch), l.lineCount)
		} else {
			tok = newToken(TokenStar, string(l.ch), l.lineCount)
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(TokenDivideEqual, string(ch)+string(l.ch), l.lineCount)
		} else if l.peekChar() == '/' {
			// skip lines with comments on them
			l.readComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.readMultiLineComment()
			return l.NextToken()
		} else {
			tok = newToken(TokenSlash, string(l.ch), l.lineCount)
		}
	case '\'':
		tok.Type = TokenCharLiteral
		literal := l.readCharLiteral()
		if literal == "\\n" {
			literal = "\n"
		}

		tok.Literal = literal
	case 0:
		tok.Literal = ""
		tok.Type = TokenEndOfFile
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok = l.readNumber()
			return tok
		} else {
			tok = newToken(TokenIllegal, string(l.ch), l.lineCount)
		}
	}

	l.readChar()
	return tok
}

// Parses the entire source code, returning the tokens it generated
func (l *Lexer) Parse() []Token {
	toks := make([]Token, 0)
	for {
		tok := l.NextToken()
		toks = append(toks, tok)

		if tok.Type == TokenEndOfFile {
			break
		}
	}
	return toks
}
