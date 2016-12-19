package ast

import (
	"github.com/bongo227/Furlang/lexer"
	types "github.com/bongo227/Furlang/types"
)

type Expression interface {
	expressionNode()
}

// Types
type (
	// Type : type
	Type struct {
		Type types.Type
	}

	// ArrayType : type[length]
	ArrayType struct {
		Length int
		Type   Type
	}

	// TypedIdent : type identifier
	TypedIdent struct {
		Type  Type
		Ident lexer.Token
	}

	// StructType : {
	//    type expressions
	//    ...
	// }
	StructType struct {
		Items []TypedIdent
	}

	// FunctionType : type identifier, type identifier, ... -> type, type, ...
	FunctionType struct {
		Parameters []TypedIdent
		Returns    []Type
	}
)

// Operations
type (
	// Binary operation i.e. +, - etc
	Binary struct {
		Lhs Expression
		Op  lexer.Token
		Rhs Expression
	}

	// Unary operation i.e. +, - etc
	Unary struct {
		Expression Expression
		Op         lexer.Token
	}

	Cast struct {
		Expression Expression
		Type       Type
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

func (b Ident) expressionNode() {}

// Values
type (
	Function struct {
		Type FunctionType
		Name Ident
		Body Block
	}

	Assignment struct {
		Type       Type
		Name       Ident
		Expression Expression
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
)

func (b Function) expressionNode()   {}
func (b Assignment) expressionNode() {}
func (b Integer) expressionNode()    {}
func (b Float) expressionNode()      {}
func (b Call) expressionNode()       {}

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

type Ast struct {
	Functions []Function
}
