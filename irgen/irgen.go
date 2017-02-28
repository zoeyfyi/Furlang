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

	instructions "github.com/bongo227/goory/instructions"
	gtypes "github.com/bongo227/goory/types"
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
	g.parentBlock = f.Entry()

	// Add arguments to function
	for _, arg := range node.Arguments {
		name := arg.Name.Value.Value()
		argType := arg.Type.Llvm()
		arg := f.AddArgument(argType, name)

		alloc := g.parentBlock.Alloca(argType)
		g.parentBlock.Store(alloc, arg)
		g.scope.AddVar(name, alloc)
	}

	g.block(node.Body)
}

// TODO: remove this
func (g *Irgen) block(node *ast.BlockStatement) {
	// Push new scope
	g.scope = g.scope.Push()

	for _, smt := range node.Statements {
		g.statement(smt)
	}
}

func (g *Irgen) newBlock(node *ast.BlockStatement) (start, end *goory.Block) {
	startBlock := g.parentBlock.Function().AddBlock()
	parent := g.parentBlock
	g.parentBlock = startBlock
	g.block(node)
	endBlock := g.parentBlock
	g.parentBlock = parent

	return startBlock, endBlock
}

func (g *Irgen) statement(node ast.Statement) {
	log.Printf("Statement of type %q", reflect.TypeOf(node).String())

	switch node := node.(type) {
	case *ast.IfStatment:
		endBlock := g.parentBlock.Function().AddBlock()
		g.ifSmt(node, nil, endBlock)
	case *ast.ReturnStatement:
		g.returnSmt(node)
	case *ast.DeclareStatement:
		g.declareSmt(node)
	case *ast.AssignmentStatement:
		g.assignmentSmt(node)
	case *ast.ForStatement:
		g.forSmt(node)
	}
}

func (g *Irgen) ifSmt(node *ast.IfStatment, block, endBlock *goory.Block) {
	parent := g.parentBlock
	if block == nil {
		block = g.parentBlock.Function().AddBlock()
	}

	// falseBlock either points to and else/else if block to be generated or
	// the last block to continue execution
	falseBlock := endBlock
	if node.Else != nil {
		falseBlock = g.parentBlock.Function().AddBlock()
	}

	// Generate true block
	g.parentBlock = block
	g.block(node.Body)
	// Didnt terminate block so continue exection at end block
	if !block.Terminated() {
		block.Br(endBlock)
	}

	// Add the conditional branch
	if node.Condition != nil {
		g.parentBlock = parent
		condition := g.expression(node.Condition)
		parent.CondBr(condition, block, falseBlock)
	}

	g.parentBlock = falseBlock

	// Check for else statement
	if node.Else != nil {
		g.ifSmt(node.Else, falseBlock, endBlock)
	}

	g.parentBlock = endBlock
}

func (g *Irgen) returnSmt(node *ast.ReturnStatement) {
	exp := g.expression(node.Result)
	g.parentBlock.Ret(exp)
}

func (g *Irgen) declareSmt(node *ast.DeclareStatement) {
	// TODO: handle function declarations
	decl := node.Statement.(*ast.VaribleDeclaration)
	name := decl.Name.Value.Value()

	log.Printf("Declaring %q", name)

	// TODO: check this didnt break anything
	// alloc := g.parentBlock.Alloca(decl.Type.Base().Llvm())
	alloc := g.parentBlock.Alloca(decl.Type.Llvm())

	g.scope.AddVar(name, alloc)

	if _, ok := decl.Type.(*types.Array); ok {
		g.arraySmt(decl.Value, alloc)
	} else {
		exp := g.expression(decl.Value)
		g.parentBlock.Store(alloc, exp)
	}

}

func (g *Irgen) arraySmt(node ast.Expression, alloc *instructions.Alloca) {
	switch node := node.(type) {
	case *ast.BraceLiteralExpression:
		for i, exp := range node.Elements {
			// Analyse element of array
			emt := g.expression(exp)

			// Get a pointer to the index of the array
			ptr := g.parentBlock.Getelementptr(node.Type.Base().Llvm(), alloc,
				goory.Constant(goory.IntType(64), 0),
				goory.Constant(goory.IntType(64), i))

			// Store element in the array
			g.parentBlock.Store(ptr, emt)
		}
	case *ast.CallExpression:
		array := g.expression(node)
		g.parentBlock.Store(alloc, array)
	default:
		log.Fatalf("Cant assign type %q to array type", reflect.TypeOf(node).String())
	}
}

func (g *Irgen) assignmentSmt(node *ast.AssignmentStatement) {
	switch leftNode := node.Left.(type) {
	case *ast.IdentExpression:
		name := leftNode.Value.Value()
		exp := g.expression(node.Right)

		alloc, ok := g.scope.GetVar(name)
		if !ok {
			log.Fatalf("%q was not in scope", name)
		}
		g.parentBlock.Store(alloc, exp)
		g.scope.AddVar(name, alloc)
	case *ast.IndexExpression:
		name := leftNode.Expression.(*ast.IdentExpression).Value.Value()
		index := g.expression(leftNode.Index)
		exp := g.expression(node.Right)

		alloc, ok := g.scope.GetVar(name)
		if !ok {
			log.Fatalf("%q was not in scope", name)
		}

		arrayType := alloc.BaseType().(gtypes.ArrayType).BaseType()
		arrayIndex := g.parentBlock.Getelementptr(arrayType, alloc,
			goory.Constant(goory.IntType(64), 0), index)

		g.parentBlock.Store(arrayIndex, exp)
		g.scope.AddVar(name, alloc)
	}

}

func (g *Irgen) forSmt(node *ast.ForStatement) {
	g.statement(node.Index)

	// Branch into for loop
	outerCondition := g.expression(node.Condition)
	start, end := g.newBlock(node.Body)
	continueBlock := g.parentBlock.Function().AddBlock()
	g.parentBlock.CondBr(outerCondition, start, continueBlock)

	// Branch to continue of exit
	g.parentBlock = end
	g.statement(node.Increment)
	innerCondition := g.expression(node.Condition)
	g.parentBlock.CondBr(innerCondition, start, continueBlock)

	// Set continue block
	g.parentBlock = continueBlock
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
	case *ast.IndexExpression:
		return g.indexExp(node)
	default:
		panic(fmt.Sprintf("Unknown expression node: %s", pp.Sprint(node)))
	}
}

func (g *Irgen) indexExp(node *ast.IndexExpression) gooryvalues.Value {
	alloc, _ := g.scope.GetVar(node.Expression.(*ast.IdentExpression).Value.Value())
	index := g.expression(node.Index)
	// TODO: Make getting base type cleaner
	ptr := g.parentBlock.Getelementptr(alloc.BaseType().(gtypes.ArrayType).BaseType(), alloc, goory.Constant(goory.IntType(32), 0), index)
	value := g.parentBlock.Load(ptr)

	return value
}

func (g *Irgen) callExp(node *ast.CallExpression) gooryvalues.Value {
	// TODO: handle lambda's (i.e. functions that are not called by name)
	funcName := node.Function.(*ast.IdentExpression).Value.Value()

	log.Printf("Function name: %q", funcName)
	function, ok := g.scope.GetFunction(funcName)
	if !ok {
		log.Fatalf("Function %q not in scope", funcName)
	}

	args := make([]gooryvalues.Value, len(node.Arguments.Elements))
	for i, element := range node.Arguments.Elements {
		args[i] = g.expression(element)
	}

	return g.parentBlock.Call(function, args...)
}

func (g *Irgen) identExp(node *ast.IdentExpression) gooryvalues.Value {
	ident := node.Value.Value()

	// TODO: do this with a map
	if ident == "true" {
		return goory.Constant(goory.BoolType(), true)
	} else if ident == "false" {
		return goory.Constant(goory.BoolType(), false)
	}

	item, ok := g.scope.GetVar(ident)
	if !ok {
		log.Fatalf("%q was not is scope", ident)
	}

	return g.parentBlock.Load(item)
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
	return g.parentBlock.Cast(exp, node.Type.Llvm())
}

func (g *Irgen) binaryExp(node *ast.BinaryExpression) gooryvalues.Value {
	left := g.expression(node.Left)
	right := g.expression(node.Right)

	log.Printf("Is fp: %t", node.IsFp)

	log.Printf("Left is %q, right is %q", left.Type().String(), right.Type().String())

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
			return g.parentBlock.Icmp(goory.IntEq, left, right)
		case lexer.NEQ:
			return g.parentBlock.Icmp(goory.IntNe, left, right)
		case lexer.GTR:
			return g.parentBlock.Icmp(goory.IntSgt, left, right)
		case lexer.LSS:
			return g.parentBlock.Icmp(goory.IntSlt, left, right)
		case lexer.REM:
			return g.parentBlock.Srem(left, right)
		}
	}

	panic("Unhandled binary operator")
}
