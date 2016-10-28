package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/dprint"
	"github.com/oleiade/lane"
)

const enableLogging = false

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

type ret struct {
	returns []expression
}

type assignment struct {
	name     string
	nameType lexer.TokenType
	value    expression
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
		ok = p.accept(lexer.COMMAN)
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
		ok = p.currentToken().Type == lexer.COMMAN
	}

	return ret{returns}
}

// Uses shunting yard algoritum to convert
func (p *parser) shuntingYard(tokens []lexer.Token) expression {
	p.log("Start ShuntingYard", true)
	defer p.log("End ShuntingYard", false)

	outputStack := lane.NewStack()
	operatorStack := lane.NewStack()
	arityStack := lane.NewStack()

	checkOperatorStack := func(op operator) bool {
		// Check is stack is empty
		if operatorStack.Empty() {
			return false
		}

		// Check if stack is an operator
		headType := operatorStack.Head().(lexer.Token).Type
		if _, isOp := opMap[headType]; !isOp {
			return false
		}

		// Check operator precendence
		headPrecendence := opMap[headType].precendence
		return (!op.right && op.precendence <= headPrecendence) ||
			(op.right && op.precendence < headPrecendence)
	}

	popOperatorStack := func() {
		operator := operatorStack.Pop()
		token := operator.(lexer.Token)

		if token.Type == lexer.IDENT {
			argCount := arityStack.Pop().(int)

			exp := call{function: token.Value.(string)}
			for i := 0; i < argCount; i++ {
				exp.args = append(exp.args, outputStack.Pop().(expression))
			}

			outputStack.Push(exp)
			return
		}

		rhs := outputStack.Pop().(expression)
		lhs := outputStack.Pop().(expression)

		switch token.Type {
		case lexer.PLUS:
			outputStack.Push(addition{lhs, rhs})
		case lexer.MINUS:
			outputStack.Push(subtraction{lhs, rhs})
		case lexer.MULTIPLY:
			outputStack.Push(multiplication{lhs, rhs})
		case lexer.FLOATDIVIDE:
			outputStack.Push(floatDivision{lhs, rhs})
		case lexer.INTDIVIDE:
			outputStack.Push(intDivision{lhs, rhs})
		case lexer.LESSTHAN:
			outputStack.Push(lessThan{lhs, rhs})
		case lexer.MORETHAN:
			outputStack.Push(moreThan{lhs, rhs})
		case lexer.EQUAL:
			outputStack.Push(equal{lhs, rhs})
		case lexer.NOTEQUAL:
			outputStack.Push(notEqual{lhs, rhs})
		}
	}

	for i, t := range tokens {
		switch t.Type {

		case lexer.TRUE:
			outputStack.Push(boolean{true})

		case lexer.FALSE:
			outputStack.Push(boolean{false})

		case lexer.INTVALUE:
			outputStack.Push(number{t.Value.(int)})

		case lexer.FLOATVALUE:
			outputStack.Push(float{t.Value.(float32)})

		case lexer.PLUS, lexer.MINUS, lexer.MULTIPLY, lexer.FLOATDIVIDE, lexer.INTDIVIDE,
			lexer.MORETHAN, lexer.LESSTHAN, lexer.EQUAL, lexer.NOTEQUAL:
			for checkOperatorStack(opMap[t.Type]) {
				popOperatorStack()
			}
			operatorStack.Push(t)

		case lexer.IDENT:
			if i < len(tokens)-1 && tokens[i+1].Type == lexer.OPENBRACKET {
				// Token is a function name, push it onto the operator stack
				operatorStack.Push(t)
				if tokens[i+2].Type == lexer.CLOSEBRACKET {
					// 0 if function dosnt have any arguments
					arityStack.Push(0)
				} else {
					// 1 if their is atleast 1 argument
					arityStack.Push(1)
				}
			} else {
				// Token is a varible name, push it onto the out queue
				outputStack.Push(name{t.Value.(string)})
			}

		case lexer.OPENBRACKET:
			operatorStack.Push(t)

		case lexer.CLOSEBRACKET:
			for operatorStack.Head().(lexer.Token).Type != lexer.OPENBRACKET {
				popOperatorStack()
			}

			operatorStack.Pop() // pop open bracket

			if operatorStack.Head().(lexer.Token).Type == lexer.IDENT {
				popOperatorStack()
			}

			// if operatorStack.Empty() {
			// 	panic("Mismatched parentheses")
			// 	// return maths{}, Error{
			// 	// 	err:        "Mismatched parentheses",
			// 	// 	tokenRange: []token{t},
			// 	// }
			// }

		case lexer.COMMAN:
			// Increment argument count
			as := arityStack.Pop().(int)
			arityStack.Push(as + 1)

			for operatorStack.Head().(lexer.Token).Type != lexer.OPENBRACKET {
				popOperatorStack()
			}

			if operatorStack.Empty() {
				panic("Misplaced comma or mismatched parentheses")
				// return maths{}, Error{
				// 	err:        "Misplaced comma or mismatched parentheses",
				// 	tokenRange: []token{t},
				// }
			}

		default:
			panic("Unexpected math token: " + t.String())
			// return maths{}, Error{
			// 	err:        fmt.Sprintf("Unexpected math token: %s", t.String()),
			// 	tokenRange: []token{t},
			// }
		}
	}

	for !operatorStack.Empty() {
		popOperatorStack()
	}

	return outputStack.Pop().(expression)
}

func (p *parser) maths() expression {
	p.log("Start Maths", true)
	defer p.log("End Maths", false)

	var buffer []lexer.Token

	for p.currentToken().Type != lexer.NEWLINE &&
		p.currentToken().Type != lexer.SEMICOLON &&
		p.currentToken().Type != lexer.OPENBODY {

		buffer = append(buffer, p.currentToken())
		p.nextToken()
	}
	p.clearNewLines()

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
	value := p.maths()

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

func (p *parser) expression() expression {
	p.log("Start Expression", true)
	defer p.log("End Expression", false)

	switch p.currentToken().Type {
	case lexer.DOUBLECOLON:
		return expression(p.function())
	case lexer.RETURN:
		return expression(p.ret())
	case lexer.IF:
		return expression(p.ifBlock())
	case lexer.OPENBODY:
		return p.block()
	case lexer.TYPE:
		return p.assignment()
	case lexer.IDENT:
		switch p.peekNextToken().Type {
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
