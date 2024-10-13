package parser

import (
	"fmt"
	"nug/pkg/ast"
	"nug/pkg/lexer"
	"testing"
	"reflect"
)

func TestParseNumberOfEntries(t *testing.T) {

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

func TestParseGetWithResponse(t *testing.T) {
	test := struct{
		input string
	}{
		input: `GET https://test.com/v1/api
                HTTP 200`,
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
						End: 44,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "HTTP",
						Status: 200,
						Capture: nil,
						Start: 44,
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

func TestParseGetWithCapture(t *testing.T) {
	test := struct{
		input string
	}{
		input: `GET https://test.com/v1/api
                HTTP 200
                [Capture]
                capture_1: value_1`,
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
						End: 44,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "HTTP",
						Status: 200,
						Capture: []ast.KeyValue{
                            {
                                Type: "KeyValue",
                                Key: "capture_1",
                                Value: "value_1",
                            },
                        },
						Start: 44,
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

func TestParseMultipleRequests(t *testing.T) {
	test := struct{
		input string
	}{
		input: `GET https://test.com/v1/api/a
                header_1: value_1
                HTTP 200
                [Capture]
                capture_1: value_1

                GET https://test.com/v1/api/b
                header_2: value_2
                header_3: value_3
                HTTP 200

                GET https://test.com/v1/api/c
                HTTP 200
                [Capture]
                capture_4: value_4

                GET https://test.com/v1/api/d`,
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
						Start: 0,
						End: 80,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "HTTP",
						Status: 200,
						Capture: []ast.KeyValue{
                            {
                                Type: "KeyValue",
                                Key: "capture_1",
                                Value: "value_1",
                            },
                        },
						Start: 80,
						End: 167,
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
						Start: 167,
						End: 281,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "HTTP",
						Status: 200,
						Capture: nil,
                        Start: 281,
						End: 310,
					},
				},

				{
					Type: "Entry",
					Req: ast.Request{
						Type: "Request",
						Line: ast.Endpoint{
							Type: "Endpoint",
							Method: "GET",
							Url: "https://test.com/v1/api/c",
						},
						Header: nil,
                        Start: 307,
						End: 353,
					},
					Res: ast.Response{
						Type: "Response",
						Version: "HTTP",
						Status: 200,
						Capture: []ast.KeyValue{
                            {
                                Type: "KeyValue",
                                Key: "capture_4",
                                Value: "value_4",
                            },
                        },
						Start: 353,
						End: 440,
					},
				},

				{
					Type: "Entry",
					Req: ast.Request{
						Type: "Request",
						Line: ast.Endpoint{
							Type: "Endpoint",
							Method: "GET",
							Url: "https://test.com/v1/api/d",
						},
						Header: nil,
                        Start: 440,
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
