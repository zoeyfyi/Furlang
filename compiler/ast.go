package compiler

import (
	"fmt"
	"reflect"

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

type boolean struct {
	value bool
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

type ifExpression struct {
	blocks []ifBlock
}

type ifBlock struct {
	condition expression
	block     block
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

// Print parse stack at every token
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
	// To do this we need to modify the shunting yard algorithum to track
	// the depth of brackets, if the line ends on the wrong depth their is
	// a parentheses mismatch, if we reach a closing bracket of the correct
	// depth we know we have reached the end
	functionArgMap = map[string]int{
		"add":  2,
		"main": 0,
		"half": 0,
	}
)

func printDebugStack(parseStack *lane.Stack) {
	if debugStack {
		tempStack := lane.NewStack()
		psSize := parseStack.Size()

		for !parseStack.Empty() {
			item := parseStack.Pop()
			fmt.Printf("\n\t%d. %s - %+v", psSize, reflect.TypeOf(item).String(), item)
			tempStack.Push(item)
			psSize--
		}

		for !tempStack.Empty() {
			parseStack.Push(tempStack.Pop())
		}
	}
}

func ast(tokens []token) (*abstractSyntaxTree, error) {
	ast := abstractSyntaxTree{}

	// Current branch of node i.e. function -> block -> assignment -> maths
	parseStack := lane.NewStack()

	// Stack and queue for coversion to infix
	outQueue := lane.NewQueue()
	stack := lane.NewStack()

	// Are we parsing arguments or returns
	arrow := false

	defer printDebugStack(parseStack)

	for i, t := range tokens {
		if debugStack {
			mag := color.New(color.FgMagenta).SprintfFunc()
			fmt.Printf("\n%d)\t- %s", i, mag(t.string()))
		}

		// Check for top-level function definitions
		if parseStack.Empty() {
			switch t.tokenType {
			case tokenNewLine:

			case tokenName:
				// Push new function on to stack
				parseStack.Push(&function{
					name: t.value.(string),
				})

			default:
				return nil, Error{
					err:        "Only names expected in top level",
					tokenRange: []token{t},
				}
			}

			printDebugStack(parseStack)
			continue
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
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				}

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
			switch parseStack.Head().(type) {
			case *function:
				arrow = true

			case *typedName:
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				}
				arrow = true

			default:
				return nil, Error{
					err:        "Unexpected arrow",
					tokenRange: []token{t},
				}
			}

		case tokenOpenBody:
			switch parseStack.Head().(type) {
			case *typedName, *function:
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				}

				// Start new block
				parseStack.Push(&block{})

			// Start of if block
			case *maths:
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				}
				parseStack.Push(&block{})

			// Start of else block
			case *ifBlock:
				parseStack.Push(&block{})

			// Start of a block
			case *block:
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
				return nil, Error{
					err:        "Unexpected assignment",
					tokenRange: []token{t},
				}
			}

		case tokenIf:
			switch parseStack.Head().(type) {
			case *block:
				parseStack.Push(&ifExpression{})
				parseStack.Push(&ifBlock{})
				parseStack.Push(&maths{})
			default:
				return nil, Error{
					err:        "Unexpected if statement",
					tokenRange: []token{t},
				}
			}

		case tokenElse:
			switch parseStack.Head().(type) {
			case *ifExpression:
				parseStack.Push(&ifBlock{})
			default:
				fmt.Println(parseStack.Head())
				return nil, Error{
					err:        "Unexpected else statement",
					tokenRange: []token{t},
				}
			}

		case tokenTrue:
			switch parseStack.Head().(type) {
			case *maths:
				outQueue.Enqueue(t)
			default:
				return nil, Error{
					err:        "Unexpected true",
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
					if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
						return nil, err
					}
					_, ok = parseStack.Head().(*block)
				}

			case *ifExpression:
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				} // Pop if
			}

		case tokenCloseBody:
			if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
				return nil, err
			} // Pop block
			if _, ok := parseStack.Head().(*function); ok {
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				} // Pop function
			} else if _, ok := parseStack.Head().(*ifBlock); ok {
				if err := popOff(parseStack, &ast, arrow, tokens, stack, outQueue); err != nil {
					return nil, err
				} // Pop ifBlock
			}
		}

		printDebugStack(parseStack)
	}

	return &ast, nil
}

// Pops the child of the parseStack and adds it to its parent node
func popOff(parseStack *lane.Stack, ast *abstractSyntaxTree,
	arrow bool,
	tokens []token,
	stack *lane.Stack,
	outQueue *lane.Queue) error {

	if parseStack.Empty() {
		panic("Empty stack")
	}

	child := parseStack.Pop()

	// If the child is a function add function to ast
	if function, ok := child.(*function); ok {
		arrow = false
		ast.functions = append(ast.functions, *function)
		return nil
	}

	// Add child to parent
	switch parent := parseStack.Head().(type) {
	case *assignment:
		parent.value = *child.(*maths)
	case *ret:
		parent.returns = append(parent.returns, expression(*child.(*maths)))
	case *ifExpression:
		parent.blocks = append(parent.blocks, *child.(*ifBlock))
	case *ifBlock:
		switch child := child.(type) {
		case *block:
			parent.block = *child
		case *maths:
			exp, err := createExpression(tokens, stack, outQueue)
			if err != nil {
				return err
			}
			parent.condition = exp
		}
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
		case *ifExpression:
			parent.expressions = append(parent.expressions, expression(*child))
		case *block:
			parent.expressions = append(parent.expressions, expression(*child))
		default:
			return Error{
				err:        "Child is of unkown type",
				tokenRange: nil,
			}
		}
	case *function:
		switch child := child.(type) {
		// End of function block
		case *block:
			parent.block = *child
		// End of argument or return definition
		case *typedName:
			if arrow {
				parent.returns = append(parent.returns, *child)
			} else {
				parent.args = append(parent.args, *child)
			}
		}
	default:
		panic("\nUnhandled parent")
	}

	return nil
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
		case tokenTrue:
			resolve.Push(boolean{true})
		case tokenFalse:
			resolve.Push(boolean{true})
		default:
			return nil, Error{
				err:        "Unexpected token",
				tokenRange: []token{t},
			}
		}
	}

	return resolve.Head().(expression), nil
}
