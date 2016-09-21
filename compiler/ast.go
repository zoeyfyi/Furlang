package compiler

import (
	"fmt"

	"github.com/fatih/color"

	lane "gopkg.in/oleiade/lane.v1"
)

const (
	typeInt32 = iota + 100
	typeFloat32
)

type typedName struct {
	nameType int
	name     string
}

type function struct {
	name    string
	args    []typedName
	arrow   bool
	returns []typedName
	block   block
}

type block struct {
	expressions []expression
}

type operator struct {
	precendence int
	right       bool
}

type name struct {
	name string
}

type ret struct {
	returns []expression
}

type assignment struct {
	name  string
	value expression
}

type maths struct {
	expression expression
}

type addition struct {
	lhs expression
	rhs expression
}

type subtraction struct {
	lhs expression
	rhs expression
}

type multiplication struct {
	lhs expression
	rhs expression
}

type floatDivision struct {
	lhs expression
	rhs expression
}

type intDivision struct {
	lhs expression
	rhs expression
}

type number struct {
	value int
}

type float struct {
	value float32
}

type call struct {
	function string
	args     []expression
}

type abstractSyntaxTree struct {
	functions []function
}

const (
	stateFunctionNone = iota
	stateFunctionArgs
	stateFunctionReturns
	stateFunctionLines

	stateLineNone
	stateLineReturn
	stateLineAssignment
	stateLineFunctionCall
)

const debugStack = false

var (
	opMap = map[int]operator{
		tokenPlus:        operator{2, false},
		tokenMinus:       operator{2, false},
		tokenMultiply:    operator{3, false},
		tokenFloatDivide: operator{3, false},
		tokenIntDivide:   operator{3, false},
	}

	// StateError occors when the parser enters an invalid state
	StateError = Error{
		err:        "Unexpected state",
		tokenRange: nil,
	}

	//TODO: compute this
	functionArgMap = map[string]int{
		"add":  2,
		"main": 0,
		"half": 0,
	}
)

// TODO: research golang concurency
// TODO: get rid of appends (almost) everywere
// TODO: sort switchs alphabeticly (is their a formatter for this?)
// TODO: fix todos!
// TODO: get rid of arrow in function

func ast(tokens []token) (*abstractSyntaxTree, error) {
	ast := abstractSyntaxTree{}

	// Current branch of node i.e. function -> block -> assignment -> maths
	parseStack := lane.NewStack()

	// Stack and queue for coversion to infix
	outQueue := lane.NewQueue()
	stack := lane.NewStack()

	for i, t := range tokens {
		if debugStack {
			mag := color.New(color.FgMagenta).SprintfFunc()
			fmt.Printf("\n%d)\t- %s", i, mag(t.string()))
		}

		// Check for top-level function definitions
		if parseStack.Empty() {
			switch t.tokenType {
			case tokenNewLine:
				continue
			case tokenName:
				// Push new function on to stack
				parseStack.Push(&function{
					name: t.value.(string),
				})
				continue

			default:
				return nil, Error{
					err:        "Only names expected in top level",
					tokenRange: []token{t},
				}
			}
		}

		switch t.tokenType {
		case tokenDoubleColon:
			if _, ok := parseStack.Head().(*function); !ok {
				return nil, Error{
					err:        "Unexpected double colon",
					tokenRange: []token{t},
				}
			}

		case tokenInt32, tokenFloat32:
			var tokenType int
			switch t.tokenType {
			case tokenInt32:
				tokenType = typeInt32
			case tokenFloat32:
				tokenType = typeFloat32
			}

			switch parseStack.Head().(type) {
			case *function:
				parseStack.Push(&typedName{
					nameType: tokenType,
				})
			default:
				return nil, Error{
					err:        "Unexpected i32",
					tokenRange: []token{t},
				}
			}

		case tokenName:
			switch e := parseStack.Head().(type) {
			// named argument or named return
			case *typedName:
				e.name = t.value.(string)

			// Function call or assignment
			case *block:
				// Look ahead 1 token to determin if this is a function call or assignment
				switch tokens[i+1].tokenType {
				// Name of assignment
				case tokenAssign:
					parseStack.Push(&assignment{
						name: t.value.(string),
					})

				// Function call
				case tokenOpenBracket:
					parseStack.Push(&maths{})
				}

			// Varible or function call
			case *maths:
				if _, found := functionArgMap[t.value.(string)]; found {
					// Token is a function name
					stack.Push(t)
				} else {
					// Token is a Varible
					outQueue.Enqueue(t)
				}

			default:
				return nil, Error{
					err:        "Unexpected name",
					tokenRange: []token{t},
				}
			}

		case tokenNumber, tokenFloat:
			switch parseStack.Head().(type) {
			case *maths:
				outQueue.Enqueue(t)
			default:
				return nil, Error{
					err:        "Unexpected number",
					tokenRange: []token{t},
				}
			}
		case tokenPlus, tokenMinus, tokenMultiply, tokenFloatDivide, tokenIntDivide:
			switch parseStack.Head().(type) {
			case *maths:
				op := opMap[t.tokenType]

				for !stack.Empty() &&
					(stack.Head().(token).tokenType == tokenPlus ||
						stack.Head().(token).tokenType == tokenMinus ||
						stack.Head().(token).tokenType == tokenMultiply ||
						stack.Head().(token).tokenType == tokenFloatDivide ||
						stack.Head().(token).tokenType == tokenIntDivide) &&
					((!op.right && op.precendence <= opMap[stack.Head().(token).tokenType].precendence) ||
						(op.right && op.precendence < opMap[stack.Head().(token).tokenType].precendence)) {

					outQueue.Enqueue(stack.Pop())
				}

				stack.Push(t)
			default:
				return nil, Error{
					err:        "Unexpected operator",
					tokenRange: []token{t},
				}
			}

		case tokenOpenBracket:
			switch parseStack.Head().(type) {
			case *maths:
				stack.Push(t)
			default:
				return nil, Error{
					err:        "Unexpected bracket",
					tokenRange: []token{t},
				}
			}

		case tokenCloseBracket:
			switch parseStack.Head().(type) {
			case *maths:
				if stack.Empty() {
					return nil, Error{
						err:        "Mismatched parentheses",
						tokenRange: []token{t},
					}
				}

				for !stack.Empty() && stack.Head().(token).tokenType != tokenOpenBracket {
					outQueue.Enqueue(stack.Pop())
				}

				if stack.Empty() {
					return nil, Error{
						err:        "Mismatched parentheses",
						tokenRange: []token{t},
					}
				}

				stack.Pop() // pop open bracket off

				if t.tokenType == tokenName {
					if _, found := functionArgMap[t.value.(string)]; found {
						outQueue.Enqueue(stack.Pop())
					}
				}
			default:
				return nil, Error{
					err:        "Unexpected token",
					tokenRange: []token{t},
				}
			}

		case tokenComma:
			// TODO: handle multiple returns
			switch parseStack.Head().(type) {
			case *typedName, *function:
				// Argument/return seperator
				popOff(parseStack, &ast)

			case *maths:
				for stack.Head().(token).tokenType != tokenOpenBracket {
					outQueue.Enqueue(stack.Pop())
				}

				if stack.Empty() {
					// Comma is out of place
					return nil, Error{
						err:        "Misplaced comma or mismatched parentheses",
						tokenRange: []token{t},
					}
				}
			default:
				return nil, Error{
					err:        "Unexpected comma",
					tokenRange: []token{t},
				}
			}

		case tokenArrow:
			switch e := parseStack.Head().(type) {
			case *function:
				e.arrow = true

			case *typedName:
				popOff(parseStack, &ast)
				function := parseStack.Head().(*function)
				function.arrow = true

			default:
				return nil, Error{
					err:        "Unexpected arrow",
					tokenRange: []token{t},
				}
			}

		case tokenOpenBody:
			switch parseStack.Head().(type) {
			case *typedName, *function:
				popOff(parseStack, &ast)

				// Start new block
				parseStack.Push(&block{})

			default:
				return nil, Error{
					err:        "Unexpected open body",
					tokenRange: []token{t},
				}
			}

		case tokenAssign:
			switch parseStack.Head().(type) {
			case *assignment:
				parseStack.Push(&maths{})

			default:
				fmt.Printf("%+v\n", *parseStack.Head().(*expression))
				return nil, Error{
					err:        "Unexpected assignment",
					tokenRange: []token{t},
				}
			}

		case tokenReturn:
			switch parseStack.Head().(type) {
			case *block:
				// TODO: support multiple returns
				// Prehaps push arg length worth of maths on to stack then when you pop them
				// of, add the expressions in reverse order. Does that work?
				parseStack.Push(&ret{})
				parseStack.Push(&maths{})

			default:
				return nil, Error{
					err:        "Unexpected return statement",
					tokenRange: []token{t},
				}
			}

		case tokenNewLine:
			switch e := parseStack.Head().(type) {
			case *maths:
				// Set maths expression
				mathExpression, err := createExpression(tokens, stack, outQueue)
				if err != nil {
					return nil, err
				}
				e.expression = mathExpression

				// Pop the stack until we reach the block expression
				_, ok := parseStack.Head().(*block)
				for !ok {
					popOff(parseStack, &ast)
					_, ok = parseStack.Head().(*block)
				}
			}

		case tokenCloseBody:
			popOff(parseStack, &ast) // Pop block
			popOff(parseStack, &ast) // Pop function
		}

		if debugStack {
			tempStack := lane.NewStack()
			psSize := parseStack.Size()

			for !parseStack.Empty() {
				item := parseStack.Pop()
				fmt.Printf("\n\t%d. %+v", psSize, item)
				tempStack.Push(item)
				psSize--
			}

			for !tempStack.Empty() {
				parseStack.Push(tempStack.Pop())
			}
		}

	}

	return &ast, nil
}

// Pops the child of the parseStack and adds it to its parent node
func popOff(parseStack *lane.Stack, ast *abstractSyntaxTree) {
	if parseStack.Empty() {
		panic("Empty stack")
	}

	child := parseStack.Pop()

	// If the child is a function add function to ast
	if function, ok := child.(*function); ok {
		ast.functions = append(ast.functions, *function)
		return
	}

	// Add child to parent
	switch parent := parseStack.Head().(type) {
	case *assignment:
		parent.value = *child.(*maths)
	case *ret:
		parent.returns = append(parent.returns, expression(*child.(*maths)))
	case *call:
		parent.args = append(parent.args, expression(*child.(*maths)))
	case *block:
		switch child := child.(type) {
		case *assignment:
			parent.expressions = append(parent.expressions, expression(*child))
		case *ret:
			parent.expressions = append(parent.expressions, expression(*child))
		case *maths:
			parent.expressions = append(parent.expressions, expression(*child))
		}
	case *function:
		switch child := child.(type) {
		// End of function block
		case *block:
			parent.block = *child
		// End of argument or return definition
		case *typedName:
			if parent.arrow {
				parent.returns = append(parent.returns, *child)
			} else {
				parent.args = append(parent.args, *child)
			}
		}
	default:
		panic("\nUnhandled parent")
	}
}

// Resolve infix to expression tree
func createExpression(tokens []token, stack *lane.Stack, outQueue *lane.Queue) (expression, error) {
	// Clear any items on stack
	for !stack.Empty() {
		head := stack.Head().(token)
		if head.tokenType == tokenOpenBracket || head.tokenType == tokenCloseBracket {
			return nil, Error{
				err:        "Mismatched parentheses",
				tokenRange: []token{tokens[0], tokens[len(tokens)-1]},
			}
		}
		outQueue.Enqueue(stack.Pop())
	}

	// Resolve out queue
	resolve := lane.NewStack()

	for !outQueue.Empty() {
		t := outQueue.Dequeue().(token)
		switch t.tokenType {
		case tokenPlus, tokenMinus, tokenMultiply, tokenFloatDivide, tokenIntDivide:
			// Token is a maths operator
			rhs, lhs := resolve.Pop().(expression), resolve.Pop().(expression)
			switch t.tokenType {
			case tokenPlus:
				resolve.Push(addition{lhs, rhs})
			case tokenMinus:
				resolve.Push(subtraction{lhs, rhs})
			case tokenMultiply:
				resolve.Push(multiplication{lhs, rhs})
			case tokenFloatDivide:
				resolve.Push(floatDivision{lhs, rhs})
			case tokenIntDivide:
				resolve.Push(intDivision{lhs, rhs})
			}
		case tokenName:
			if argCount, found := functionArgMap[t.value.(string)]; found {
				// Token is a function call
				args := make([]expression, argCount)
				for i := 0; i < argCount; i++ {
					exp, ok := resolve.Pop().(expression)
					if !ok {
						return nil, Error{
							err:        "Expected function to have arguments",
							tokenRange: []token{t},
						}
					}
					args[i] = exp
				}

				resolve.Push(call{t.value.(string), args})
			} else {
				// Token is a varible
				resolve.Push(name{t.value.(string)})
			}
		case tokenNumber:
			resolve.Push(number{t.value.(int)})
		case tokenFloat:
			resolve.Push(float{t.value.(float32)})
		default:
			return nil, Error{
				err:        "Unexpected token",
				tokenRange: []token{t},
			}
		}
	}

	return resolve.Head().(expression), nil
}
