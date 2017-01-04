package analysis

import "github.com/bongo227/Furlang/ast"
import "github.com/bongo227/Furlang/types"

import "fmt"
import "reflect"

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
				return f.Type.Returns[0], nil
			}
		}
	default:
		return nil, fmt.Errorf("Unknown type: %s", reflect.TypeOf(node).String())
	}
}

func (a *Analysis) function(node *ast.Function) *ast.Function {
	newFunction := &ast.Function{
		Name: node.Name,
		Type: node.Type,
		Body: ast.Block{
			Expressions: make([]ast.Expression, len(node.Body.Expressions)),
		},
	}

	for i, e := range node.Body.Expressions {
		newFunction.Body.Expressions[i] = a.expression(e)
	}

	return newFunction
}

func (a *Analysis) expression(node ast.Expression) ast.Expression {
	switch node := (node).(type) {
	case ast.Assignment:
		return a.assigment(&node)
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

func (a *Analysis) assigment(node *ast.Assignment) ast.Expression {
	newAssign := ast.Assignment{
		Name:       node.Name,
		Expression: a.expression(node.Expression),
	}

	// Get type of assigment expression
	nodeType, err := a.typ(newAssign.Expression)
	if err != nil {
		panic(err)
	}

	// Infer assigment type
	if node.Type == nil {
		newAssign.Type = ast.Basic{
			Type: nodeType,
		}
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
	typ, _ := a.typ(*node)

	// If left part of the node doesnt match the type of the node cast it
	if leftTyp, _ := a.typ(node.Lhs); leftTyp != typ {
		node.Lhs = ast.Cast{
			Expression: node.Lhs,
			Type: ast.Basic{
				Type: typ,
			},
		}
	}

	// If the right part of the node doesnt match the type of the node cast it
	if rightTyp, _ := a.typ(node.Rhs); rightTyp != typ {
		node.Rhs = ast.Cast{
			Expression: node.Rhs,
			Type: ast.Basic{
				Type: typ,
			},
		}
	}

	return *node
}
