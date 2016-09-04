package compiler

import "strconv"

const (
	tokenName = iota
	tokenNumber

	tokenInt32
	tokenReturn

	tokenArrow
	tokenAssign
	tokenDoubleColon

	tokenComma
	tokenNewLine
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
	tokenColon
	tokenEqual
)

type token struct {
	tokenType int
	value     interface{}
}

// ParseTokens parses the program and returns a sequantial list of tokens
func parseTokens(in string) []token {
	multiSymbolMap := map[int][]int{
		tokenArrow:       []int{tokenMinus, tokenMoreThan},
		tokenAssign:      []int{tokenColon, tokenEqual},
		tokenDoubleColon: []int{tokenColon, tokenColon},
	}

	symbolMap := map[string]int{
		"\n": tokenNewLine,
		",":  tokenComma,
		"{":  tokenOpenBody,
		"}":  tokenCloseBody,
		"(":  tokenOpenBracket,
		")":  tokenCloseBracket,
		"+":  tokenPlus,
		"-":  tokenMinus,
		"*":  tokenMultiply,
		"/":  tokenDivide,
		"<":  tokenLessThan,
		">":  tokenMoreThan,
		":":  tokenColon,
		"=":  tokenEqual,
	}

	var buffer string
	var tokens []token

	// Create list of single character tokens

	parseBuffer := func(buffer *string, tokens *[]token) {
		if i, err := strconv.Atoi(*buffer); err == nil {
			*tokens = append(*tokens, token{tokenNumber, i})
		} else if *buffer == "i32" {
			*tokens = append(*tokens, token{tokenInt32, *buffer})
		} else if *buffer == "return" {
			*tokens = append(*tokens, token{tokenReturn, *buffer})
		} else {
			*tokens = append(*tokens, token{tokenName, *buffer})
		}

		*buffer = ""
	}

characterLoop:
	for _, char := range in {

		if string(char) == " " && buffer != "" {
			parseBuffer(&buffer, &tokens)
			continue characterLoop
		} else if string(char) == " " {
			continue characterLoop
		}

		for symbol, symbolToken := range symbolMap {
			if string(char) == symbol {
				if buffer != "" {
					parseBuffer(&buffer, &tokens)
				}

				tokens = append(tokens, token{symbolToken, string(char)})
				continue characterLoop
			}
		}

		buffer += string(char)
	}

	for i := 0; i < len(tokens); i++ {
		for msk, msv := range multiSymbolMap {
			if tokens[i].tokenType == msv[0] {
				equal := true
				for offset, val := range msv {
					if tokens[i+offset].tokenType != val {
						equal = false
					}
				}

				if equal {
					lower := append(tokens[:i], token{msk, nil})
					tokens = append(lower, tokens[i+len(msv):]...)
				}
			}
		}
	}

	return tokens
}
