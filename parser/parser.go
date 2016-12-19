package parser

import (
	"fmt"
	"strconv"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
)

// Parser creates an abstract syntax tree from a sequence of tokens
type Parser struct {
	tokens []lexer.Token
	index  int
}

// NewParser creates a new parser
func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) token() lexer.Token {
	return p.tokens[p.index]
}

func (p *Parser) next() lexer.Token {
	p.index++
	return p.tokens[p.index]
}

func (p *Parser) peek() lexer.Token {
	return p.tokens[p.index+1]
}

func (p *Parser) expect(typ lexer.TokenType) lexer.Token {
	token := p.token()
	if token.Type() != typ {
		panic(fmt.Sprintf("Expected: %s, Got: %s", typ.String(), token.Type().String()))
	}

	p.next()
	return token
}

func (p *Parser) accept(typ lexer.TokenType) (lexer.Token, bool) {
	token := p.token()
	if token.Type() != typ {
		return token, false
	}

	p.next()
	return token, true
}

func (p *Parser) clearNewLines() {
	_, ok := p.accept(lexer.SEMICOLON)
	for ok && p.index != len(p.tokens)-1 {
		_, ok = p.accept(lexer.SEMICOLON)
	}
}

func (p *Parser) ident() string {
	return p.expect(lexer.IDENT).Value()
}

func (p *Parser) integer() ast.Integer {
	value, err := strconv.ParseInt(p.expect(lexer.INT).Value(), 10, 16)
	if err != nil {
		panic(err)
	}

	return ast.Integer{int64(value)}
}

func (p *Parser) float() ast.Float {
	value, err := strconv.ParseFloat(p.expect(lexer.FLOAT).Value(), 64)
	if err != nil {
		panic(err)
	}

	return ast.Float{float64(value)}
}

func (p *Parser) maths() ast.Expression {
	var buffer []lexer.Token
	depth := 0
	for p.token().Type() != lexer.SEMICOLON &&
		p.token().Type() != lexer.LBRACE &&
		p.token().Type() != lexer.RBRACE &&
		!(p.token().Type() == lexer.COMMA && depth == 0) {

		switch p.token().Type() {
		case lexer.LPAREN:
			depth++
		case lexer.RPAREN:
			depth--
		}

		buffer = append(buffer, p.token())
		p.next()
	}

	return p.shuntingYard(buffer)
}

func (p *Parser) Expression() ast.Expression {
	return p.maths()
	// switch p.token().Type() {
	// default:
	// 	return p.maths()
	// 	// case lexer.INT:
	// 	// 	return p.integer()
	// 	// case lexer.FLOAT:
	// 	// 	return p.float()
	// }

	// panic("Unknown expression")
}

func (p *Parser) Parse() ast.Expression {
	return p.Expression()
}
