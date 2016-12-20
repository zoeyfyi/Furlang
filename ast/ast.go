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
	Type interface {
		typeNode()
	}

	// Basic : type
	Basic struct {
		// Ident on the initial pass this is the type identifier during
		// semantic analysis we use this value to derive the actual type
		Ident Ident
		Type  types.Type
	}

	// ArrayType : type[length]
	ArrayType struct {
		Type   Type
		Length Integer
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

func (t Basic) typeNode()        {}
func (t ArrayType) typeNode()    {}
func (t StructType) typeNode()   {}
func (t FunctionType) typeNode() {}

// Operations
type (
	// Binary operation i.e. +, - etc
	Binary struct {
		Lhs Expression
		Op  lexer.TokenType
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
func (b List) expressionNode()  {}
func (b Block) expressionNode() {}

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
