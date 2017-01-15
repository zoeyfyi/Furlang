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

// Analysis checks the semantics of the abstract syntax tree and adds any allowed omisions
// such as type inference
type Analysis struct {
	root *ast.Ast
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
func (a *Analysis) typ(node ast.Expression) (types.Type, error) {
	switch node := node.(type) {
	case *ast.LiteralExpression:
		switch node.Value.Type() {
		case lexer.INT:
			return intType, nil
		case lexer.FLOAT:
			return floatType, nil
		}
		return nil, fmt.Errorf("Unregcognized literal type: %s", node.Value.Type().String())

	case *ast.BinaryExpression:
		lType, _ := a.typ(node.Left)
		rType, _ := a.typ(node.Right)
		if lType == floatType || rType == floatType {
			return floatType, nil
		}
		return intType, nil

	case *ast.CallExpression:
		return a.typ(node.Function)

	// case ast.ArrayList:
	// 	return node.Type, nil

	default:
		return nil, fmt.Errorf("Unknown type: %s", reflect.TypeOf(node).String())
	}
}

func (a *Analysis) block(node *ast.BlockStatement) {
	for _, e := range node.Statements {
		a.statement(e)
	}
}

func (a *Analysis) function(node *ast.FunctionDeclaration) {
	a.block(node.Body)
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
	}
}

func (a *Analysis) expression(node ast.Expression) ast.Expression {
	switch node := (node).(type) {
	case *ast.BinaryExpression:
		a.binary(node)
	case *ast.CallExpression:
		a.call(node)
	default:
		fmt.Printf("Unhandled: %s\n", reflect.TypeOf(node).String())
	}

	return node
}

func (a *Analysis) returnNode(node *ast.ReturnStatement) {
	a.expression(node.Result)
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

func (a *Analysis) assigment(node *ast.AssignmentStatement) {
	// Get type of assigment expression
	leftType, err := a.typ(node.Left)
	if err != nil {
		panic(err)
	}
	rightType, err := a.typ(node.Right)
	if err != nil {
		panic(err)
	}

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
	// for _, f := range a.root.Functions {
	// 	if f.Name.Value == node.Function.Value {
	// 		for i, arg := range node.Arguments {
	// 			// Check if parameters match argument type
	// 			fType := f.Type.Parameters[i].Type
	// 			if typ, _ := a.typ(arg); typ.Llvm() != fType.Llvm() {
	// 				newCall.Arguments[i] = ast.Cast{
	// 					Expression: arg,
	// 					Type:       fType,
	// 				}
	// 			} else {
	// 				newCall.Arguments[i] = arg
	// 			}
	// 		}
	// 		break
	// 	}
	// }

	// return newCall
}

func (a *Analysis) binary(node *ast.BinaryExpression) {
	log.Printf("Binary %s node", node.Operator.String())

	typ, _ := a.typ(node)

	// If left part of the node doesnt match the type of the node cast it
	if leftTyp, _ := a.typ(node.Left); leftTyp != typ {
		node.Left = &ast.CastExpression{
			Expression: node.Left,
			Type:       typ,
		}
	}

	// If the right part of the node doesnt match the type of the node cast it
	if rightTyp, _ := a.typ(node.Right); rightTyp != typ {
		node.Right = &ast.CastExpression{
			Expression: node.Right,
			Type:       typ,
		}
	}
}
