package lexer

import (
	"dao/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `var five = 5;
	var ten int = 10;
	
	var add = func(x int, y int) int {
	   x + y;
	};
	
	var result = add(five, ten);

	!-/*5;
	5 < 10 > 5;

	if 5 > 10 {
		return false;
	} else {
		return true;
	}

	10 == 10;
	10 != 9;
	"foobar"
	"foo bar"
	" foo bar "
	[1,2];
	var a int = 2; a = a + 1
	nil
	%
	break
	for
	 `

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.ID, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.VAR, "var"},
		{token.ID, "ten"},
		{token.ID, "int"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.VAR, "var"},
		{token.ID, "add"},
		{token.ASSIGN, "="},
		{token.FUNC, "func"},
		{token.LPAREN, "("},
		{token.ID, "x"},
		{token.ID, "int"},
		{token.COMMA, ","},
		{token.ID, "y"},
		{token.ID, "int"},
		{token.RPAREN, ")"},
		{token.ID, "int"},
		{token.LBRACE, "{"},
		{token.ID, "x"},
		{token.PLUS, "+"},
		{token.ID, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.VAR, "var"},
		{token.ID, "result"},
		{token.ASSIGN, "="},
		{token.ID, "add"},
		{token.LPAREN, "("},
		{token.ID, "five"},
		{token.COMMA, ","},
		{token.ID, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.INT, "5"},
		{token.GT, ">"},
		{token.INT, "10"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NEQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.STRING, " foo bar "},

		{token.LBRACKET, "["}, // []
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},

		{token.VAR, "var"}, // =
		{token.ID, "a"},
		{token.ID, "int"},
		{token.ASSIGN, "="},
		{token.INT, "2"},
		{token.SEMICOLON, ";"},
		{token.ID, "a"},
		{token.ASSIGN, "="},
		{token.ID, "a"},
		{token.PLUS, "+"},
		{token.INT, "1"},
		{token.NIL, "nil"},
		{token.MOD, "%"},
		{token.BREAK, "break"},
		{token.FOR, "for"},

		{token.EOF, ""},
	}
	// !-/*5; 5 < 10 > 5;

	l := New(input)
	for i, tt := range tests {
		tok := l.Next()
		if tok.Type != tt.expectedType {
			t.Fatalf("testes[%d] - tokentype goes wrong, expected %q but got %q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("testes[%d] - literal goes wrong, expected %q but got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
