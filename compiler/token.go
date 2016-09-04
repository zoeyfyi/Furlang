package compiler

import (
	"fmt"
	"strconv"
)

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

var (
	symbolMap = map[string]int{
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

	nameMap = map[string]int{
		"return": tokenReturn,
		"i32":    tokenInt32,
	}

	multiSymbolMap = map[int][]int{
		tokenArrow:       []int{tokenMinus, tokenMoreThan},
		tokenAssign:      []int{tokenColon, tokenEqual},
		tokenDoubleColon: []int{tokenColon, tokenColon},
	}
)

type token struct {
	tokenType int
	value     interface{}
}

func (t token) printToken() string {
	switch t.tokenType {
	case tokenArrow:
		return fmt.Sprintf("Token type: tokenArrow\n")
	case tokenAssign:
		return fmt.Sprintf("Token type: tokenAssign\n")
	case tokenCloseBody:
		return fmt.Sprintf("Token type: tokenCloseBody\n")
	case tokenComma:
		return fmt.Sprintf("Token type: tokenComma\n")
	case tokenDoubleColon:
		return fmt.Sprintf("Token type: tokenDoubleColon\n")
	case tokenInt32:
		return fmt.Sprintf("Token type: tokenInt32\n")
	case tokenName:
		return fmt.Sprintf("Token type: tokenName, Value: %s\n", t.value.(string))
	case tokenNewLine:
		return fmt.Sprintf("Token type: tokenNewLine\n")
	case tokenNumber:
		return fmt.Sprintf("Token type: tokenNumber, Value: %d\n", t.value.(int))
	case tokenOpenBody:
		return fmt.Sprintf("Token type: tokenOpenBody\n")
	case tokenPlus:
		return fmt.Sprintf("Token type: tokenPlus\n")
	case tokenReturn:
		return fmt.Sprintf("Token type: tokenReturn\n")
	}

	return "Undefined token"
}

// ParseTokens parses the program and returns a sequantial list of tokens
func parseTokens(in string) []token {
	var buffer string
	var tokens []token

	// Create list of single character tokens

	parseBuffer := func(buffer *string, tokens *[]token) {
		if i, err := strconv.Atoi(*buffer); err == nil {
			*tokens = append(*tokens, token{tokenNumber, i})
		} else if val, found := nameMap[*buffer]; found {
			*tokens = append(*tokens, token{val, *buffer})
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
