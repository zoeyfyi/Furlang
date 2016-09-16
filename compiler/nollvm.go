// +build nollvm

package compiler

import "fmt"

type llvmFunction struct {
	functions map[string]interface{}
	names     map[string]interface{}
	builder   interface{}
	tempCount *int
}

func (lf llvmFunction) nextTempName() string {
	*(lf.tempCount)++
	return fmt.Sprintf("tmp%d", *(lf.tempCount))
}

type expression interface{}

func Llvm(ast *abstractSyntaxTree) string {
	return "Compiled without llvm"
}
