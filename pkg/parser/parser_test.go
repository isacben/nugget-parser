package parser

import (
	"fmt"
	"nug/pkg/lexer"
	"testing"
)

func TestParseProgram(t *testing.T) {

	fmt.Println("testing...")
	tests := [...]struct {
		input      string
		entriesLen int
	}{
		{input: `GET https://test.com`, entriesLen: 1},
		{input: `GET https://test.com/items?date=2024-01-01
			GET https://test.com/#`, entriesLen: 2},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)

		program, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("failed to parse program: %v", err)
		}

		rv := *program.RootValue
		val := rv

		checkParserErrors(t, p)

		if len(val.Entries) != test.entriesLen {
			t.Fatalf("the length of the entries is not correct: got %d", len(val.Entries))
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("Parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("Parser error: %q", msg)
	}
	t.FailNow()
}
