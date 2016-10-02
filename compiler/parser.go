package compiler

import (
	"os"

	"github.com/bongo227/cmap"
)

const (
	typeInt32 = iota + 100
	typeFloat32
)

type typedName struct {
	nameType int
	name     string
}

type function struct {
	name    string
	args    []typedName
	returns []typedName
	block   block
}

type block struct {
	expressions []expression
}

type operator struct {
	precendence int
	right       bool
}

type name struct {
	name string
}

type ret struct {
	returns []expression
}

type assignment struct {
	name  string
	value expression
}

type maths struct {
	expression expression
}

type boolean struct {
	value bool
}

type addition struct {
	lhs expression
	rhs expression
}

type subtraction struct {
	lhs expression
	rhs expression
}

type multiplication struct {
	lhs expression
	rhs expression
}

type floatDivision struct {
	lhs expression
	rhs expression
}

type intDivision struct {
	lhs expression
	rhs expression
}

type number struct {
	value int
}

type float struct {
	value float32
}

type call struct {
	function string
	args     []expression
}

type ifExpression struct {
	blocks []ifBlock
}

type ifBlock struct {
	condition expression
	block     block
}

type SyntaxTree struct {
	functions []function
}

func (s *SyntaxTree) Print() {
	cmap.Dump(*s, "Ast")
}

func (s *SyntaxTree) Write(f *os.File) {
	f.WriteString(cmap.SDump(*s, "Ast"))
}

type Parser struct {
	tokens            []Token
	currentTokenIndex int
}

func (p *Parser) currentToken() Token {
	return p.tokens[p.currentTokenIndex]
}

func (p *Parser) previousToken() Token {
	if p.currentTokenIndex == 0 {
		panic("At first token (no previous token)")
	}

	return p.tokens[p.currentTokenIndex-1]
}

func (p *Parser) clearNewLines() {
	for p.currentToken().tokenType == tokenNewLine {
		p.nextToken()
	}
}

func (p *Parser) nextToken() Token {
	if p.currentTokenIndex < len(p.tokens) {
		p.currentTokenIndex++
		return p.currentToken()
	}

	panic("Ran out of tokens")
}

func (p *Parser) expect(tokenType int) Token {
	if p.currentToken().tokenType == tokenType {
		prev := p.currentToken()
		p.nextToken()
		return prev
	}

	panic("Unexpected: " + p.currentToken().String())
}

// typed name in the format: type name, where name is optional
func (p *Parser) typedName() typedName {
	ftype := p.expect(tokenType).value.(int)

	// type has a name
	if p.currentToken().tokenType == tokenName {
		return typedName{
			name:     p.currentToken().value.(string),
			nameType: ftype,
		}
	}

	// type with no name
	return typedName{
		nameType: ftype,
	}
}

// List in to format of type name, type name ...
func (p *Parser) typeList() []typedName {
	var names []typedName

	// If its an empty list return nil
	if p.currentToken().tokenType != tokenType {
		return nil
	}

	// While their is a comma continue
	ok := true
	for ok {
		names = append(names, p.typedName())
		ok = p.currentToken().tokenType == tokenComma
	}

	return names
}

func (p *Parser) block() block {
	block := block{}
	p.expect(tokenOpenBody)
	p.clearNewLines()

	for p.currentToken().tokenType != tokenCloseBody {
		block.expressions = append(block.expressions, p.expression())
	}

	p.expect(tokenCloseBody)
	return block
}

func (p *Parser) function() function {
	name := p.previousToken().value.(string)
	p.expect(tokenDoubleColon)
	args := p.typeList()
	p.expect(tokenArrow)
	returns := p.typeList()
	block := p.block()

	return function{name, args, returns, block}
}

func (p *Parser) ret() ret {
	p.expect(tokenReturn)

	var returns []expression

	// While their is a comma continue
	ok := true
	for ok {
		returns = append(returns, p.expression())
		ok = p.currentToken().tokenType == tokenComma
	}

	return ret{returns}
}

func (p *Parser) interger() number {
	value := p.expect(tokenNumber).value.(int)
	p.clearNewLines()
	return number{value}
}

func (p *Parser) expression() expression {
	switch p.currentToken().tokenType {
	case tokenDoubleColon:
		return expression(p.function())
	case tokenReturn:
		return expression(p.ret())
	case tokenNumber:
		return expression(p.interger())
	default:
		p.nextToken()
	}

	p.clearNewLines()
	return nil
}

// NewParser creates a new parser
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:            tokens,
		currentTokenIndex: 0,
	}
}

// Parse parses the file and returns the syntax tree
func (p *Parser) Parse() SyntaxTree {
	p.nextToken()
	return SyntaxTree{
		functions: []function{p.expression().(function)},
	}
}
