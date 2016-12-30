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

func (p *Parser) eof() bool {
	return p.index >= len(p.tokens)-1
}

func (p *Parser) expect(typ lexer.TokenType) lexer.Token {
	token := p.token()
	if token.Type() != typ {
		panic(fmt.Sprintf("Expected: %s, Got: %s", typ.String(), token.Type().String()))
	}

	if !p.eof() {
		p.next()
	}

	return token
}

func (p *Parser) accept(typ lexer.TokenType) (lexer.Token, bool) {
	if p.eof() {
		return lexer.Token{}, false
	}

	token := p.token()
	if token.Type() != typ {
		return token, false
	}

	p.next()
	return token, true
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
		outputStack.Push(ast.Binary{lhs, token.Type(), rhs})
	}

	// TODO: simplify this condition
	notEnded := func(token lexer.Token, depth int) bool {
		return token.Type() != lexer.SEMICOLON &&
			token.Type() != lexer.LBRACE &&
			token.Type() != lexer.RBRACE &&
			!((token.Type() == lexer.COMMA || token.Type() == lexer.RPAREN) && depth == 0)
	}

	// fmt.Println("=== Begin maths ===")
	// fmt.Println("Start token:", p.token().String())

	depth := 0
	for notEnded(p.token(), depth) {
		token := p.token()

		// fmt.Println("current: ", token.String())

		switch token.Type() {
		case lexer.INT:
			outputStack.Push(p.integer())

		case lexer.FLOAT:
			outputStack.Push(p.float())

		case lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO, lexer.REM,
			lexer.GTR, lexer.LSS, lexer.GEQ, lexer.LEQ, lexer.EQL, lexer.NEQ:

			for !operatorStack.Empty() &&
				!operatorStack.Head().(lexer.Token).IsOperator() &&
				token.Precedence() <= operatorStack.Head().(lexer.Token).Precedence() {

				popOperatorStack()
			}
			operatorStack.Push(token)
			p.next()

		case lexer.IDENT:
			switch {
			// Token is a function name, push it onto the operator stack
			case notEnded(p.peek(), depth) && p.peek().Type() == lexer.LPAREN:
				outputStack.Push(p.call())

			// Token is a array index
			case notEnded(p.peek(), depth) && p.peek().Type() == lexer.LBRACK:
				outputStack.Push(p.arrayValue())

			// Token is a varible name, push it onto the out queue
			default:
				outputStack.Push(p.ident())
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

		// fmt.Println("next: ", p.token().String())
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

func (p *Parser) typ() ast.Type {
	switch {
	case p.peek().Type() == lexer.LBRACK:
		base := p.expect(lexer.IDENT)
		p.expect(lexer.LBRACK)
		length := p.integer()
		p.expect(lexer.RBRACK)
		return &ast.ArrayType{
			Type:   ast.NewBasic(base.Value()),
			Length: length,
		}
	default:
		return ast.NewBasic(p.expect(lexer.IDENT).Value())
	}
}

func (p *Parser) namedTyp() ast.TypedIdent {
	typ := p.typ()
	name := p.ident()

	return ast.TypedIdent{
		Ident: name,
		Type:  typ,
	}
}

func (p *Parser) argList() (args []ast.TypedIdent) {
	for p.token().Type() != lexer.ARROW {
		args = append(args, p.namedTyp())
		p.accept(lexer.COMMA)
	}

	return args
}

func (p *Parser) returnList() (returns []ast.Type) {
	p.expect(lexer.ARROW)
	for p.token().Type() != lexer.LBRACE {
		returns = append(returns, p.typ())
		p.accept(lexer.COMMA)
	}

	return returns
}

func (p *Parser) function() ast.Function {
	name := p.ident()
	p.expect(lexer.DOUBLE_COLON)
	args := p.argList()
	returns := p.returnList()
	body := p.block()

	return ast.Function{
		Name: name,
		Type: ast.FunctionType{
			Parameters: args,
			Returns:    returns,
		},
		Body: body,
	}
}

func (p *Parser) assignment() ast.Assignment {
	typ := p.typ()
	ident := p.ident()
	p.expect(lexer.ASSIGN)
	expression := p.Value()
	return ast.Assignment{typ, ident, expression}
}

func (p *Parser) inferAssigment() ast.Assignment {
	ident := p.ident()
	p.expect(lexer.DEFINE)
	expression := p.Value()
	return ast.Assignment{nil, ident, expression}
}

func (p *Parser) reAssigment() ast.Assignment {
	ident := p.ident()
	p.expect(lexer.ASSIGN)
	expression := p.Value()
	return ast.Assignment{nil, ident, expression}
}

func (p *Parser) increment() ast.Assignment {
	ident := p.ident()
	p.next()
	return ast.Assignment{
		Name: ident,
		Expression: ast.Binary{
			Lhs: ident,
			Op:  lexer.ADD,
			Rhs: ast.Integer{
				Value: 1,
			},
		},
	}
}

func (p *Parser) decrement() ast.Assignment {
	ident := p.ident()
	p.next()
	return ast.Assignment{
		Name: ident,
		Expression: ast.Binary{
			Lhs: ident,
			Op:  lexer.SUB,
			Rhs: ast.Integer{
				Value: 1,
			},
		},
	}
}

func (p *Parser) addAssign() ast.Assignment {
	ident := p.ident()
	p.next()
	value := p.Value()
	return ast.Assignment{
		Name: ident,
		Expression: ast.Binary{
			Lhs: ident,
			Op:  lexer.ADD,
			Rhs: value,
		},
	}
}

func (p *Parser) subAssign() ast.Assignment {
	ident := p.ident()
	p.next()
	value := p.Value()
	return ast.Assignment{
		Name: ident,
		Expression: ast.Binary{
			Lhs: ident,
			Op:  lexer.SUB,
			Rhs: value,
		},
	}
}

func (p *Parser) block() ast.Block {
	p.expect(lexer.LBRACE)

	var expressions []ast.Expression
	for p.token().Type() != lexer.RBRACE {
		expressions = append(expressions, p.Expression())
	}

	p.expect(lexer.RBRACE)

	return ast.Block{expressions}
}

func (p *Parser) ifBlock() *ast.If {
	p.expect(lexer.IF)
	condition := p.maths()
	body := p.block()
	var elseIf *ast.If

	// Check for else/else if
	_, isElse := p.accept(lexer.ELSE)
	if isElse {
		if p.token().Type() == lexer.IF {
			elseIf = p.ifBlock()
		} else {
			elseIf = &ast.If{
				Block: p.block(),
			}
		}
	}

	return &ast.If{
		Condition: condition,
		Block:     body,
		Else:      elseIf,
	}
}

func (p *Parser) forBlock() *ast.For {
	p.expect(lexer.FOR)
	index := p.Expression()
	condition := p.maths()
	p.expect(lexer.SEMICOLON)
	increment := p.Expression()
	block := p.block()

	return &ast.For{
		Index:     index,
		Condition: condition,
		Increment: increment,
		Block:     block,
	}
}

func (p *Parser) cast() ast.Cast {
	p.expect(lexer.LPAREN)
	typ := p.typ()
	p.expect(lexer.RPAREN)
	exp := p.Value()

	return ast.Cast{
		Type:       typ,
		Expression: exp,
	}
}

func (p *Parser) list() ast.List {
	p.expect(lexer.LBRACE)

	var expressions []ast.Expression
	ok := p.token().Type() != lexer.RBRACE
	for ok {
		expressions = append(expressions, p.Value())
		_, ok = p.accept(lexer.COMMA)
	}

	p.expect(lexer.RBRACE)

	return ast.List{
		Expressions: expressions,
	}
}

func (p *Parser) Value() ast.Expression {
	switch p.token().Type() {
	case lexer.LPAREN:
		return p.cast()
	case lexer.LBRACE:
		return p.list()
	default:
		return p.maths()
	}

}

func (p *Parser) Expression() ast.Expression {
	var exp ast.Expression

	switch p.token().Type() {
	case lexer.RETURN:
		exp = p.ret()
	case lexer.IF:
		exp = p.ifBlock()
	case lexer.FOR:
		exp = p.forBlock()
	case lexer.LPAREN:
		exp = p.cast()
	case lexer.LBRACE:
		exp = p.block()

	case lexer.IDENT:
		switch p.peek().Type() {
		case lexer.DOUBLE_COLON:
			exp = p.function()
		case lexer.INC:
			exp = p.increment()
		case lexer.DEC:
			exp = p.decrement()
		case lexer.ADD_ASSIGN:
			exp = p.addAssign()
		case lexer.SUB_ASSIGN:
			exp = p.subAssign()
		case lexer.ASSIGN:
			exp = p.reAssigment()
		case lexer.DEFINE:
			exp = p.inferAssigment()
		case lexer.LPAREN:
			exp = p.call()
		default:
			exp = p.assignment()
		}
	default:
		exp = p.maths()
	}

	p.accept(lexer.SEMICOLON)
	return exp
}

func (p *Parser) Parse() ast.Ast {
	var functions []ast.Function

	for !p.eof() {
		functions = append(functions, p.function())
		p.accept(lexer.SEMICOLON)
	}

	return ast.Ast{
		Functions: functions,
	}
}
