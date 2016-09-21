// +build never

{


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
				// case tokenReturn:
				// 	switch state {
				// 	case stateLineNone:
				// 		state = stateLineReturn
				// 		currectExpression = ret{}
				// 	}

				// Assignment, Function call or varible
				// case tokenName:
				// 	switch state {
				// 	case stateLineNone:
				// 		switch innerTokens[i+1].tokenType {
				// 		// Assignment expression
				// 		case tokenAssign:
				// 			state = stateLineAssignment
				// 			newAssignment := assignment{}
				// 			newAssignment.name = t.value.(string)
				// 			currectExpression = newAssignment

				// 		// Function call expression
				// 		case tokenOpenBracket:
				// 			state = stateLineFunctionCall
				// 		}

				// 	case stateLineAssignment, stateLineFunctionCall, stateLineReturn:
				// 		if _, found := functionArgMap[t.value.(string)]; found {
				// 			// Token is a function name
				// 			stack.Push(t)
				// 		} else {
				// 			// Token is a varible
				// 			outQueue.Enqueue(t)
				// 		}
				// 	default:
				// 		return nil, StateError
				// 	}

				// Assignment
				// case tokenAssign:
				// 	if _, ok := currectExpression.(assignment); !ok {
				// 		return nil, Error{
				// 			err:        "Misplaced assignment",
				// 			tokenRange: []token{t},
				// 		}
				// 	}

				// Newline
				// case tokenNewLine:
				// 	switch state {
				// 	case stateLineNone:
				// 		// Ignore blank lines

					// case stateLineAssignment:
					// 	exp, err := createExpression(stack, outQueue)
					// 	if err != nil {
					// 		return nil, err
					// 	}

					// 	assignmentExpression := currectExpression.(assignment)
					// 	assignmentExpression.value = exp
					// 	parseFunction.lines = append(parseFunction.lines, assignmentExpression)

					// case stateLineFunctionCall:
					// 	exp, err := createExpression(stack, outQueue)
					// 	if err != nil {
					// 		return nil, err
					// 	}

					// 	parseFunction.lines = append(parseFunction.lines, exp)

					// case stateLineReturn:
					// 	exp, err := createExpression(stack, outQueue)
					// 	if err != nil {
					// 		return nil, err
					// 	}

					// 	retExpression := currectExpression.(ret)
					// 	retExpression.returns = append(retExpression.returns, exp)
					// 	parseFunction.lines = append(parseFunction.lines, retExpression)

					// default:
					// 	return nil, StateError
					// }
					// state = stateLineNone
					// continue tokenLoop

				// Number token
				// case tokenNumber, tokenFloat:
				// 	switch state {
				// 	case stateLineNone:
				// 		return nil, Error{
				// 			err:        "Unexpected number",
				// 			tokenRange: []token{t},
				// 		}
				// 	case stateLineAssignment, stateLineFunctionCall, stateLineReturn:
				// 		outQueue.Enqueue(t)
				// 	default:
				// 		return nil, StateError
				// 	}

				// Comma token
				// case tokenComma:
				// 	switch state {
				// 	case stateLineNone:
				// 		return nil, Error{
				// 			err:        "Unexpected comma",
				// 			tokenRange: []token{t},
				// 		}
				// 	case stateLineReturn:
				// 		for stack.Head().(token).tokenType != tokenOpenBracket {
				// 			outQueue.Enqueue(stack.Pop())
				// 		}

				// 		if stack.Empty() && len(parseFunction.returns) < len(currectExpression.(ret).returns) {
				// 			// Comma is seperating multiple return values
				// 			retExpression := currectExpression.(ret)
				// 			exp, err := createExpression(stack, outQueue)
				// 			if err != nil {
				// 				return nil, err
				// 			}
				// 			retExpression.returns = append(retExpression.returns, exp)
				// 		} else if stack.Empty() {
				// 			// Comma is out of place
				// 			return nil, Error{
				// 				err:        "Misplaced comma or mismatched parentheses",
				// 				tokenRange: []token{t},
				// 			}
				// 		}

				// 	case stateLineAssignment, stateLineFunctionCall:
				// 		for stack.Head().(token).tokenType != tokenOpenBracket {
				// 			outQueue.Enqueue(stack.Pop())
				// 		}

				// 		if stack.Empty() {
				// 			// Comma is out of place
				// 			return nil, Error{
				// 				err:        "Misplaced comma or mismatched parentheses",
				// 				tokenRange: []token{t},
				// 			}
				// 		}
				// 	default:
				// 		return nil, StateError
				// 	}

				// Mathmatical operator
				// case tokenPlus, tokenMinus, tokenMultiply, tokenIntDivide, tokenFloatDivide:
				// 	op := opMap[t.tokenType]

				// 	for !stack.Empty() &&
				// 		(stack.Head().(token).tokenType == tokenPlus ||
				// 			stack.Head().(token).tokenType == tokenMinus ||
				// 			stack.Head().(token).tokenType == tokenMultiply ||
				// 			stack.Head().(token).tokenType == tokenFloatDivide ||
				// 			stack.Head().(token).tokenType == tokenIntDivide) &&
				// 		((!op.right && op.precendence <= opMap[stack.Head().(token).tokenType].precendence) ||
				// 			(op.right && op.precendence < opMap[stack.Head().(token).tokenType].precendence)) {

				// 		outQueue.Enqueue(stack.Pop())
				// 	}

				// 	stack.Push(t)

				// // Open bracket
				// case tokenOpenBracket:
				// 	stack.Push(t)

				// Close bracket
				// case tokenCloseBracket:
				// 	if stack.Empty() {
				// 		return nil, Error{
				// 			err:        "Mismatched parentheses",
				// 			tokenRange: []token{t},
				// 		}
				// 	}

				// 	for !stack.Empty() && stack.Head().(token).tokenType != tokenOpenBracket {
				// 		outQueue.Enqueue(stack.Pop())
				// 	}

				// 	if stack.Empty() {
				// 		return nil, Error{
				// 			err:        "Mismatched parentheses",
				// 			tokenRange: []token{t},
				// 		}
				// 	}

				// 	stack.Pop() // pop open bracket off

				// 	if t.tokenType == tokenName {
				// 		if _, found := functionArgMap[t.value.(string)]; found {
				// 			outQueue.Enqueue(stack.Pop())
				// 		}
				// 	}

				// // Default case
				// default:
				// 	return nil, Error{
				// 		err:        "Unexpected token",
				// 		tokenRange: []token{t},
				// 	}
				// }

			}
		}
	}