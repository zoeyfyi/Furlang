package compiler

import (
	"fmt"
	"strconv"
)

type token struct {
	tokenType int
	value     interface{}
	line      int
	column    int
	length    int
}

// token type constants
const (
	tokenName = iota
	tokenNumber
	tokenFloat

	tokenType
	tokenInt32
	tokenFloat32
	tokenReturn
	tokenIf
	tokenElse
	tokenTrue
	tokenFalse
	tokenFor
	tokenRange

	tokenArrow
	tokenAssign
	tokenDoubleColon
	tokenIncrement
	tokenDecrement
	tokenDoubleEqual
	tokenNotEqual

	tokenComma
	tokenSemiColon
	tokenNewLine
	tokenOpenBody
	tokenOpenBracket
	tokenCloseBody
	tokenCloseBracket
	tokenPlus
	tokenMinus
	tokenMultiply
	tokenFloatDivide
	tokenIntDivide
	tokenLessThan
	tokenMoreThan
	tokenColon
	tokenEqual
	tokenBang
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
		"/":  tokenFloatDivide,
		"<":  tokenLessThan,
		">":  tokenMoreThan,
		":":  tokenColon,
		";":  tokenSemiColon,
		"=":  tokenEqual,
		"!":  tokenBang,
	}

	typeMap = map[string]int{
		"i32": typeInt32,
		"f32": typeFloat32,
	}

	nameMap = map[string]int{
		"return": tokenReturn,
		"if":     tokenIf,
		"else":   tokenElse,
		"true":   tokenTrue,
		"false":  tokenFalse,
		"for":    tokenFor,
		"range":  tokenRange,
	}

	multiSymbolMap = map[int][]int{
		tokenArrow:       []int{tokenMinus, tokenMoreThan},
		tokenAssign:      []int{tokenColon, tokenEqual},
		tokenDoubleColon: []int{tokenColon, tokenColon},
		tokenIntDivide:   []int{tokenFloatDivide, tokenFloatDivide},
		tokenIncrement:   []int{tokenPlus, tokenPlus},
		tokenDecrement:   []int{tokenMinus, tokenMinus},
		tokenDoubleEqual: []int{tokenEqual, tokenEqual},
		tokenNotEqual:    []int{tokenBang, tokenEqual},
	}
)

// tokenType returns the string representation of tokenType
func tokenTypeString(tt int) string {
	switch tt {
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
	case tokenFloat:
		return "tokenFloat"
	case tokenOpenBody:
		return "tokenOpenBody"
	case tokenPlus:
		return "tokenPlus"
	case tokenMinus:
		return "tokenMinus"
	case tokenMultiply:
		return "tokenMultiply"
	case tokenFloatDivide:
		return "tokenFloatDivide"
	case tokenIntDivide:
		return "tokenIntDivide"
	case tokenOpenBracket:
		return "tokenOpenBracket"
	case tokenCloseBracket:
		return "tokenCloseBracket"
	case tokenReturn:
		return "tokenReturn"
	case tokenIf:
		return "tokenIf"
	case tokenElse:
		return "tokenElse"
	case tokenTrue:
		return "tokenTrue"
	case tokenFalse:
		return "tokenFalse"
	case tokenType:
		return "tokenType"
	case tokenBang:
		return "tokenBang"
	case tokenDoubleEqual:
		return "tokenDoubleEqual"
	case tokenNotEqual:
		return "tokenNotEqual"
	default:
		return "Undefined token"
	}
}

// Converts token to printable string
func (t token) String() string {
	tokenString := tokenTypeString(t.tokenType)
	return fmt.Sprintf("%s, line: %d, column: %d", tokenString, t.line, t.column)
}

// Parsers what ever is in the the buffer
func parseBuffer(buffer *string, tokens *[]token, line int, column int) {
	bufferLength := len([]rune(*buffer))

	if *buffer != "" {
		var ttype int
		var value interface{}

		if i, err := strconv.Atoi(*buffer); err == nil {
			// Buffer contains a number
			ttype = tokenNumber
			value = i
		} else if i, err := strconv.ParseFloat(*buffer, 32); err == nil {
			// Buffer contains a float
			ttype = tokenFloat
			value = float32(i)
		} else if val, found := typeMap[*buffer]; found {
			// Buffer contains a type identifier
			ttype = tokenType
			value = val
		} else if val, found := nameMap[*buffer]; found {
			// Buffer contains a control name
			ttype = val
			value = *buffer
		} else {
			// Buffer contains a name
			ttype = tokenName
			value = *buffer
		}

		*tokens = append(*tokens, token{
			tokenType: ttype,
			value:     value,
			line:      line,
			column:    column - bufferLength + 1,
			length:    bufferLength,
		})

		*buffer = ""
	}

}

// Lexer returns a sequential list of tokens from the input string
func lexer(in string) (tokens []token) {
	buffer := ""

	// Parse all single character tokens, names and numbers
	lineIndex := 1
	columnIndex := 0
characterLoop:
	for _, char := range in {
		columnIndex++

		// Handle whitespace
		if string(char) == " " {
			parseBuffer(&buffer, &tokens, lineIndex, columnIndex-1)
			continue characterLoop
		}

		// Handle symbol character
		for symbol, symbolToken := range symbolMap {
			if string(char) == symbol {
				parseBuffer(&buffer, &tokens, lineIndex, columnIndex-1)
				tokens = append(tokens, token{
					tokenType: symbolToken,
					value:     string(char),
					line:      lineIndex,
					column:    columnIndex,
					length:    1,
				})
				if symbolToken == tokenNewLine {
					lineIndex++
					columnIndex = 0
				}
				continue characterLoop
			}
		}

		// Any other character (number/letter)
		buffer += string(char)
	}

	// Parse anything left in buffer
	parseBuffer(&buffer, &tokens, lineIndex, columnIndex)

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
				lower := append(tokens[:i], token{
					tokenType: symbolsToken,
					value:     nil,
					line:      tokens[i].line,
					column:    tokens[i].column,
					length:    2, //TODO: make this work with varible length multisymbols
				})
				tokens = append(lower, tokens[i+len(symbols):]...)
			}
		}
	}

	return tokens
}
