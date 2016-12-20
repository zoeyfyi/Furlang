package parser

import (
	"fmt"
	"strconv"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/oleiade/lane"
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

// TODO: remove this
func (p *Parser) back() {
	p.index--
}

func (p *Parser) peek() lexer.Token {
	return p.tokens[p.index+1]
}

// TODO: Can we remove this?
func (p *Parser) peekpeek() lexer.Token {
	return p.tokens[p.index+2]
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

func (p *Parser) ident() ast.Ident {
	return ast.Ident{p.expect(lexer.IDENT).Value()}
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

func (p *Parser) call() ast.Call {
	function := p.ident()
	p.expect(lexer.LPAREN)

	var args []ast.Expression
	for p.token().Type() != lexer.RPAREN {
		args = append(args, p.maths())
		p.accept(lexer.COMMA)
	}

	fmt.Println(args)
	p.expect(lexer.RPAREN)

	return ast.Call{function, args}
}

func (p *Parser) arrayValue() ast.ArrayValue {
	array := p.ident()
	p.expect(lexer.LBRACK)
	index := p.maths()
	p.expect(lexer.RBRACK)

	return ast.ArrayValue{array, ast.Index{index}}
}

// Maths parses binary expressions
func (p *Parser) maths() ast.Expression {
	outputStack := lane.NewStack()
	operatorStack := lane.NewStack()
	arityStack := lane.NewStack()

	popOperatorStack := func() {
		token := operatorStack.Pop().(lexer.Token)
		rhs := outputStack.Pop().(ast.Expression)
		lhs := outputStack.Pop().(ast.Expression)
		outputStack.Push(ast.Binary{lhs, token, rhs})
	}

	// TODO: simplify this condition
	notEnded := func(token lexer.Token, depth int) bool {
		return token.Type() != lexer.SEMICOLON &&
			token.Type() != lexer.LBRACE &&
			token.Type() != lexer.RBRACE &&
			!((token.Type() == lexer.COMMA || token.Type() == lexer.RPAREN) && depth == 0)
	}

	depth := 0
	for notEnded(p.token(), depth) {
		token := p.token()

		switch token.Type() {
		case lexer.INT:
			outputStack.Push(p.integer())
			p.back()

		case lexer.FLOAT:
			outputStack.Push(p.float())
			p.back()

		case lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO, lexer.REM,
			lexer.GTR, lexer.LSS, lexer.GEQ, lexer.LEQ, lexer.EQL, lexer.NEQ:

			for !operatorStack.Empty() &&
				!operatorStack.Head().(lexer.Token).IsOperator() &&
				token.Precedence() <= operatorStack.Head().(lexer.Token).Precedence() {

				popOperatorStack()
			}
			operatorStack.Push(token)

		case lexer.IDENT:
			switch {
			// Token is a function name, push it onto the operator stack
			case notEnded(p.peek(), depth) && p.peek().Type() == lexer.LPAREN:
				outputStack.Push(p.call())
				p.back()

			// Token is a array index
			case notEnded(p.peek(), depth) && p.peek().Type() == lexer.LBRACK:
				outputStack.Push(p.arrayValue())

			// Token is a varible name, push it onto the out queue
			default:
				outputStack.Push(p.ident())
				p.back()
			}

		case lexer.LPAREN:
			depth++
			operatorStack.Push(token)

		case lexer.RPAREN:
			depth--
			for operatorStack.Head().(lexer.Token).Type() != lexer.LPAREN {
				popOperatorStack()
			}
			operatorStack.Pop() // pop open bracket

		case lexer.COMMA:
			// Increment argument count
			as := arityStack.Pop().(int)
			arityStack.Push(as + 1)

			for operatorStack.Head().(lexer.Token).Type() != lexer.LPAREN {
				popOperatorStack()
			}

			if operatorStack.Empty() {
				panic("Misplaced comma or mismatched parentheses")
			}

		default:
			panic("Unexpected math token: " + token.String())
		}

		p.next()
	}

	for !operatorStack.Empty() {
		popOperatorStack()
	}

	return outputStack.Pop().(ast.Expression)
}

func (p *Parser) ret() ast.Return {
	p.expect(lexer.RETURN)
	return ast.Return{p.maths()}
}

func (p *Parser) typ() *ast.Type {
	return &ast.Type{Token: p.expect(lexer.IDENT)}
}

func (p *Parser) assignment() ast.Assignment {
	typ := p.typ()
	ident := p.ident()
	p.expect(lexer.ASSIGN)
	expression := p.maths()

	return ast.Assignment{typ, ident, expression}
}

func (p *Parser) inferAssigment() ast.Assignment {
	ident := p.ident()
	p.expect(lexer.DEFINE)
	expression := p.maths()

	return ast.Assignment{nil, ident, expression}
}

func (p *Parser) block() ast.Block {
	p.expect(lexer.LBRACE)

	var expressions []ast.Expression
	for p.token().Type() != lexer.RBRACE {
		expressions = append(expressions, p.Expression())
		// p.clearNewLines()
	}

	p.expect(lexer.RBRACE)

	return ast.Block{expressions}
}

func (p *Parser) Expression() ast.Expression {
	switch p.token().Type() {
	case lexer.RETURN:
		return p.ret()
	case lexer.LBRACE:
		return p.block()
	case lexer.IDENT:
		switch p.peek().Type() {
		case lexer.IDENT:
			return p.assignment()
		case lexer.DEFINE:
			return p.inferAssigment()
		default:
			return p.maths()
		}
	default:
		return p.maths()
	}

}

func (p *Parser) Parse() ast.Expression {
	return p.Expression()
}
