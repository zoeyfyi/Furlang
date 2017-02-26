package analysis

import (
	"reflect"

	"fmt"
	"log"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
	"github.com/k0kubun/pp"
)

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

	for i, f := range a.root.Functions {
		a.root.Functions[i] = a.functionDcl(f).(*ast.FunctionDeclaration)
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
		if lType == floatType || rType == floatType {
			return floatType
		}
		return intType

	case *ast.CallExpression:
		switch nodeType := a.typ(node.Function).(type) {
		case *types.Function:
			return nodeType.Return()
		case *types.Basic:
			return nodeType
		}

		log.Fatalf("Unexpected function type on call expression")
		return nil

	case *ast.CastExpression:
		return node.Type

	case *ast.IdentExpression:
		ident := node.Value.Value()
		log.Printf("Looking for %q in scope", ident)

		switch ident {
		case "i8":
			return types.IntType(8)
		case "i16":
			return types.IntType(16)
		case "i32":
			return types.IntType(32)
		case "i64":
			return types.IntType(64)
		}

		return a.typ(a.currentBlock.Scope.Lookup(ident))

	case *ast.IndexExpression:
		return a.typ(node.Expression)

	case *ast.BraceLiteralExpression:
		return node.Type

	case *ast.VaribleDeclaration:
		if node.Type != nil {
			return node.Type
		}
		return a.typ(node.Value)

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

// block runs analysis on each statement in the block
func (a *Analysis) blockSmt(node *ast.BlockStatement) ast.Statement {
	newBlockSmt := &ast.BlockStatement{}

	a.currentBlock = node

	newBlockSmt.Statements = make([]ast.Statement, len(node.Statements))
	for i := range node.Statements {
		newBlockSmt.Statements[i] = a.statement(node.Statements[i])
	}

	return newBlockSmt
}

// function runs analysis on the body of the function
func (a *Analysis) functionDcl(node *ast.FunctionDeclaration) ast.Declare {
	newFunctionDcl := &ast.FunctionDeclaration{}

	a.currentFunction = node

	newFunctionDcl.Name = node.Name
	newFunctionDcl.Arguments = node.Arguments
	newFunctionDcl.Body = a.blockSmt(node.Body).(*ast.BlockStatement)
	newFunctionDcl.Return = node.Return

	return newFunctionDcl
}

// varible runs analysis on the type and value of a varible declaration
func (a *Analysis) varibleDcl(node *ast.VaribleDeclaration) ast.Declare {
	newVaribleDcl := &ast.VaribleDeclaration{}

	newVaribleDcl.Name = node.Name

	log.Println("var dcl: " + node.Name.Value.Value())
	log.Println(a.typ(node.Value))

	if node.Type == nil {
		newVaribleDcl.Type = a.typ(node.Value)
		newVaribleDcl.Value = a.expression(node.Value)
	} else {
		newVaribleDcl.Type = node.Type
		newVaribleDcl.Value = &ast.CastExpression{
			Type:       node.Type,
			Expression: a.expression(node.Value),
		}
	}

	if a.currentBlock != nil {
		a.currentBlock.Scope.Replace(node.Name.Value.Value(), newVaribleDcl)
	}

	return newVaribleDcl
}

func (a *Analysis) declare(node ast.Declare) ast.Declare {
	switch node := (node).(type) {
	case *ast.VaribleDeclaration:
		return a.varibleDcl(node)
	case *ast.FunctionDeclaration:
		return a.functionDcl(node)
	default:
		log.Printf("Unhandled %q node\n", reflect.TypeOf(node).String())
	}

	return node
}

func (a *Analysis) statement(node ast.Statement) ast.Statement {
	switch node := (node).(type) {
	case *ast.AssignmentStatement:
		return a.assigmentSmt(node)
	case *ast.ForStatement:
		return a.forSmt(node)
	case *ast.IfStatment:
		return a.ifSmt(node)
	case *ast.BlockStatement:
		return a.blockSmt(node)
	case *ast.ReturnStatement:
		return a.returnSmt(node)
	case *ast.DeclareStatement:
		return &ast.DeclareStatement{
			Statement: a.declare(node.Statement),
		}
	default:
		log.Printf("Unhandled %q node\n", reflect.TypeOf(node).String())
	}

	return node
}

func (a *Analysis) expression(node ast.Expression) ast.Expression {
	switch node := (node).(type) {
	case *ast.BinaryExpression:
		return a.binaryExp(node)
	case *ast.CallExpression:
		return a.callExp(node)
	case *ast.BraceLiteralExpression:
		return a.braceLiteralExp(node)
	default:
		log.Printf("Unhandled %q node\n", reflect.TypeOf(node).String())
	}

	return node
}

func (a *Analysis) braceLiteralExp(node *ast.BraceLiteralExpression) ast.Expression {
	newBraceLiteralExp := &ast.BraceLiteralExpression{}

	newBraceLiteralExp.Type = node.Type

	newBraceLiteralExp.Elements = make([]ast.Expression, len(node.Elements))
	for i, elm := range node.Elements {
		exp := a.expression(elm)
		// TODO: add non-reflect equal check
		if !reflect.DeepEqual(a.typ(exp), node.Type.Base()) {
			exp = &ast.CastExpression{
				Type:       node.Type.Base(),
				Expression: exp,
			}
		}
		newBraceLiteralExp.Elements[i] = exp
	}

	return newBraceLiteralExp
}

func (a *Analysis) returnSmt(node *ast.ReturnStatement) ast.Statement {
	newReturnSmt := &ast.ReturnStatement{}

	newReturnSmt.Result = a.expression(node.Result)
	pp.Print(newReturnSmt.Result)

	resultType := a.typ(newReturnSmt.Result)
	if resultType != a.currentFunction.Return {
		log.Printf("Casting return statment\n")
		newReturnSmt.Result = &ast.CastExpression{
			Expression: newReturnSmt.Result,
			Type:       a.currentFunction.Return,
		}
	}

	return newReturnSmt
}

func (a *Analysis) forSmt(node *ast.ForStatement) ast.Statement {
	log.Println("For")

	newForSmt := &ast.ForStatement{}

	newForSmt.Index = a.statement(node.Index)
	newForSmt.Condition = a.expression(node.Condition)
	newForSmt.Increment = a.statement(node.Increment)
	newForSmt.Body = a.blockSmt(node.Body).(*ast.BlockStatement)

	return newForSmt
}

func (a *Analysis) ifSmt(node *ast.IfStatment) ast.Statement {
	log.Println("If")

	newIfSmt := &ast.IfStatment{}

	if node.Condition != nil {
		newIfSmt.Condition = a.expression(node.Condition)
	}

	newIfSmt.Body = a.blockSmt(node.Body).(*ast.BlockStatement)

	if node.Else != nil {
		newIfSmt.Else = a.ifSmt(node.Else).(*ast.IfStatment)
	}

	return newIfSmt
}

func (a *Analysis) assigmentSmt(node *ast.AssignmentStatement) ast.Statement {
	newAssigmentSmt := &ast.AssignmentStatement{}

	newAssigmentSmt.Left = node.Left
	newAssigmentSmt.Right = node.Right

	// Get type of assigment expression
	leftType := a.typ(node.Left)
	rightType := a.typ(node.Right)

	// Expression doesnt match assigment type
	// TODO: do we need llvm types of can we check base types
	if leftType.Llvm() != rightType.Llvm() {
		// Cast it
		newAssigmentSmt.Right = ast.Expression(&ast.CastExpression{
			Expression: node.Right,
			Type:       leftType,
		})
	}

	return newAssigmentSmt
}

func (a *Analysis) callExp(node *ast.CallExpression) ast.Expression {

	switch nodeType := a.typ(node.Function).(type) {
	// Regular function call
	case *types.Function:
		newCallExp := &ast.CallExpression{}
		newCallExp.Function = node.Function

		// Cast arguments
		newCallExp.Arguments = &ast.ParenLiteralExpression{
			Elements: make([]ast.Expression, len(node.Arguments.Elements)),
		}
		for i, arg := range node.Arguments.Elements {
			newArg := a.expression(arg)
			newCallExp.Arguments.Elements[i] = newArg

			// Auto cast arguments
			defType := nodeType.Arguments()[i] // definition type
			typ := a.typ(newArg)               // actual type
			if typ.Llvm() != defType.Llvm() {
				log.Printf("Casting argument %d", i)
				newCallExp.Arguments.Elements[i] = &ast.CastExpression{
					Expression: newArg,
					Type:       defType,
				}
			}
		}

		return newCallExp

	// Value cast
	case *types.Basic:
		newCastExp := &ast.CastExpression{}

		// TODO: check for multiple arguments
		newCastExp.Expression = a.expression(node.Arguments.Elements[0])
		newCastExp.Type = nodeType

		return newCastExp
	default:
		log.Fatalf("Call node node type had invalid type %q",
			reflect.TypeOf(a.typ(node.Function)).String())
	}

	return node // Unreachable
}

func (a *Analysis) binaryExp(node *ast.BinaryExpression) ast.Expression {
	log.Printf("Binary %s node", node.Operator.String())

	newBinaryExp := &ast.BinaryExpression{
		Left:     a.expression(node.Left),
		Operator: node.Operator,
		Right:    a.expression(node.Right),
	}

	// Gets the overall type of node
	typ := a.typ(newBinaryExp)
	newBinaryExp.IsFp = typ.(*types.Basic).Info()&types.IsFloat != 0

	// If left part of the node doesnt match the type of the node cast it
	if leftTyp := a.typ(newBinaryExp.Left); leftTyp != typ {
		newBinaryExp.Left = &ast.CastExpression{
			Expression: newBinaryExp.Left,
			Type:       typ,
		}
	}

	// If the right part of the node doesnt match the type of the node cast it
	if rightTyp := a.typ(newBinaryExp.Right); rightTyp != typ {
		newBinaryExp.Right = &ast.CastExpression{
			Expression: newBinaryExp.Right,
			Type:       typ,
		}
	}

	return newBinaryExp
}
