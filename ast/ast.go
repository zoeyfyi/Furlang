package ast

import (
	"fmt"

	"github.com/bongo227/Furlang/lexer"
	types "github.com/bongo227/Furlang/types"
	goorytypes "github.com/bongo227/goory/types"
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
		Llvm() goorytypes.Type
	}

	// Basic : type
	Basic struct {
		Type types.Type
	}

	// ArrayType : type[length]
	ArrayType struct {
		Type   Type
		Length Integer
	}

	// TypedIdent : type identifier
	TypedIdent struct {
		Type  Type
		Ident Ident
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

func NewBasic(ident string) *Basic {
	switch ident {
	case "int":
		return &Basic{
			Type: types.IntType(0),
		}
	case "i8":
		return &Basic{
			Type: types.IntType(8),
		}
	case "i16":
		return &Basic{
			Type: types.IntType(16),
		}
	case "i32":
		return &Basic{
			Type: types.IntType(32),
		}
	case "i64":
		return &Basic{
			Type: types.IntType(64),
		}
	case "bool":
		return &Basic{
			Type: types.BasicBool,
		}
	}

	panic(fmt.Sprintf("Unrecognized basic type: %s", ident))
}

func (t Basic) typeNode()             {}
func (t Basic) Llvm() goorytypes.Type { return t.Type.Llvm() }

func (t ArrayType) typeNode()             {}
func (t ArrayType) Llvm() goorytypes.Type { return t.Type.Llvm() }

// func (t StructType) typeNode()             {}
// func (t StructType) Llvm() goorytypes.Type { return t.Type.Llvm() }

// func (t FunctionType) typeNode()             {}
// func (t FunctionType) Llvm() goorytypes.Type { return t.Type.Llvm() }

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
