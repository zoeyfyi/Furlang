package compiler

import "bitbucket.com/bongo227/cmap"

// Tokens compiles a list of tokens representing the input program
func Tokens(data string) string {
	tokens := lexer(data)

	s := ""
	for _, t := range tokens {
		s += t.string() + "\n"
	}

	return s
}

// AbstractSyntaxTree returns the abstract sytax tree in a pretty printed tree
func AbstractSyntaxTree(data string) string {
	tokens := lexer(data)
	functions := ast(tokens)
	cmap.Dump(functions, "functions")

	// TODO: Make this return the real ast
	return ""
}

// Compile produces llvm ir code from the input program
func Compile(data string) string {
	tokens := lexer(data)
	functions := ast(tokens)
	llvm := Llvm(functions)

	return llvm
}
