package compiler

import "strconv"

const (
	tokenName = iota
	tokenInt32
	tokenReturn
	tokenNumber

	tokenNewLine

	tokenArrow
	tokenAssign
	tokenComma
	tokenDoubleColon

	tokenOpenBody
	tokenOpenBracket
	tokenCloseBody
	tokenCloseBracket

	tokenPlus
	tokenMinus
	tokenMultiply
	tokenDivide

	tokenLessThan
	tokenMoreThan
)

type token struct {
	tokenType int
	value     interface{}
}

// ParseTokens parses the program and returns a sequantial list of tokens
func parseTokens(in string) []token {
	var tokens []token

	buffer := ""
	lineNumber := 0
	lineIndex := 0

	parseBuffer := func(tokens []token, buffer string) []token {
		if buffer != "" {
			if i, err := strconv.Atoi(buffer); err == nil {
				return append(tokens, token{tokenNumber, i})
			} else if buffer == "i32" {
				return append(tokens, token{tokenInt32, nil})
			} else if buffer == "return" {
				return append(tokens, token{tokenReturn, nil})
			} else if buffer == "+" {
				return append(tokens, token{tokenPlus, nil})
			} else if buffer == "-" {
				return append(tokens, token{tokenMinus, nil})
			} else {
				return append(tokens, token{tokenName, buffer})
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
			tokens = append(tokens, token{tokenNewLine, nil})
			buffer = ""
			continue

		case " ":
			tokens = parseBuffer(tokens, buffer)
			buffer = ""
			continue

		case ":":
			if buffer == ":" {
				tokens = append(tokens, token{tokenDoubleColon, nil})
				buffer = ""
			} else {
				tokens = parseBuffer(tokens, buffer)
				buffer = ":"
			}
			continue

		case "=":
			if buffer == ":" {
				tokens = append(tokens, token{tokenAssign, nil})
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
				tokens = append(tokens, token{tokenArrow, nil})
				buffer = ""
			} else {
				tokens = parseBuffer(tokens, buffer)
				tokens = append(tokens, token{tokenLessThan, nil})
				buffer = ""
			}
			continue

		case ",":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, token{tokenComma, nil})
			buffer = ""
			continue

		case "{":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, token{tokenOpenBody, nil})
			buffer = ""
			continue

		case "}":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, token{tokenCloseBody, nil})
			buffer = ""
			continue

		case "(":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, token{tokenOpenBracket, nil})
			buffer = ""
			continue

		case ")":
			tokens = parseBuffer(tokens, buffer)
			tokens = append(tokens, token{tokenCloseBracket, nil})
			buffer = ""
			continue

		}

		// Check if buffer is a token
		switch buffer {
		case "+":
			tokens = append(tokens, token{tokenPlus, nil})
			buffer = ""
		case "-":
			tokens = append(tokens, token{tokenMinus, nil})
			buffer = ""
		}

		buffer += string(char)
		lineIndex++

	}

	return tokens
}
