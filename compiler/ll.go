// +build llvm

package compiler

import "fmt"

type llvmFunction struct {
	functions map[string]Value
	names     map[string]Value
	builder   Builder
	tempCount *int
}

func (lf llvmFunction) nextTempName() string {
	*(lf.tempCount)++
	return fmt.Sprintf("tmp%d", *(lf.tempCount))
}
