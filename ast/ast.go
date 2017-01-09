package ast

import "github.com/bongo227/Furlang/lexer"

type Node interface {
	// First returns the first token beloning to the node
	First() lexer.Token
	// Last returnst the last token beloning to the node
	Last() lexer.Token
}

type Ast struct {
	Functions []*FunctionDeclaration
}
