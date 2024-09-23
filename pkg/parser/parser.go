package parser

/** nugget gramar
 *
 *  <NUGGET>		::= [ <request> *(<request>) ]
 *  <request>		::= [ <command> "\n" [ <expression> *('\n' <expression>) ]
 *  <expression>    	::= <property> | <tag>
 *  <command>		::= <string> " " <string> | <number>
 *  <property>		::= <string> ":" <string> | <command>
 *  <tag>			::= "[" <string> "]" */

// Examples:
// command -> GET http://example.com
// property -> x-api-version: 2024-06-30
// tag -> [Capture]

/*
*  <NUGGET>		::= [ <entry> *(<entry>) ]
 * <entry>              ::= <request> "\n" <response>
 * <request>		::= <line>
			    [ <header> *(<header>) ]
 * <line>		::= <method> <string>
 * <header>		::= <key-value>
 * <response>		::= "HTTP" <number>
 			    "[Capture]"
			    [ <capture> *(<capture>)]
 * <capture>		::= <key-value>
 * <key-value>		::= <string> ":" <string> | "\""<string>"\""
 * <method>		::= "POST" | "GET" */

/*

nugget-file
	entry*
	lt*
entry
	request
	response?
request
	lt*
	method sp value-string lt
	header*
	body?
response
	lt*
	HTTP sp status lt
	captures
method
	POST | GET
status
	[0-9]
header lt* key-value lt body
	lt*
	json-value lt
captures
	lt*
	[Captures] lt
	capture*
key-value
	[A-Za-z0-9]|_|-|.|[|]|@|$) : value-string
capture
	lt*
	key-string : quoted-string-text lt
quoted-string-text:
	~["k\]+
lt
	sp* comment? [\n]?

*/

// Parser holds a Lexer, errors, the currentToken, and the peek peekToken (next token).
// Parser methods handle iterating through tokens and building and AST.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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

	entry := p.parseEntry()
	if entry == nil {
		p.parseError(fmt.Sprintf(
			"error: parsing nugget: expected an entry, got: %v:",
			p.currentToken.Literal,
		))
		return ast.RootNode{}, errors.New(p.Errors())
	}
	rootNode.RootValue = &entry

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
func (p *Parser) parseEntry() ast.Value {
	switch p.currentToken.Type {
	case token.Get, token.Post:
		return p.parseRequest()
	default:
		return p.parseJSONLiteral()
	}
}

// parseRequest is called when an type of request identifier (POST, GET, etc.) token is found
func (p *Parser) parseRequest() ast.Value {
	req := ast.Request{Type: "Request"} // Struct of type Request
	reqState := ast.ReqStart            // Request state of the state machine

	for !p.currentTokenTypeIs(token.EOF) {
		switch reqState {
		case ast.ReqStart:
			if p.currentTokenTypeIs(token.Get) || p.currentTokenTypeIs(token.Post) {
				reqState = ast.ReqOpen
				req.Start = p.currentToken.Start
				//p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"error: parsing nugget: expected `POST` or `GET` token, got: %s",
					p.currentToken.Literal,
				))
				return nil
			}
		case ast.ReqOpen:
			fmt.Println("I'm in parseNuggetRequest OPEN...")
			fmt.Println(p.currentToken.Type)
			fmt.Println(p.peekToken.Type)
			if p.peekTokenTypeIs(token.Post) || p.peekTokenTypeIs(token.Get) || p.peekTokenTypeIs(token.EOF) {
				p.nextToken()
				req.End = p.currentToken.End
				return req
			}
			line := p.parseLine()
			req.Line = line 
			//req.Children = append(req.Children, cmd)
			reqState = ast.ReqCommand
		case ast.ReqCommand:
			if p.currentTokenTypeIs(token.Get) {
				req.End = p.currentToken.Start
				reqState = ast.ReqStart
				//	return req
			} else {
				return nil
			}
			//prop := p.parseProperty()
			//obj.Children = append(obj.Children, prop)
			//objState = ast.ObjProperty
			//case ast.ObjProperty:
			//	if p.currentTokenTypeIs(token.RightBrace) {
			//		p.nextToken()
			//		obj.End = p.currentToken.Start
			//		return obj
			//	} else if p.currentTokenTypeIs(token.Comma) {
			//		objState = ast.ObjComma
			//		p.nextToken()
			//	} else {
			//		p.parseError(fmt.Sprintf(
			//			"Error parsing property. Expected RightBrace or Comma token, got: %s",
			//			p.currentToken.Literal,
			//		))
			//		return nil
			//	}
			//case ast.ObjComma:
			//	prop := p.parseProperty()
			//	if prop.Value != nil {
			//		obj.Children = append(obj.Children, prop)
			//		objState = ast.ObjProperty
			//	}
		}
	}

	req.End = p.currentToken.Start

	return req
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

// parseCommand is used to parse an object command and doing so handles setting command keyword and the parameter
func (p *Parser) parseLine() ast.Endpoint{
	fmt.Println("Got in parseCommand")
	fmt.Println("Current token: ", p.currentToken.Type)
	endpoint:= ast.Endpoint{Type: "Endpoint"}
	cmdState := ast.LineStart

	for !p.currentTokenTypeIs(token.EOF) {
		switch cmdState {
		case ast.LineStart:
			if p.currentTokenTypeIs(token.Get) {
				fmt.Println("In LineStart...")
				endpoint.Method = p.parseString()
				fmt.Println("Endpoint: ", endpoint)
				cmdState = ast.LineMethod
				p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"error: parse command start: expected GET, got %s",
					p.currentToken.Literal,
				))
			}
		case ast.LineMethod:
			if p.currentTokenTypeIs(token.String) {
				cmdState = ast.LineNewLine
			} else {
				p.parseError(fmt.Sprintf(
					"error: parsing command: expected new line token, got: %s",
					p.currentToken.Literal,
				))
			}
		case ast.LineNewLine:
			param := p.parseString()
			endpoint.Url = param
			p.nextToken()
			fmt.Println("Entered CommandNewLine, current token is: ", p.currentToken.Type)
			return endpoint
		}
	}
	return endpoint
}

// TODO: all the tedius escaping, etc still needs to be applied here
func (p *Parser) parseString() string {
	return p.currentToken.Literal
}

func (p *Parser) peekTokenTypeIs(t token.Type) bool {
	return p.peekToken.Type == t
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
