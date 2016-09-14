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

type function struct {
	name    string
	names   []typedName
	args    []typedName
	returns []typedName
	lines   []expression
}

type block struct {
	start int
	end   int
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

type functionDefinition struct {
	name          string
	start         int
	end           int
	argumentCount int
}

func ast(tokens []token) (functions []function, err error) {
	// Find the function positions and names
	current := functionDefinition{}
	// TODO: allocate correct ammount of functions
	functionDefinitions := make(map[string]functionDefinition, 1000)
	// TODO: compute this based on order of calls
	var functionOrder []string
	arrow := false

	for i, t := range tokens {
		switch t.tokenType {
		case tokenDoubleColon:
			if tokens[i-1].tokenType != tokenName {
				return nil, Error{
					err:        "Expected function to start with name",
					tokenRange: []token{tokens[i-1]},
				}
			}

			current.name = tokens[i-1].value.(string)
			current.start = i - 1
		case tokenInt32:
			if !arrow {
				current.argumentCount++
			}
		case tokenFloat32:
			if !arrow {
				current.argumentCount++
			}
		case tokenArrow:
			arrow = true
		case tokenCloseBody:
			current.end = i
			functionDefinitions[current.name] = current
			functionOrder = append(functionOrder, current.name)
			current = functionDefinition{}
		}
	}

	// Parse functions
	for _, fkey := range functionOrder {
		definition := functionDefinitions[fkey]

		fTokens := tokens[definition.start:definition.end]
		function := function{}

		// Set function name
		function.name = fTokens[0].value.(string)

		// Set function arguments and returns
		arrow := false
		startBody := -1
		currentTypedName := typedName{}
		for i, t := range fTokens[2:] {
			if currentTypedName.nameType != 0 &&
				(t.tokenType == tokenComma || t.tokenType == tokenOpenBody || t.tokenType == tokenArrow) {
				if arrow {
					function.returns = append(function.returns, currentTypedName)
				} else {
					function.args = append(function.args, currentTypedName)
				}
			}

			switch t.tokenType {
			case tokenInt32:
				currentTypedName.nameType = typeInt32
			case tokenFloat32:
				currentTypedName.nameType = typeFloat32
			case tokenArrow:
				arrow = true
				continue
			case tokenOpenBody:
				startBody = i + 2
				break
			case tokenName:
				currentTypedName.name = t.value.(string)
			}
		}

		// Parse each line
		var tokenBuffer []token
		for _, t := range fTokens[startBody+1:] {
			if t.tokenType == tokenNewLine && tokenBuffer != nil {
				// Line has ended pass line
				var lineExpression expression
				switch tokenBuffer[0].tokenType {
				case tokenReturn:
					// Line is a return statement
					retExpression := ret{}

					lastComma := 0
					returnTokens := tokenBuffer[1:]
					bracketDepth := 0
					for i, t := range returnTokens {
						if (t.tokenType == tokenComma && bracketDepth == 0) ||
							i == len(returnTokens)-1 {

							exp, err := infixToTree(returnTokens[lastComma:i+1], functionDefinitions)
							if err != nil {
								return nil, err
							}
							retExpression.returns = append(retExpression.returns, exp)
							lastComma = i
						} else if t.tokenType == tokenOpenBracket {
							bracketDepth++
						} else if t.tokenType == tokenCloseBracket {
							bracketDepth--
						}
					}

					lineExpression = retExpression
				case tokenName:
					if tokenBuffer[1].tokenType == tokenAssign {
						// Line is a assignment
						tree, err := infixToTree(tokenBuffer[2:], functionDefinitions)
						if err != nil {
							return nil, err
						}

						lineExpression = assignment{
							name:  tokenBuffer[0].value.(string),
							value: tree,
						}

					} else {
						// Line is a function call
						tree, err := infixToTree(tokenBuffer, functionDefinitions)
						lineExpression = tree
						if err != nil {
							return nil, err
						}
					}
				}

				function.lines = append(function.lines, lineExpression)

				tokenBuffer = nil
			} else {
				// Append to buffer
				if t.tokenType != tokenNewLine && t.tokenType != tokenCloseBody {
					tokenBuffer = append(tokenBuffer, t)
				}
			}
		}

		// Add function to slice
		functions = append(functions, function)
	}

	return functions, nil
}

func infixToTree(tokens []token, functionDefinitions map[string]functionDefinition) (expression, error) {
	opMap := map[int]operator{
		tokenPlus:        operator{2, false},
		tokenMinus:       operator{2, false},
		tokenMultiply:    operator{3, false},
		tokenFloatDivide: operator{3, false},
		tokenIntDivide:   operator{3, false},
	}

	isOp := func(t token) bool {
		return t.tokenType == tokenPlus ||
			t.tokenType == tokenMinus ||
			t.tokenType == tokenMultiply ||
			t.tokenType == tokenFloatDivide ||
			t.tokenType == tokenIntDivide
	}

	outQueue := lane.NewQueue()
	stack := lane.NewStack()

	for i, t := range tokens {
		if t.tokenType == tokenNumber {
			outQueue.Enqueue(t)
		} else if i+1 < len(tokens) && t.tokenType == tokenName &&
			tokens[i+1].tokenType == tokenOpenBracket {

			if _, found := functionDefinitions[t.value.(string)]; found {
				// Token is a function name
				stack.Push(t)
			} else {
				// Token is a varible * somthing in brackets
				outQueue.Enqueue(t)
			}
		} else if t.tokenType == tokenComma {
			for stack.Head().(token).tokenType != tokenOpenBracket {
				outQueue.Enqueue(stack.Pop())
			}

			if stack.Empty() {
				return nil, Error{
					err:        "Misplaced comma or mismatched parentheses",
					tokenRange: []token{tokens[0], tokens[len(tokens)-1]},
				}
			}
		} else if isOp(t) {
			op := opMap[t.tokenType]

			for !stack.Empty() &&
				isOp(stack.Head().(token)) &&
				((!op.right && op.precendence <= opMap[stack.Head().(token).tokenType].precendence) ||
					(op.right && op.precendence < opMap[stack.Head().(token).tokenType].precendence)) {

				outQueue.Enqueue(stack.Pop())
			}

			stack.Push(t)
		} else if t.tokenType == tokenOpenBracket {
			stack.Push(t)
		} else if t.tokenType == tokenCloseBracket {
			if stack.Empty() {
				return nil, Error{
					err:        "Mismatched parentheses",
					tokenRange: []token{tokens[0], tokens[len(tokens)-1]},
				}
			}

			for !stack.Empty() && stack.Head().(token).tokenType != tokenOpenBracket {
				outQueue.Enqueue(stack.Pop())
			}

			if stack.Empty() {
				return nil, Error{
					err:        "Mismatched parentheses",
					tokenRange: []token{tokens[0], tokens[len(tokens)-1]},
				}
			}

			stack.Pop() // pop open bracket off

			if t.tokenType == tokenName {
				if _, found := functionDefinitions[t.value.(string)]; found {
					outQueue.Enqueue(stack.Pop())
				}
			}
		} else {
			outQueue.Enqueue(t)
		}
	}

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

	resolve := lane.NewStack()

	for !outQueue.Empty() {
		t := outQueue.Dequeue().(token)
		if isOp(t) {
			var exp expression
			rhs, lhs := resolve.Pop().(expression), resolve.Pop().(expression)

			switch t.tokenType {
			case tokenPlus:
				exp = addition{lhs, rhs}
			case tokenMinus:
				exp = subtraction{lhs, rhs}
			case tokenMultiply:
				exp = multiplication{lhs, rhs}
			case tokenFloatDivide:
				exp = floatDivision{lhs, rhs}
			case tokenIntDivide:
				exp = intDivision{lhs, rhs}
			}

			resolve.Push(exp)
		} else if t.tokenType == tokenName {
			// Token is a function
			if def, found := functionDefinitions[t.value.(string)]; found {
				var args []expression
				for i := 0; i < def.argumentCount; i++ {
					exp, ok := resolve.Pop().(expression)
					if !ok {
						return nil, Error{
							err:        "Expected function to have arguments",
							tokenRange: []token{t},
						}
					}
					args = append(args, exp)
				}

				resolve.Push(call{t.value.(string), args})
			} else {
				resolve.Push(name{t.value.(string)})
			}

		} else if t.tokenType == tokenNumber {
			resolve.Push(number{t.value.(int)})
		} else if t.tokenType == tokenFloat {
			resolve.Push(float{t.value.(float32)})
		} else {
			return nil, Error{
				err:        "Unexpected token",
				tokenRange: []token{t},
			}
		}
	}

	return resolve.Head().(expression), nil
}
