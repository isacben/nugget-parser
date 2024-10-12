package parser

import (
	"fmt"
	"nug/pkg/ast"
	"nug/pkg/lexer"
	"testing"
	"reflect"
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

func TestParseSingleGet(t *testing.T) {
	test := struct{
		input string
	}{
		input: `GET https://test.com/v1/api`,
    }

	result := ast.RootNode{
		Type: ast.NuggetRoot,
		RootValue: &ast.Nugget{
			Type: "Nugget",
			Entries: []ast.Entry{
				{
					Type: "Entry",
					Req: ast.Request{
						Type: "Request",
						Line: ast.Endpoint{
							Type: "Endpoint",
							Method: "GET",
							Url: "https://test.com/v1/api",
						},
						Header: nil,
						Start: 0,
						End: 0,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "",
						Status: 0,
						Capture: nil,
						Start: 0,
						End: 0,
					},
				},
			},
		},
	}

	l := lexer.New(test.input)
	p := New(l)

	program, err := p.ParseProgram()
	if err != nil {
		t.Fatal("error: ", err)
	}

	if !reflect.DeepEqual(*program.RootValue, *result.RootValue) {
		t.Fatalf("error: expected %+v, got: %+v", *result.RootValue, *program.RootValue)
	}
}

func TestParseMultipleGetsWithHeaders(t *testing.T) {
	test := struct{
		input string
	}{
		input: `
			GET https://test.com/v1/api/a
			header_1: value_1
			GET https://test.com/v1/api/b
			header_2: value_2
			header_3: value_3`,
	}

	result := ast.RootNode{
		Type: ast.NuggetRoot,
		RootValue: &ast.Nugget{
			Type: "Nugget",
			Entries: []ast.Entry{
				{
					Type: "Entry",
					Req: ast.Request{
						Type: "Request",
						Line: ast.Endpoint{
							Type: "Endpoint",
							Method: "GET",
							Url: "https://test.com/v1/api/a",
						},
						Header: []ast.KeyValue{
                            {
                                Type: "KeyValue",
                                Key: "header_1",
                                Value: "value_1",
                            },
                        },
						Start: 4,
						End: 58,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "",
						Status: 0,
						Capture: nil,
						Start: 0,
						End: 58,
					},
				},
				{
					Type: "Entry",
					Req: ast.Request{
						Type: "Request",
						Line: ast.Endpoint{
							Type: "Endpoint",
							Method: "GET",
							Url: "https://test.com/v1/api/b",
						},
						Header: []ast.KeyValue{
                            {
                                Type: "KeyValue",
                                Key: "header_2",
                                Value: "value_2",
                            },
                            {
                                Type: "KeyValue",
                                Key: "header_3",
                                Value: "value_3",
                            },
                        },
						Start: 58,
						End: 0,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "",
						Status: 0,
						Capture: nil,
						Start: 0,
						End: 0,
					},
				},
			},
		},
	}

	l := lexer.New(test.input)
	p := New(l)

	program, err := p.ParseProgram()
	if err != nil {
		t.Fatal("error: ", err)
	}

	if !reflect.DeepEqual(*program.RootValue, *result.RootValue) {
		t.Fatalf("error: expected %+v, got: %+v", *result.RootValue, *program.RootValue)
	}
}
