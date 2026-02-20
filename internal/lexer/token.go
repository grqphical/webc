package lexer

type TokenType string

const (
	TokenIdent TokenType = "IDENTIFIER"

	TokenIntLiteral   TokenType = "INTEGER"
	TokenFloatLiteral TokenType = "FLOAT"
	TokenCharLiteral  TokenType = "CHAR"

	TokenLParen TokenType = "("
	TokenRParen TokenType = ")"
	TokenLBrace TokenType = "{"
	TokenRBrace TokenType = "}"

	TokenSemicolon TokenType = ";"
	TokenComma     TokenType = ","

	TokenEqual       TokenType = "="
	TokenDash        TokenType = "-"
	TokenPlus        TokenType = "+"
	TokenStar        TokenType = "*"
	TokenSlash       TokenType = "/"
	TokenLessThan    TokenType = "<"
	TokenGreaterThan TokenType = ">"
	TokenBang        TokenType = "!"

	TokenPlusEqual      TokenType = "+="
	TokenMinusEqual     TokenType = "-="
	TokenTimesEqual     TokenType = "*="
	TokenDivideEqual    TokenType = "/="
	TokenNotEqual       TokenType = "!="
	TokenLessOrEqual    TokenType = "<="
	TokenGreaterOrEqual TokenType = ">="
	TokenEqualEqual     TokenType = "=="

	TokenIntKeyword   TokenType = "int"
	TokenFloatKeyword TokenType = "float"
	TokenCharKeyword  TokenType = "char"
	TokenReturn       TokenType = "return"
	TokenConst        TokenType = "const"
	TokenExtern       TokenType = "extern"
	TokenVoid         TokenType = "void"

	TokenEndOfFile TokenType = "EOF"
	TokenIllegal   TokenType = "ILLEGAL"
)

var keywords = map[string]TokenType{
	"int":    TokenIntKeyword,
	"float":  TokenFloatKeyword,
	"char":   TokenCharKeyword,
	"return": TokenReturn,
	"const":  TokenConst,
	"extern": TokenExtern,
	"void":   TokenVoid,
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
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
