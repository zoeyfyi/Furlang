package compiler

import (
	"fmt"

	"llvm.org/llvm/bindings/go/llvm"
)

type expression interface {
	compile(llvmFunction) llvm.Value
}

// ================ Name expression ================ //
type name struct {
	name string
}

func newName(nameToken token) name {
	if nameToken.tokenType != tokenName {
		panic("Exprected name token")
	}

	return name{nameToken.value.(string)}
}

func (t name) compile(function llvmFunction) llvm.Value {
	val, ok := function.names[t.name]
	if !ok {
		panic(fmt.Sprintf("Variable '%s' not in scope", t.name))
	}

	return val
}

// ================ Return expression ================ //
type ret struct {
	returns []expression
}

func newRet(returnTokens []token) ret {
	retExpression := ret{}
	lastComma := 0

	// Remove return token
	returnTokens = returnTokens[1:]

	for i, t := range returnTokens {
		if t.tokenType == tokenComma || i == len(returnTokens)-1 {
			exp := parseExpression(returnTokens[lastComma : i+1])
			retExpression.returns = append(retExpression.returns, exp)
			lastComma = i
		}
	}

	return retExpression
}

func (t ret) compile(function llvmFunction) llvm.Value {
	val := t.returns[0].compile(function)

	return function.builder.CreateRet(val)
}

// ================ Assignment expression ================ //
type assignment struct {
	name  string
	value expression
}

func newAssignment(tokens []token) assignment {
	assignmentExpression := assignment{}

	assignmentExpression.name = tokens[0].value.(string)
	assignmentExpression.value = parseExpression(tokens[2:])

	return assignmentExpression
}

func (t assignment) compile(function llvmFunction) llvm.Value {
	val := t.value.compile(function)
	function.names[t.name] = val

	return val
}

// ================ Addition expression ================ //
type addition struct {
	lhs expression
	rhs expression
}

func newAddition(tokens []token) addition {
	lhs, rhs := newMath(tokens)
	return addition{lhs, rhs}
}

func (t addition) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateAdd(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

// ================ Subtraction expression ================ //
type subtraction struct {
	lhs expression
	rhs expression
}

func newSubtraction(tokens []token) subtraction {
	lhs, rhs := newMath(tokens)
	return subtraction{lhs, rhs}
}

func (t subtraction) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateSub(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

// ================ Multiply expression ================ //
type multiplication struct {
	lhs expression
	rhs expression
}

func newMulitply(tokens []token) multiplication {
	lhs, rhs := newMath(tokens)
	return multiplication{lhs, rhs}
}

func (t multiplication) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateMul(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

// ================ Division expression ================ //
type division struct {
	lhs expression
	rhs expression
}

func newDivision(tokens []token) division {
	lhs, rhs := newMath(tokens)
	return division{lhs, rhs}
}

func (t division) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateSDiv(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

// ================ Math expression ================ //

func newMath(tokens []token) (lhs expression, rhs expression) {
	switch tokens[0].tokenType {
	case tokenNumber:
		lhs = newNumber(tokens[0])
	case tokenName:
		lhs = newName(tokens[0])
	}

	if len(tokens) > 3 {
		rhs = parseExpression(tokens[2:])
	} else {
		switch tokens[2].tokenType {
		case tokenNumber:
			rhs = newNumber(tokens[2])
		case tokenName:
			rhs = newName(tokens[2])
		}
	}

	return lhs, rhs
}

// ================ Number expression ================ //
type number struct {
	value int
}

func newNumber(t token) number {
	if t.tokenType != tokenNumber {
		panic("expected number token")
	}

	return number{t.value.(int)}
}

func (t number) compile(function llvmFunction) llvm.Value {
	return llvm.ConstInt(
		llvm.Int32Type(),
		uint64(t.value),
		false)
}

// ================ Function call expression ================ //
type call struct {
	function string
	args     []int
}

func newCall(tokens []token) call {
	callExpression := call{}

	if tokens[0].tokenType != tokenName {
		panic("Expected function name (tokenName)")
	}

	callExpression.function = tokens[0].value.(string)

	for i := 2; i < len(tokens); i += 2 {
		callExpression.args = append(callExpression.args, tokens[i].value.(int))
	}

	return callExpression
}

func (t call) compile(function llvmFunction) llvm.Value {
	args := []llvm.Value{}
	for _, a := range t.args {
		args = append(args, number{a}.compile(function))
	}

	return function.builder.CreateCall(
		function.functions[t.function],
		args,
		function.nextTempName())
}
