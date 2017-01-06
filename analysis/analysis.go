package analysis

import (
	"reflect"

	"github.com/bongo227/Furlang/ast"
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

	newAst := ast.Ast{
		Functions: make([]ast.Function, len(a.root.Functions)),
	}

	for i, f := range a.root.Functions {
		newAst.Functions[i] = *a.function(&f)
	}

	return &newAst
}

// Gets the type of a node
func (a *Analysis) typ(node ast.Expression) (types.Type, error) {
	switch node := node.(type) {
	case ast.Integer:
		return intType, nil
	case ast.Float:
		return floatType, nil
	case ast.Binary:
		lType, _ := a.typ(node.Lhs)
		rType, _ := a.typ(node.Rhs)
		if lType == floatType || rType == floatType {
			return floatType, nil
		}
		return intType, nil
	case ast.Call:
		for _, f := range a.root.Functions {
			if f.Name.Value == node.Function.Value {
				return f.Type.Return, nil
			}
		}

		return nil, fmt.Errorf("No function named %q", node.Function.Value)
	default:
		return nil, fmt.Errorf("Unknown type: %s", reflect.TypeOf(node).String())
	}
}

func (a *Analysis) block(node *ast.Block) *ast.Block {
	newBlock := &ast.Block{
		Expressions: make([]ast.Expression, len(node.Expressions)),
	}

	for i, e := range node.Expressions {
		newBlock.Expressions[i] = a.expression(e)
	}

	return newBlock
}

func (a *Analysis) function(node *ast.Function) *ast.Function {
	newFunction := &ast.Function{
		Name: node.Name,
		Type: node.Type,
		Body: *a.block(&node.Body),
	}

	return newFunction
}

func (a *Analysis) expression(node ast.Expression) ast.Expression {

	switch node := (node).(type) {
	case ast.Assignment:
		return a.assigment(&node)
	case *ast.For:
		return a.forNode(node)
	case *ast.If:
		return a.ifNode(node)
	case ast.Block:
		return a.block(&node)
	case ast.Binary:
		return a.binary(&node)
	case ast.Call:
		return a.call(&node)
	case ast.Return:
		return a.returnNode(&node)
	default:
		fmt.Printf("Unhandled: %s\n", reflect.TypeOf(node).String())
	}

	return node
}

func (a *Analysis) returnNode(node *ast.Return) ast.Expression {
	return ast.Return{
		Value: a.expression(node.Value),
	}
}

func (a *Analysis) forNode(node *ast.For) ast.For {
	log.Println("For")

	node.Index = a.expression(node.Index).(ast.Assignment)
	node.Condition = a.expression(node.Condition)
	node.Increment = a.expression(node.Increment)
	node.Block = *a.block(&node.Block)

	return *node
}

func (a *Analysis) ifNode(node *ast.If) ast.If {
	log.Println("If")

	if node.Condition != nil {
		node.Condition = a.expression(node.Condition)
	}
	node.Block = *a.block(&node.Block)

	if node.Else != nil {
		newElse := a.ifNode(node.Else)
		node.Else = &newElse
	}

	return *node
}

func (a *Analysis) assigment(node *ast.Assignment) ast.Assignment {
	newAssign := ast.Assignment{
		Name:       node.Name,
		Expression: a.expression(node.Expression),
		Declare:    node.Declare,
	}

	// Get type of assigment expression
	nodeType, err := a.typ(newAssign.Expression)
	if err != nil {
		panic(err)
	}

	// Infer assigment type
	if node.Type == nil {
		newAssign.Type = nodeType
		return newAssign
	}

	newAssign.Type = node.Type

	// Expression doesnt match assigment type
	if node.Type.Llvm() != nodeType.Llvm() {
		// Cast it
		expression := ast.Expression(ast.Cast{
			Expression: node.Expression,
			Type:       node.Type,
		})
		newAssign.Expression = expression
	}

	return newAssign
}

func (a *Analysis) call(node *ast.Call) ast.Expression {
	newCall := ast.Call{
		Function:  node.Function,
		Arguments: make([]ast.Expression, len(node.Arguments)),
	}

	for _, f := range a.root.Functions {
		if f.Name.Value == node.Function.Value {
			for i, arg := range node.Arguments {
				// Check if parameters match argument type
				fType := f.Type.Parameters[i].Type
				if typ, _ := a.typ(arg); typ.Llvm() != fType.Llvm() {
					newCall.Arguments[i] = ast.Cast{
						Expression: arg,
						Type:       fType,
					}
				} else {
					newCall.Arguments[i] = arg
				}
			}
			break
		}
	}

	return newCall
}

func (a *Analysis) binary(node *ast.Binary) ast.Expression {
	log.Printf("Binary %s node", node.Op.String())

	typ, _ := a.typ(*node)
	node.IsFp = typ == floatType

	// If left part of the node doesnt match the type of the node cast it
	if leftTyp, _ := a.typ(node.Lhs); leftTyp != typ {
		node.Lhs = ast.Cast{
			Expression: node.Lhs,
			Type:       typ,
		}
	}

	// If the right part of the node doesnt match the type of the node cast it
	if rightTyp, _ := a.typ(node.Rhs); rightTyp != typ {
		node.Rhs = ast.Cast{
			Expression: node.Rhs,
			Type:       typ,
		}
	}

	return *node
}
