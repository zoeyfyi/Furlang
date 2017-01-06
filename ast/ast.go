package ast

import (
	"github.com/bongo227/Furlang/lexer"
	types "github.com/bongo227/Furlang/types"
)

// TODO: split into expressions and Values
// expressions are things like typedefs, blocks ets
// values are integers, lists etc

type Expression interface {
	expressionNode()
}

// Types
type (
	// TypedIdent : type identifier
	TypedIdent struct {
		Type  types.Type
		Ident Ident
	}

	// FunctionType : type identifier, type identifier, ... -> type, type, ...
	FunctionType struct {
		Parameters []TypedIdent
		Return     types.Type
	}
)

// Operations
type (
	// Binary operation i.e. +, - etc
	Binary struct {
		Lhs  Expression
		Op   lexer.TokenType
		Rhs  Expression
		IsFp bool
	}

	// Unary operation i.e. +, - etc
	Unary struct {
		Expression Expression
		Op         lexer.Token
	}

	Cast struct {
		Expression Expression
		Type       types.Type
	}
)

func (b Binary) expressionNode() {}
func (b Unary) expressionNode()  {}
func (b Cast) expressionNode()   {}

// Constructs
type (
	Ident struct {
		Value string
	}

	ParenList struct {
		Expressions []Expression
	}

	SquareList struct {
		Expressions []Expression
	}

	List struct {
		Expressions []Expression
	}

	ArrayList struct {
		Type types.Type
		List List
	}

	Index struct {
		Index Expression
	}

	Slice struct {
		Low  Expression
		High Expression
	}

	Block struct {
		Expressions []Expression
	}
)

func (b Ident) expressionNode()     {}
func (b List) expressionNode()      {}
func (b ArrayList) expressionNode() {}
func (b Block) expressionNode()     {}

// Values
type (
	Function struct {
		Type FunctionType
		Name Ident
		Body Block
	}

	Assignment struct {
		Type       types.Type
		Name       Ident
		Expression Expression
		Declare    bool
	}

	Integer struct {
		Value int64
	}

	Float struct {
		Value float64
	}

	Call struct {
		Function  Ident
		Arguments []Expression
	}

	ArrayValue struct {
		Array Ident
		Index Index
	}

	Return struct {
		Value Expression
	}
)

func (b Function) expressionNode()   {}
func (b Assignment) expressionNode() {}
func (b Integer) expressionNode()    {}
func (b Float) expressionNode()      {}
func (b Call) expressionNode()       {}
func (b ArrayValue) expressionNode() {}
func (b Return) expressionNode()     {}

// Flow control
type (
	If struct {
		Condition Expression
		Block     Block
		Else      *If
	}

	For struct {
		Index     Expression
		Condition Expression
		Increment Expression
		Block     Block
	}
)

func (b If) expressionNode()  {}
func (b For) expressionNode() {}

type Ast struct {
	Functions []Function
}
