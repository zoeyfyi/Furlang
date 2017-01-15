package parser

import (
	"fmt"

	"runtime/debug"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/k0kubun/pp"
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

// InternalError represents an error in the parser package that is unexpected
type InternalError struct {
	Message string
	Stack   string
}

func (p *Parser) newError(message string) *Error {
	return &Error{
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Message
}

func (p *Parser) newInternalError(message string) *InternalError {
	return &InternalError{
		Message: "Internal error: " + message,
		Stack:   string(debug.Stack()),
	}
}

func (e *InternalError) Error() string {
	return e.Message
}

// Parse is a convience method for parsing a raw string of code
// func Parse(code string) (ast.Expression, error) {
// 	lex := lexer.NewLexer([]byte(code))

// 	tokens, err := lex.Lex()
// 	if err != nil {
// 		return nil, err
// 	}

// 	parser := NewParser(tokens)
// 	return parser.expression(), nil
// }

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
		return lexer.Token{}, false
	}

	p.next()
	return token, true
}

func bindingPower(token lexer.Token) int {
	switch token.Type() {
	case lexer.ADD, lexer.SUB:
		return 10
	case lexer.MUL, lexer.QUO:
		return 20
	case lexer.LSS, lexer.LEQ, lexer.GTR, lexer.GEQ:
		return 60
	case lexer.LPAREN, lexer.LBRACK:
		return 150
	}

	return 0
}

func (p *Parser) nud(token lexer.Token) ast.Expression {
	switch token.Type() {
	case lexer.IDENT:
		return &ast.IdentExpression{
			Value: token,
		}
	case lexer.INT, lexer.FLOAT:
		return &ast.LiteralExpression{
			Value: token,
		}
	case lexer.ADD, lexer.SUB:
		return &ast.UnaryExpression{
			Operator:   token,
			Expression: p.expression(100),
		}
	case lexer.LPAREN:
		if rparen, ok := p.accept(lexer.RPAREN); ok {
			return &ast.ParenLiteralExpression{
				LeftParen:  token,
				Elements:   []ast.Expression{},
				RightParen: rparen,
			}
		}

		e := p.expression(0)
		if _, ok := p.accept(lexer.RPAREN); ok {
			return e
		}

		elements := []ast.Expression{e}
		_, ok := p.accept(lexer.COMMA)
		for ok {
			elements = append(elements, p.expression(0))
			_, ok = p.accept(lexer.COMMA)
		}

		return &ast.ParenLiteralExpression{
			LeftParen:  token,
			Elements:   elements,
			RightParen: p.expect(lexer.RPAREN),
		}
	}

	panic("nud Undefined for token type: " + token.Type().String())
}

func (p *Parser) led(token lexer.Token, tree ast.Expression) ast.Expression {
	switch token.Type() {
	case lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO,
		lexer.LSS, lexer.LEQ, lexer.GTR, lexer.GEQ:

		e := p.expression(bindingPower(token))
		fmt.Println("Hello their!")
		return &ast.BinaryExpression{
			Left:     tree,
			Operator: token,
			Right:    e,
		}
	case lexer.LPAREN:
		fmt.Println("led lparen")
		elements := []ast.Expression{}
		ok := p.token().Type() != lexer.RPAREN
		for ok {
			elements = append(elements, p.expression(0))
			_, ok = p.accept(lexer.COMMA)
		}

		return &ast.CallExpression{
			Function: tree,
			Arguments: &ast.ParenLiteralExpression{
				LeftParen:  token,
				Elements:   elements,
				RightParen: p.expect(lexer.RPAREN),
			},
		}
	case lexer.LBRACK:
		return &ast.IndexExpression{
			Expression: tree,
			LeftBrack:  token,
			Index:      p.expression(0),
			RightBrack: p.expect(lexer.RBRACK),
		}
	}

	panic("led Undefined for token type: " + p.token().Type().String())
}

func (p *Parser) expression(rightBindingPower int) ast.Expression {
	t := p.token()
	p.next()
	left := p.nud(t)
	for rightBindingPower < bindingPower(p.token()) {
		pp.Println(rightBindingPower,
			bindingPower(p.token()),
			t.Type().String(),
			p.token().Type().String(),
			left)
		t = p.token()
		p.next()
		left = p.led(t, left)
	}

	return left
}

func (p *Parser) assigment() *ast.AssignmentStatement {
	return &ast.AssignmentStatement{
		Left:   p.expression(0),
		Assign: p.expect(lexer.ASSIGN),
		Right:  p.expression(0),
	}
}

func (p *Parser) returnSmt() *ast.ReturnStatement {
	return &ast.ReturnStatement{
		Return: p.expect(lexer.RETURN),
		Result: p.expression(0),
	}
}

func (p *Parser) block() *ast.BlockStatement {
	lbrace := p.expect(lexer.LBRACE)

	statements := []ast.Statement{}
	rbrace, ok := p.accept(lexer.RBRACE)
	for !ok {
		statements = append(statements, p.statement())
		p.expect(lexer.SEMICOLON)
		rbrace, ok = p.accept(lexer.RBRACE)
	}

	return &ast.BlockStatement{
		LeftBrace:  lbrace,
		Statements: statements,
		RightBrace: rbrace,
	}
}

func (p *Parser) ifSmt() *ast.IfStatment {
	ifToken, hasCondition := p.accept(lexer.IF)
	var condition ast.Expression
	if hasCondition {
		condition = p.expression(0)
	}

	block := p.block()

	var elseSmt *ast.IfStatment
	if _, ok := p.accept(lexer.ELSE); ok {
		elseSmt = p.ifSmt()
	}

	return &ast.IfStatment{
		If:        ifToken,
		Condition: condition,
		Body:      block,
		Else:      elseSmt,
	}
}

func (p *Parser) statement() ast.Statement {
	switch p.token().Type() {
	case lexer.RETURN:
		return p.returnSmt()
	case lexer.LBRACE:
		return p.block()
	case lexer.IF:
		return p.ifSmt()
	default:
		return p.assigment()
	}
}
