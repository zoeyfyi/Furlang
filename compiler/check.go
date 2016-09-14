package compiler

func check(functions []function) error {
	for _, f := range functions {
		for _, line := range f.lines {
			switch line := line.(type) {
			case ret:
				for i, r := range line.returns {
					eType, err := checkMaths(r, f)
					if err != nil {
						return err
					}
					if eType != f.returns[i].nameType {
						return Error{
							err:        "Wrong return type",
							tokenRange: []token{},
						}
					}
				}
			case assignment:
				eType, err := checkMaths(line.value, f)
				if err != nil {
					return err
				}
				f.names = append(f.names, typedName{eType, line.name})
			}
		}
	}

	return nil
}

func checkMaths(e expression, f function) (int, error) {
	checkOp := func(lhs expression, rhs expression, f function) (int, error) {
		lhsType, err := checkMaths(lhs, f)
		if err != nil {
			return 0, err
		}
		rhsType, err := checkMaths(rhs, f)
		if err != nil {
			return 0, err
		}

		if lhsType != rhsType {
			return 0, Error{
				err:        "Mismatched types: ",
				tokenRange: []token{},
			}
		}

		return lhsType, nil
	}

	switch e := e.(type) {
	case addition:
		return checkOp(e.lhs, e.rhs, f)
	case subtraction:
		return checkOp(e.lhs, e.rhs, f)
	case multiplication:
		return checkOp(e.lhs, e.rhs, f)
	case floatDivision:
		return checkOp(e.lhs, e.rhs, f)
	case intDivision:
		return checkOp(e.lhs, e.rhs, f)
	case number:
		return typeInt32, nil
	case float:
		return typeFloat32, nil
	case name:
		for _, tn := range f.names {
			if e.name == tn.name {
				return tn.nameType, nil
			}
		}

		return 0, Error{
			err:        "Varible not defined",
			tokenRange: []token{},
		}
	}

	return 0, Error{
		err:        "Unrecognised token",
		tokenRange: []token{},
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
