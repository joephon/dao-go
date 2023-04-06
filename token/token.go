package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

func New(tt TokenType, li byte) Token {
	return Token{Type: tt, Literal: string(li)}
}

func GetTokenType(id string) TokenType {
	if tok, ok := keywords[id]; ok {
		return tok
	}
	return ID
}

var keywords = map[string]TokenType{
	"func":   FUNC,
	"var":    VAR,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"nil":    NIL,
	"return": RETURN,
	"break":  BREAK,
	"for":    FOR,
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// 标识符+字面量
	ID     = "ID"  // add, foobar, x, y, ...
	INT    = "INT" // 1343456
	STRING = "STRING"

	// 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MOD      = "%"

	EQ  = "=="
	NEQ = "!="

	LT = "<"
	GT = ">"

	// 分隔符
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// 关键字
	FUNC   = "FUNC"
	VAR    = "VAR"
	IF     = "IF"
	ELSE   = "ELSE"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	NIL    = "NIL"
	RETURN = "RETURN"
	BREAK  = "BREAK"
	FOR    = "FOR"
)
