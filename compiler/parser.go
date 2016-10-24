package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/bongo227/dprint"
	"github.com/oleiade/lane"
)

const enableLogging = false

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
	opMap = map[int]operator{
		tokenMoreThan:    operator{3, false},
		tokenLessThan:    operator{3, false},
		tokenPlus:        operator{4, false},
		tokenMinus:       operator{4, false},
		tokenMultiply:    operator{5, false},
		tokenFloatDivide: operator{5, false},
		tokenIntDivide:   operator{5, false},
	}
)

type parser struct {
	tokens            []token
	currentTokenIndex int
	depth             int
}

const debugLog = false

func (p *parser) log(statement string, start bool) {
	if enableLogging {
		if !start {
			p.depth--
		}

		fmt.Println(strings.Repeat(" ", p.depth), statement)

		if start {
			p.depth++
		}
	}
}

func (p *parser) currentToken() token {
	return p.tokens[p.currentTokenIndex]
}

func (p *parser) previousToken() token {
	if p.currentTokenIndex == 0 {
		panic("At first token (no previous token)")
	}

	return p.tokens[p.currentTokenIndex-1]
}

func (p *parser) clearNewLines() {
	ok := p.accept(tokenNewLine)
	for ok && p.currentTokenIndex != len(p.tokens)-1 {
		ok = p.accept(tokenNewLine)
	}
}

func (p *parser) nextToken() token {
	if p.currentTokenIndex < len(p.tokens) {
		p.currentTokenIndex++
		return p.currentToken()
	}

	panic("Ran out of tokens")
}

func (p *parser) peekNextToken() token {
	return p.tokens[p.currentTokenIndex+1]
}

func (p *parser) expect(tokenType int) token {
	if p.currentToken().tokenType == tokenType {
		prev := p.currentToken()

		if p.currentTokenIndex < len(p.tokens)-1 {
			p.nextToken()
		}

		return prev
	}

	panic("Unexpected: " + p.currentToken().String() + "; Expected: " + tokenTypeString(tokenType))
}

func (p *parser) accept(tokenType int) bool {
	if p.currentToken().tokenType == tokenType && p.currentTokenIndex == len(p.tokens)-1 {
		return true
	}

	if p.currentToken().tokenType == tokenType {
		p.nextToken()
		return true
	}

	return false
}

// typed name in the format: type name, where name is optional
func (p *parser) typedName() typedName {
	p.log("Start TypedName", true)
	defer p.log("End TypedName", false)

	ftype := p.expect(tokenType).value.(int)

	// type has a name
	if p.currentToken().tokenType == tokenName {
		return typedName{
			name:     p.expect(tokenName).value.(string),
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
	if p.currentToken().tokenType != tokenType {
		return nil
	}

	// While their is a comma continue
	ok := true
	for ok {
		names = append(names, p.typedName())
		ok = p.accept(tokenComma)
	}

	return names
}

func (p *parser) block() block {
	p.log("Start Block", true)
	defer p.log("End Block", false)

	block := block{}
	p.expect(tokenOpenBody)
	p.clearNewLines()

	for p.currentToken().tokenType != tokenCloseBody {
		block.expressions = append(block.expressions, p.expression())
	}

	p.expect(tokenCloseBody)
	p.clearNewLines()
	return block
}

func (p *parser) function() function {
	p.log("Start Function", true)
	defer p.log("End Function", false)

	name := p.previousToken().value.(string)
	p.expect(tokenDoubleColon)
	args := p.typeList()
	p.expect(tokenArrow)
	returns := p.typeList()
	block := p.block()

	return function{name, args, returns, block}
}

func (p *parser) ret() ret {
	p.log("Start Return", true)
	defer p.log("End Return", false)

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

func (p *parser) interger() number {
	p.log("Start Number", true)
	defer p.log("End Number", false)

	value := p.expect(tokenNumber).value.(int)
	p.clearNewLines()
	return number{value}
}

// Uses shunting yard algoritum to convert
func (p *parser) shuntingYard(tokens []token) expression {
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
		headType := operatorStack.Head().(token).tokenType
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
		token := operator.(token)

		if token.tokenType == tokenName {
			argCount := arityStack.Pop().(int)

			exp := call{function: token.value.(string)}
			for i := 0; i < argCount; i++ {
				exp.args = append(exp.args, outputStack.Pop().(expression))
			}

			outputStack.Push(exp)
			return
		}

		rhs := outputStack.Pop().(expression)
		lhs := outputStack.Pop().(expression)

		switch token.tokenType {
		case tokenPlus:
			outputStack.Push(addition{lhs, rhs})
		case tokenMinus:
			outputStack.Push(subtraction{lhs, rhs})
		case tokenMultiply:
			outputStack.Push(multiplication{lhs, rhs})
		case tokenFloatDivide:
			outputStack.Push(floatDivision{lhs, rhs})
		case tokenIntDivide:
			outputStack.Push(intDivision{lhs, rhs})
		case tokenLessThan:
			outputStack.Push(lessThan{lhs, rhs})
		case tokenMoreThan:
			outputStack.Push(moreThan{lhs, rhs})
		}
	}

	for i, t := range tokens {
		switch t.tokenType {

		case tokenTrue:
			outputStack.Push(boolean{true})

		case tokenFalse:
			outputStack.Push(boolean{false})

		case tokenNumber:
			outputStack.Push(number{t.value.(int)})

		case tokenFloat:
			outputStack.Push(float{t.value.(float32)})

		case tokenPlus, tokenMinus, tokenMultiply, tokenFloatDivide, tokenIntDivide,
			tokenMoreThan, tokenLessThan:
			for checkOperatorStack(opMap[t.tokenType]) {
				popOperatorStack()
			}
			operatorStack.Push(t)

		case tokenName:
			if i < len(tokens)-1 && tokens[i+1].tokenType == tokenOpenBracket {
				// Token is a function name, push it onto the operator stack
				operatorStack.Push(t)
				if tokens[i+2].tokenType == tokenCloseBracket {
					// 0 if function dosnt have any arguments
					arityStack.Push(0)
				} else {
					// 1 if their is atleast 1 argument
					arityStack.Push(1)
				}
			} else {
				// Token is a varible name, push it onto the out queue
				outputStack.Push(name{t.value.(string)})
			}

		case tokenOpenBracket:
			operatorStack.Push(t)

		case tokenCloseBracket:
			for operatorStack.Head().(token).tokenType != tokenOpenBracket {
				popOperatorStack()
			}

			operatorStack.Pop() // pop open bracket

			if operatorStack.Head().(token).tokenType == tokenName {
				popOperatorStack()
			}

			// if operatorStack.Empty() {
			// 	panic("Mismatched parentheses")
			// 	// return maths{}, Error{
			// 	// 	err:        "Mismatched parentheses",
			// 	// 	tokenRange: []token{t},
			// 	// }
			// }

		case tokenComma:
			// Increment argument count
			as := arityStack.Pop().(int)
			arityStack.Push(as + 1)

			for operatorStack.Head().(token).tokenType != tokenOpenBracket {
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

	var buffer []token

	for p.currentToken().tokenType != tokenNewLine &&
		p.currentToken().tokenType != tokenSemiColon &&
		p.currentToken().tokenType != tokenOpenBody {

		buffer = append(buffer, p.currentToken())
		p.nextToken()
	}
	p.clearNewLines()

	return p.shuntingYard(buffer)
}

func (p *parser) assignment() expression {
	p.log("Start Assignment", true)
	defer p.log("End Assignment", false)

	name := p.expect(tokenName).value.(string)
	p.expect(tokenAssign)
	value := p.maths()

	return assignment{
		name:  name,
		value: value,
	}
}

func (p *parser) ifBranch() ifBlock {
	var condition expression
	var block block

	switch p.currentToken().tokenType {
	case tokenIf:
		p.expect(tokenIf)
		condition = p.maths()
		block = p.block()

	case tokenElse:
		p.expect(tokenElse)
		block = p.block()
	}

	return ifBlock{
		block:     block,
		condition: condition,
	}
}

func (p *parser) ifBlock() ifExpression {
	ifBranch := p.ifBranch()
	if p.currentToken().tokenType == tokenElse {
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

	switch p.currentToken().tokenType {
	case tokenDoubleColon:
		return expression(p.function())
	case tokenReturn:
		return expression(p.ret())
	case tokenIf:
		return expression(p.ifBlock())
	case tokenOpenBody:
		return p.block()
	case tokenName:
		switch p.peekNextToken().tokenType {
		case tokenAssign:
			return p.assignment()
		default:
			return p.maths()
		}
	default:
		return p.maths()
	}
}

// NewParser creates a new parser
func newParser(tokens []token) *parser {
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
