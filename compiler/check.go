package compiler

import "fmt"

func check(ast *abstractSyntaxTree) error {
	// for i := 0; i < len(ast.functions); i++ {
	// 	f := &ast.functions[i]
	// 	f.names = append(f.names, f.args...)
	// 	for _, line := range f.lines {
	// 		switch line := line.(type) {
	// 		case ret:
	// 			for i, r := range line.returns {
	// 				eType, err := checkMaths(r, ast.functions, i)
	// 				if err != nil {
	// 					return err
	// 				}
	// 				if eType != f.returns[i].nameType {
	// 					return Error{
	// 						err:        "Wrong return type",
	// 						tokenRange: []token{},
	// 					}
	// 				}
	// 			}
	// 		case assignment:
	// 			eType, err := checkMaths(line.value, ast.functions, i)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			f.names = append(f.names, typedName{eType, line.name})
	// 		}
	// 	}
	// }

	return nil
}

func checkOp(lhs expression, rhs expression, functions []function, fkey int) (int, error) {
	lhsType, err := checkMaths(lhs, functions, fkey)
	if err != nil {
		return 0, err
	}
	rhsType, err := checkMaths(rhs, functions, fkey)
	if err != nil {
		return 0, err
	}

	if lhsType != rhsType {
		return 0, Error{
			err:        "Mismatched types: ",
			tokenRange: nil,
		}
	}

	return lhsType, nil
}

// TODO:
func checkMaths(e expression, functions []function, fkey int) (int, error) {
	switch e := e.(type) {
	case addition:
		return checkOp(e.lhs, e.rhs, functions, fkey)
	case subtraction:
		return checkOp(e.lhs, e.rhs, functions, fkey)
	case multiplication:
		return checkOp(e.lhs, e.rhs, functions, fkey)
	case floatDivision:
		return checkOp(e.lhs, e.rhs, functions, fkey)
	case intDivision:
		return checkOp(e.lhs, e.rhs, functions, fkey)
	case number:
		return typeInt32, nil
	case float:
		return typeFloat32, nil
	case name:
		// for _, tn := range functions[fkey].names {
		// 	if e.name == tn.name {
		// 		return tn.nameType, nil
		// 	}
		// }

		return 0, Error{
			err:        fmt.Sprintf("Varible %s not defined", e.name),
			tokenRange: nil,
		}
	case call:
		var functionCall function
		for _, f := range functions {
			if f.name == e.function {
				functionCall = f
			}
		}

		for i, arg := range e.args {
			expType, err := checkMaths(arg, functions, fkey)
			if err != nil {
				return 0, err
			}
			if expType != functionCall.args[i].nameType {
				return 0, Error{
					err:        "Wrong argument type",
					tokenRange: nil,
				}
			}

		}

		// Handle mulitple types
		return functions[fkey].returns[0].nameType, nil
	}

	fmt.Printf("%+v\n", e)
	return 0, Error{
		err:        "Unrecognised token",
		tokenRange: nil,
	}
}

func checkType(e expression, t int, nameTypes map[string]int) error {
	// TODO: Add token range to errors
	switch e := e.(type) {
	case number:
		if t != typeInt32 {
			return Error{
				err:        "Mismatched types:",
				tokenRange: []token{},
			}
		}
	case float:
		if t != typeFloat32 {
			return Error{
				err:        "Mismatched types:",
				tokenRange: []token{},
			}
		}
	case name:
		nt, found := nameTypes[e.name]

		if !found {
			return Error{
				err:        "Varible dosnt exsists",
				tokenRange: []token{},
			}
		}

		if nt != t {
			return Error{
				err:        "Mismatched types:",
				tokenRange: []token{},
			}
		}
	}

	return nil
}
