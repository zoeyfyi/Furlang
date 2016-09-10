package compiler

import "strconv"

type token struct {
	tokenType int
	value     interface{}
}

// Token type constants
const (
	tokenName = iota
	tokenNumber

	tokenInt32
	tokenFloat32
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

// Lexer maps
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
		"f32":    tokenFloat32,
	}

	multiSymbolMap = map[int][]int{
		tokenArrow:       []int{tokenMinus, tokenMoreThan},
		tokenAssign:      []int{tokenColon, tokenEqual},
		tokenDoubleColon: []int{tokenColon, tokenColon},
	}
)

// Converts token to printable string
func (t token) string() string {
	switch t.tokenType {
	case tokenArrow:
		return "tokenArrow"
	case tokenAssign:
		return "tokenAssign"
	case tokenCloseBody:
		return "tokenCloseBody"
	case tokenComma:
		return "tokenComma"
	case tokenDoubleColon:
		return "tokenDoubleColon"
	case tokenInt32:
		return "tokenInt32"
	case tokenFloat32:
		return "tokenFloat32"
	case tokenName:
		return "tokenName"
	case tokenNewLine:
		return "tokenNewLine"
	case tokenNumber:
		return "tokenNumber"
	case tokenOpenBody:
		return "tokenOpenBody"
	case tokenPlus:
		return "tokenPlus"
	case tokenMinus:
		return "tokenMinus"
	case tokenMultiply:
		return "tokenMultiply"
	case tokenDivide:
		return "tokenDivide"
	case tokenOpenBracket:
		return "tokenOpenBracket"
	case tokenCloseBracket:
		return "tokenCloseBracket"
	case tokenReturn:
		return "tokenReturn"
	default:
		return "Undefined token"
	}
}

// Parsers what ever is in the the buffer
func parseBuffer(buffer *string, tokens *[]token) {

	if *buffer != "" {
		if i, err := strconv.Atoi(*buffer); err == nil {
			// Buffer contains a number
			*tokens = append(*tokens, token{tokenNumber, i})
		} else if val, found := nameMap[*buffer]; found {
			// Buffer contains a control name
			*tokens = append(*tokens, token{val, *buffer})
		} else {
			// Buffer contains a name
			*tokens = append(*tokens, token{tokenName, *buffer})
		}

		*buffer = ""
	}

}

// Returns a sequential list of tokens from the input string
func lexer(in string) (tokens []token) {
	buffer := ""

	// Parse all single character tokens, names and numbers
characterLoop:
	for _, char := range in {
		// Handle whitespace
		if string(char) == " " {
			parseBuffer(&buffer, &tokens)
			continue characterLoop
		}

		// Handle symbol character
		for symbol, symbolToken := range symbolMap {
			if string(char) == symbol {
				parseBuffer(&buffer, &tokens)
				tokens = append(tokens, token{symbolToken, string(char)})
				continue characterLoop
			}
		}

		// Any other character (number/letter)
		buffer += string(char)
	}

	// Group single character tokens
	for i := 0; i < len(tokens); i++ {
		for symbolsToken, symbols := range multiSymbolMap {
			// Check if tokens can be grouped
			equal := true
			for offset, val := range symbols {
				if len(tokens) > i+offset && tokens[i+offset].tokenType != val {
					equal = false
				}
			}

			// Collapse tokens in group into a single token
			if equal {
				lower := append(tokens[:i], token{symbolsToken, nil})
				tokens = append(lower, tokens[i+len(symbols):]...)
			}
		}
	}

	return tokens
}
