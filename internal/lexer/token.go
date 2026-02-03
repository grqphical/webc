package lexer

type TokenType string

const (
	TK_IDENT TokenType = "IDENTIFIER"

	TK_INTEGER_LITERAL TokenType = "INTEGER"
	TK_FLOAT_LITERAL   TokenType = "FLOAT"
	TK_CHAR_LITERAL    TokenType = "CHAR"

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

	TK_PLUS_EQUAL   TokenType = "+="
	TK_MINUS_EQUAL  TokenType = "-="
	TK_TIMES_EQUAL  TokenType = "*="
	TK_DIVIDE_EQUAL TokenType = "/="

	TK_INT    TokenType = "int"
	TK_FLOAT  TokenType = "float"
	TK_CHAR   TokenType = "char"
	TK_RETURN TokenType = "return"

	TK_EOF     TokenType = "EOF"
	TK_ILLEGAL TokenType = "ILLEGAL"
)

var keywords = map[string]TokenType{
	"int":    TK_INT,
	"float":  TK_FLOAT,
	"char":   TK_CHAR,
	"return": TK_RETURN,
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TK_IDENT
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

func newToken(t TokenType, literal string, line int) Token {
	return Token{
		Type:    t,
		Literal: literal,
		Line:    line,
	}
}
