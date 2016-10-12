// +build llvm

package compiler

import (
	"fmt"

	"strings"

	"llvm.org/llvm/bindings/go/llvm"
)

type llvmFunction struct {
	function  llvm.Value
	functions map[string]llvm.Value
	scope     [](map[string]llvm.Value)
	level     int
	blocks    []llvm.BasicBlock
	builder   llvm.Builder
	tempCount *int
}

func (lf llvmFunction) nextTempName() string {
	*(lf.tempCount)++
	return fmt.Sprintf("tmp%d", *(lf.tempCount))
}

type expression interface {
	compile(llvmFunction) llvm.Value
}

// Llvm compiles functions to llvm ir
func Llvm(ast *SyntaxTree) string {
	context := llvm.NewContext()
	module := context.NewModule("ben")
	builder := llvm.NewBuilder()

	functions := make(map[string]llvm.Value)
	scope := make([](map[string]llvm.Value), 1000)
	for i := 0; i < 1000; i++ {
		scope[i] = make(map[string]llvm.Value)
	}

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

	for _, f := range ast.functions {
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
			scope[0][a.name] = function.Param(i)
		}

		// Create entry block and set builder cursor
		entry := llvm.AddBasicBlock(function, "entry")
		builder.SetInsertPointAtEnd(entry)

		// Compile all expressions
		lfunction := llvmFunction{
			function:  function,
			functions: functions,
			scope:     scope,
			level:     0,
			builder:   builder,
			tempCount: &tempCount,
		}
		for _, l := range f.block.expressions {
			l.compile(lfunction)
		}
	}

	// Remove weird line which stops code compiling
	s := module.String()
	return strings.Replace(s, "source_filename = \"ben\"\n", "", 1)
}

// TODO: rework compile methods
func (t function) compile(function llvmFunction) (val llvm.Value) {
	return llvm.Value{}
}

func (t name) compile(function llvmFunction) (val llvm.Value) {
	found := false
	for i := function.level; i >= 0; i-- {
		if v, ok := function.scope[i][t.name]; ok {
			val = v
			found = true
			break
		}
	}

	if !found {
		panic(fmt.Sprintf("Variable '%s' not in scope", t.name))
	}

	return val
}

func (t maths) compile(function llvmFunction) llvm.Value {
	return t.expression.compile(function)
}

func (t boolean) compile(function llvmFunction) llvm.Value {
	if t.value {
		return llvm.ConstInt(llvm.Int1Type(), 1, false)
	} else {
		return llvm.ConstInt(llvm.Int1Type(), 0, false)
	}
}

func (t block) compileBlock(function llvmFunction) llvm.BasicBlock {
	// Save refrence to parent block
	parentBlock := function.builder.GetInsertBlock()

	// Create a new block
	function.level++
	childBlock := llvm.AddBasicBlock(function.function, function.nextTempName())
	function.builder.SetInsertPointAtEnd(childBlock)

	// Compile all expressions in child block
	for _, e := range t.expressions {
		e.compile(function)
	}

	// Restore parent block position
	function.builder.SetInsertPointAtEnd(parentBlock)

	// Check if block has instructions
	if childBlock.FirstInstruction().IsNil() {
		childBlock.EraseFromParent()
		return llvm.BasicBlock{}
	}

	return childBlock
}

func (t block) compile(function llvmFunction) llvm.Value {
	// Create a branch to block
	newBlock := t.compileBlock(function)
	if newBlock.IsNil() {
		return llvm.Value{}
	}

	return function.builder.CreateBr(newBlock)
}

func (t ret) compile(function llvmFunction) llvm.Value {
	val := t.returns[0].compile(function)

	return function.builder.CreateRet(val)
}

func (t assignment) compile(function llvmFunction) llvm.Value {
	val := t.value.compile(function)
	function.scope[function.level][t.name] = val

	return val
}

func (t ifExpression) compile(function llvmFunction) llvm.Value {
	ifBranch := t.blocks[0].block.compileBlock(function)
	elseBranch := t.blocks[1].block.compileBlock(function)
	condition := t.blocks[0].condition.compile(function)
	return function.builder.CreateCondBr(condition, ifBranch, elseBranch)
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
		args = append(args, a.compile(function))
	}

	return function.builder.CreateCall(
		function.functions[t.function],
		args,
		function.nextTempName())
}
