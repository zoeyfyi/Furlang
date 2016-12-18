//go:generate stringer -type=TokenType

package lexer

import "fmt"

// Token is a token
type Token struct {
	typ    TokenType
	value  interface{}
	line   int
	column int
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
	ARROW: "<-",
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
}

func (t Token) String() string {
	if int(t.typ) >= 0 && int(t.typ) < len(tokens) {
		return tokens[t.typ]
	}
	return fmt.Sprintf("token(%d)", t)
}

const (
	LowestPrecedence  = 0
	UnaryPrecedence   = 6
	HighestPrecedence = 7
)

// Precedence returns the operators precedence
func (t Token) Precedence() int {
	switch t.typ {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, EQL, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrecedence
}

var keywords map[string]TokenType

func init() {
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
