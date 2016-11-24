package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/dprint"
	"github.com/davecgh/go-spew/spew"
)

const enableLogging = true

type typedName struct {
	nameType lexer.TokenType
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

type cast struct {
	cast  lexer.TokenType
	value expression
}

type ret struct {
	returns []expression
}

type assignment struct {
	name     string
	nameType lexer.TokenType
	value    expression
}

type increment struct {
	name string
}

type decrement struct {
	name string
}

type incrementExpression struct {
	name   string
	amount expression
}

type decrementExpression struct {
	name   string
	amount expression
}

type boolean struct {
	value bool
}

type binaryOperator struct {
	lhs expression
	rhs expression
}

type addition binaryOperator
type subtraction binaryOperator
type multiplication binaryOperator
type floatDivision binaryOperator
type intDivision binaryOperator
type lessThan binaryOperator
type moreThan binaryOperator
type equal binaryOperator
type notEqual binaryOperator
type mod binaryOperator

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

type forExpression struct {
	index     expression
	condition expression
	increment expression
	block     block
}

type array struct {
	baseType lexer.TokenType
	length   int
	values   []expression
}

type arrayValue struct {
	name  string
	index expression
}

type syntaxTree struct {
	functions []function
}

func (s *syntaxTree) print() {
	dprint.Tree(*s)
}

func (s *syntaxTree) Write(f *os.File) {
	f.WriteString(dprint.STree(*s))
}

var (
	// Maps tokens onto operators
	opMap = map[lexer.TokenType]operator{
		lexer.NOTEQUAL:    operator{3, false},
		lexer.EQUAL:       operator{3, false},
		lexer.MORETHAN:    operator{3, false},
		lexer.LESSTHAN:    operator{3, false},
		lexer.PLUS:        operator{4, false},
		lexer.MINUS:       operator{4, false},
		lexer.MULTIPLY:    operator{5, false},
		lexer.FLOATDIVIDE: operator{5, false},
		lexer.INTDIVIDE:   operator{5, false},
		lexer.MOD:         operator{5, false},
	}
)

type parser struct {
	tokens            []lexer.Token
	currentTokenIndex int
	depth             int
}

func (p *parser) log(statement string, start bool) {
	if enableLogging {
		if !start {
			p.depth--
		}

		fmt.Println(strings.Repeat(" ", p.depth*2), statement)

		if start {
			p.depth++
		}
	}
}

func (p *parser) currentToken() lexer.Token {
	return p.tokens[p.currentTokenIndex]
}

func (p *parser) previousToken() lexer.Token {
	if p.currentTokenIndex == 0 {
		panic("At first token (no previous token)")
	}

	return p.tokens[p.currentTokenIndex-1]
}

func (p *parser) clearNewLines() {
	ok := p.accept(lexer.NEWLINE)
	for ok && p.currentTokenIndex != len(p.tokens)-1 {
		ok = p.accept(lexer.NEWLINE)
	}
}

func (p *parser) nextToken() lexer.Token {
	if p.currentTokenIndex < len(p.tokens) {
		p.currentTokenIndex++
		return p.currentToken()
	}

	panic("Ran out of tokens")
}

func (p *parser) peekNextToken() lexer.Token {
	return p.tokens[p.currentTokenIndex+1]
}

func (p *parser) expect(tokenType lexer.TokenType) lexer.Token {
	if p.currentToken().Type == tokenType {
		prev := p.currentToken()

		if p.currentTokenIndex < len(p.tokens)-1 {
			p.nextToken()
		}

		return prev
	}

	panic("Unexpected: " + p.currentToken().String() + "; Expected: " + tokenType.String())
}

func (p *parser) accept(tokenType lexer.TokenType) bool {
	if p.currentToken().Type == tokenType && p.currentTokenIndex == len(p.tokens)-1 {
		return true
	}

	if p.currentToken().Type == tokenType {
		p.nextToken()
		return true
	}

	return false
}

// typed name in the format: type name, where name is optional
func (p *parser) typedName() typedName {
	p.log("Start TypedName", true)
	defer p.log("End TypedName", false)

	ftype := p.expect(lexer.TYPE).Value.(lexer.TokenType)

	// type has a name
	if p.currentToken().Type == lexer.IDENT {
		return typedName{
			name:     p.expect(lexer.IDENT).Value.(string),
			nameType: ftype,
		}
	}

	// type with no name
	return typedName{
		nameType: ftype,
	}
}

// List in to format of type name, type name ...
func (p *parser) typeList() []typedName {
	p.log("Start TypeList", true)
	defer p.log("End TypeList", false)

	var names []typedName

	// If its an empty list return nil
	if p.currentToken().Type != lexer.TYPE {
		return nil
	}

	// While their is a comma continue
	ok := true
	for ok {
		names = append(names, p.typedName())
		ok = p.accept(lexer.COMMA)
	}

	return names
}

func (p *parser) block() block {
	p.log("Start Block", true)
	defer p.log("End Block", false)

	block := block{}
	p.expect(lexer.OPENBODY)
	p.clearNewLines()

	for p.currentToken().Type != lexer.CLOSEBODY {
		block.expressions = append(block.expressions, p.expression())
	}

	p.expect(lexer.CLOSEBODY)
	p.clearNewLines()
	return block
}

func (p *parser) function() function {
	p.log("Start Function", true)
	defer p.log("End Function", false)

	name := p.previousToken().Value.(string)
	p.expect(lexer.DOUBLECOLON)
	args := p.typeList()
	p.expect(lexer.ARROW)
	returns := p.typeList()
	block := p.block()

	return function{name, args, returns, block}
}

func (p *parser) ret() ret {
	p.log("Start Return", true)
	defer p.log("End Return", false)

	p.expect(lexer.RETURN)

	var returns []expression

	// While their is a comma continue
	ok := true
	for ok {
		returns = append(returns, p.expression())
		ok = p.currentToken().Type == lexer.COMMA
	}

	return ret{returns}
}

func (p *parser) maths() expression {
	p.log("Start Maths", true)
	defer p.log("End Maths", false)

	var buffer []lexer.Token

	depth := 0

	for p.currentToken().Type != lexer.NEWLINE &&
		p.currentToken().Type != lexer.SEMICOLON &&
		p.currentToken().Type != lexer.OPENBODY &&
		p.currentToken().Type != lexer.CLOSEBODY &&
		!(p.currentToken().Type == lexer.COMMA && depth == 0) {

		token := p.currentToken()

		if token.Type == lexer.OPENBRACKET {
			depth++
		}

		if token.Type == lexer.CLOSEBRACKET {
			depth--
		}

		buffer = append(buffer, p.currentToken())
		p.nextToken()
	}
	p.clearNewLines()

	spew.Dump(buffer)

	return p.shuntingYard(buffer)
}

func (p *parser) assignment() expression {
	p.log("Start Assignment", true)
	defer p.log("End Assignment", false)

	assignType := p.expect(lexer.TYPE).Value.(lexer.TokenType)
	name := p.expect(lexer.IDENT).Value.(string)
	p.expect(lexer.ASSIGN)
	value := p.maths()

	return assignment{
		name:     name,
		nameType: assignType,
		value:    value,
	}
}

func (p *parser) inferAssignment() expression {
	p.log("Start Infer Assignment", true)
	defer p.log("End Infer Assignment", false)

	name := p.expect(lexer.IDENT).Value.(string)
	p.expect(lexer.INFERASSIGN)
	value := p.expression()

	return assignment{
		name:     name,
		value:    value,
		nameType: lexer.ILLEGAL,
	}
}

func (p *parser) ifBranch() ifBlock {
	p.log("ifBranch", true)
	defer p.log("ifBranch", false)

	var condition expression
	var block block

	switch p.currentToken().Type {
	case lexer.IF:
		p.expect(lexer.IF)
		condition = p.maths()
		block = p.block()

	case lexer.ELSE:
		p.expect(lexer.ELSE)
		block = p.block()
	}

	return ifBlock{
		block:     block,
		condition: condition,
	}
}

func (p *parser) ifBlock() ifExpression {
	p.log("ifBlock", true)
	defer p.log("ifBlock", false)

	ifBranch := p.ifBranch()
	if p.currentToken().Type == lexer.ELSE {
		elseBranch := p.ifBranch()

		return ifExpression{
			blocks: []ifBlock{ifBranch, elseBranch},
		}
	}

	return ifExpression{
		blocks: []ifBlock{ifBranch},
	}
}

func (p *parser) cast() cast {
	p.log("Start Cast", true)
	defer p.log("End Cast", false)

	p.expect(lexer.OPENBRACKET)
	castType := p.expect(lexer.TYPE).Value.(lexer.TokenType)
	p.expect(lexer.CLOSEBRACKET)

	return cast{
		cast:  castType,
		value: p.maths(),
	}
}

func (p *parser) increment() expression {
	p.log("Start Increment", true)
	defer p.log("End Increment", false)

	name := p.expect(lexer.IDENT).Value.(string)
	var expr expression
	switch p.currentToken().Type {
	case lexer.INCREMENT:
		p.nextToken()
		expr = increment{name}
	case lexer.DECREMENT:
		p.nextToken()
		expr = decrement{name}
	case lexer.INCREMENTEQUAL:
		p.nextToken()
		expr = incrementExpression{name, p.maths()}
	case lexer.DECREMENTEQUAL:
		p.nextToken()
		expr = decrementExpression{name, p.maths()}
	}

	p.clearNewLines()

	return expr
}

func (p *parser) forExpression() forExpression {
	p.log("Start for", true)
	defer p.log("End for", false)

	p.expect(lexer.FOR)

	index := p.inferAssignment()
	p.expect(lexer.SEMICOLON)

	condition := p.maths()
	p.expect(lexer.SEMICOLON)

	increment := p.increment()

	block := p.block()

	return forExpression{index, condition, increment, block}
}

func (p *parser) array() array {
	p.log("Start array", true)
	defer p.log("End array", false)

	p.expect(lexer.OPENSQUAREBRACKET)
	length := p.expect(lexer.INTVALUE).Value.(int)
	p.expect(lexer.CLOSESQUAREBRACKET)
	arrayType := p.expect(lexer.TYPE)
	p.expect(lexer.OPENBODY)

	array := array{
		length:   length,
		baseType: arrayType.Value.(lexer.TokenType),
	}

	for p.currentToken().Type != lexer.CLOSEBODY {
		array.values = append(array.values, p.maths())
		p.accept(lexer.COMMA)
	}

	p.expect(lexer.CLOSEBODY)
	p.clearNewLines()

	return array
}

func (p *parser) expression() expression {
	p.log("Start Expression", true)
	defer p.log("End Expression", false)

	switch p.currentToken().Type {
	case lexer.DOUBLECOLON:
		return p.function()
	case lexer.RETURN:
		return p.ret()
	case lexer.IF:
		return p.ifBlock()
	case lexer.FOR:
		return p.forExpression()
	case lexer.OPENBODY:
		return p.block()
	case lexer.OPENSQUAREBRACKET:
		return p.array()
	case lexer.TYPE:
		return p.assignment()
	case lexer.OPENBRACKET:
		return p.cast()
	case lexer.IDENT:
		switch p.peekNextToken().Type {
		case lexer.INCREMENT, lexer.DECREMENT, lexer.INCREMENTEQUAL, lexer.DECREMENTEQUAL:
			return p.increment()
		case lexer.INFERASSIGN:
			return p.inferAssignment()
		default:
			return p.maths()
		}
	default:
		return p.maths()
	}
}

// NewParser creates a new parser
func newParser(tokens []lexer.Token) *parser {
	return &parser{
		tokens:            tokens,
		currentTokenIndex: 0,
	}
}

// Parse parses the file and returns the syntax tree
func (p *parser) Parse() syntaxTree {
	var functions []function

	for p.currentTokenIndex < len(p.tokens)-1 {
		p.nextToken()
		nextFunction := p.expression()
		functions = append(functions, nextFunction.(function))
		p.clearNewLines()
	}

	return syntaxTree{functions}
}
