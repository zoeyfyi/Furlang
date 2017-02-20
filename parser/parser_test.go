package parser

import (
	"testing"

	"reflect"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
	"github.com/k0kubun/pp"
)

func TestParserExpressions(t *testing.T) {
	cases := []struct {
		source string
		ast    interface{}
	}{
		{
			`123`,
			&ast.LiteralExpression{
				Value: lexer.NewToken(lexer.INT, "123", 1, 1),
			},
		},

		{
			`321 + 123`,
			&ast.BinaryExpression{
				Left: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "321", 1, 1),
				},
				Operator: lexer.NewToken(lexer.ADD, "", 1, 5),
				Right: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 7),
				},
			},
		},

		{
			`-123`,
			&ast.UnaryExpression{
				Operator: lexer.NewToken(lexer.SUB, "", 1, 1),
				Expression: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 2),
				},
			},
		},

		{
			`+123`,
			&ast.UnaryExpression{
				Operator: lexer.NewToken(lexer.ADD, "", 1, 1),
				Expression: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 2),
				},
			},
		},

		{
			`()`,
			&ast.ParenLiteralExpression{
				LeftParen:  lexer.NewToken(lexer.LPAREN, "", 1, 1),
				Elements:   []ast.Expression{},
				RightParen: lexer.NewToken(lexer.RPAREN, "", 1, 2),
			},
		},

		{
			`(123)`,
			&ast.LiteralExpression{
				Value: lexer.NewToken(lexer.INT, "123", 1, 2),
			},
		},

		{
			`(123, 321)`,
			&ast.ParenLiteralExpression{
				LeftParen: lexer.NewToken(lexer.LPAREN, "", 1, 1),
				Elements: []ast.Expression{
					&ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "123", 1, 2),
					},
					&ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "321", 1, 7),
					},
				},
				RightParen: lexer.NewToken(lexer.RPAREN, "", 1, 10),
			},
		},

		{
			`call()`,
			&ast.CallExpression{
				Function: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "call", 1, 1),
				},
				Arguments: &ast.ParenLiteralExpression{
					LeftParen:  lexer.NewToken(lexer.LPAREN, "", 1, 5),
					Elements:   []ast.Expression{},
					RightParen: lexer.NewToken(lexer.RPAREN, "", 1, 6),
				},
			},
		},

		{
			`call(123)`,
			&ast.CallExpression{
				Function: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "call", 1, 1),
				},
				Arguments: &ast.ParenLiteralExpression{
					LeftParen: lexer.NewToken(lexer.LPAREN, "", 1, 5),
					Elements: []ast.Expression{
						&ast.LiteralExpression{
							Value: lexer.NewToken(lexer.INT, "123", 1, 6),
						},
					},
					RightParen: lexer.NewToken(lexer.RPAREN, "", 1, 9),
				},
			},
		},

		{
			`call(123, 321)`,
			&ast.CallExpression{
				Function: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "call", 1, 1),
				},
				Arguments: &ast.ParenLiteralExpression{
					LeftParen: lexer.NewToken(lexer.LPAREN, "", 1, 5),
					Elements: []ast.Expression{
						&ast.LiteralExpression{
							Value: lexer.NewToken(lexer.INT, "123", 1, 6),
						},
						&ast.LiteralExpression{
							Value: lexer.NewToken(lexer.INT, "321", 1, 11),
						},
					},
					RightParen: lexer.NewToken(lexer.RPAREN, "", 1, 14),
				},
			},
		},

		{
			`123 * (321 + 45)`,
			&ast.BinaryExpression{
				Left: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 1),
				},
				Operator: lexer.NewToken(lexer.MUL, "", 1, 5),
				Right: &ast.BinaryExpression{
					Left: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "321", 1, 8),
					},
					Operator: lexer.NewToken(lexer.ADD, "", 1, 12),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "45", 1, 14),
					},
				},
			},
		},

		{
			`test[12]`,
			&ast.IndexExpression{
				Expression: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "test", 1, 1),
				},
				LeftBrack: lexer.NewToken(lexer.LBRACK, "", 1, 5),
				Index: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "12", 1, 6),
				},
				RightBrack: lexer.NewToken(lexer.RBRACK, "", 1, 8),
			},
		},
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.source))
		tokens, err := lexer.Lex()
		if err != nil {
			t.Error(err)
		}

		parser := NewParser(tokens, false)
		t.Log("=== Start tokens ===")
		for _, tok := range parser.tokens {
			t.Log(tok.String())
		}
		t.Log("=== End tokens ===\n\n")
		ast := parser.expression(0)
		if !reflect.DeepEqual(c.ast, ast) {
			t.Errorf("Source:\n%q\nExpected:\n%s\nGot:\n%s\n",
				c.source, pp.Sprint(c.ast), pp.Sprint(ast))
		}
	}

}

func TestParserStatements(t *testing.T) {
	cases := []struct {
		source string
		ast    ast.Statement
	}{
		{
			`ben = 123`,
			&ast.AssignmentStatement{
				Left: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "ben", 1, 1),
				},
				Assign: lexer.NewToken(lexer.ASSIGN, "", 1, 5),
				Right: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 7),
				},
			},
		},

		{
			`return 123`,
			&ast.ReturnStatement{
				Return: lexer.NewToken(lexer.RETURN, "return", 1, 1),
				Result: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 8),
				},
			},
		},

		{
			`{}`,
			&ast.BlockStatement{
				Scope:      nil,
				LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 1),
				Statements: []ast.Statement{},
				RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 2),
			},
		},

		{
			`{
				ben = 123
			}`,
			&ast.BlockStatement{
				LeftBrace: lexer.NewToken(lexer.LBRACE, "", 1, 1),
				Statements: []ast.Statement{
					&ast.AssignmentStatement{
						Left: &ast.IdentExpression{
							Value: lexer.NewToken(lexer.IDENT, "ben", 1, 7),
						},
						Assign: lexer.NewToken(lexer.ASSIGN, "", 1, 11),
						Right: &ast.LiteralExpression{
							Value: lexer.NewToken(lexer.INT, "123", 1, 13),
						},
					},
				},
				RightBrace: lexer.NewToken(lexer.RBRACE, "", 2, 4),
			},
		},

		{
			`if x > 3 {}`,
			&ast.IfStatment{
				If: lexer.NewToken(lexer.IF, "if", 1, 1),
				Condition: &ast.BinaryExpression{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "x", 1, 4),
					},
					Operator: lexer.NewToken(lexer.GTR, "", 1, 6),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "3", 1, 8),
					},
				},
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 10),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 11),
				},
				Else: nil,
			},
		},

		{
			`if x == 3 {}`,
			&ast.IfStatment{
				If: lexer.NewToken(lexer.IF, "if", 1, 1),
				Condition: &ast.BinaryExpression{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "x", 1, 4),
					},
					Operator: lexer.NewToken(lexer.EQL, "", 1, 6),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "3", 1, 9),
					},
				},
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 11),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 12),
				},
				Else: nil,
			},
		},

		{
			`if x > 3 {} else if x < 3 {}`,
			&ast.IfStatment{
				If: lexer.NewToken(lexer.IF, "if", 1, 1),
				Condition: &ast.BinaryExpression{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "x", 1, 4),
					},
					Operator: lexer.NewToken(lexer.GTR, "", 1, 6),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "3", 1, 8),
					},
				},
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 10),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 11),
				},
				Else: &ast.IfStatment{
					If: lexer.NewToken(lexer.IF, "if", 1, 18),
					Condition: &ast.BinaryExpression{
						Left: &ast.IdentExpression{
							Value: lexer.NewToken(lexer.IDENT, "x", 1, 21),
						},
						Operator: lexer.NewToken(lexer.LSS, "", 1, 23),
						Right: &ast.LiteralExpression{
							Value: lexer.NewToken(lexer.INT, "3", 1, 25),
						},
					},
					Body: &ast.BlockStatement{
						LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 27),
						Statements: []ast.Statement{},
						RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 28),
					},
					Else: nil,
				},
			},
		},

		{
			`if x > 3 {} else {}`,
			&ast.IfStatment{
				If: lexer.NewToken(lexer.IF, "if", 1, 1),
				Condition: &ast.BinaryExpression{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "x", 1, 4),
					},
					Operator: lexer.NewToken(lexer.GTR, "", 1, 6),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "3", 1, 8),
					},
				},
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 10),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 11),
				},
				Else: &ast.IfStatment{
					If:        lexer.Token{},
					Condition: nil,
					Body: &ast.BlockStatement{
						LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 18),
						Statements: []ast.Statement{},
						RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 19),
					},
					Else: nil,
				},
			},
		},

		{
			`for i = 0; i < 10; i = i + 1 {}`,
			&ast.ForStatement{
				For: lexer.NewToken(lexer.FOR, "for", 1, 1),
				Index: &ast.AssignmentStatement{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "i", 1, 5),
					},
					Assign: lexer.NewToken(lexer.ASSIGN, "", 1, 7),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "0", 1, 9),
					},
				},
				Semi1: lexer.NewToken(lexer.SEMICOLON, "", 1, 10),
				Condition: &ast.BinaryExpression{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "i", 1, 12),
					},
					Operator: lexer.NewToken(lexer.LSS, "", 1, 14),
					Right: &ast.LiteralExpression{
						Value: lexer.NewToken(lexer.INT, "10", 1, 16),
					},
				},
				Semi2: lexer.NewToken(lexer.SEMICOLON, "", 1, 18),
				Increment: &ast.AssignmentStatement{
					Left: &ast.IdentExpression{
						Value: lexer.NewToken(lexer.IDENT, "i", 1, 20),
					},
					Assign: lexer.NewToken(lexer.ASSIGN, "", 1, 22),
					Right: &ast.BinaryExpression{
						Left: &ast.IdentExpression{
							Value: lexer.NewToken(lexer.IDENT, "i", 1, 24),
						},
						Operator: lexer.NewToken(lexer.ADD, "", 1, 26),
						Right: &ast.LiteralExpression{
							Value: lexer.NewToken(lexer.INT, "1", 1, 28),
						},
					},
				},
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 30),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 31),
				},
			},
		},
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.source))
		tokens, err := lexer.Lex()
		if err != nil {
			t.Error(err)
		}

		parser := NewParser(tokens, false)
		t.Log("=== Start tokens ===")
		for _, tok := range parser.tokens {
			t.Log(tok.String())
		}
		t.Log("=== End tokens ===\n\n")
		ast := parser.statement()
		if !reflect.DeepEqual(c.ast, ast) {
			t.Errorf("Source:\n%q\nExpected:\n%s\nGot:\n%s\n",
				c.source, pp.Sprint(c.ast), pp.Sprint(ast))
		}
	}
}

func TestFunctionDeclarations(t *testing.T) {
	cases := []struct {
		source string
		ast    *ast.FunctionDeclaration
	}{
		{
			`proc function :: -> {}`,
			&ast.FunctionDeclaration{
				Name: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "function", 1, 6),
				},
				DoubleColon: lexer.NewToken(lexer.DOUBLE_COLON, "", 1, 15),
				Arguments:   []*ast.ArgumentDeclaration{},
				Return:      nil,
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 21),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 22),
				},
			},
		},

		{
			`proc function :: -> int {}`,
			&ast.FunctionDeclaration{
				Name: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "function", 1, 6),
				},
				DoubleColon: lexer.NewToken(lexer.DOUBLE_COLON, "", 1, 15),
				Arguments:   []*ast.ArgumentDeclaration{},
				Return:      types.GetType("int"),
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 25),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 26),
				},
			},
		},

		{
			`proc function :: int a, int b -> int {}`,
			&ast.FunctionDeclaration{
				Name: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "function", 1, 6),
				},
				DoubleColon: lexer.NewToken(lexer.DOUBLE_COLON, "", 1, 15),
				Arguments: []*ast.ArgumentDeclaration{
					&ast.ArgumentDeclaration{
						Name: &ast.IdentExpression{
							Value: lexer.NewToken(lexer.IDENT, "a", 1, 22),
						},
						Type: types.GetType("int"),
					},
					&ast.ArgumentDeclaration{
						Name: &ast.IdentExpression{
							Value: lexer.NewToken(lexer.IDENT, "b", 1, 29),
						},
						Type: types.GetType("int"),
					},
				},
				Return: types.GetType("int"),
				Body: &ast.BlockStatement{
					LeftBrace:  lexer.NewToken(lexer.LBRACE, "", 1, 38),
					Statements: []ast.Statement{},
					RightBrace: lexer.NewToken(lexer.RBRACE, "", 1, 39),
				},
			},
		},
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.source))
		tokens, err := lexer.Lex()
		if err != nil {
			t.Error(err)
		}
		parser := NewParser(tokens, false)
		tree := parser.declaration()

		testFunc := func(expect interface{}, got interface{}) {
			if !reflect.DeepEqual(expect, got) {
				t.Errorf("Source:\n%q\nExpected:\n%s\nGot:\n%s\n",
					c.source, pp.Sprint(expect), pp.Sprint(got))
			}
		}

		// Test all parts of the function declaration except scope
		testFunc(c.ast.Name, tree.(*ast.FunctionDeclaration).Name)
		testFunc(c.ast.DoubleColon, tree.(*ast.FunctionDeclaration).DoubleColon)
		testFunc(c.ast.Arguments, tree.(*ast.FunctionDeclaration).Arguments)
		testFunc(c.ast.Return, tree.(*ast.FunctionDeclaration).Return)
		testFunc(c.ast.Body.LeftBrace, tree.(*ast.FunctionDeclaration).Body.LeftBrace)
		testFunc(c.ast.Body.Statements, tree.(*ast.FunctionDeclaration).Body.Statements)
		testFunc(c.ast.Body.RightBrace, tree.(*ast.FunctionDeclaration).Body.RightBrace)
	}
}

func TestVaribleDeclarations(t *testing.T) {
	cases := []struct {
		source string
		ast    *ast.VaribleDeclaration
	}{
		{
			`ben := 123`,
			&ast.VaribleDeclaration{
				Type: nil,
				Name: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "ben", 1, 1),
				},
				Value: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 8),
				},
			},
		},
		{
			`int ben = 123`,
			&ast.VaribleDeclaration{
				Type: types.GetType("int"),
				Name: &ast.IdentExpression{
					Value: lexer.NewToken(lexer.IDENT, "ben", 1, 5),
				},
				Value: &ast.LiteralExpression{
					Value: lexer.NewToken(lexer.INT, "123", 1, 11),
				},
			},
		},
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.source))
		tokens, err := lexer.Lex()
		if err != nil {
			t.Error(err)
		}

		parser := NewParser(tokens, false)
		t.Log("=== Start tokens ===")
		for _, tok := range parser.tokens {
			t.Log(tok.String())
		}
		t.Log("=== End tokens ===\n\n")
		tree := parser.declaration()
		if !reflect.DeepEqual(c.ast, tree) {

			t.Errorf("Source:\n%q\nExpected:\n%s\nGot:\n%s\n",
				c.source, pp.Sprint(c.ast), pp.Sprint(tree))
		}
	}
}
