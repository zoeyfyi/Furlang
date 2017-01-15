//go:generate stringer -type=TokenType

package lexer

import "fmt"

// Token is a token
type Token struct {
	typ    TokenType
	value  string
	line   int
	column int
}

// NewToken is a convenice method for making new tokens for testing etc.
func NewToken(typ TokenType, value string, line, column int) Token {
	return Token{typ, value, line, column}
}

// TokenType is the type of a token
type TokenType int

// TokenType constants
const (
	// Special
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	literals_begin
	IDENT
	INT
	FLOAT
	CHAR
	STRING
	literals_end

	// Operators
	operators_begin
	ADD
	SUB
	MUL
	QUO
	REM
	AND
	OR
	XOR
	SHL
	SHR
	AND_NOT
	ADD_ASSIGN
	SUB_ASSIGN
	MUL_ASSIGN
	QUO_ASSIGN
	REM_ASSIGN
	AND_ASSIGN
	OR_ASSIGN
	XOR_ASSIGN
	SHL_ASSIGN
	SHR_ASSIGN
	AND_NOT_ASSIGN
	LAND
	LOR
	ARROW
	INC
	DEC
	EQL
	LSS
	GTR
	ASSIGN
	NOT
	NEQ
	LEQ
	GEQ
	DEFINE
	ELLIPSIS
	LPAREN
	LBRACK
	LBRACE
	COMMA
	PERIOD
	RPAREN
	RBRACK
	RBRACE
	SEMICOLON
	COLON
	DOUBLE_COLON
	operators_end

	// Keywords
	keywords_begin
	BREAK
	CASE
	CONST
	CONTINUE
	DEFAULT
	DEFER
	ELSE
	FALLTHROUGH
	FOR
	FUNC
	PROC
	IF
	IMPORT
	RETURN
	SELECT
	STRUCT
	SWITCH
	TYPE
	VAR
	// I8TYPE
	// I16TYPE
	// I32TYPE
	// I64TYPE
	// INTTYPE
	// F32TYPE
	// F64TYPE
	// FLOATTYPE
	keywords_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	CHAR:   "CHAR",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND:     "&",
	OR:      "|",
	XOR:     "^",
	SHL:     "<<",
	SHR:     ">>",
	AND_NOT: "&^",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	AND_ASSIGN:     "&=",
	OR_ASSIGN:      "|=",
	XOR_ASSIGN:     "^=",
	SHL_ASSIGN:     "<<=",
	SHR_ASSIGN:     ">>=",
	AND_NOT_ASSIGN: "&^=",

	LAND:  "&&",
	LOR:   "||",
	ARROW: "->",
	INC:   "++",
	DEC:   "--",

	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	NEQ:      "!=",
	LEQ:      "<=",
	GEQ:      ">=",
	DEFINE:   ":=",
	ELLIPSIS: "...",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:       ")",
	RBRACK:       "]",
	RBRACE:       "}",
	SEMICOLON:    ";",
	COLON:        ":",
	DOUBLE_COLON: "::",

	BREAK:    "break",
	CASE:     "case",
	CONST:    "const",
	CONTINUE: "continue",

	DEFAULT:     "default",
	DEFER:       "defer",
	ELSE:        "else",
	FALLTHROUGH: "fallthrough",
	FOR:         "for",

	FUNC:   "func",
	PROC:   "proc",
	IF:     "if",
	IMPORT: "import",

	RETURN: "return",

	SELECT: "select",
	STRUCT: "struct",
	SWITCH: "switch",
	TYPE:   "type",
	VAR:    "var",

	// I8TYPE:    "i8",
	// I16TYPE:   "i16",
	// I32TYPE:   "i32",
	// I64TYPE:   "i64",
	// INTTYPE:   "int",
	// F32TYPE:   "f32",
	// F64TYPE:   "f64",
	// FLOATTYPE: "float",
}

func (t TokenType) String() string {
	if int(t) >= 0 && int(t) < len(tokens) {
		return tokens[t]
	}
	return fmt.Sprintf("token(%d)", t)
}

func (t Token) String() string {
	return fmt.Sprintf("Type: %s, Line: %d, Column: %d, Value: %q", t.typ, t.line, t.column, t.value)
}

const (
	// LowestPrecedence is the lowest precedence
	LowestPrecedence = 0
	// UnaryPrecedence is the precedence of unary operators
	UnaryPrecedence = 6
	// HighestPrecedence is the highest precedence
	HighestPrecedence = 7
)

// Precedence returns the operators precedence
func (t Token) Precedence() int {
	switch t.typ {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrecedence
}

// Line returns the line number of the token
func (t Token) Line() int {
	return t.line
}

// Column returns the column number of the token
func (t Token) Column() int {
	return t.column
}

// Type returns the type of the token
func (t Token) Type() TokenType {
	return t.typ
}

// Value returns the value of the token
func (t Token) Value() string {
	return t.value
}

// Map of keywords to their respective types
var keywords map[string]TokenType

func init() {
	// Initialize keyword map
	keywords = make(map[string]TokenType)
	for i := keywords_begin + 1; i < keywords_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup returns the keyword token if the string is a keyword
func Lookup(ident string) TokenType {
	if t, ok := keywords[ident]; ok {
		return t
	}

	return IDENT
}

// IsLiteral returns true if token is a literal
func (t Token) IsLiteral() bool {
	return literals_begin < t.typ && t.typ < literals_end
}

// IsOperator returns true if token is a operator
func (t Token) IsOperator() bool {
	return operators_begin < t.typ && t.typ < operators_end
}

// IsKeyword returns true if token is a keyword
func (t Token) IsKeyword() bool {
	return keywords_begin < t.typ && t.typ < keywords_end
}
