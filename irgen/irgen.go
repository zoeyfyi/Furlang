package irgen

import (
	"fmt"
	"strconv"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
	"github.com/bongo227/goory"

	"log"

	gooryvalues "github.com/bongo227/goory/value"
	"github.com/k0kubun/pp"
)

type Irgen struct {
	tree        *ast.Ast
	module      *goory.Module
	values      map[string]*gooryvalues.Value
	parentBlock *goory.Block
	parentScope *ast.Scope
}

func NewIrgen(tree *ast.Ast) *Irgen {
	return &Irgen{
		tree:   tree,
		module: goory.NewModule("test"),
		values: make(map[string]*gooryvalues.Value),
	}
}

func (g *Irgen) Generate() string {
	for _, f := range g.tree.Functions {
		g.function(f)
	}

	return g.module.LLVM()
}

func (g *Irgen) function(node *ast.FunctionDeclaration) {
	// Create new function in module
	f := g.module.NewFunction(node.Name.Value.Value(), node.Return.Llvm())

	// Add arguments to function
	for _, arg := range node.Arguments {
		f.AddArgument(arg.Type.Llvm(), arg.Name.Value.Value())
	}

	g.parentBlock = f.Entry()
	g.block(node.Body)
}

func (g *Irgen) block(node *ast.BlockStatement) {
	g.parentScope = node.Scope
	for _, smt := range node.Statements {
		g.statement(smt)
	}
}

func (g *Irgen) statement(node ast.Statement) {
	switch node := node.(type) {
	case *ast.ReturnStatement:
		g.returnSmt(node)
	}
}

func (g *Irgen) returnSmt(node *ast.ReturnStatement) {
	exp := g.expression(node.Result)
	g.parentBlock.Ret(exp)
}

func (g *Irgen) expression(node ast.Expression) gooryvalues.Value {
	switch node := node.(type) {
	case *ast.CastExpression:
		return g.castExp(node)
	case *ast.BinaryExpression:
		return g.binaryExp(node)
	case *ast.LiteralExpression:
		return g.literalExp(node)
	default:
		panic(fmt.Sprintf("Unknown expression node: %s", pp.Sprint(node)))
	}
}

func (g *Irgen) literalExp(node *ast.LiteralExpression) gooryvalues.Value {
	switch node.Value.Type() {
	case lexer.INT:
		value, _ := strconv.Atoi(node.Value.Value())
		return goory.Constant(types.IntType(0).Llvm(), value)
	case lexer.FLOAT:
		value, _ := strconv.ParseFloat(node.Value.Value(), 64)
		return goory.Constant(types.FloatType(0).Llvm(), value)
	default:
		panic("Unknown literal type")
	}
}

func (g *Irgen) castExp(node *ast.CastExpression) gooryvalues.Value {
	exp := g.expression(node.Expression)
	log.Printf("Casting to: %s", node.Type.Llvm())
	return g.parentBlock.Cast(exp, node.Type.Llvm())
}

func (g *Irgen) binaryExp(node *ast.BinaryExpression) gooryvalues.Value {
	left := g.expression(node.Left)
	right := g.expression(node.Right)

	if node.IsFp {
		switch node.Operator.Type() {
		case lexer.ADD:
			return g.parentBlock.Fadd(left, right)
		case lexer.SUB:
			return g.parentBlock.Fsub(left, right)
		case lexer.MUL:
			return g.parentBlock.Fmul(left, right)
		case lexer.QUO:
			return g.parentBlock.Fdiv(left, right)
		case lexer.EQL:
			return g.parentBlock.Fcmp(goory.FloatOeq, left, right)
		case lexer.NEQ:
			return g.parentBlock.Fcmp(goory.FloatOne, left, right)
		case lexer.GTR:
			return g.parentBlock.Fcmp(goory.FloatOgt, left, right)
		case lexer.LSS:
			return g.parentBlock.Fcmp(goory.FloatOlt, left, right)
		}
	} else {
		switch node.Operator.Type() {
		case lexer.ADD:
			return g.parentBlock.Add(left, right)
		case lexer.SUB:
			return g.parentBlock.Sub(left, right)
		case lexer.MUL:
			return g.parentBlock.Mul(left, right)
		case lexer.QUO:
			return g.parentBlock.Div(left, right)
		case lexer.EQL:
			return g.parentBlock.Icmp(goory.FloatOeq, left, right)
		case lexer.NEQ:
			return g.parentBlock.Icmp(goory.FloatOne, left, right)
		case lexer.GTR:
			return g.parentBlock.Icmp(goory.FloatOgt, left, right)
		case lexer.LSS:
			return g.parentBlock.Icmp(goory.FloatOlt, left, right)
		}
	}

	panic("Unhandled binary operator")
}
