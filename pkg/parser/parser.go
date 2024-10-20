package parser

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
		rootNode.Type = ast.NuggetRoot
	}

	nugget := p.parseNugget()

	if p.Errors() != "" {
		return ast.RootNode{}, errors.New(p.Errors())
	}

	if len(nugget.Entries) == 0 {
		p.parseError(fmt.Sprintf(
            "line %v, expected a request, got: `%v`",
			p.currentToken.Line+1, p.currentToken.Literal,
		))
		return ast.RootNode{}, errors.New(p.Errors())
	}

	rootNode.RootValue = &nugget
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

func (p *Parser) parseNugget() ast.Nugget {
	nugget := ast.Nugget{Type: "Nugget"}
	nuggetState := ast.NuggetStart
	var entries []ast.Entry

	for !p.currentTokenTypeIs(token.EOF) {
		switch nuggetState {
		case ast.NuggetStart:
			if p.currentTokenTypeIs(token.Get) || p.currentTokenTypeIs(token.Post) {
				entry := p.parseEntry()
				entries = append(entries, entry)
				nuggetState = ast.NuggetEntry
			} else {
				p.parseError(fmt.Sprintf(
					"line %v, expected HTTP method, got: %s",
					p.currentToken.Line+1, p.currentToken.Literal,
				))
				return ast.Nugget{}
			}
		case ast.NuggetEntry:
			if p.currentTokenTypeIs(token.Get) || p.currentTokenTypeIs(token.Post) {
				nuggetState = ast.NuggetStart
			} else {
				return ast.Nugget{}
			}
		}
	}

	nugget.Entries = entries
	return nugget
}

// parseEntry is ouur dynamic entrypoint to parsing JSON values. All scenarios for
// this parser fall under these 3 actions.
func (p *Parser) parseEntry() ast.Entry {
	entry := ast.Entry{Type: "Entry"}

	entry.Req = p.parseRequest()
	entry.Res = p.parseResponse()

	return entry
}

// parseRequest is called when an type of request identifier (POST, GET, etc.) token is found
func (p *Parser) parseRequest() ast.Request {
	req := ast.Request{Type: "Request"} // Struct of type Request
	reqState := ast.ReqStart            // Request state of the state machine

	for !p.currentTokenTypeIs(token.EOF) {
		switch reqState {
		case ast.ReqStart:
			if p.currentTokenTypeIs(token.Get) || p.currentTokenTypeIs(token.Post) {
				reqState = ast.ReqOpen
				req.Start = p.currentToken.Start
			} else {
				p.parseError(fmt.Sprintf(
					"line %v, expected HTTP method, got: %s",
					p.currentToken.Line+1, p.currentToken.Literal,
				))
				return ast.Request{}
			}

		case ast.ReqOpen:
			// we haven't advanced to the next token
			if p.peekTokenTypeIs(token.EOF) {
				req.End = p.currentToken.End
				p.nextToken()
				return req
			}
			reqState = ast.ReqLine
			line := p.parseLine()
			req.Line = line
            req.End = p.currentToken.End
			p.nextToken()

		case ast.ReqLine:
			// if the next token is a string, it might be a header
			if !p.currentTokenTypeIs(token.String) {
				req.End = p.currentToken.Start
				return req
			}

			header := p.parseKeyValue()
			req.Header = append(req.Header, header)
            req.End = p.currentToken.End
			p.nextToken()
		}
	}

	return req
}

func (p *Parser) parseResponse() ast.Response {
	res := ast.Response{Type: "Response"} // Struct of type Response

	if !p.currentTokenTypeIs(token.Http) {
		return res 
	} 

	res.Version = p.parseString()
	res.Start = p.currentToken.Start
	p.nextToken()

	if !p.currentTokenTypeIs(token.Number) {
		p.parseError(fmt.Sprintf(
			"line %v, expected number, got: `%s`",
			p.currentToken.Line+1, p.currentToken.Literal,
		))
		return ast.Response{}
	}

	res.Status, _ = strconv.Atoi(p.currentToken.Literal)

	res.End = p.currentToken.End
	p.nextToken()

	if p.currentTokenTypeIs(token.Capture) {
		p.nextToken()
		for !p.currentTokenTypeIs(token.EOF) {
			if !p.currentTokenTypeIs(token.String) {
				res.End = p.currentToken.Start
				return res
			}

			capture := p.parseKeyValue()
			res.Capture = append(res.Capture, capture)
            res.End = p.currentToken.End
			p.nextToken()
		}
	}

	return res
}

// parseCommand is used to parse an object command and doing so handles setting command keyword and the parameter
func (p *Parser) parseLine() ast.Endpoint {
	endpoint := ast.Endpoint{Type: "Endpoint"}
	lineState := ast.LineStart

	for !p.currentTokenTypeIs(token.EOF) {
		switch lineState {
		case ast.LineStart:
			if p.currentTokenTypeIs(token.Get) {
				endpoint.Method = p.currentToken.Literal
				lineState = ast.LineMethod
				p.nextToken()
			} else {
				p.parseError(fmt.Sprintf(
					"line %v, expected HTTP method, got: %s",
					p.currentToken.Line+1, p.currentToken.Literal,
				))
			}

		case ast.LineMethod:
			if p.currentTokenTypeIs(token.String) {
				lineState = ast.LineNewLine
			} else {
				p.parseError(fmt.Sprintf(
					"line %v, expected url, got: `%s`",
					p.currentToken.Line+1, p.currentToken.Literal,
				))
			}

		case ast.LineNewLine:
			param := p.parseString()
			endpoint.Url = param
			return endpoint
		}
	}
	return endpoint
}

func (p *Parser) parseKeyValue() ast.KeyValue {
	kv := ast.KeyValue{Type: "KeyValue"}

	strToken := p.parseString()
	if strToken[len(strToken)-1] != ':' {
		p.parseError(fmt.Sprintf(
			"line %v, expected `:`, got:`%s`",
			p.currentToken.Line+1, p.peekToken.Literal,
		))
		p.nextToken()
		return ast.KeyValue{}
	}

	kv.Key = strToken[:len(strToken)-1]
	p.nextToken()

	if !p.currentTokenTypeIs(token.String) {
		p.parseError(fmt.Sprintf(
			"line %v, expected string, got: `%s`",
			p.currentToken.Line+1, p.currentToken.Literal,
		))
		return ast.KeyValue{}
	}

	kv.Value = p.parseString()
	return kv
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
