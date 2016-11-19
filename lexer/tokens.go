//go:generate stringer -type=TokenType

package lexer

import "fmt"

// TokenType is the type of a token
type TokenType int

// TokenType constants
const (
	// Values
	IDENT TokenType = iota
	INTVALUE
	FLOATVALUE
	TRUE
	FALSE
	ILLEGAL

	// Types
	TYPE
	INT
	INT8
	INT16
	INT32
	INT64
	FLOAT
	FLOAT32
	FLOAT64

	// key words
	RETURN
	IF
	ELSE
	FOR
	RANGE

	// Symbols
	ARROW
	INFERASSIGN
	DOUBLECOLON
	INCREMENT
	INCREMENTEQUAL
	DECREMENT
	DECREMENTEQUAL
	EQUAL
	NOTEQUAL
	COMMAN
	SEMICOLON
	NEWLINE
	OPENBODY
	CLOSEBODY
	OPENBRACKET
	CLOSEBRACKET
	PLUS
	MINUS
	MULTIPLY
	FLOATDIVIDE
	INTDIVIDE
	LESSTHAN
	MORETHAN
	COLON
	ASSIGN
	BANG
	MOD
)

// Position is the line and column and width of a token
type Position struct {
	Line   int
	Column int
	Width  int
}

// Token is one or more characters that are grouped together to add meaning
type Token struct {
	Type  TokenType
	Pos   Position
	Value interface{}
}

func (t Token) String() string {
	return fmt.Sprintf("Type: %s, Line: %d, Column: %d, Width: %d, Value: %+s", t.Type.String(), t.Pos.Line, t.Pos.Column, t.Pos.Width, t.Value)
}
