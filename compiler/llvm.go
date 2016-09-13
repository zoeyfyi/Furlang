// +build llvm

package compiler

import (
	"fmt"

	"strings"

	"llvm.org/llvm/bindings/go/llvm"
)

type value llvm.Value
type build llvm.Builder

type expression interface {
	compile(llvmFunction) value
}

// Llvm compiles functions to llvm ir
func Llvm(funcs []function) string {
	context := llvm.NewContext()
	module := context.NewModule("ben")
	builder := llvm.NewBuilder()

	functions := make(map[string]llvm.Value)
	names := make(map[string]llvm.Value)
	tempCount := 0

	toType := func(t int) llvm.Type {
		switch t {
		case typeInt32:
			return llvm.Int32Type()
		case typeFloat32:
			return llvm.FloatType()
		default:
			panic(fmt.Sprintf("Unkown type '%d'", t))
		}
	}

	for _, f := range funcs {
		// Add function definintion to module
		var argumentTypes []llvm.Type
		for _, a := range f.args {
			argumentTypes = append(argumentTypes, toType(a.nameType))
		}

		returnType := toType(f.returns[0].nameType)
		functionType := llvm.FunctionType(returnType, argumentTypes, false)
		function := llvm.AddFunction(module, f.name, functionType)

		// Add function to function map
		functions[f.name] = function

		// Add function parameters to names map
		for i, a := range f.args {
			names[a.name] = function.Param(i)
		}

		// Create entry block and set builder cursor
		entry := llvm.AddBasicBlock(function, "entry")
		builder.SetInsertPointAtEnd(entry)

		// Compile all expressions
		lfunction := llvmFunction{functions, names, builder, &tempCount}
		for _, l := range f.lines {
			l.compile(lfunction)
		}
	}

	// Remove weird line which stops code compiling
	s := module.String()
	return strings.Replace(s, "source_filename = \"ben\"\n", "", 1)
}

func (t name) compile(function llvmFunction) llvm.Value {
	val, ok := function.names[t.name]
	if !ok {
		panic(fmt.Sprintf("Variable '%s' not in scope", t.name))
	}

	return val
}

func (t ret) compile(function llvmFunction) llvm.Value {
	val := t.returns[0].compile(function)

	return function.builder.CreateRet(val)
}

func (t assignment) compile(function llvmFunction) llvm.Value {
	val := t.value.compile(function)
	function.names[t.name] = val

	return val
}

func (t addition) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateAdd(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

func (t subtraction) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateSub(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

func (t multiplication) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateMul(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

func (t floatDivision) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateFDiv(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

func (t intDivision) compile(function llvmFunction) llvm.Value {
	return function.builder.CreateUDiv(
		t.lhs.compile(function),
		t.rhs.compile(function),
		function.nextTempName())
}

func (t number) compile(function llvmFunction) llvm.Value {
	return llvm.ConstInt(llvm.Int32Type(), uint64(t.value), false)
}

func (t float) compile(function llvmFunction) llvm.Value {
	return llvm.ConstFloat(
		llvm.FloatType(),
		float64(t.value))
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
