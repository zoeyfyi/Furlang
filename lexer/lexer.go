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

func NewLexer(source []byte) *Lexer {
	return &Lexer{
		source: append(source, byte('\n')),
	}
}

func (l *Lexer) nextRune() {
	if l.readingOffset < len(l.source) {
		l.offset = l.readingOffset
		if l.currentRune == '\n' {
			l.lineOffset = l.offset
		}
		r, w := rune(l.source[l.readingOffset]), 1
		switch {
		case r == 0:
			panic("illegal NULL character")
		case r >= utf8.RuneSelf:
			r, w := utf8.DecodeRune(l.source[l.readingOffset:])
			if r == utf8.RuneError && w == 1 {
				panic("illegal UTF-8 encoding")
			} else if r == 0xFEFF && l.offset > 0 {
				panic("illegal byte order mark")
			}
		}
		l.readingOffset += w
		l.currentRune = r
	} else {
		l.offset = len(l.source)
		if l.currentRune == '\n' {
			l.lineOffset = l.offset
		}
		// Reached end of file
		l.currentRune = -1
	}
}

func (l *Lexer) clearWhitespace() {
	for l.currentRune == ' ' ||
		l.currentRune == '\t' ||
		l.currentRune == '\n' && !l.insertSemi ||
		l.currentRune == '\r' {
		l.nextRune()
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

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
func (l *Lexer) escape(quote rune) {
	// offset := l.offset

	var n int
	var base, max uint32
	switch l.currentRune {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		l.nextRune()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		l.nextRune()
		n, base, max = 2, 16, 255
	case 'u':
		l.nextRune()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		l.nextRune()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		if l.currentRune < 0 {
			panic("Escape sequence not terminated")
		}
		panic("Unkown exacape sequence")
	}

	var x uint32
	for n > 0 {
		d := uint32(asDigit(l.currentRune))
		if d >= base {
			if l.currentRune < 0 {
				panic(fmt.Sprintf("illegal character %#U in escape sequence", l.currentRune))
			}
			panic("Escape sequence not terminated")
		}
		x = x*base + d
		l.nextRune()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		panic("escape sequence is invalid Unicode code point")
	}
}

// string consumes a string and returns its value
func (l *Lexer) string() string {
	offset := l.offset - 1

	for {
		if l.currentRune == '\n' || l.currentRune < 0 {
			panic("string literal not terminated")
		}
		l.nextRune()
		if l.currentRune == '"' {
			break
		}
		if l.currentRune == '\\' {
			l.escape('"')
		}
	}

	return string(l.source[offset:l.offset])
}

func (l *Lexer) number() (TokenType, string) {
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
				panic("Illegal hex number")
			}
		} else {
			// octal
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
				panic("Illegal octal number")
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
	return tok, string(l.source[offset:l.offset])
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

func (l *Lexer) Lex() (tokens []Token) {
	log.Println("Starting lex")

	l.nextRune()
	if l.currentRune == 0xFEFF {
		l.nextRune()
	}

	line := 1
	column := 1

	for l.offset < len(l.source) {
		l.clearWhitespace()

		tok := Token{}
		tok.column = l.offset + 1*line - column + 1

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
			tok.typ, tok.value = l.number()
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
				// tok.column--
				column = l.offset
			case '"':
				tok.typ = STRING
				tok.value = l.string()
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
					panic(fmt.Sprintf("illegal character %#U", l.currentRune))
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

	//tokens = append(tokens, Token{typ: EOF})
	return tokens
}
