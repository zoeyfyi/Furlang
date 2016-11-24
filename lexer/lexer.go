package lexer

import "strconv"

// Lexer maps
var (
	symbolMap = map[string]TokenType{
		"\n": NEWLINE,
		",":  COMMA,
		"{":  OPENBODY,
		"}":  CLOSEBODY,
		"(":  OPENBRACKET,
		")":  CLOSEBRACKET,
		"[":  OPENSQUAREBRACKET,
		"]":  CLOSESQUAREBRACKET,
		"+":  PLUS,
		"-":  MINUS,
		"*":  MULTIPLY,
		"/":  FLOATDIVIDE,
		"<":  LESSTHAN,
		">":  MORETHAN,
		":":  COLON,
		";":  SEMICOLON,
		"=":  ASSIGN,
		"!":  BANG,
		"%":  MOD,
	}

	typeMap = map[string]TokenType{
		// Integers
		"int": INT,
		"i8":  INT8,
		"i16": INT16,
		"i32": INT32,
		"i64": INT64,
		// Floats
		"float": FLOAT,
		"f32":   FLOAT32,
		"f64":   FLOAT64,
	}

	nameMap = map[string]TokenType{
		"return": RETURN,
		"if":     IF,
		"else":   ELSE,
		"true":   TRUE,
		"false":  FALSE,
		"for":    FOR,
		"range":  RANGE,
	}

	multiSymbolMap = map[TokenType][]TokenType{
		ARROW:          []TokenType{MINUS, MORETHAN},
		INFERASSIGN:    []TokenType{COLON, ASSIGN},
		DOUBLECOLON:    []TokenType{COLON, COLON},
		INTDIVIDE:      []TokenType{FLOATDIVIDE, FLOATDIVIDE},
		INCREMENT:      []TokenType{PLUS, PLUS},
		DECREMENT:      []TokenType{MINUS, MINUS},
		EQUAL:          []TokenType{ASSIGN, ASSIGN},
		NOTEQUAL:       []TokenType{BANG, ASSIGN},
		INCREMENTEQUAL: []TokenType{PLUS, ASSIGN},
		DECREMENTEQUAL: []TokenType{MINUS, ASSIGN},
	}
)

// Lexer struct holds the lexers internal state
type Lexer struct {
	input string
}

// NewLexer creates a new lexer for the input string
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
	}
}

// Parsers what ever is in the the buffer
func parseBuffer(buffer *string, tokens *[]Token, line int, column int) {
	bufferLength := len([]rune(*buffer))

	if *buffer != "" {
		var ttype TokenType
		var value interface{}

		if i, err := strconv.Atoi(*buffer); err == nil {
			// Buffer contains a number
			ttype = INTVALUE
			value = i
		} else if i, err := strconv.ParseFloat(*buffer, 32); err == nil {
			// Buffer contains a float
			ttype = FLOATVALUE
			value = float32(i)
		} else if val, found := typeMap[*buffer]; found {
			// Buffer contains a type identifier
			ttype = TYPE
			value = val
		} else if val, found := nameMap[*buffer]; found {
			// Buffer contains a control name
			ttype = val
			value = *buffer
		} else {
			// Buffer contains a name
			ttype = IDENT
			value = *buffer
		}

		*tokens = append(*tokens, Token{
			Type: ttype,
			Pos: Position{
				Line:   line,
				Column: column - bufferLength + 1,
				Width:  bufferLength,
			},
			Value: value,
		})

		*buffer = ""
	}
}

// Lex returns a sequential list of tokens
func (l *Lexer) Lex() (tokens []Token) {
	buffer := ""

	// Parse all single character tokens, names and numbers
	lineIndex := 1
	columnIndex := 0
characterLoop:
	for _, char := range l.input {
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
				tokens = append(tokens, Token{
					Type:  symbolToken,
					Value: string(char),
					Pos: Position{
						Line:   lineIndex,
						Column: columnIndex,
						Width:  1,
					},
				})
				if symbolToken == NEWLINE {
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
				if len(tokens) > i+offset && tokens[i+offset].Type != val {
					equal = false
				}
			}

			// Collapse tokens in group into a single token
			if equal {
				lower := append(tokens[:i], Token{
					Type:  symbolsToken,
					Value: nil,
					Pos: Position{
						Line:   tokens[i].Pos.Line,
						Column: tokens[i].Pos.Column,
						Width:  2,
					},
				})
				tokens = append(lower, tokens[i+len(symbols):]...)
			}
		}
	}

	return tokens
}
