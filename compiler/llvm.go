package compiler

import (
	"fmt"

	"strings"

	"llvm.org/llvm/bindings/go/llvm"
)

type llvmFunction struct {
	functions map[string]llvm.Value
	names     map[string]llvm.Value
	builder   llvm.Builder
	tempCount *int
}

func (lf llvmFunction) nextTempName() string {
	*(lf.tempCount)++
	return fmt.Sprintf("tmp%d", *(lf.tempCount))
}

func stringToType(t string) llvm.Type {
	switch t {
	case "i32":
		return llvm.Int32Type()
	default:
		panic(fmt.Sprintf("Unkown type '%s'", t))
	}
}

// Llvm compiles functions to llvm ir
func Llvm(funcs []Function) string {
	context := llvm.NewContext()
	module := context.NewModule("ben")
	builder := llvm.NewBuilder()

	functions := make(map[string]llvm.Value)
	names := make(map[string]llvm.Value)
	tempCount := 0

	for _, f := range funcs {
		// Add function definintion to module
		var argumentTypes []llvm.Type
		for _, a := range f.args {
			argumentTypes = append(argumentTypes, stringToType(a.t))
		}

		returnType := stringToType(f.returns[0].t)
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
