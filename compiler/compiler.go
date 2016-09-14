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
func AbstractSyntaxTree(data string) (string, error) {
	tokens := lexer(data)
	functions, err := ast(tokens)
	if err != nil {
		return "", err
	}

	return cmap.SDump(functions, "functions"), nil
}

// Compile produces llvm ir code from the input program
func Compile(data string) (string, error) {
	tokens := lexer(data)
	functions, err := ast(tokens)
	if err != nil {
		return "", err
	}

	llvm := Llvm(functions)

	return llvm, err
}
