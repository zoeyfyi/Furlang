package lexer

import (
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	source        []byte
	currentRune   rune
	offset        int
	readingOffset int
	lineOffset    int
	insertSemi    bool
}

// Error describes an error during the lexing
type Error struct {
	Line    int
	Column  int
	Message string
}

// newError creates a new lexer error
func (l *Lexer) newError(message string) *Error {
	return &Error{
		Line:    l.lineOffset,
		Column:  l.offset,
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Message
}

// NewLexer creates a new lexer
func NewLexer(source []byte) *Lexer {
	return &Lexer{
		source: append(source, byte('\n')),
	}
}

// nextRune gets the next rune in source or returns an error if their is a problem with the character
func (l *Lexer) nextRune() error {
	// Check if we are not at the end of file
	if l.readingOffset < len(l.source) {
		// Move the offset forward
		l.offset = l.readingOffset

		// Read the rune
		r, width := rune(l.source[l.readingOffset]), 1

		// Check for Null character
		if r == 0 {
			return l.newError("Illegal NULL character")
		}

		// Check for non-UTF8 character
		if r >= utf8.RuneSelf {
			// Decode rune
			r, width := utf8.DecodeRune(l.source[l.readingOffset:])

			// Check encoding
			if r == utf8.RuneError && width == 1 {
				return l.newError("Illegal UTF-8 encoding")
			}

			// Byte order mark not at start of file
			if r == 0xFEFF && l.offset > 0 {
				return l.newError("Illegal byte order mark")
			}
		}

		// Update lexer
		l.readingOffset += width
		l.currentRune = r

		return nil
	}

	// Update offsets
	l.offset = len(l.source)
	if l.currentRune == '\n' {
		l.lineOffset = l.offset
	}

	// Set end of file rune
	l.currentRune = -1
	return nil
}

// clearWhitespace removes spaces newlines and tabs
func (l *Lexer) clearWhitespace() {
	for l.currentRune == ' ' ||
		l.currentRune == '\t' ||
		l.currentRune == '\n' && !l.insertSemi ||
		l.currentRune == '\r' {
		l.nextRune()
	}
}

// isLetter returns true if the rune is a a-z or A-Z or _ or a unicode letter
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

// IsDigit returns true if the rune is 0-9 or a unicode digit
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

// asDigit returns the current rune as an integer digit (also works for hexedecimal digits)
func asDigit(c rune) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	}
	return 16
}

// mantissa consumes a mantissa
func (l *Lexer) mantissa(base int) {
	for asDigit(l.currentRune) < base {
		l.nextRune()
	}
}

// ident consumes an identifyer and returns its string
func (l *Lexer) ident() string {
	offset := l.offset
	for isLetter(l.currentRune) || isDigit(l.currentRune) {
		l.nextRune()
	}
	return string(l.source[offset:l.offset])
}

// escape consumes an escaped character
func (l *Lexer) escape(quote rune) error {
	// offset := l.offset

	var n int
	var base, max uint32
	switch l.currentRune {
	// Special character escapes
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		l.nextRune()
	// Octal
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	// Hex
	case 'x':
		l.nextRune()
		n, base, max = 2, 16, 255
	// Small unicode
	case 'u':
		l.nextRune()
		n, base, max = 4, 16, unicode.MaxRune
	// Full unicode
	case 'U':
		l.nextRune()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		if l.currentRune < 0 {
			return l.newError("Escape sequence not terminated")
		}

		return l.newError("Unknown escape sequence")
	}

	// Consume additional runes
	var x uint32
	for n > 0 {
		d := uint32(asDigit(l.currentRune))
		if d >= base {
			if l.currentRune < 0 {
				return l.newError(fmt.Sprintf("illegal character %#U in escape sequence", l.currentRune))
			}
			return l.newError("Escape sequence not terminated")
		}

		x = x*base + d
		l.nextRune()
		n--
	}

	// Check if unicode character is valid
	if x > max || 0xD800 <= x && x < 0xE000 {
		return l.newError("escape sequence is invalid Unicode code point")
	}

	return nil
}

// string consumes a string and returns its value
func (l *Lexer) string() (string, error) {
	offset := l.offset - 1

	for {
		// Newline of end of line
		if l.currentRune == '\n' || l.currentRune < 0 {
			return "", l.newError("string literal not terminated")
		}

		l.nextRune()

		// End of string
		if l.currentRune == '"' {
			break
		}

		// Start of escape sequence
		if l.currentRune == '\\' {
			l.escape('"')
		}
	}

	// Return the string value
	return string(l.source[offset:l.offset]), nil
}

func (l *Lexer) number() (TokenType, string, error) {
	offset := l.offset
	tok := INT

	if l.currentRune == '0' {
		// int or float
		offset := l.offset
		l.nextRune()
		if l.currentRune == 'x' || l.currentRune == 'X' {
			// hex
			l.nextRune()
			l.mantissa(16)
			if l.offset-offset <= 2 {
				return ILLEGAL, "", l.newError("Illegal hex number")
			}
		} else {
			// assume octal
			octal := true
			l.mantissa(8)
			if l.currentRune == '8' || l.currentRune == '9' {
				// not an octal
				octal = false
				l.mantissa(10)
			}
			if l.currentRune == '.' {
				goto fraction
			}
			if !octal {
				return ILLEGAL, "", l.newError("Illegal octal number")
			}
		}
		goto exit
	}

	l.mantissa(10)

fraction:
	if l.currentRune == '.' {
		tok = FLOAT
		l.nextRune()
		l.mantissa(10)
	}

exit:
	return tok, string(l.source[offset:l.offset]), nil
}

// switch helper functions deside between 2-4 runes in the case of multi symbol runes

func (l *Lexer) switch2(tok0, tok1 TokenType) TokenType {
	if l.currentRune == '=' {
		l.nextRune()
		return tok1
	}
	return tok0
}

func (l *Lexer) switch3(tok0, tok1 TokenType, ch2 rune, tok2 TokenType) TokenType {
	if l.currentRune == '=' {
		l.nextRune()
		return tok1
	}
	if l.currentRune == ch2 {
		l.nextRune()
		return tok2
	}
	return tok0
}

func (l *Lexer) switch4(tok0, tok1 TokenType, ch2 rune, tok2, tok3 TokenType) TokenType {
	if l.currentRune == '=' {
		l.nextRune()
		return tok1
	}
	if l.currentRune == ch2 {
		l.nextRune()
		if l.currentRune == '=' {
			l.nextRune()
			return tok3
		}
		return tok2
	}
	return tok0
}

// Lex runs the lexer across the source and returns a slice of tokens or an error
func (l *Lexer) Lex() ([]Token, error) {
	log.Println("Starting lex")

	var tokens []Token

	l.nextRune()
	if l.currentRune == 0xFEFF {
		l.nextRune()
	}

	line := 1
	column := 1

	for l.offset < len(l.source) {
		l.clearWhitespace()

		tok := Token{}
		// TODO: fix this, this should be an if
		if line == 1 {
			tok.column = l.offset - column + 2
		} else {
			tok.column = l.offset - column + 1
		}

		currentRune := l.currentRune
		switch {
		case isLetter(currentRune):
			log.Println("Ident")
			tok.value = l.ident()
			tok.typ = Lookup(tok.value)
			switch tok.typ {
			case IDENT, BREAK, CONTINUE, FALLTHROUGH, RETURN:
				l.insertSemi = true
			}
		case isDigit(currentRune):
			log.Println("Number")
			l.insertSemi = true

			typ, value, err := l.number()
			if err != nil {
				return nil, err
			}

			tok.typ = typ
			tok.value = value
		default:
			l.nextRune()
			switch currentRune {
			case -1:
				if l.insertSemi {
					tok.typ = SEMICOLON
					tok.value = "\n"
				} else {
					tok.typ = EOF
				}
				tok.column--
				column = l.offset
			case '\n':
				l.insertSemi = false
				tok.typ = SEMICOLON
				tok.value = "\n"
				column = l.offset
			case '"':
				tok.typ = STRING
				value, err := l.string()
				if err != nil {
					return nil, err
				}
				tok.value = value
				l.insertSemi = true
			case ':':
				tok.typ = l.switch3(COLON, DEFINE, ':', DOUBLE_COLON)
			case '.':
				if l.currentRune == '.' {
					l.nextRune()
					if l.currentRune == '.' {
						l.nextRune()
						tok.typ = ELLIPSIS
					} else {
						tok.typ = PERIOD
					}
				}
			case ',':
				tok.typ = COMMA
			case ';':
				tok.typ = SEMICOLON
			case '(':
				tok.typ = LPAREN
			case ')':
				tok.typ = RPAREN
				l.insertSemi = true
			case '[':
				tok.typ = LBRACK
			case ']':
				tok.typ = RBRACK
				l.insertSemi = true
			case '{':
				tok.typ = LBRACE
				l.insertSemi = false
			case '}':
				tok.typ = RBRACE
				l.insertSemi = true
			case '+':
				tok.typ = l.switch3(ADD, ADD_ASSIGN, '+', INC)
				if tok.typ == INC {
					l.insertSemi = true
				}
			case '-':
				if l.currentRune == '>' {
					// Consume arrow head
					l.nextRune()
					tok.typ = ARROW
				} else {
					tok.typ = l.switch3(SUB, SUB_ASSIGN, '-', DEC)
					if tok.typ == DEC {
						l.insertSemi = true
					}
				}
			case '*':
				tok.typ = l.switch2(MUL, MUL_ASSIGN)
			case '/':
				tok.typ = l.switch2(QUO, QUO_ASSIGN)
			case '%':
				tok.typ = l.switch2(REM, REM_ASSIGN)
			case '^':
				tok.typ = l.switch2(XOR, XOR_ASSIGN)
			case '<':
				tok.typ = l.switch4(LSS, LEQ, '<', SHL, SHL_ASSIGN)
			case '>':
				tok.typ = l.switch4(GTR, GEQ, '>', SHR, SHR_ASSIGN)
			case '=':
				tok.typ = l.switch2(ASSIGN, EQL)
			case '!':
				tok.typ = l.switch2(NOT, NEQ)
			case '&':
				if l.currentRune == '^' {
					l.nextRune()
					tok.typ = l.switch2(AND_NOT, AND_NOT_ASSIGN)
				} else {
					tok.typ = l.switch3(AND, AND_ASSIGN, '&', LAND)
				}
			case '|':
				tok.typ = l.switch3(OR, OR_ASSIGN, '|', LOR)
			default:
				if l.currentRune == 0xFEFF {
					return nil, l.newError(fmt.Sprintf("illegal character %#U", l.currentRune))
				}
				tok.typ = ILLEGAL
			}

		}

		// Append token
		tok.line = line
		tokens = append(tokens, tok)
		if currentRune == -1 || currentRune == '\n' {
			line++
		}
	}

	return tokens, nil
}
