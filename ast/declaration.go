package ast

import (
	"go/types"

	"github.com/bongo227/Furlang/lexer"
)

type Declare interface {
	declareNode()
}

// FunctionDeclaration is a declare node in the form:
// ident :: type ident, ... -> type { statement; ... }
type FunctionDeclaration struct {
	Name      *IdentExpression
	Define    lexer.Token
	Arguments map[string]types.Type
	Return    types.Type
	Body      *BlockStatement
}
