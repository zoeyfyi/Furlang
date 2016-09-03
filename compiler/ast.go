package compiler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type typedName struct {
	t    string
	name string
}

type Function struct {
	Name      string
	Arguments []typedName
	Returns   []typedName
	Lines     []expression
}

func Ast(tokens []Token) (funcs []Function) {
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

	parseFunction := func(tokens []Token, functionNames []string) (function Function) {
		t := tokens[0].(TokenName)
		function.Name = t.name

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
					function.Lines = append(function.Lines, parseExpression(tokenBuffer, functionNames))
				}

				tokenBuffer = nil
			} else {
				tokenBuffer = append(tokenBuffer, t)
			}
		}

		function.Arguments = arguments
		function.Returns = returns

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
		functionNames = append(functionNames, f.Name)
		funcs = append(funcs, f)
	}

	return funcs
}

func Dump(item interface{}, name string) {
	itype := reflect.TypeOf(item)
	ivalue := reflect.ValueOf(item)

	printThis(name, itype, ivalue, "")
}

func printThis(name string, itype reflect.Type, ivalue reflect.Value, depth string) {
	cname := color.New(color.FgHiCyan).SprintfFunc()
	ctype := color.New(color.FgBlue).SprintfFunc()
	cstring := color.New(color.FgYellow).SprintfFunc()
	cint := color.New(color.FgHiGreen).SprintfFunc()

	typeName := strings.Replace(itype.String(), "compiler.", "", 100)

	if itype.Kind() != reflect.Interface {
		s := ""
		if name != "" {
			s = cname("%s", name) + ctype(" %s", typeName)
			depth += strings.Repeat(" ", len(name)+1+len(typeName))
		} else {
			s = ctype("%s", typeName)
			depth += strings.Repeat(" ", len(typeName))
		}

		fmt.Print(s)
	}

	if itype.Kind() == reflect.Slice {

		if ivalue.Len() == 0 {
			fmt.Println()
		}

		for i := 0; i < ivalue.Len(); i++ {
			if ivalue.Len() == 1 {
				fmt.Printf(" ─── ")
				printThis("", ivalue.Index(i).Type(), ivalue.Index(i), depth+"     ")
			} else if i == 0 {
				fmt.Printf(" ─┬─ ")
				printThis("", ivalue.Index(i).Type(), ivalue.Index(i), depth+"  │  ")
			} else if i == ivalue.Len()-1 {
				fmt.Printf("%s  └─ ", depth)
				printThis("", ivalue.Index(i).Type(), ivalue.Index(i), depth+"     ")
			} else {
				fmt.Printf("%s  ├─ ", depth)
				printThis("", ivalue.Index(i).Type(), ivalue.Index(i), depth+"  │  ")
			}
		}
	} else if itype.Kind() == reflect.String {
		s := cstring(" %s", strconv.Quote(ivalue.String()))
		fmt.Printf("%s\n", s)
	} else if itype.Kind() == reflect.Int {
		s := cint(" %d", ivalue.Int())
		fmt.Printf("%s\n", s)
	} else if itype.Kind() == reflect.Struct {
		if ivalue.NumField() == 0 {
			fmt.Println()
		}

		for i := 0; i < ivalue.NumField(); i++ {
			if ivalue.NumField() == 1 {
				fmt.Printf(" ─── ")
				printThis(itype.Field(i).Name, ivalue.Field(i).Type(), ivalue.Field(i), depth+"     ")
			} else if i == 0 {
				fmt.Printf(" ─┬─ ")
				printThis(itype.Field(i).Name, ivalue.Field(i).Type(), ivalue.Field(i), depth+"  │  ")
			} else if i == ivalue.NumField()-1 {
				fmt.Printf("%s  └─ ", depth)
				printThis(itype.Field(i).Name, ivalue.Field(i).Type(), ivalue.Field(i), depth+"     ")
			} else {
				fmt.Printf("%s  ├─ ", depth)
				printThis(itype.Field(i).Name, ivalue.Field(i).Type(), ivalue.Field(i), depth+"  │  ")
			}
		}
	} else if itype.Kind() == reflect.Interface {
		val := ivalue.Interface()
		printThis(name, reflect.TypeOf(val), reflect.ValueOf(val), depth)
	} else {
		panic("Unrecognized type")
	}
}

func parseExpression(tokens []Token, functionNames []string) (e Expression) {
	// Expression is a return (return expression)
	if _, found := tokens[0].(TokenReturn); found {
		returnExpression := Return{}
		for _, r := range tokens[1:] {
			if _, found := r.(TokenName); found {
				returnExpression.Returns = append(returnExpression.Returns, Name{r.(TokenName).name})
			} else if _, found := r.(TokenNumber); found {
				returnExpression.Returns = append(returnExpression.Returns, Number{r.(TokenNumber).number})
			}
		}

		return returnExpression
		// Expression is an assignment (name := expression)
	} else if symbol, found := tokens[1].(TokenSymbol); found && symbol.symbol == ":=" {
		assignmentExpression := Assignment{}
		assignmentExpression.Assignee = tokens[0].(TokenName).name
		assignmentExpression.Assign = parseExpression(tokens[2:], functionNames)

		return assignmentExpression
		// Expression is an addition (expression + expression)
	} else if symbol, found := tokens[1].(TokenSymbol); found && symbol.symbol == "+" {
		additionExpression := Addition{}
		// Check if lhs is name or number
		switch t := tokens[0].(type) {
		case TokenNumber:
			additionExpression.Lhs = Number{t.number}
		case TokenName:
			additionExpression.Lhs = Name{t.name}
		default:
			panic("Unkown token on left hand side of '+'")
		}

		if len(tokens) > 3 {
			additionExpression.Rhs = parseExpression(tokens[2:], functionNames)
		} else {
			// Check if rhs is name or number
			switch t := tokens[2].(type) {
			case TokenNumber:
				additionExpression.Rhs = Number{t.number}
			case TokenName:
				additionExpression.Rhs = Name{t.name}
			default:
				panic("Unkown token on right hand side of '+'")
			}
		}

		return additionExpression
		// Expression is a function call ( name(number, number, ...) )
	} else if name, found := tokens[0].(TokenName); found && stringInSlice(name.name, functionNames) {
		functionCallExpression := FunctionCall{}
		functionCallExpression.Function = name.name
		for i := 2; i < len(tokens); i += 2 {
			functionCallExpression.Arguments = append(functionCallExpression.Arguments, tokens[i].(TokenNumber).number)
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

// func debugPrint(functions []Function) {
// 	funcLength := 0

// 	for _, f := range functions {
// 		if funcLength < len(f.name.name) {
// 			funcLength = len(f.name.name)
// 		}
// 	}

// 	ff := strconv.Itoa(funcLength) // Func format

// 	for _, f := range functions {
// 		s := fmt.Sprintf("%-"+ff+"s -> %-7s -> ", f.name.name, "args")
// 		fmt.Printf(s)
// 		for i, arg := range f.arguments {
// 			if i == 0 {
// 				fmt.Printf("%s -> %s", arg.name.name, arg.nType.typeName)

// 			} else {
// 				fmt.Printf("\n%s-> %s -> %s", pad(len(s)-3), arg.name.name, arg.nType.typeName)
// 			}
// 		}

// 		s = fmt.Sprintf("\n%-"+ff+"s -> returns -> ", pad(len(f.name.name)))
// 		fmt.Printf(s)
// 		for i, arg := range f.returns {
// 			if i == 0 {
// 				fmt.Printf("%s -> %s", arg.name.name, arg.nType.typeName)

// 			} else {
// 				fmt.Printf("\n%s-> %s -> %s", pad(len(s)-2), arg.name.name, arg.nType.typeName)
// 			}
// 		}

// 		fmt.Printf("\n\n")
// 	}
// }

func pad(n int) (s string) {
	for i := 0; i < n; i++ {
		s += " "
	}

	return s
}
