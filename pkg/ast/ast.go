package ast

// These are the available root node types. In JSON it will either be an
// object or an array at the base.
const (
	NuggetRoot RootNodeType = iota
)

// RootNodeType is a type alias for an int
type RootNodeType int

// RootNode is what starts every parsed AST. There is a `Type` field so that
// you can ask which root node type starts the tree.
type RootNode struct {
	RootValue *Nugget
	Type      RootNodeType
}

type Nugget struct {
	Type    string // "Nugget"
	Entries []Entry
}

type Entry struct {
	Type string // "Entry"
	Req  Request
	Res  Response
}

// Object represents a nugget request. It holds a slice of Property as its children,
// a Type ("Request"), and start & end code points for displaying.
type Request struct {
	Type   string // "Request"
	Line   Endpoint
	Header []KeyValue
	Start  int
	End    int
}

type Response struct {
	Type    string // "Response"
	Version string
	Status  int
	Capture []KeyValue
	Start   int
	End     int
}

type Endpoint struct {
	Type   string // "Endpoint"
	Method string
	Url    string
}

type KeyValue struct {
	Type  string // "KeyValue"
	Key   string
	Value string
}

// state is a type alias for int and used to create the available value states below
type state int

// Available states for each type used in parsing
const (
	// Nuget states
	NuggetStart state = iota
	NuggetEntry

	// Entry states
	EntryStart
	EntryRequest
	EntryResponse

	// Request states
	ReqStart
	ReqOpen
	ReqLine

	// Response states
	ResStart
	ResOpen
	ResStatus

	// Line States
	LineStart
	LineMethod
	LineNewLine

	// Header States
	HeaderStart
	HeaderKey
	HeaderValue

	// String states
	StringStart
	StringQuoteOrChar
	Escape

	// Number states
	NumberStart
	NumberMinus
	NumberZero
	NumberDigit
	NumberPoint
	NumberDigitFraction
	NumberExp
	NumberExpDigitOrSign
)
