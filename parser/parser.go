package parser

import (
	"fmt"

	"log"

	"errors"

	"runtime/debug"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
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
func Parse(code string) (ast.Expression, error) {
	lex := lexer.NewLexer([]byte(code))

	tokens, err := lex.Lex()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	return parser.expression(), nil
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

func (p *Parser) typ() types.Type {
	baseType := types.GetType(p.ident().Value.Value())

	if p.token().Type() == lexer.LBRACK {
		p.next()
		elementCount := p.expression()
		return types.NewArray(baseType, 10)
	}

	return baseType
}

func (p *Parser) ident() *ast.IdentExpression {
	return &ast.IdentExpression{
		Value: p.expect(lexer.IDENT),
	}
}

func (p *Parser) literal() *ast.LiteralExpression {
	if tok, ok := p.accept(lexer.INT); ok {
		return &ast.LiteralExpression{
			Value: tok,
		}
	} else if tok, ok := p.accept(lexer.FLOAT); ok {
		return &ast.LiteralExpression{
			Value: tok,
		}
	}

	panic(p.newError("Expected literal value"))
}

func (p *Parser) braceLiteral() *ast.BraceLiteralExpression {
	return &ast.BraceLiteralExpression{
		Type:       p.typ(),
		LeftBrace:  p.expect(lexer.LBRACE),
		Elements:   p.expressionList(lexer.RBRACE),
		RightBrace: p.expect(lexer.RBRACE),
	}
}

func (p *Parser) parenLiteral() *ast.ParenLiteralExpression {
	return &ast.ParenLiteralExpression{
		LeftParen:  p.expect(lexer.LPAREN),
		Elements:   p.expressionList(lexer.RPAREN),
		RightParen: p.expect(lexer.RPAREN),
	}
}

func (p *Parser) indexNode() *ast.IndexExpression {
	return &ast.IndexExpression{
		Expression: p.expression(),
		LeftBrack:  p.expect(lexer.LBRACK),
		Index:      p.expression(),
		RightBrack: p.expect(lexer.RBRACK),
	}
}

func (p *Parser) slice() *ast.SliceExpression {
	return &ast.SliceExpression{
		Expression: p.expression(),
		LeftBrack:  p.expect(lexer.LBRACK),
		Low:        p.expression(),
		Colon:      p.expect(lexer.COLON),
		High:       p.expression(),
		RightBrack: p.expect(lexer.RBRACK),
	}
}

func (p *Parser) call() *ast.CallExpression {
	return &ast.CallExpression{
		Function:  p.expression(),
		Arguments: p.parenLiteral(),
	}
}

func (p *Parser) cast() *ast.CastExpression {
	return &ast.CastExpression{
		LeftParen:  p.expect(lexer.LPAREN),
		Type:       p.typ(),
		RightParen: p.expect(lexer.RPAREN),
		Expression: p.expression(),
	}
}

func (p *Parser) assignment() *ast.AssignmentStatement {
	return &ast.AssignmentStatement{
		Left:   p.expression(),
		Assign: p.expect(lexer.ASSIGN),
		Right:  p.expression(),
	}
}

func (p *Parser) returnNode() *ast.ReturnStatement {
	return &ast.ReturnStatement{
		Return: p.expect(lexer.RETURN),
		Result: p.expression(),
	}
}

func (p *Parser) block() *ast.BlockStatement {
	return &ast.BlockStatement{
		LeftBrace:  p.expect(lexer.LBRACE),
		Statements: p.statementList(lexer.RBRACE),
		RightBrace: p.expect(lexer.RBRACE),
	}
}

func (p *Parser) ifNode() *ast.IfStatment {
	return &ast.IfStatment{
		If:        p.expect(lexer.IF),
		Condition: p.expression(),
		Body:      p.block(),
		Else: (func() *ast.IfStatment {
			// Check for else/else if
			elseToken, ok := p.accept(lexer.ELSE)
			if ok {
				if p.token().Type() == lexer.IF {
					// Else if block
					return p.ifNode()
				}

				// Else block
				return &ast.IfStatment{
					If:        elseToken,
					Condition: nil,
					Body:      p.block(),
					Else:      nil,
				}
			}

			return nil
		})(),
	}
}

func (p *Parser) forNode() *ast.ForStatement {
	return &ast.ForStatement{
		For:       p.expect(lexer.FOR),
		Index:     p.statement(),
		Semi1:     p.expect(lexer.SEMICOLON),
		Condition: p.expression(),
		Semi2:     p.expect(lexer.SEMICOLON),
		Increment: p.statement(),
		Body:      p.block(),
	}
}

func (p *Parser) function() *ast.FunctionDeclaration {
	return &ast.FunctionDeclaration{
		Name:        p.ident(),
		DoubleColon: p.expect(lexer.DOUBLE_COLON),
		Arguments:   p.argumentList(),
		Return:      p.typ(),
		Body:        p.block(),
	}
}

func (p *Parser) expressionList(delimiter lexer.TokenType) []ast.Expression {
	var expressions []ast.Expression
	for p.token().Type() != delimiter {
		expressions = append(expressions, p.expression())
		p.accept(lexer.COMMA)
	}

	return expressions
}

func (p *Parser) statementList(delimiter lexer.TokenType) []ast.Statement {
	var statements []ast.Statement
	for p.token().Type() != delimiter {
		statements = append(statements, p.statement())
		p.accept(lexer.SEMICOLON)
	}

	return statements
}

func (p *Parser) argumentList() map[ast.IdentExpression]types.Type {
	arguments := make(map[ast.IdentExpression]types.Type)

	for p.token().Type() != lexer.ARROW {
		arguments[*p.ident()] = p.typ()
		p.accept(lexer.COMMA)
	}

	return arguments
}

func (p *Parser) expression() ast.Expression {
	switch {
	case p.token().Type() == lexer.INT, p.token().Type() == lexer.FLOAT:
		return p.literal()
	case p.token().Type() == lexer.LPAREN:
		return p.parenLiteral()
	}
}

func (p *Parser) statement() ast.Statement {
	switch {

	}
}

func (p *Parser) Parse() (tree *ast.Ast, err error) {
	log.Println("Starting parse")

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			tree = nil

			switch r := r.(type) {
			case *Error:
				err = r
			case *InternalError:
				err = r
			default:
				err = fmt.Errorf("Unhandled internal error: %q", r)
			}

			if parserError, ok := r.(*Error); ok {
				// Parser error
				err = parserError
			} else {
				// Unknown error
				err = errors.New("Internal error")
			}
		}
	}()

	var functions []*ast.FunctionDeclaration

	for !p.eof() {
		functions = append(functions, p.function())
		p.accept(lexer.SEMICOLON)
	}

	tree = &ast.Ast{
		Functions: functions,
	}

	return tree, err
}
