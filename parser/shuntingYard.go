package parser

import (
	"fmt"
	"strconv"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/davecgh/go-spew/spew"
	"github.com/oleiade/lane"
)

func checkOperatorStack(operatorStack *lane.Stack, op lexer.Token) bool {
	// Check is stack is empty
	if operatorStack.Empty() {
		return false
	}

	// Check if stack is an operator
	head := operatorStack.Head().(lexer.Token)
	if head.IsOperator() {
		return false
	}

	// Check operator precendence
	return op.Precedence() <= head.Precedence()
}

func popOperatorStack(operatorStack, arityStack, outputStack *lane.Stack) {
	operator := operatorStack.Pop()
	token := operator.(lexer.Token)

	spew.Dump(outputStack.Head())
	rhs := outputStack.Pop().(ast.Expression)
	lhs := outputStack.Pop().(ast.Expression)

	switch token.Type() {
	case lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO, lexer.REM,
		lexer.LSS, lexer.GTR, lexer.LEQ, lexer.GEQ, lexer.EQL, lexer.NEQ:
		outputStack.Push(ast.Binary{lhs, token, rhs})
	default:
		panic(fmt.Sprintf("Unkown operator: %s", token.Type().String()))
	}
}

// Uses shunting yard algoritum to convert
func (p *Parser) shuntingYard(tokens []lexer.Token) ast.Expression {
	outputStack := lane.NewStack()
	operatorStack := lane.NewStack()
	arityStack := lane.NewStack()

	for i, t := range tokens {
		switch t.Type() {
		// case lexer.TRUE:
		// 	outputStack.Push(boolean{true})

		// case lexer.FALSE:
		// 	outputStack.Push(boolean{false})

		case lexer.INT:
			value, err := strconv.ParseInt(t.Value(), 10, 64)
			if err != nil {
				panic(err)
			}
			outputStack.Push(ast.Integer{value})

		case lexer.FLOAT:
			value, err := strconv.ParseFloat(t.Value(), 64)
			if err != nil {
				panic(err)
			}
			outputStack.Push(ast.Float{value})

		case lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO, lexer.REM,
			lexer.GTR, lexer.LSS, lexer.GEQ, lexer.LEQ, lexer.EQL, lexer.NEQ:
			for checkOperatorStack(operatorStack, t) {
				popOperatorStack(operatorStack, arityStack, outputStack)
			}
			operatorStack.Push(t)

		case lexer.IDENT:
			switch {
			// Token is a function name, push it onto the operator stack
			case i < len(tokens)-1 && tokens[i+1].Type() == lexer.LPAREN:
				operatorStack.Push(t)
				if tokens[i+2].Type() == lexer.RPAREN {
					// 0 if function dosnt have any arguments
					arityStack.Push(0)
				} else {
					// 1 if their is atleast 1 argument
					arityStack.Push(1)
				}

			// Token is a array index
			case i < len(tokens)-1 && tokens[i+1].Type() == lexer.LBRACK:
				operatorStack.Push(t)
				arityStack.Push(1)

			// Token is a varible name, push it onto the out queue
			default:
				outputStack.Push(ast.Ident{t.Value()})
			}

		case lexer.LPAREN:
			operatorStack.Push(t)

		case lexer.RPAREN:
			for operatorStack.Head().(lexer.Token).Type() != lexer.LPAREN {
				fmt.Println(operatorStack.Head())
				popOperatorStack(operatorStack, arityStack, outputStack)
				fmt.Println("op stack poped")
			}
			fmt.Println(operatorStack.Head())
			operatorStack.Pop() // pop open bracket

			// Check for function
			if !operatorStack.Empty() &&
				operatorStack.Head().(lexer.Token).Type() == lexer.IDENT {

				// Pop function name
				token := operatorStack.Pop().(lexer.Token)
				argCount := arityStack.Pop().(int)
				exp := ast.Call{Function: ast.Ident{token.Value()}}

				for i := 0; i < argCount; i++ {
					exp.Arguments = append(exp.Arguments, outputStack.Pop().(ast.Expression))
				}

				fmt.Printf("%+v\n", operatorStack.Head())

				outputStack.Push(exp)
			}

		case lexer.LBRACK:
			operatorStack.Push(t)

		case lexer.RBRACK:
			operatorStack.Pop() // pop open sqaure bracket

			// Pop array name
			token := operatorStack.Pop().(lexer.Token)
			exp := ast.ArrayValue{
				Array: ast.Ident{token.Value()},
				Index: ast.Index{outputStack.Pop().(ast.Expression)},
			}

			outputStack.Push(exp)

		case lexer.COMMA:
			// Increment argument count
			as := arityStack.Pop().(int)
			arityStack.Push(as + 1)

			for operatorStack.Head().(lexer.Token).Type() != lexer.LPAREN {
				popOperatorStack(operatorStack, arityStack, outputStack)
			}

			if operatorStack.Empty() {
				panic("Misplaced comma or mismatched parentheses")
			}

		default:
			panic("Unexpected math token: " + t.String())
		}
	}

	for !operatorStack.Empty() {
		popOperatorStack(operatorStack, arityStack, outputStack)
	}

	return outputStack.Pop().(ast.Expression)
}
