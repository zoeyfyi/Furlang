package irgen

import (
	"fmt"
	"strconv"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
	"github.com/bongo227/goory"

	"log"

	"reflect"

	gooryvalues "github.com/bongo227/goory/value"
	"github.com/k0kubun/pp"
)

type Irgen struct {
	tree        *ast.Ast
	module      *goory.Module
	parentBlock *goory.Block
	scope       *Scope
}

func NewIrgen(tree *ast.Ast) *Irgen {
	return &Irgen{
		tree:   tree,
		module: goory.NewModule("test"),
		scope:  NewScope(),
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
	fName := node.Name.Value.Value()
	f := g.module.NewFunction(fName, node.Return.Llvm())

	g.scope.AddFunction(fName, f)

	// Add arguments to function
	for _, arg := range node.Arguments {
		name := arg.Name.Value.Value()
		arg := f.AddArgument(arg.Type.Llvm(), name)
		g.scope.AddVar(name, arg)
	}

	g.parentBlock = f.Entry()
	g.block(node.Body)
}

func (g *Irgen) block(node *ast.BlockStatement) {
	// Push new scope
	g.scope = g.scope.Push()

	for _, smt := range node.Statements {
		g.statement(smt)
	}
}

func (g *Irgen) statement(node ast.Statement) {
	log.Printf("Statement of type %q", reflect.TypeOf(node).String())

	switch node := node.(type) {
	case *ast.IfStatment:
		block := g.parentBlock.Function().AddBlock()
		endBlock := g.parentBlock.Function().AddBlock()
		g.ifSmt(node, block, endBlock)
	case *ast.ReturnStatement:
		g.returnSmt(node)
	case *ast.DeclareStatement:
		g.declareSmt(node)
	}
}

func (g *Irgen) ifSmt(node *ast.IfStatment, block, endBlock *goory.Block) {
	// parent := g.parentBlock
	if block == nil {
		block = g.parentBlock.Function().AddBlock()
	}

	// falseBlock either points to and else/else if block to be generated or
	// the last block to continue execution
	falseBlock := endBlock
	if node.Else != nil {
		g.parentBlock.Function().AddBlock()
	}

	// Generate true block
	g.parentBlock = block
	g.block(node.Body)
	// Didnt terminate block so continue exection at end block
	if !block.Terminated() {
		g.parentBlock.Br(endBlock)
	}
	g.parentBlock = falseBlock

	// Add the conditional branch
	if node.Condition != nil {
		condition := g.expression(node.Condition)
		g.parentBlock.CondBr(condition, block, falseBlock)
	} else {
		g.parentBlock.Br(block)
	}

	// Check for else statement
	if node.Else != nil {
		g.ifSmt(node.Else, falseBlock, endBlock)
	}

	g.parentBlock = endBlock
}

func (g *Irgen) declareSmt(node *ast.DeclareStatement) {
	// TODO: handle function declarations
	decl := node.Statement.(*ast.VaribleDeclaration)
	name := decl.Name.Value.Value()
	exp := g.expression(decl.Value)
	g.scope.AddVar(name, exp)
}

func (g *Irgen) returnSmt(node *ast.ReturnStatement) {
	exp := g.expression(node.Result)
	g.parentBlock.Ret(exp)
}

func (g *Irgen) expression(node ast.Expression) gooryvalues.Value {
	switch node := node.(type) {
	case *ast.BinaryExpression:
		return g.binaryExp(node)
	case *ast.CastExpression:
		return g.castExp(node)
	case *ast.LiteralExpression:
		return g.literalExp(node)
	case *ast.IdentExpression:
		return g.identExp(node)
	case *ast.CallExpression:
		return g.callExp(node)
	default:
		panic(fmt.Sprintf("Unknown expression node: %s", pp.Sprint(node)))
	}
}

func (g *Irgen) callExp(node *ast.CallExpression) gooryvalues.Value {
	// TODO: handle lambda's (i.e. functions that are not called by name)
	funcName := node.Function.(*ast.IdentExpression).Value.Value()
	function, _ := g.scope.GetFunction(funcName)

	args := make([]gooryvalues.Value, len(node.Arguments.Elements))
	for i, element := range node.Arguments.Elements {
		args[i] = g.expression(element)
	}

	return g.parentBlock.Call(function, args...)
}

func (g *Irgen) identExp(node *ast.IdentExpression) gooryvalues.Value {
	ident := node.Value.Value()
	item, ok := g.scope.GetVar(ident)
	if !ok {
		log.Fatalf("%q was not is scope", ident)
	}

	return item
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
