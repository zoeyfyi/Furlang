package parser

import (
	"fmt"
	"reflect"

	"log"

	"strconv"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
)

// Parser creates an abstract syntax tree from a sequence of tokens
type Parser struct {
	tokens []lexer.Token
	scope  *ast.Scope
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

// func (p *Parser) newInternalError(message string) *InternalError {
// 	return &InternalError{
// 		Message: "Internal error: " + message,
// 		Stack:   string(debug.Stack()),
// 	}
// }

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

// NewParser creates a new parser, if scope is false all block scopes will be nil
func NewParser(tokens []lexer.Token, scope bool) *Parser {
	p := &Parser{
		tokens: tokens,
	}

	if scope {
		p.scope = ast.NewScope()
	}

	return p
}

func (p *Parser) enterScope() {
	if p.scope != nil {
		p.scope = p.scope.Enter()
	}
}

func (p *Parser) exitScope() {
	if p.scope != nil {
		p.scope = p.scope.Exit()
	}
}

func (p *Parser) insertScope(name string, node ast.Node) {
	log.Printf("Inserting %q into scope", name)
	if p.scope != nil {
		p.scope.Insert(name, node)
	}
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
	case lexer.LPAREN, lexer.LBRACK:
		return 150
	case lexer.ADD, lexer.SUB:
		return 110
	case lexer.MUL, lexer.QUO, lexer.REM:
		return 120
	case lexer.LSS, lexer.LEQ, lexer.GTR, lexer.GEQ,
		lexer.EQL, lexer.NEQ:
		return 60
	case lexer.LBRACE:
		return 20
	}

	return 0
}

func (p *Parser) nud(token lexer.Token) ast.Expression {
	// TODO: break this down into a token interace that has a nud method.
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
		lexer.LSS, lexer.LEQ, lexer.GTR, lexer.GEQ,
		lexer.EQL, lexer.NEQ, lexer.REM:

		log.Printf("Binding with power %d", bindingPower(token))
		e := p.expression(bindingPower(token))
		return &ast.BinaryExpression{
			Left:     tree,
			Operator: token,
			Right:    e,
		}
	case lexer.LPAREN:
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

	case lexer.LBRACE:
		indexExp, ok := tree.(*ast.IndexExpression)
		if !ok {
			log.Fatalf("Expected left hand side of { to be type \"*ast.IndexExpression\", got %q",
				reflect.TypeOf(tree).String())
		}

		// Convert index expression into array type
		typeName := indexExp.Expression.(*ast.IdentExpression).Value.Value()
		arraySize := indexExp.Index.(*ast.LiteralExpression).Value.Value()
		size, _ := strconv.Atoi(arraySize)
		arrayType := types.NewArray(types.GetType(typeName), int64(size))

		elements := []ast.Expression{}

		for p.token().Type() != lexer.RBRACE {
			elements = append(elements, p.expression(0))
			p.accept(lexer.COMMA)
		}

		return &ast.BraceLiteralExpression{
			Type:       arrayType,
			LeftBrace:  token,
			Elements:   elements,
			RightBrace: p.expect(lexer.RBRACE),
		}
	}

	panic("led Undefined for token type: " + p.token().Type().String())
}

func (p *Parser) expression(rightBindingPower int) ast.Expression {
	log.Printf("Start expression (%d)", rightBindingPower)

	t := p.token()
	p.next()
	left := p.nud(t)
	log.Printf("lbp: %d, rbp: %d", bindingPower(p.token()), rightBindingPower)
	for rightBindingPower < bindingPower(p.token()) {
		// TODO: figure out how to remove this check
		if _, ok := left.(*ast.IndexExpression); !ok && p.token().Type() == lexer.LBRACE {
			return left
		}
		t = p.token()
		p.next()
		left = p.led(t, left)
	}

	log.Print("Finished expression")

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
	p.enterScope()
	lbrace := p.expect(lexer.LBRACE)

	statements := []ast.Statement{}
	rbrace, ok := p.accept(lexer.RBRACE)
	for !ok {
		statements = append(statements, p.statement())
		p.expect(lexer.SEMICOLON)
		rbrace, ok = p.accept(lexer.RBRACE)
	}

	blockScope := p.scope
	p.exitScope()

	return &ast.BlockStatement{
		Scope:      blockScope,
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

	log.Print("Parsed condition")

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

func (p *Parser) forSmt() *ast.ForStatement {
	return &ast.ForStatement{
		For:       p.expect(lexer.FOR),
		Index:     p.statement(),
		Semi1:     p.expect(lexer.SEMICOLON),
		Condition: p.expression(0),
		Semi2:     p.expect(lexer.SEMICOLON),
		Increment: p.statement(),
		Body:      p.block(),
	}
}

func (p *Parser) incrementSmt() *ast.AssignmentStatement {
	exp := p.expression(0)

	var op lexer.TokenType
	var opRight ast.Expression
	token := p.token().Type()
	p.next()
	switch token {
	case lexer.INC:
		op = lexer.ADD
		opRight = &ast.LiteralExpression{
			Value: lexer.NewToken(lexer.INT, "1", 0, 0),
		}

	case lexer.DEC:
		op = lexer.SUB
		opRight = &ast.LiteralExpression{
			Value: lexer.NewToken(lexer.INT, "1", 0, 0),
		}

	case lexer.ADD_ASSIGN:
		op = lexer.ADD
		opRight = p.expression(0)

	case lexer.SUB_ASSIGN:
		op = lexer.SUB
		opRight = p.expression(0)

	case lexer.MUL_ASSIGN:
		op = lexer.MUL
		opRight = p.expression(0)

	case lexer.QUO_ASSIGN:
		op = lexer.QUO
		opRight = p.expression(0)

	case lexer.REM_ASSIGN:
		op = lexer.REM
		opRight = p.expression(0)
	}

	return &ast.AssignmentStatement{
		Left: exp,
		Right: &ast.BinaryExpression{
			IsFp:     false,
			Left:     exp,
			Operator: lexer.NewToken(op, "", 0, 0),
			Right:    opRight,
		},
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
	case lexer.FOR:
		return p.forSmt()
	default:
		// TODO: covert this into pratt pass
		switch p.peek().Type() {
		case lexer.INC, lexer.DEC, lexer.ADD_ASSIGN, lexer.SUB_ASSIGN, lexer.MUL_ASSIGN,
			lexer.QUO_ASSIGN, lexer.REM_ASSIGN:
			return p.incrementSmt()
		case lexer.ASSIGN:
			return p.assigment()
		default:
			return &ast.DeclareStatement{
				Statement: p.varibleDcl(),
			}
		}
	}
}

func (p *Parser) functionDcl() *ast.FunctionDeclaration {
	p.expect(lexer.PROC)

	// Parse function name
	name := &ast.IdentExpression{
		Value: p.expect(lexer.IDENT),
	}

	colon := p.expect(lexer.DOUBLE_COLON)

	// Parse function arguments
	arguments := []*ast.ArgumentDeclaration{}
	_, ok := p.accept(lexer.ARROW)
	for !ok {
		typeExp := p.expression(0)
		typ := types.GetType(typeExp.(*ast.IdentExpression).Value.Value())
		name := p.expect(lexer.IDENT)
		ident := &ast.IdentExpression{Value: name}

		arguments = append(arguments, &ast.ArgumentDeclaration{
			Name: ident,
			Type: typ,
		})

		_, ok = p.accept(lexer.ARROW)
		if !ok {
			p.expect(lexer.COMMA)
		}
	}

	// Get the return type
	var returnTyp types.Type
	if p.token().Type() != lexer.LBRACE {
		returnTypeExp := p.expression(0)
		returnTyp = types.GetType(returnTypeExp.(*ast.IdentExpression).Value.Value())
	}

	// Parse the function body
	block := p.block()

	// Inject function arguments into block scope
	for _, arg := range arguments {
		if block.Scope != nil {
			block.Scope.Insert(arg.Name.Value.Value(), &ast.VaribleDeclaration{
				Name:  arg.Name,
				Type:  arg.Type,
				Value: nil,
			})
		}
	}

	p.expect(lexer.SEMICOLON)

	funcDcl := &ast.FunctionDeclaration{
		Name:        name,
		DoubleColon: colon,
		Arguments:   arguments,
		Return:      returnTyp,
		Body:        block,
	}

	// Insert function into root scope
	p.insertScope(name.Value.Value(), funcDcl)

	return funcDcl
}

func (p *Parser) varibleDcl() *ast.VaribleDeclaration {
	var typ types.Type
	if p.peek().Type() != lexer.DEFINE {
		typExp := p.expression(0)
		// TODO: make type parser (simplifys brace literal expressions)
		typ = types.GetType(typExp.(*ast.IdentExpression).Value.Value())
	}

	name := &ast.IdentExpression{
		Value: p.expect(lexer.IDENT),
	}

	_, ok := p.accept(lexer.DEFINE)
	if !ok {
		p.expect(lexer.ASSIGN)
	}

	varDcl := &ast.VaribleDeclaration{
		Type:  typ,
		Name:  name,
		Value: p.expression(0),
	}

	p.insertScope(name.Value.Value(), varDcl)
	return varDcl
}

func (p *Parser) declaration() ast.Declare {
	switch p.token().Type() {
	case lexer.PROC:
		return p.functionDcl()
	default:
		return p.varibleDcl()
	}
}

func (p *Parser) Parse() *ast.Ast {
	var functions []*ast.FunctionDeclaration
	for !p.eof() {
		functions = append(functions, p.declaration().(*ast.FunctionDeclaration))
	}

	return &ast.Ast{
		Functions: functions,
		Scope:     p.scope,
	}
}

func ParseExpression(code string) (ast.Expression, error) {
	tokens, err := lexer.NewLexer([]byte(code)).Lex()
	if err != nil {
		return nil, err
	}

	ast := NewParser(tokens, true).expression(0)
	return ast, nil
}

func ParseStatement(code string) (ast.Statement, error) {
	tokens, err := lexer.NewLexer([]byte(code)).Lex()
	if err != nil {
		return nil, err
	}

	ast := NewParser(tokens, true).statement()
	return ast, nil
}

func ParseDeclaration(code string) (ast.Declare, error) {
	tokens, err := lexer.NewLexer([]byte(code)).Lex()
	if err != nil {
		return nil, err
	}

	ast := NewParser(tokens, true).declaration()
	return ast, nil
}
