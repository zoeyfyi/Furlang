package compiler

import (
	"fmt"

	"github.com/bongo227/Furlang/lexer"
	"github.com/davecgh/go-spew/spew"
	"github.com/oleiade/lane"
)

func checkOperatorStack(operatorStack *lane.Stack, op operator) bool {
	// Check is stack is empty
	if operatorStack.Empty() {
		return false
	}

	// Check if stack is an operator
	headType := operatorStack.Head().(lexer.Token).Type
	if _, isOp := opMap[headType]; !isOp {
		return false
	}

	// Check operator precendence
	headPrecendence := opMap[headType].precendence
	return (!op.right && op.precendence <= headPrecendence) ||
		(op.right && op.precendence < headPrecendence)
}

func popOperatorStack(operatorStack, arityStack, outputStack *lane.Stack) {
	operator := operatorStack.Pop()
	token := operator.(lexer.Token)

	spew.Dump(outputStack.Head())
	rhs := outputStack.Pop().(expression)
	lhs := outputStack.Pop().(expression)

	switch token.Type {
	case lexer.PLUS:
		outputStack.Push(addition{lhs, rhs})
	case lexer.MINUS:
		outputStack.Push(subtraction{lhs, rhs})
	case lexer.MULTIPLY:
		outputStack.Push(multiplication{lhs, rhs})
	case lexer.FLOATDIVIDE:
		outputStack.Push(floatDivision{lhs, rhs})
	case lexer.INTDIVIDE:
		outputStack.Push(intDivision{lhs, rhs})
	case lexer.LESSTHAN:
		outputStack.Push(lessThan{lhs, rhs})
	case lexer.MORETHAN:
		outputStack.Push(moreThan{lhs, rhs})
	case lexer.EQUAL:
		outputStack.Push(equal{lhs, rhs})
	case lexer.NOTEQUAL:
		outputStack.Push(notEqual{lhs, rhs})
	case lexer.MOD:
		outputStack.Push(mod{lhs, rhs})
	}
}

// Uses shunting yard algoritum to convert
func (p *parser) shuntingYard(tokens []lexer.Token) expression {
	p.log("Start ShuntingYard", true)
	defer p.log("End ShuntingYard", false)

	outputStack := lane.NewStack()
	operatorStack := lane.NewStack()
	arityStack := lane.NewStack()

	for i, t := range tokens {
		switch t.Type {

		case lexer.TRUE:
			outputStack.Push(boolean{true})

		case lexer.FALSE:
			outputStack.Push(boolean{false})

		case lexer.INTVALUE:
			outputStack.Push(number{t.Value.(int)})

		case lexer.FLOATVALUE:
			outputStack.Push(float{t.Value.(float32)})

		case lexer.PLUS, lexer.MINUS, lexer.MULTIPLY, lexer.FLOATDIVIDE, lexer.INTDIVIDE,
			lexer.MORETHAN, lexer.LESSTHAN, lexer.EQUAL, lexer.NOTEQUAL, lexer.MOD:
			for checkOperatorStack(operatorStack, opMap[t.Type]) {
				popOperatorStack(operatorStack, arityStack, outputStack)
			}
			operatorStack.Push(t)

		case lexer.IDENT:
			if i < len(tokens)-1 && tokens[i+1].Type == lexer.OPENBRACKET {
				// Token is a function name, push it onto the operator stack
				operatorStack.Push(t)
				if tokens[i+2].Type == lexer.CLOSEBRACKET {
					// 0 if function dosnt have any arguments
					arityStack.Push(0)
				} else {
					// 1 if their is atleast 1 argument
					arityStack.Push(1)
				}
			} else if i < len(tokens)-1 && tokens[i+1].Type == lexer.OPENSQUAREBRACKET {
				// Token is a array index
				operatorStack.Push(t)
				arityStack.Push(1)
			} else {
				// Token is a varible name, push it onto the out queue
				outputStack.Push(name{t.Value.(string)})
			}

		case lexer.OPENBRACKET:
			operatorStack.Push(t)

		case lexer.OPENSQUAREBRACKET:
			operatorStack.Push(t)

		case lexer.CLOSEBRACKET:
			for operatorStack.Head().(lexer.Token).Type != lexer.OPENBRACKET {
				popOperatorStack(operatorStack, arityStack, outputStack)
			}

			operatorStack.Pop() // pop open bracket

			// Check for function
			if operatorStack.Head().(lexer.Token).Type == lexer.IDENT {
				// Pop function name
				token := operatorStack.Pop().(lexer.Token)
				argCount := arityStack.Pop().(int)
				exp := call{
					function: token.Value.(string),
				}

				for i := 0; i < argCount; i++ {
					exp.args = append(exp.args, outputStack.Pop().(expression))
				}

				fmt.Printf("%+v\n", operatorStack.Head())

				outputStack.Push(exp)
			}

		case lexer.CLOSESQUAREBRACKET:
			operatorStack.Pop() // pop open sqaure bracket

			// Pop array name
			token := operatorStack.Pop().(lexer.Token)
			argCount := arityStack.Pop().(int)
			exp := arrayValue{
				name: token.Value.(string),
			}

			for i := 0; i < argCount; i++ {
				exp.index = outputStack.Pop().(expression)
				fmt.Println(exp.index)
			}

			outputStack.Push(exp)

		case lexer.COMMA:
			// Increment argument count
			as := arityStack.Pop().(int)
			arityStack.Push(as + 1)

			for operatorStack.Head().(lexer.Token).Type != lexer.OPENBRACKET {
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

	return outputStack.Pop().(expression)
}
