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

func Llvm(funcs []Function) string {
	context := llvm.NewContext()
	module := context.NewModule("ben")
	builder := llvm.NewBuilder()

	functions := make(map[string]llvm.Value)
	names := make(map[string]llvm.Value)
	tempCount := 0

	for _, f := range funcs {
		var argumentTypes []llvm.Type
		for _, a := range f.Arguments {
			switch a.Type.Name {
			case "i32":
				argumentTypes = append(argumentTypes, llvm.Int32Type())
			}
		}

		var returnType llvm.Type
		switch f.Returns[0].Type.Name {
		case "i32":
			returnType = llvm.Int32Type()
		}

		functionType := llvm.FunctionType(returnType, argumentTypes, false)
		function := llvm.AddFunction(module, f.Name, functionType)

		functions[f.Name] = function

		for i, a := range f.Arguments {
			names[a.Name] = function.Param(i)
		}

		entry := llvm.AddBasicBlock(function, "entry")
		builder.SetInsertPointAtEnd(entry)

		lfunction := llvmFunction{functions, names, builder, &tempCount}

		for _, l := range f.Lines {
			l.Compile(lfunction)
		}
	}

	// Remove weird line which stops code compiling
	s := module.String()
	return strings.Replace(s, "source_filename = \"ben\"\n", "", 1)
}
