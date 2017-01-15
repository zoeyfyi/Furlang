package parser

import (
	"testing"

	"reflect"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/k0kubun/pp"
)

func TestParser(t *testing.T) {
	// int[5] a = {1, 2, 3, 4, 5}

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
