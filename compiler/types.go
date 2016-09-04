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

func (t ret) compile(function llvmFunction) llvm.Value {
	val := t.returns[0].compile(function)

	return function.builder.CreateRet(val)
}

// ================ Assignment expression ================ //
type assignment struct {
	name  string
	value expression
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

func (t subtraction) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateSub(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

// ================ Number expression ================ //
type number struct {
	value int
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
