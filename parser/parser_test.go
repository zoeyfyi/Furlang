package parser

import (
	"testing"

	"reflect"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
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

		parser := NewParser(tokens)
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
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.source))
		tokens, err := lexer.Lex()
		if err != nil {
			t.Error(err)
		}

		parser := NewParser(tokens)
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
