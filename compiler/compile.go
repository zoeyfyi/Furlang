package compiler

import (
	"fmt"

	"bitbucket.com/bongo227/cmap"
)

// Tokens compiles a list of tokens representing the input program
func Tokens(data string) string {
	tokens := parseTokens(data)

	s := ""
	for _, t := range tokens {
		switch t.tokenType {
		case tokenArrow:
			s += fmt.Sprintf("Token type: tokenArrow\n")
		case tokenAssign:
			s += fmt.Sprintf("Token type: tokenAssign\n")
		case tokenCloseBody:
			s += fmt.Sprintf("Token type: tokenCloseBody\n")
		case tokenComma:
			s += fmt.Sprintf("Token type: tokenComma\n")
		case tokenDoubleColon:
			s += fmt.Sprintf("Token type: tokenDoubleColon\n")
		case tokenInt32:
			s += fmt.Sprintf("Token type: tokenInt32\n")
		case tokenName:
			s += fmt.Sprintf("Token type: tokenName, Value: %s\n", t.value.(string))
		case tokenNewLine:
			s += fmt.Sprintf("Token type: tokenNewLine\n")
		case tokenNumber:
			s += fmt.Sprintf("Token type: tokenNumber, Value: %d\n", t.value.(int))
		case tokenOpenBody:
			s += fmt.Sprintf("Token type: tokenOpenBody\n")
		case tokenPlus:
			s += fmt.Sprintf("Token type: tokenPlus\n")
		case tokenReturn:
			s += fmt.Sprintf("Token type: tokenReturn\n")
		}
	}

	return s
}

// AbstractSyntaxTree returns the abstract sytax tree in a pretty printed tree
func AbstractSyntaxTree(data string) string {
	tokens := parseTokens(data)
	functions := ast(tokens)
	cmap.Dump(functions, "functions")

	// TODO: Make this return the real ast
	return ""
}

// Compile produces llvm ir code from the input program
func Compile(data string) string {
	tokens := parseTokens(data)
	functions := ast(tokens)
	llvm := Llvm(functions)

	return llvm
}
