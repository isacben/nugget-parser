# nugget parser

A parser for the nugget CLI application.

https://github.com/isacben/nugget
 
> [!NOTE]
> This parser is in development and is not used in nugget yet.

## Gramar

```
<NUGGET>    ::= [ <entry> *(<entry>) ]
<entry>     ::= <request> "\n" <response>
<request>   ::= <line>
                [ <header> *(<header>) ]
<line>      ::= <method> <string>
<header>    ::= <key-value>
<response>  ::= "HTTP" <number>
                "[Capture]"
                [ <capture> *(<capture>)]
<capture>   ::= <key-value>
<key-value> ::= <string> ":" <string> | "\""<string>"\""
<method>    ::= "POST" | "GET" */
```

Another representation:

```
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
```

## Todos

Implement:

- [ ] implement request body support (json parsing)
- [ ] implement usage of captured variables
- [ ] fix go package to be able to import it in nugget

Nice to have:

- [ ] allow spaces between key and `:` character for the `key-value` token
- [ ] allog HTTP methods in lowe case
- [ ] refactor error formatting with line numbers
- [ ] add tests with error messages
- [ ] simplify if-else code
