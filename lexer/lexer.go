package lexer

import (
	"fmt"
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
	offset := l.offset

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

func (l *Lexer) Lex(source []byte) (tokens []Token) {
	l.source = source

	l.nextRune()
	if l.currentRune == 0xFEFF {
		l.nextRune()
	}

	l.clearWhitespace()
	line := 0
	column := 0

	insertSemi := false
	switch {
	case isLetter(l.currentRune):
		ident := l.ident()
		tok := Lookup(ident)
		switch tok {
		case IDENT, BREAK, CONTINUE, FALLTHROUGH, RETURN:
			insertSemi = true
		}
	case isDigit(l.currentRune):
		insertSemi = true
		tok, lit := l.number()
	default:
		l.nextRune()
		switch l.currentRune {
		case -1:
			if l.insertSemi {
				l.insertSemi = false
				tokens = append(tokens, Token{
					typ:    SEMICOLON,
					line:   line,
					column: column,
					value:  "\n",
				})
			}
		case '\n':
			tokens = append(tokens, Token{
				typ:    SEMICOLON,
				line:   line,
				column: column,
				value:  "\n",
			})
		case '"':
			insertSemi = true
			tokens = append(tokens, Token{
				typ:    STRING,
				line:   line,
				column: column,
				value:  l.string(),
			})
		case ':':
			tokens = append(tokens, Token{
				typ:    l.switch3(COLON, DEFINE, ':', DOUBLE_COLON),
				line:   line,
				column: column,
			})
		case '.':
			if l.currentRune == '.' {
				l.nextRune()
				if l.currentRune == '.' {
					l.nextRune()
					tokens = append(tokens, Token{
						typ:    ELLIPSIS,
						line:   line,
						column: column,
					})
				} else {
					tokens = append(tokens, Token{
						typ:    PERIOD,
						line:   line,
						column: column,
					})
				}
			}
		case ',':
			tokens = append(tokens, Token{
				typ:    COMMA,
				line:   line,
				column: column,
			})
		case ';':
			tokens = append(tokens, Token{
				typ:    SEMICOLON,
				line:   line,
				column: column,
			})
		case '(':
			tokens = append(tokens, Token{
				typ:    LPAREN,
				line:   line,
				column: column,
			})
		case ')':
			tokens = append(tokens, Token{
				typ:    RPAREN,
				line:   line,
				column: column,
			})
		case '[':
			tokens = append(tokens, Token{
				typ:    LBRACK,
				line:   line,
				column: column,
			})
		case ']':
			tokens = append(tokens, Token{
				typ:    RBRACK,
				line:   line,
				column: column,
			})
		case '{':
			tokens = append(tokens, Token{
				typ:    LBRACE,
				line:   line,
				column: column,
			})
		case '}':
			tokens = append(tokens, Token{
				typ:    RBRACE,
				line:   line,
				column: column,
			})
		}
	}

	return tokens
}
