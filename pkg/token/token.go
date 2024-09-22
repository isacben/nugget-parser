package token

import "fmt"

// Type alias for a string
type Type string

// Types of tokens
const (
	// Unrecognize token or character
	Ilegal Type = "ILEGAL"

	// End of file
	EOF Type = "EOF"

	// Literals
	String Type = "STRING"
	Number Type = "NUMBER"

	// Structural tokens
	Whitespace   Type = "WHITESPACE"
	NewLine      Type = "NEWLINE"

	// Comments
	Comment Type = "COMMENT"

	// Methods
	Post   Type = "POST"
	Get    Type = "GET"
)

type Token struct {
	Type    Type
	Literal string
	Line    int
	Start   int
	End     int
}

var validMethods= map[string]Type{
	"POST":     Post,
	"GET":      Get,
}

func LookupMethod(identifier string) (Type, error) {
	if token, ok := validMethods[identifier]; ok {
		return token, nil
	}
	return "", fmt.Errorf("error: expected a valid method, found: %s", identifier)
}
