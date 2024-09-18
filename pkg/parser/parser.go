package parser

/** nugget gramar
 *
 *  <NUGGET>		::= [ <request> *(<request>) ]
 *  <request>		::= [ <command> "\n" [ <expression> *('\n' <expression>) ]
 *  <expression>    ::= <property> | <tag>
 *  <command>		::= <string> " " <string> | <number>
 *  <property>		::= <string> ":" <string> | <command>
 *  <tag>			::= "[" <string> "]" */

// Examples:
// command -> GET http://example.com
// property -> x-api-version: 2024-06-30
// tag -> [Capture]

// Parser holds a Lexer, errors, the currentToken, and the peek peekToken (next token).
// Parser methods handle iterating through tokens and building and AST.

import (
	"errors"
	"fmt"
	"strings"
	"strconv"

	"nug/pkg/ast"
	"nug/pkg/lexer"
	"nug/pkg/token"
)

type Parser struct {
	lexer        *lexer.Lexer
	errors       []string
	currentToken token.Token
	peekToken    token.Token
}

// New takes a Lexer, creates a Parser with that Lexer, sets the current and
// peek tokens, and returns the Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{lexer: l}

	// Read two tokens, so currentToken and peekToken are both set.
	p.nextToken()
	p.nextToken()

	return p
}

// ParseProgram parses tokens and creates an AST. It returns the RootNode
// which holds a slice of Values (and in turn, the rest of the tree)
func (p *Parser) ParseProgram() (ast.RootNode, error) {
	var rootNode ast.RootNode
	if p.currentTokenTypeIs(token.Get) || p.currentTokenTypeIs(token.Post) {
		rootNode.Type = ast.RequestRoot
	}

	val := p.parseValue()
	if val == nil {
		p.parseError(fmt.Sprintf(
			"error: parsing nugget: expected a request, got: %v:",
			p.currentToken.Literal,
		))
		return ast.RootNode{}, errors.New(p.Errors())
	}
	rootNode.RootValue = &val

	return rootNode, nil
}

// nextToken sets our current token to the peek token and the peek token to
// p.lexer.NextToken() which ends up scanning and returning the next token
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) currentTokenTypeIs(t token.Type) bool {
	return p.currentToken.Type == t
}

// parseValue is our dynamic entrypoint to parsing JSON values. All scenarios for
// this parser fall under these 3 actions.
func (p *Parser) parseValue() ast.Value {
	switch p.currentToken.Type {
	case token.Get:
		return p.parseNuggetRequest()
	default:
		return p.parseJSONLiteral()
	}
}

// parseJSONObject is called when an open left brace `{` token is found
func (p *Parser) parseNuggetRequest() ast.Value {
	obj := ast.Object{Type: "Request"}
	objState := ast.ObjStart

	for !p.currentTokenTypeIs(token.EOF) {
		switch objState {
		case ast.ObjStart:
			if p.currentTokenTypeIs(token.Get) || p.currentTokenTypeIs(token.Post) {
				objState = ast.ObjOpen
				obj.Start = p.currentToken.Start
				p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"Error parsing JSON object Expected `{` token, got: %s",
					p.currentToken.Literal,
				))
				return nil
			}
		case ast.ObjOpen:
			if p.currentTokenTypeIs(token.RightBrace) {
				p.nextToken()
				obj.End = p.currentToken.End
				return obj
			}
			prop := p.parseProperty()
			obj.Children = append(obj.Children, prop)
			objState = ast.ObjProperty
		case ast.ObjProperty:
			if p.currentTokenTypeIs(token.RightBrace) {
				p.nextToken()
				obj.End = p.currentToken.Start
				return obj
			} else if p.currentTokenTypeIs(token.Comma) {
				objState = ast.ObjComma
				p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"Error parsing property. Expected RightBrace or Comma token, got: %s",
					p.currentToken.Literal,
				))
				return nil
			}
		case ast.ObjComma:
			prop := p.parseProperty()
			if prop.Value != nil {
				obj.Children = append(obj.Children, prop)
				objState = ast.ObjProperty
			}
		}
	}

	obj.End = p.currentToken.Start

	return obj
}

// parseJSONLiteral switches on the current token's type, sets the Value on a return val and returns it.
func (p *Parser) parseJSONLiteral() ast.Literal {
	val := ast.Literal{Type: "Literal"}

	// Regardless of what the current token type is - after it's been assigned, we must consume the token
	defer p.nextToken()

	switch p.currentToken.Type {
	case token.String:
		val.Value = p.parseString()
		return val
	case token.Number:
		v, _ := strconv.Atoi(p.currentToken.Literal)
		val.Value = v
		return val
	default:
		val.Value = "null"
		return val
	}
}

// parseProperty is used to parse an object property and in doing so handles setting the `key`:`value` pair.
func (p *Parser) parseProperty() ast.Property {
	prop := ast.Property{Type: "Property"}
	propertyState := ast.PropertyStart

	for !p.currentTokenTypeIs(token.EOF) {
		switch propertyState {
		case ast.PropertyStart:
			if p.currentTokenTypeIs(token.String) {
				key := ast.Identifier{
					Type:  "Identifier",
					Value: p.parseString(),
				}
				prop.Key = key
				propertyState = ast.PropertyKey
				p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"Error parsing property start. Expected String token, got: %s",
					p.currentToken.Literal,
				))
			}
		case ast.PropertyKey:
			if p.currentTokenTypeIs(token.Colon) {
				propertyState = ast.PropertyColon
				p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"Error parsing property. Expected Colon token, got: %s",
					p.currentToken.Literal,
				))
			}
		case ast.PropertyColon:
			val := p.parseValue()
			prop.Value = val
			return prop
		}
	}

	return prop
}

// TODO: all the tedius escaping, etc still needs to be applied here
func (p *Parser) parseString() string {
	return p.currentToken.Literal
}

// parseError is very similar to `peekError`, except it simply takes a string message that
// gets appended to the parser's errors
func (p *Parser) parseError(msg string) {
	p.errors = append(p.errors, msg)
}

// Errors is simply a helper function that returns the parser's errors
func (p *Parser) Errors() string {
	return strings.Join(p.errors, ", ")
}
