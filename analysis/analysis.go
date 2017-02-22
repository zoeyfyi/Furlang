package analysis

import (
	"reflect"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
)

import "fmt"

import "log"

var (
	intType   = types.IntType(0)
	floatType = types.FloatType(0)
)

// Analysis checks the semantics of the abstract syntax tree and adds any allowed
// omisions such as type inference
type Analysis struct {
	root            *ast.Ast
	currentBlock    *ast.BlockStatement
	currentFunction *ast.FunctionDeclaration
}

func NewAnalysis(root *ast.Ast) *Analysis {
	return &Analysis{
		root: root,
	}
}

func (a *Analysis) Analalize() *ast.Ast {
	log.Println("Analasis Started")

	for _, f := range a.root.Functions {
		a.function(f)
	}

	return a.root
}

// Gets the type of a node
func (a *Analysis) typ(node ast.Node) types.Type {
	switch node := node.(type) {
	case *ast.LiteralExpression:
		switch node.Value.Type() {
		case lexer.INT:
			return intType
		case lexer.FLOAT:
			return floatType
		}
		panic(fmt.Sprintf("Unregcognized literal type: %s", node.Value.Type().String()))

	case *ast.BinaryExpression:
		lType := a.typ(node.Left)
		rType := a.typ(node.Right)
		log.Printf("lType: %t, rType: %t", lType == floatType, rType == floatType)
		if lType == floatType || rType == floatType {
			return floatType
		}
		return intType

	case *ast.CallExpression:
		return a.typ(node.Function)

	case *ast.CastExpression:
		return node.Type

	case *ast.IdentExpression:
		return a.typ(a.currentBlock.Scope.Lookup(node.Value.Value()))

	case *ast.VaribleDeclaration:
		return node.Type

	case *ast.FunctionDeclaration:
		argTypes := make([]types.Type, len(node.Arguments))
		for i, arg := range node.Arguments {
			argTypes[i] = arg.Type
		}

		return types.NewFunction(node.Return, argTypes...)

	default:
		panic(fmt.Sprintf("Unknown type: %s", reflect.TypeOf(node).String()))
	}
}

func (a *Analysis) block(node *ast.BlockStatement) {
	a.currentBlock = node

	for _, e := range node.Statements {
		a.statement(e)
	}
}

func (a *Analysis) function(node *ast.FunctionDeclaration) {
	a.currentFunction = node
	a.block(node.Body)
}

func (a *Analysis) varible(node *ast.VaribleDeclaration) {
	exp := a.expression(node.Value)
	if node.Type == nil {
		node.Type = a.typ(node.Value)
	} else {
		node.Value = &ast.CastExpression{
			Type:       node.Type,
			Expression: exp,
		}
	}
}

func (a *Analysis) declare(node ast.Declare) {
	switch node := (node).(type) {
	case *ast.VaribleDeclaration:
		a.varible(node)
	}
}

func (a *Analysis) statement(node ast.Statement) {
	switch node := (node).(type) {
	case *ast.AssignmentStatement:
		a.assigment(node)
	case *ast.ForStatement:
		a.forNode(node)
	case *ast.IfStatment:
		a.ifNode(node)
	case *ast.BlockStatement:
		a.block(node)
	case *ast.ReturnStatement:
		a.returnNode(node)
	case *ast.DeclareStatement:
		a.declare(node.Statement)
	}
}

func (a *Analysis) expression(node ast.Expression) ast.Expression {
	switch node := (node).(type) {
	case *ast.BinaryExpression:
		a.binary(node)
	case *ast.CallExpression:
		a.call(node)
	case *ast.LiteralExpression:
		// Do nothing
	default:
		fmt.Printf("Unhandled: %s\n", reflect.TypeOf(node).String())
	}

	return node
}

func (a *Analysis) returnNode(node *ast.ReturnStatement) {
	a.expression(node.Result)
	if typ := a.typ(node.Result); typ != a.currentFunction.Return {
		node.Result = &ast.CastExpression{
			Expression: node.Result,
			Type:       a.currentFunction.Return,
		}
	}
}

func (a *Analysis) forNode(node *ast.ForStatement) {
	log.Println("For")

	a.statement(node.Index)
	a.expression(node.Condition)
	a.statement(node.Increment)
	a.block(node.Body)
}

func (a *Analysis) ifNode(node *ast.IfStatment) {
	log.Println("If")

	if node.Condition != nil {
		a.expression(node.Condition)
	}

	a.block(node.Body)

	if node.Else != nil {
		a.ifNode(node.Else)
	}
}

func (a *Analysis) declareVar(node *ast.VaribleDeclaration) {
	valueType := a.typ(node.Value)

	if node.Type == nil {
		node.Type = valueType
	} else if node.Type.Llvm() != valueType.Llvm() {
		node.Value = &ast.CastExpression{
			Expression: node.Value,
			Type:       valueType,
		}
	}
}

func (a *Analysis) assigment(node *ast.AssignmentStatement) {
	// Get type of assigment expression
	leftType := a.typ(node.Left)
	rightType := a.typ(node.Right)

	// Expression doesnt match assigment type
	if leftType.Llvm() != rightType.Llvm() {
		// Cast it
		node.Right = ast.Expression(&ast.CastExpression{
			Expression: node.Right,
			Type:       leftType,
		})
	}
}

func (a *Analysis) call(node *ast.CallExpression) {
	funcType := a.typ(node.Function).(*types.Function)
	for i, arg := range node.Arguments.Elements {
		aType := funcType.Arguments()[i]
		if typ := a.typ(arg); typ.Llvm() != aType.Llvm() {
			node.Arguments.Elements[i] = &ast.CastExpression{
				Expression: arg,
				Type:       aType,
			}
		}
	}
}

func (a *Analysis) binary(node *ast.BinaryExpression) {
	log.Printf("Binary %s node", node.Operator.String())

	// Gets the overall type of node
	typ := a.typ(node)
	node.IsFp = typ.(*types.Basic).Info()&types.IsFloat != 0

	// If left part of the node doesnt match the type of the node cast it
	if leftTyp := a.typ(node.Left); leftTyp != typ {
		node.Left = &ast.CastExpression{
			Expression: node.Left,
			Type:       typ,
		}
	}

	// If the right part of the node doesnt match the type of the node cast it
	if rightTyp := a.typ(node.Right); rightTyp != typ {
		node.Right = &ast.CastExpression{
			Expression: node.Right,
			Type:       typ,
		}
	}
}
