package main

import (
	"encoding/json"
	"fmt"
	"nug/pkg/lexer"
	"nug/pkg/parser"
)

func main() {

	var s = `GET http://airwallex.com/v1/api 
	GET https://airwallex.com/v1/api/issuing/card/create
	someHeader
	anotherHeader
	GET http://test.com/todos?done=false
	andaAotherHeader`

	l := lexer.New(s)

	p := parser.New(l)
	tree, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("error: parser error: %v\n", err)
	}

	fmt.Printf("%+v\n", *tree.RootValue)

	jtree, _ := json.MarshalIndent(*tree.RootValue, "  ", "    ")
	fmt.Println(string(jtree))
}
