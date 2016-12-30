package irgen

import (
	"fmt"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/goory"
	goorytypes "github.com/bongo227/goory/types"
	gooryvalues "github.com/bongo227/goory/value"
)

type scope map[string]gooryvalues.Value

type Irgen struct {
	root            ast.Ast
	currentFunction ast.Function

	scopes       []scope
	currentScope int

	module *goory.Module
	block  *goory.Block
}

func NewIrgen(ast ast.Ast) *Irgen {
	scopes := make([]scope, 1000)
	for i := range scopes {
		scopes[i] = make(scope)
	}

	return &Irgen{
		root:         ast,
		scopes:       scopes,
		currentScope: 0,
		module:       goory.NewModule("furlang"),
	}
}

// Finds a scoped value
func (g *Irgen) find(v string) gooryvalues.Value {
	// Start at current scope and work backwords until the value is found
	search := g.currentScope
	for search >= 0 {
		if value, ok := g.scopes[search][v]; ok {
			return value
		}
		search--
	}

	panic(fmt.Sprintf("Varible not in scope: %s", v))
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
		g.ret(node)
	case ast.Assignment:
		g.assignment(node)
	case ast.Integer:
		return g.integer(node)
	case ast.Ident:
		return g.find(node.Value)
	default:
		panic(fmt.Sprintf("Node not handled: %+v", node))
	}

	return nil
}

func (g *Irgen) assignment(node ast.Assignment) {
	typ := g.typ(node.Type)
	alloc := g.block.Alloca(typ)
	value := g.expression(node.Expression)
	value = g.block.Cast(value, typ)
	// Store value in current scope
	g.scopes[g.currentScope][node.Name.Value] = value
	g.block.Store(alloc, value)
}

func (g *Irgen) ret(node ast.Return) {
	value := g.expression(node.Value)
	value = g.block.Cast(value, g.currentFunction.Type.Returns[0].Llvm())
	g.block.Ret(value)
}

func (g *Irgen) integer(node ast.Integer) gooryvalues.Value {
	return goory.Constant(goory.IntType(64), node.Value)
}
