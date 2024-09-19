package token

import "fmt"

// Type alias for a string
type Type string

const (
	// Unrecognize token or character
	Ilegal Type = "ILEGAL"

	// End of file
	EOF Type = "EOF"

	// Literals
	String Type = "STRING"
	Number Type = "NUMBER"

	// Structural tokens
	LeftBrace    Type = "{"
	RightBrace   Type = "}"
	LeftBracket  Type = "["
	RightBracket Type = "]"
	Colon        Type = ":"
	Whitespace   Type = "WHITESPACE"
	NewLine      Type = "NEWLINE"

	// Comments
	LineComment Type = "#"

	// Commands
	Post   Type = "POST"
	Get    Type = "GET"
	Header Type = "HEADER"
	Http   Type = "HTTP"

	// Tags
	Captures Type = "CAPTURES"
)

type Token struct {
	Type    Type
	Literal string
	Line    int
	Start   int
	End     int
}

var validIdentifiers = map[string]Type{
	"POST":     Post,
	"GET":      Get,
	"header":   Header,
	"HTTP":     Http,
	"Captures": Captures,
}

func LookupIdentifier(identifier string) (Type, error) {
	if token, ok := validIdentifiers[identifier]; ok {
		return token, nil
	}
	return "", fmt.Errorf("error: expected a valid identifier, found: %s", identifier)
}
