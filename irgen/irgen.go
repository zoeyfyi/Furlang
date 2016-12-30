package irgen

import (
	"fmt"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/goory"
	goorytypes "github.com/bongo227/goory/types"
	gooryvalues "github.com/bongo227/goory/value"
)

type Irgen struct {
	root            ast.Ast
	currentFunction ast.Function

	module *goory.Module
	block  *goory.Block
}

func NewIrgen(ast ast.Ast) *Irgen {
	return &Irgen{
		root:   ast,
		module: goory.NewModule("furlang"),
	}
}

func (g *Irgen) Generate() string {
	for _, function := range g.root.Functions {
		g.function(function)
	}

	return g.module.LLVM()
}

func (g *Irgen) typ(node ast.Type) goorytypes.Type {
	return node.Llvm()
}

func (g *Irgen) function(node ast.Function) {
	g.currentFunction = node

	// Add function to module
	function := g.module.NewFunction(node.Name.Value, g.typ(node.Type.Returns[0]))
	for _, arg := range node.Type.Parameters {
		function.AddArgument(g.typ(arg.Type), arg.Ident.Value)
	}
	g.block = function.Entry()

	// Add expressions to function body
	for _, exp := range node.Body.Expressions {
		g.expression(exp)
	}
}

func (g *Irgen) expression(node ast.Expression) gooryvalues.Value {
	switch node := node.(type) {
	case ast.Return:
		return g.ret(node)
	case ast.Integer:
		return g.integer(node)
	}

	panic(fmt.Sprintf("Node not handled: %+v", node))
}

func (g *Irgen) ret(node ast.Return) gooryvalues.Value {
	value := g.expression(node.Value)
	value = g.block.Cast(value, g.currentFunction.Type.Returns[0].Llvm())
	g.block.Ret(value)
	return nil
}

func (g *Irgen) integer(node ast.Integer) gooryvalues.Value {
	return goory.Constant(goory.IntType(64), node.Value)
}
