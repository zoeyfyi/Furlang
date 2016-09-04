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
			function.lines = append(function.lines, parseExpression(tokenBuffer))
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

func parseExpression(tokens []token) (e expression) {
	// Remove new line tokens
	for tokens[0].tokenType == tokenNewLine {
		tokens = tokens[1:]
	}

	if len(tokens) == 1 && tokens[0].tokenType == tokenName {
		return newName(tokens[0])
	} else if len(tokens) == 1 && tokens[0].tokenType == tokenNumber {
		return newNumber(tokens[0])
	} else if tokens[0].tokenType == tokenReturn {
		return newRet(tokens)
	} else if tokens[1].tokenType == tokenAssign {
		return newAssignment(tokens)
	} else if tokens[1].tokenType == tokenPlus {
		return newAddition(tokens)
	} else if tokens[1].tokenType == tokenMinus {
		return newSubtraction(tokens)
	} else if tokens[1].tokenType == tokenMultiply {
		return newMulitply(tokens)
	} else if tokens[1].tokenType == tokenDivide {
		return newDivision(tokens)
	} else if tokens[0].tokenType == tokenName && tokens[1].tokenType == tokenOpenBracket {
		return newCall(tokens)
	}

	panic("Unkown expression")

}

func stringInSlice(item string, slice []string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}

	return false
}
