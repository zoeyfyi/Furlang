package compiler

import (
	"fmt"
	"strconv"
)

// TokenName are tokens representing functions and varibles
type TokenName struct {
	name string
}

func (t TokenName) String() string {
	return fmt.Sprintf("token type: TokenName\t token.name: '%s'\n", t.name)
}

// TokenNumber are tokens representing numbers
type TokenNumber struct {
	number int
}

func (t TokenNumber) String() string {
	return fmt.Sprintf("token type: TokenNumber\t token.number: '%d'\n", t.number)
}

// TokenType are tokens representing a type
type TokenType struct {
	typeName string
}

func (t TokenType) String() string {
	return fmt.Sprintf("token type: TokenType\t token.typeName: '%s'\n", t.typeName)
}

// TokenSymbol are tokens representing a symbol such as :: , . \n etc
type TokenSymbol struct {
	symbol string
}

func (t TokenSymbol) String() string {
	if t.symbol == "\n" {
		t.symbol = "\\n"
	}
	return fmt.Sprintf("token type: TokenSymbol\t token.symbol: '%s'\n", t.symbol)
}

type TokenReturn struct{}

func (t TokenReturn) String() string {
	return fmt.Sprintf("token type: TokenReturn\n")
}

// Token represents one or more characters in a program
type Token interface {
	String() string
}

// ParseTokens parses the program and returns a sequantial list of tokens
func ParseTokens(in string) []Token {
	var tokens []Token

	buffer := ""
	lineNumber := 0
	lineIndex := 0

	parseBuffer := func(tokens []Token, buffer string) []Token {
		if buffer != "" {
			if i, err := strconv.Atoi(buffer); err == nil {
				return append(tokens, TokenNumber{i})
			} else if buffer == "i32" {
				return append(tokens, TokenType{buffer})
			} else if buffer == "return" {
				return append(tokens, TokenReturn{})
			} else if buffer == "+" {
				return append(tokens, TokenSymbol{"+"})
			} else {
				return append(tokens, TokenName{buffer})
			}
		}

		return tokens

	}

	for _, char := range in {

		// Handle special case characters
		switch string(char) {

		case "\n":
			lineNumber++
			lineIndex = 0
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, TokenSymbol{"\n"})
			buffer = ""
			continue

		case " ":
			tokens = parseBuffer(tokens, buffer)
			buffer = ""
			continue

		case ":":
			if buffer == ":" {
				tokens = append(tokens, TokenSymbol{"::"})
				buffer = ""
			} else {
				tokens = parseBuffer(tokens, buffer)
				buffer = ":"
			}
			continue

		case "=":
			if buffer == ":" {
				tokens = append(tokens, TokenSymbol{":="})
				buffer = ""
			} else {
				tokens = parseBuffer(tokens, buffer)
				buffer = "="
			}
			continue

		case "-":
			tokens = parseBuffer(tokens, buffer)
			buffer = "-"
			continue

		case "+":
			tokens = parseBuffer(tokens, buffer)
			buffer = "+"
			continue

		case ">":
			if buffer == "-" {
				tokens = append(tokens, TokenSymbol{"->"})
				buffer = ""
			} else {
				tokens = parseBuffer(tokens, buffer)
				tokens = append(tokens, TokenSymbol{">"})
				buffer = ""
			}
			continue

		case ",":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, TokenSymbol{","})
			buffer = ""
			continue

		case "{":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, TokenSymbol{"{"})
			buffer = ""
			continue

		case "}":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, TokenSymbol{"}"})
			buffer = ""
			continue

		case "(":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, TokenSymbol{"("})
			buffer = ""
			continue

		case ")":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, TokenSymbol{")"})
			buffer = ""
			continue

		}

		// Check if buffer is a token
		switch buffer {
		case "+":
			tokens = append(tokens, TokenSymbol{"+"})
			buffer = ""
		}

		buffer += string(char)
		lineIndex++

	}

	return tokens
}
