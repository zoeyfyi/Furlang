package compiler

import (
	"os"

	"github.com/bongo227/cmap"
	"github.com/oleiade/lane"
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

var (
	// Maps tokens onto operators
	opMap = map[int]operator{
		tokenPlus:        operator{2, false},
		tokenMinus:       operator{2, false},
		tokenMultiply:    operator{3, false},
		tokenFloatDivide: operator{3, false},
		tokenIntDivide:   operator{3, false},
	}
)

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

// Uses shunting yard algoritum to convert
func (p *Parser) shuntingYard(tokens []Token) expression {
	outputStack := lane.NewStack()
	operatorStack := lane.NewStack()
	arityStack := lane.NewStack()

	checkOperatorStack := func(op operator) bool {
		return !operatorStack.Empty() &&
			(operatorStack.Head().(Token).tokenType == tokenPlus ||
				operatorStack.Head().(Token).tokenType == tokenMinus ||
				operatorStack.Head().(Token).tokenType == tokenMultiply ||
				operatorStack.Head().(Token).tokenType == tokenFloatDivide ||
				operatorStack.Head().(Token).tokenType == tokenIntDivide) &&
			((!op.right && op.precendence <= opMap[operatorStack.Head().(Token).tokenType].precendence) ||
				(op.right && op.precendence < opMap[operatorStack.Head().(Token).tokenType].precendence))
	}

	popOperatorStack := func() {
		second := outputStack.Pop().(expression)
		first := outputStack.Pop().(expression)
		operator := operatorStack.Pop()
		token := operator.(Token)

		switch token.tokenType {
		case tokenPlus:
			outputStack.Push(addition{
				lhs: first,
				rhs: second,
			})
		case tokenMinus:
			outputStack.Push(subtraction{
				lhs: first,
				rhs: second,
			})
		case tokenMultiply:
			outputStack.Push(multiplication{
				lhs: first,
				rhs: second,
			})
		case tokenFloatDivide:
			outputStack.Push(floatDivision{
				lhs: first,
				rhs: second,
			})
		case tokenIntDivide:
			outputStack.Push(intDivision{
				lhs: first,
				rhs: second,
			})
		case tokenName:
			exp := call{function: token.value.(string)}
			for i := 0; i < arityStack.Pop().(int); i++ {
				exp.args = append(exp.args, outputStack.Pop().(expression))
			}
		}
	}

	for i, t := range tokens {
		switch t.tokenType {

		case tokenNumber:

			outputStack.Push(number{t.value.(int)})

		case tokenFloat:
			outputStack.Push(float{t.value.(float32)})

		case tokenPlus, tokenMinus, tokenMultiply, tokenFloatDivide, tokenIntDivide:
			op := opMap[t.tokenType]

			for checkOperatorStack(op) {
				popOperatorStack()
			}

			operatorStack.Push(t)

		case tokenName:
			if tokens[i+1].tokenType == tokenOpenBracket {
				// Token is a function name, push it onto the operator stack
				operatorStack.Push(t)

				// Push 0 if function dosnt have any arguments
				// Push 1 if their is atleast 1 argument
				if tokens[i+2].tokenType == tokenCloseBracket {
					arityStack.Push(0)
				} else {
					arityStack.Push(1)
				}
			} else {
				// Token is a varible name, push it onto the out queue
				outputStack.Push(name{t.value.(string)})
			}

		case tokenOpenBracket:
			operatorStack.Push(t)

		case tokenCloseBracket:
			if operatorStack.Empty() {
				panic("Mismatched parentheses")
				// return maths{}, Error{
				// 	err:        "Mismatched parentheses",
				// 	tokenRange: []Token{t},
				// }
			}

		case tokenComma:
			for operatorStack.Head().(Token).tokenType != tokenOpenBracket {
				popOperatorStack()
			}

			if operatorStack.Empty() {
				panic("Misplaced comma or mismatched parentheses")
				// return maths{}, Error{
				// 	err:        "Misplaced comma or mismatched parentheses",
				// 	tokenRange: []Token{t},
				// }
			}

		default:
			panic("Unexpected math token")
			// return maths{}, Error{
			// 	err:        fmt.Sprintf("Unexpected math token: %s", t.String()),
			// 	tokenRange: []Token{t},
			// }
		}
	}

	for !operatorStack.Empty() {
		popOperatorStack()
	}

	return outputStack.Pop().(expression)
}

func (p *Parser) maths() expression {
	var buffer []Token

	for p.currentToken().tokenType != tokenNewLine &&
		p.currentToken().tokenType != tokenSemiColon {

		buffer = append(buffer, p.currentToken())
		p.nextToken()
	}
	p.clearNewLines()

	return p.shuntingYard(buffer)
}

func (p *Parser) expression() expression {
	switch p.currentToken().tokenType {
	case tokenDoubleColon:
		return expression(p.function())
	case tokenReturn:
		return expression(p.ret())
	default:
		return p.maths()
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
