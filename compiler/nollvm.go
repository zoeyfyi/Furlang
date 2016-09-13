// +build nollvm

package compiler

type value interface{}
type builder interface{}

type expression interface{}

// Llvm returns a empty program because it wasnt compiled with llvm support
func Llvm(funcs []function) string {
	return "Compiler build with no LLVM"
}
