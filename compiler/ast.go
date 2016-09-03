package compiler

import "reflect"

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

func ast(tokens []Token) (funcs []function) {
	// Parses a list of types and names with a format of 'type a, type b, type c, ...'
	parseTypedList := func(tokens []Token) (names []typedName, lastPosition int) {

		for i, t := range tokens {
			switch t := t.(type) {
			case TokenSymbol:
				if t.symbol != "," {
					return names, i
				}
			case TokenType:
				names = append(names, typedName{t: t.typeName})
			case TokenName:
				names[len(names)-1].name = t.name
			}
		}

		return nil, 0
	}

	getTokenPosition := func(toke Token, tokens []Token) (pos int) {
		for i, t := range tokens {
			if reflect.DeepEqual(toke, t) {
				return i
			}
		}

		return -1
	}

	parseFunction := func(tokens []Token, functionNames []string) (function function) {
		t := tokens[0].(TokenName)
		function.name = t.name

		// Ofset of 2 skips the function name and the double colons

		colonPosition := getTokenPosition(TokenSymbol{"::"}, tokens)
		arguments, _ := parseTypedList(tokens[colonPosition+1:])

		arrowPosition := getTokenPosition(TokenSymbol{"->"}, tokens)
		returns, _ := parseTypedList(tokens[arrowPosition+1:])

		blockPosition := getTokenPosition(TokenSymbol{"{"}, tokens)
		var tokenBuffer []Token
		for _, t := range tokens[blockPosition+1:] {
			if token, found := t.(TokenSymbol); found && token.symbol == "\n" {

				if tokenBuffer != nil {
					function.lines = append(function.lines, parseExpression(tokenBuffer, functionNames))
				}

				tokenBuffer = nil
			} else {
				tokenBuffer = append(tokenBuffer, t)
			}
		}

		function.args = arguments
		function.returns = returns

		return function
	}

	var functions []int
	for i, t := range tokens {
		switch t := t.(type) {
		case TokenSymbol:
			if t.symbol == "::" {
				functions = append(functions, i-1)
			}

			if t.symbol == "}" {
				functions = append(functions, i)
			}
		}
	}

	var functionNames []string
	for i := 0; i < len(functions); i += 2 {
		f := parseFunction(tokens[functions[i]:functions[i+1]], functionNames)
		functionNames = append(functionNames, f.name)
		funcs = append(funcs, f)
	}

	return funcs
}

func parseExpression(tokens []Token, functionNames []string) (e expression) {
	if _, found := tokens[0].(TokenReturn); found {
		// Expression is a return (return expression)

		returnExpression := ret{}
		for _, r := range tokens[1:] {
			var exp expression
			switch r := r.(type) {
			case TokenName:
				exp = name{r.name}
			case TokenNumber:
				exp = number{r.number}
			default:
				panic("Unexpected expression in return statement")
			}

			returnExpression.returns = append(returnExpression.returns, exp)
		}

		return returnExpression
	} else if symbol, found := tokens[1].(TokenSymbol); found && symbol.symbol == ":=" {
		// Expression is an assignment (name := expression)

		assignmentExpression := assignment{}
		assignmentExpression.name = tokens[0].(TokenName).name
		assignmentExpression.value = parseExpression(tokens[2:], functionNames)

		return assignmentExpression
	} else if symbol, found := tokens[1].(TokenSymbol); found && symbol.symbol == "+" {
		// Expression is an addition (expression + expression)

		// Check if lhs is name or number
		additionExpression := addition{}
		switch t := tokens[0].(type) {
		case TokenNumber:
			additionExpression.lhs = number{t.number}
		case TokenName:
			additionExpression.lhs = name{t.name}
		default:
			panic("Unkown token on left hand side of '+'")
		}

		if len(tokens) > 3 {
			additionExpression.rhs = parseExpression(tokens[2:], functionNames)
		} else {
			// Check if rhs is name or number
			switch t := tokens[2].(type) {
			case TokenNumber:
				additionExpression.rhs = number{t.number}
			case TokenName:
				additionExpression.rhs = name{t.name}
			default:
				panic("Unkown token on right hand side of '+'")
			}
		}

		return additionExpression
		// Expression is a function call ( name(number, number, ...) )
	} else if name, found := tokens[0].(TokenName); found && stringInSlice(name.name, functionNames) {
		functionCallExpression := call{}
		functionCallExpression.function = name.name
		for i := 2; i < len(tokens); i += 2 {
			functionCallExpression.args = append(functionCallExpression.args, tokens[i].(TokenNumber).number)
		}

		return functionCallExpression
	}

	return nil
}

func stringInSlice(item string, slice []string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}

	return false
}

func pad(n int) (s string) {
	for i := 0; i < n; i++ {
		s += " "
	}

	return s
}
