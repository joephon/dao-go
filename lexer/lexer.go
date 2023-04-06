package lexer

import "dao/token"

type Lexer struct {
	input   string
	pos     int
	nextPos int
	ch      byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.eat()
	return l
}

func (l *Lexer) Next() token.Token {
	var tok token.Token
	l.eatBlank()

	switch l.ch {
	case '=':
		if l.peak() == '=' {
			ch := l.ch
			l.eat()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = token.New(token.ASSIGN, l.ch)
		}
	case '+':
		tok = token.New(token.PLUS, l.ch)
	case '-':
		tok = token.New(token.MINUS, l.ch)
	case '*':
		tok = token.New(token.ASTERISK, l.ch)
	case '/':
		tok = token.New(token.SLASH, l.ch)
	case '%':
		tok = token.New(token.MOD, l.ch)
	case '!':
		if l.peak() == '=' {
			ch := l.ch
			l.eat()
			tok = token.Token{Type: token.NEQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = token.New(token.BANG, l.ch)
		}
	case '<':
		tok = token.New(token.LT, l.ch)
	case '>':
		tok = token.New(token.GT, l.ch)
	case '(':
		tok = token.New(token.LPAREN, l.ch)
	case ')':
		tok = token.New(token.RPAREN, l.ch)
	case '{':
		tok = token.New(token.LBRACE, l.ch)
	case '}':
		tok = token.New(token.RBRACE, l.ch)
	case '[':
		tok = token.New(token.LBRACKET, l.ch)
	case ']':
		tok = token.New(token.RBRACKET, l.ch)
	case ',':
		tok = token.New(token.COMMA, l.ch)
	case ';':
		tok = token.New(token.SEMICOLON, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.eatID()
			tok.Type = token.GetTokenType(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.eatNumber()
			return tok
		} else {
			tok = token.New(token.ILLEGAL, l.ch)
		}
	}

	l.eat()
	return tok
}

func (l *Lexer) eat() {
	if l.nextPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.nextPos]
	}
	l.pos = l.nextPos
	l.nextPos += 1
}

func (l *Lexer) peak() byte {
	if l.nextPos >= len(l.input) {
		return 0
	} else {
		return l.input[l.nextPos]
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isBlank(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r'
}

func (l *Lexer) eatID() string {
	pos := l.pos
	for isLetter((l.ch)) {
		l.eat()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) eatBlank() {
	for isBlank(l.ch) {
		l.eat()
	}
}

func (l *Lexer) eatNumber() string {
	pos := l.pos
	for isDigit((l.ch)) {
		l.eat()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readString() string {
	pos := l.pos + 1
	for {
		l.eat()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}

	return l.input[pos:l.pos]
}
