package compiler

type typedName struct {
	t    string
	name string
}

type function struct {
	name    string
	args    []typedName
	returns []typedName
	lines   []expression
}

func ast(tokens []token) (funcs []function) {
	// Position in slice of functions
	var functionPositions []int
	for i, t := range tokens {
		switch t.tokenType {
		case tokenDoubleColon:
			functionPositions = append(functionPositions, i-1)

		case tokenCloseBody:
			functionPositions = append(functionPositions, i)

		}
	}

	// Parse functions
	var functionNames []string
	for i := 0; i < len(functionPositions); i += 2 {
		f := parseFunction(tokens[functionPositions[i]:functionPositions[i+1]], functionNames)
		functionNames = append(functionNames, f.name)
		funcs = append(funcs, f)
	}

	return funcs
}

// Gets the first position of a token in a slice with the corrasponding token type
func getTokenPosition(tokenType int, tokens []token) (pos int) {
	for i, t := range tokens {
		if t.tokenType == tokenType {
			return i
		}
	}

	return -1
}

// Create the function definition from a slice of tokens
func parseFunction(tokens []token, functionNames []string) (function function) {
	if tokens[0].tokenType != tokenName {
		panic("Expected function to begin with name")
	}

	function.name = tokens[0].value.(string)

	// Ofset of 2 skips the function name and the double colons
	colonPosition := getTokenPosition(tokenDoubleColon, tokens)
	arguments, _ := parseTypedList(tokens[colonPosition+1:])

	arrowPosition := getTokenPosition(tokenArrow, tokens)
	returns, _ := parseTypedList(tokens[arrowPosition+1:])

	blockPosition := getTokenPosition(tokenOpenBody, tokens)

	var tokenBuffer []token
	for _, t := range tokens[blockPosition+1:] {
		if t.tokenType == tokenNewLine && tokenBuffer != nil {
			function.lines = append(function.lines, parseExpression(tokenBuffer, functionNames))
			tokenBuffer = nil
		} else {
			tokenBuffer = append(tokenBuffer, t)
		}
	}

	function.args = arguments
	function.returns = returns

	return function
}

// Parses a list of types and names with a format of 'type a, type b, type c, ...'
func parseTypedList(tokens []token) (names []typedName, lastPosition int) {
	lastToken := -1

	for i, t := range tokens {
		switch t.tokenType {
		case tokenComma:
			if lastToken != tokenName {
				panic("Unexpected comma")
			}
		case tokenInt32:
			if lastToken == tokenName {
				panic("Unexpected type")
			}

			names = append(names, typedName{t: "i32"})
		case tokenName:
			if lastToken != tokenInt32 {
				panic("Unexpected name")
			}

			names[len(names)-1].name = t.value.(string)
		default:
			return names, i
		}

		lastToken = t.tokenType
	}

	return nil, 0
}

func parseExpression(tokens []token, functionNames []string) (e expression) {
	for tokens[0].tokenType == tokenNewLine {
		tokens = tokens[1:]
	}

	if tokens[0].tokenType == tokenReturn {
		// Expression is a return expression
		returnExpression := ret{}

		// Create a slice of expressions to return
		for _, r := range tokens[1:] {
			var exp expression

			switch r.tokenType {
			case tokenName:
				exp = name{r.value.(string)}
			case tokenNumber:
				exp = number{r.value.(int)}
			default:
				panic("Unexpected expression in return statement")
			}

			returnExpression.returns = append(returnExpression.returns, exp)
		}

		return returnExpression
	} else if tokens[1].tokenType == tokenAssign {
		// Expression is an assignment

		assignmentExpression := assignment{}
		assignmentExpression.name = tokens[0].value.(string)
		assignmentExpression.value = parseExpression(tokens[2:], functionNames)

		return assignmentExpression
	} else if tokens[1].tokenType == tokenPlus {
		// Expression is an addition (expression + expression)

		// Check if lhs is name or number
		additionExpression := addition{}
		switch tokens[0].tokenType {
		case tokenNumber:
			additionExpression.lhs = number{tokens[0].value.(int)}
		case tokenName:
			additionExpression.lhs = name{tokens[0].value.(string)}
		default:
			panic("Unkown token on left hand side of '+'")
		}

		if len(tokens) > 3 {
			additionExpression.rhs = parseExpression(tokens[2:], functionNames)
		} else {
			// Check if rhs is name or number
			switch tokens[2].tokenType {
			case tokenNumber:
				additionExpression.rhs = number{tokens[2].value.(int)}
			case tokenName:
				additionExpression.rhs = name{tokens[2].value.(string)}
			default:
				panic("Unkown token on right hand side of '+'")
			}
		}

		return additionExpression
	} else if tokens[1].tokenType == tokenMinus {
		// Check if lhs is name or number
		subtractionExpression := subtraction{}
		switch tokens[0].tokenType {
		case tokenNumber:
			subtractionExpression.lhs = number{tokens[0].value.(int)}
		case tokenName:
			subtractionExpression.lhs = name{tokens[0].value.(string)}
		default:
			panic("Unkown token on left hand side of '-'")
		}

		if len(tokens) > 3 {
			subtractionExpression.rhs = parseExpression(tokens[2:], functionNames)
		} else {
			// Check if rhs is name or number
			switch tokens[2].tokenType {
			case tokenNumber:
				subtractionExpression.rhs = number{tokens[2].value.(int)}
			case tokenName:
				subtractionExpression.rhs = name{tokens[2].value.(string)}
			default:
				panic("Unkown token on right hand side of '-'")
			}
		}

		return subtractionExpression
	} else if tokens[0].tokenType == tokenName && stringInSlice(tokens[0].value.(string), functionNames) {
		// Expression is a function call ( name(number, number, ...) )
		functionCallExpression := call{}
		functionCallExpression.function = tokens[0].value.(string)
		for i := 2; i < len(tokens); i += 2 {
			functionCallExpression.args = append(functionCallExpression.args, tokens[i].value.(int))
		}

		return functionCallExpression
	} else {
		panic("Unkown expression")
	}
}

func stringInSlice(item string, slice []string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}

	return false
}
