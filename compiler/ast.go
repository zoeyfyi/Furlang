package compiler

import lane "gopkg.in/oleiade/lane.v1"

const (
	typeInt32 = iota + 100
	typeFloat32
)

type typedName struct {
	nameType int
	name     string
}

// TODO: do we need names in here?
type function struct {
	name     string
	position block
	names    []typedName
	args     []typedName
	returns  []typedName
	lines    []expression
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

type block struct {
	start int
	end   int
}

type functionBlock struct {
	position      block
	name          string
	argumentCount int
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

var (
	opMap = map[int]operator{
		tokenPlus:        operator{2, false},
		tokenMinus:       operator{2, false},
		tokenMultiply:    operator{3, false},
		tokenFloatDivide: operator{3, false},
		tokenIntDivide:   operator{3, false},
	}

	StateError = Error{
		err:        "Unexpected state",
		tokenRange: nil,
	}
)

func ast(tokens []token) (*abstractSyntaxTree, error) {
	ast := abstractSyntaxTree{}
	functionArgMap := make(map[string]int, 1000) //TODO: compute this
	{
		state := stateFunctionNone
		var currentFunction *function
		var currentType *typedName

		for i, t := range tokens {
			switch t.tokenType {
			case tokenDoubleColon:
				state = stateFunctionArgs

			case tokenInt32:
				currentType = &typedName{}
				currentType.nameType = typeInt32

			case tokenFloat32:
				currentType = &typedName{}
				currentType.nameType = typeFloat32

			case tokenName:
				switch state {
				case stateFunctionNone:
					currentFunction = &function{}
					currentFunction.name = t.value.(string)

				case stateFunctionArgs:
					// TODO: handle if nil
					if currentType != nil {
						currentType.name = t.value.(string)
					}
				case stateFunctionReturns:
					if currentType != nil {
						currentFunction.name = t.value.(string)
					}
				}

			case tokenComma:
				// TODO: handle stateFunctionNone, stateFunctionLines
				switch state {
				case stateFunctionArgs:
					if currentType != nil {
						currentFunction.args = append(currentFunction.args, *currentType)
					}
				case stateFunctionReturns:
					if currentType != nil {
						currentFunction.returns = append(currentFunction.returns, *currentType)
					}
				}

				// Zero out currentType
				currentType.name = ""
				currentType.nameType = 0

			case tokenArrow:
				// TODO: handle other states
				switch state {
				case stateFunctionArgs:
					if currentType != nil {
						currentFunction.args = append(currentFunction.args, *currentType)
					}

					currentType = nil
					state = stateFunctionReturns
				}

			case tokenOpenBody:
				// TODO: handle other states
				switch state {
				case stateFunctionReturns:
					if currentType != nil {
						currentFunction.returns = append(currentFunction.returns, *currentType)
					}

					state = stateFunctionLines
					currentFunction.position.start = i + 1

					functionArgMap[currentFunction.name] = len(currentFunction.args)
				}

			case tokenCloseBody:
				// TODO: handle other state
				switch state {
				case stateFunctionLines:
					currentFunction.position.end = i
					ast.functions = append(ast.functions, *currentFunction)
					state = stateFunctionNone
				}

			}
		}

		//  Remove Empty function
		if currentFunction == nil {
			ast.functions = ast.functions[:len(ast.functions)-1]
		}

	}

	{
		createExpression := func(stack *lane.Stack, outQueue *lane.Queue) (expression, error) {
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
							args[len(args)] = exp
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

		for i := 0; i < len(ast.functions); i++ {
			parseFunction := &ast.functions[i]

			state := stateLineNone
			var currectExpression expression
			// TODO: make custom stack and queue
			outQueue := lane.NewQueue()
			stack := lane.NewStack()

			// TODO: handle other cases
			innerTokens := tokens[parseFunction.position.start:parseFunction.position.end]
		tokenLoop:
			for i, t := range innerTokens {
				switch t.tokenType {
				// Return expression
				case tokenReturn:
					switch state {
					case stateLineNone:
						state = stateLineReturn
						currectExpression = ret{}
					}

				// Assignment, Function call or varible
				case tokenName:
					switch state {
					case stateLineNone:
						switch innerTokens[i+1].tokenType {
						// Assignment expression
						case tokenAssign:
							state = stateLineAssignment
							newAssignment := assignment{}
							newAssignment.name = t.value.(string)
							currectExpression = newAssignment

						// Function call expression
						case tokenOpenBracket:
							state = stateLineFunctionCall
						}

					case stateLineAssignment, stateLineFunctionCall, stateLineReturn:
						if _, found := functionArgMap[t.value.(string)]; found {
							// Token is a function name
							stack.Push(t)
						} else {
							// Token is a varible
							outQueue.Enqueue(t)
						}
					default:
						return nil, StateError
					}

				// Assignment
				case tokenAssign:
					if _, ok := currectExpression.(assignment); !ok {
						return nil, Error{
							err:        "Misplaced assignment",
							tokenRange: []token{t},
						}
					}

				// Newline
				case tokenNewLine:
					switch state {
					case stateLineNone:
						// Ignore blank lines

					case stateLineAssignment:
						exp, err := createExpression(stack, outQueue)
						if err != nil {
							return nil, err
						}

						assignmentExpression := currectExpression.(assignment)
						assignmentExpression.value = exp
						parseFunction.lines = append(parseFunction.lines, assignmentExpression)

					case stateLineFunctionCall:
						exp, err := createExpression(stack, outQueue)
						if err != nil {
							return nil, err
						}

						parseFunction.lines = append(parseFunction.lines, exp)

					case stateLineReturn:
						exp, err := createExpression(stack, outQueue)
						if err != nil {
							return nil, err
						}

						retExpression := currectExpression.(ret)
						retExpression.returns = append(retExpression.returns, exp)
						parseFunction.lines = append(parseFunction.lines, retExpression)

					default:
						return nil, StateError
					}
					state = stateLineNone
					continue tokenLoop

				// Number token
				case tokenNumber, tokenFloat:
					switch state {
					case stateLineNone:
						return nil, Error{
							err:        "Unexpected number",
							tokenRange: []token{t},
						}
					case stateLineAssignment, stateLineFunctionCall, stateLineReturn:
						outQueue.Enqueue(t)
					default:
						return nil, StateError
					}

				// Comma token
				case tokenComma:
					switch state {
					case stateLineNone:
						return nil, Error{
							err:        "Unexpected comma",
							tokenRange: []token{t},
						}
					case stateLineReturn:
						for stack.Head().(token).tokenType != tokenOpenBracket {
							outQueue.Enqueue(stack.Pop())
						}

						if stack.Empty() && len(parseFunction.returns) < len(currectExpression.(ret).returns) {
							// Comma is seperating multiple return values
							retExpression := currectExpression.(ret)
							exp, err := createExpression(stack, outQueue)
							if err != nil {
								return nil, err
							}
							retExpression.returns = append(retExpression.returns, exp)
						} else if stack.Empty() {
							// Comma is out of place
							return nil, Error{
								err:        "Misplaced comma or mismatched parentheses",
								tokenRange: []token{t},
							}
						}

					case stateLineAssignment, stateLineFunctionCall:
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
						return nil, StateError
					}

				// Mathmatical operator
				case tokenPlus, tokenMinus, tokenMultiply, tokenIntDivide, tokenFloatDivide:
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

				// Open bracket
				case tokenOpenBracket:
					stack.Push(t)

				// Close bracket
				case tokenCloseBracket:
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

				// Default case
				default:
					return nil, Error{
						err:        "Unexpected token",
						tokenRange: []token{t},
					}
				}

			}
		}
	}

	return &ast, nil

}
