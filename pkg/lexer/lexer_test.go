package lexer

import (
	"fmt"
	"testing"

	"nug/pkg/token"
)

func TestNextToken(t *testing.T) {
	input := `POST http://test.com/api/v1?var1=val1&var2=val2
		HTTP 200
	`

	tests := []token.Token{
		{Type: token.Post, Literal: "POST", Line: 0},
		{Type: token.String, Literal: "http://test.com/api/v1?var1=val1&var2=val2", Line: 0},
		{Type: token.Http, Literal: "HTTP", Line: 1},
		{Type: token.Number, Literal: "200", Line: 1},
		{Type: token.EOF, Literal: "", Line: 2},
	}

	l := New(input)

	assertLexerMatches(t, l, tests)
}

func assertLexerMatches(t *testing.T, l *Lexer, tests []token.Token) {
	for i, expectedToken := range tests {
		actualToken := l.NextToken()

		if actualToken.Type != expectedToken.Type {
			t.Fatalf("tests[%d] - tokentype wrong. Expected: %s, Got: %s", i, formatTokenOutputString(expectedToken), formatTokenOutputString(actualToken))
		}
		if actualToken.Literal != expectedToken.Literal {
			t.Fatalf("tests[%d] - literal wrong. Expected: %s, Got: %s", i, formatTokenOutputString(expectedToken), formatTokenOutputString(actualToken))
		}
		if actualToken.Line != expectedToken.Line {
			t.Fatalf("tests[%d] - line wrong. Expected: %s, Got: %s", i, formatTokenOutputString(expectedToken), formatTokenOutputString(actualToken))
		}
	}
}

func formatTokenOutputString(t token.Token) string {
	result := fmt.Sprintf("Type:%q; Literal:%q; Line:%d", t.Type, t.Literal, t.Line)
	return result
}
