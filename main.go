package main

import (
	"encoding/json"
	"fmt"
	"nug/pkg/lexer"
	"nug/pkg/parser"
	"os"
)

func main() {

	var s = `
	GET http://airwallex.com/v1/api 
	wrongheader: "wrongHeaderValue"
	HTTP 200
	[Capture]
	test: hola

    GET http://airwallex.com
    header: value
	`

	l := lexer.New(s)

	p := parser.New(l)
	tree, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("error: parser error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", *tree.RootValue)

	jtree, _ := json.MarshalIndent(*tree.RootValue, "  ", "    ")
	fmt.Println(string(jtree))
}
