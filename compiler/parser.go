package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/dprint"
)

const enableLogging = true

type typedName struct {
	nameType lexer.Type
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
	cast  lexer.Type
	value expression
}

type ret struct {
	returns []expression
}

type assignment struct {
	name     string
	nameType lexer.Type
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
	baseType lexer.Type
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

// func (p *parser) increment() expression {
// 	p.log("Start Increment", true)
// 	defer p.log("End Increment", false)

// 	name := p.expect(lexer.IDENT).Value.(string)
// 	var expr expression
// 	switch p.currentToken().Type {
// 	case lexer.INCREMENT:
// 		p.nextToken()
// 		expr = increment{name}
// 	case lexer.DECREMENT:
// 		p.nextToken()
// 		expr = decrement{name}
// 	case lexer.INCREMENTEQUAL:
// 		p.nextToken()
// 		expr = incrementExpression{name, p.maths()}
// 	case lexer.DECREMENTEQUAL:
// 		p.nextToken()
// 		expr = decrementExpression{name, p.maths()}
// 	}

// 	p.clearNewLines()

// 	return expr
// }

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
