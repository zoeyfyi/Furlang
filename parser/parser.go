package parser

import (
	"fmt"
	"strconv"

	"log"

	"errors"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
	"github.com/oleiade/lane"
)

// Parser creates an abstract syntax tree from a sequence of tokens
type Parser struct {
	tokens []lexer.Token
	index  int
}

// Error represents an error in the parser package
type Error struct {
	Message string
}

func (p *Parser) newError(message string) *Error {
	return &Error{
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Message
}

// Parse is a convience method for parsing a raw string of code
func Parse(code string) (ast.Expression, error) {
	lex := lexer.NewLexer([]byte(code))

	tokens, err := lex.Lex()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	return parser.Expression(), nil
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

func (p *Parser) eof() bool {
	return p.index >= len(p.tokens)-1
}

func (p *Parser) expect(typ lexer.TokenType) lexer.Token {
	token := p.token()
	if token.Type() != typ {
		err := p.newError(fmt.Sprintf("Expected: %s, Got: %s", typ.String(), token.Type().String()))
		panic(err)
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

func (p *Parser) arrayType() types.Type {
	arrayType := p.ident().Value
	p.expect(lexer.LBRACK)
	length := p.integer()
	p.expect(lexer.RBRACK)

	return types.NewArray(types.GetType(arrayType), length.Value)
}

func (p *Parser) arrayListOrValue() ast.Expression {
	// Parse array type or array value
	ident := p.ident()
	p.expect(lexer.LBRACK)
	value := p.Value()
	p.expect(lexer.RBRACK)

	if p.token().Type() != lexer.LBRACE {
		// Not array list, must be array value
		return ast.ArrayValue{
			Array: ident,
			Index: ast.Index{
				Index: value,
			},
		}
	}

	// Assemble the array type
	elementType := types.GetType(ident.Value)
	arrayType := types.NewArray(elementType, value.(ast.Integer).Value)
	list := p.list()

	return ast.ArrayList{
		Type: arrayType,
		List: list,
	}
}

// Maths parses binary expressions
func (p *Parser) maths() ast.Expression {
	log.Println("Maths")

	outputStack := lane.NewStack()
	operatorStack := lane.NewStack()

	popOperatorStack := func() {
		token := operatorStack.Pop().(lexer.Token)
		rhs := outputStack.Pop().(ast.Expression)
		lhs := outputStack.Pop().(ast.Expression)
		outputStack.Push(ast.Binary{
			Lhs: lhs,
			Op:  token.Type(),
			Rhs: rhs,
		})
	}

	// TODO: simplify this condition
	notEnded := func(token lexer.Token, depth int) bool {
		return token.Type() != lexer.SEMICOLON &&
			token.Type() != lexer.LBRACE &&
			token.Type() != lexer.RBRACE &&
			token.Type() != lexer.RBRACK &&
			token.Type() != lexer.COMMA &&
			!(token.Type() == lexer.RPAREN && depth == 0)
	}

	log.Printf("Shunting yard: %s", p.peek().String())

	depth := 0
	for notEnded(p.token(), depth) {
		token := p.token()

		log.Printf("Shunting yard: %s", token.String())

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
			// Token is a function name
			case notEnded(p.peek(), depth) && p.peek().Type() == lexer.LPAREN:
				outputStack.Push(p.call())

			// Token is a array index
			case notEnded(p.peek(), depth) && p.peek().Type() == lexer.LBRACK:
				outputStack.Push(p.arrayListOrValue())

			// Token is a varible name, push it onto the out queue
			default:
				outputStack.Push(p.ident())
			}

		case lexer.LPAREN:
			depth++
			operatorStack.Push(token)
			p.next()

		case lexer.RPAREN:
			depth--
			for operatorStack.Head().(lexer.Token).Type() != lexer.LPAREN {
				popOperatorStack()
			}
			operatorStack.Pop() // pop open bracket
			p.next()

		default:
			panic("Unexpected math token: " + token.String())
		}
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

func (p *Parser) typ() types.Type {
	if p.peek().Type() == lexer.LBRACK {
		return p.arrayType()
	}

	return types.GetType(p.expect(lexer.IDENT).Value())
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

func (p *Parser) returnList() (returns []types.Type) {
	p.expect(lexer.ARROW)
	for p.token().Type() != lexer.LBRACE {
		returns = append(returns, p.typ())
		p.accept(lexer.COMMA)
	}

	return returns
}

func (p *Parser) function() ast.Function {
	log.Println("Function")

	name := p.ident()
	p.expect(lexer.DOUBLE_COLON)
	args := p.argList()
	returns := p.returnList()
	body := p.block()

	// TODO: multiple returns
	return ast.Function{
		Name: name,
		Type: ast.FunctionType{
			Parameters: args,
			Return:     returns[0],
		},
		Body: body,
	}
}

func (p *Parser) assignment() ast.Assignment {
	log.Println("Assignment")

	typ := p.typ()
	ident := p.ident()
	p.expect(lexer.ASSIGN)
	expression := p.Value()

	// If we are assigning an array give the list the array name
	if exp, ok := expression.(ast.ArrayList); ok {
		exp.Name = ident
	}

	return ast.Assignment{
		Type:       typ,
		Name:       ident,
		Expression: expression,
		Declare:    true,
	}
}

func (p *Parser) inferAssigment() ast.Assignment {
	log.Println("Infer Assignment")

	ident := p.ident()
	p.expect(lexer.DEFINE)
	expression := p.Value()

	// If we are assigning an array give the list the array name
	if exp, ok := expression.(ast.ArrayList); ok {
		expression = ast.ArrayList{
			Name: ident,
			Type: exp.Type,
			List: exp.List,
		}

	}

	return ast.Assignment{
		Type:       nil,
		Name:       ident,
		Expression: expression,
		Declare:    true,
	}
}

func (p *Parser) reAssigment() ast.Assignment {
	log.Println("Re-Assignment")

	ident := p.ident()
	p.expect(lexer.ASSIGN)
	expression := p.Value()

	// If we are assigning an array give the list the array name
	if exp, ok := expression.(ast.ArrayList); ok {
		exp.Name = ident
	}

	return ast.Assignment{
		Type:       nil,
		Name:       ident,
		Expression: expression,
		Declare:    false,
	}
}

func (p *Parser) increment() ast.Assignment {
	ident := p.ident()
	p.next()

	expression := ast.Expression(ast.Binary{
		Lhs: ident,
		Op:  lexer.ADD,
		Rhs: ast.Integer{
			Value: 1,
		},
	})

	return ast.Assignment{
		Name:       ident,
		Expression: expression,
	}
}

func (p *Parser) decrement() ast.Assignment {
	ident := p.ident()
	p.next()

	expression := ast.Expression(ast.Binary{
		Lhs: ident,
		Op:  lexer.SUB,
		Rhs: ast.Integer{
			Value: 1,
		},
	})

	return ast.Assignment{
		Name:       ident,
		Expression: expression,
	}
}

func (p *Parser) addAssign() ast.Assignment {
	ident := p.ident()
	p.next()
	value := p.Value()

	expression := ast.Expression(ast.Binary{
		Lhs: ident,
		Op:  lexer.ADD,
		Rhs: value,
	})

	return ast.Assignment{
		Name:       ident,
		Expression: expression,
	}
}

func (p *Parser) subAssign() ast.Assignment {
	ident := p.ident()
	p.next()
	value := p.Value()

	expression := ast.Expression(ast.Binary{
		Lhs: ident,
		Op:  lexer.SUB,
		Rhs: value,
	})

	return ast.Assignment{
		Name:       ident,
		Expression: expression,
	}
}

func (p *Parser) block() ast.Block {
	log.Println("Block")

	p.expect(lexer.LBRACE)

	var expressions []ast.Expression
	for p.token().Type() != lexer.RBRACE {
		expressions = append(expressions, p.Expression())
	}

	p.expect(lexer.RBRACE)

	return ast.Block{
		Expressions: expressions,
	}
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
	current := p.token().Type()
	next := p.peek().Type()

	log.Printf("Value current: %s", current.String())

	switch {
	case current == lexer.LPAREN && next == lexer.IDENT:
		return p.cast()
	case current == lexer.LBRACE:
		return p.list()
	case current == lexer.IDENT && next == lexer.LBRACK:
		return p.arrayListOrValue()
	default:
		return p.maths()
	}

}

func (p *Parser) Expression() ast.Expression {
	log.Printf("Expression, current: {%s}, next: {%s}", p.token().String(), p.peek().String())

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

func (p *Parser) Parse() (tree *ast.Ast, err error) {
	log.Println("Starting parse")

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			tree = nil

			if parserError, ok := r.(*Error); ok {
				// Parser error
				err = parserError
			} else {
				// Unknown error
				err = errors.New("Internal error")
			}
		}
	}()

	var functions []ast.Function

	for !p.eof() {
		functions = append(functions, p.function())
		p.accept(lexer.SEMICOLON)
	}

	tree = &ast.Ast{
		Functions: functions,
	}

	return tree, err
}
