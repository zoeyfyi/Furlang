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
		// call(123)
		{
			source: `call(123)`,
			ast: ast.Call{
				Function: ast.Ident{
					Value: "call",
				},
				Arguments: []ast.Expression{
					ast.Integer{
						Value: 123,
					},
				},
			},
		},
		// int[5] items = {1, 2, 3, 4, 5}
		{
			source: `int[5] items = {1, 2, 3, 4, 5}`,
			ast: ast.Assignment{
				Type: &ast.ArrayType{
					Type: &ast.Basic{
						Ident: ast.Ident{
							Value: "int",
						},
						Type: nil,
					},
					Length: ast.Integer{
						Value: 5,
					},
				},
				Name: ast.Ident{
					Value: "items",
				},
				Expression: ast.List{
					Expressions: []ast.Expression{
						ast.Integer{
							Value: 1,
						},
						ast.Integer{
							Value: 2,
						},
						ast.Integer{
							Value: 3,
						},
						ast.Integer{
							Value: 4,
						},
						ast.Integer{
							Value: 5,
						},
					},
				},
			},
		},
		// {
		// 		call(123)
		// 		int[2] pair = {5, 6}
		// }
		{
			source: `{
				call(123)
				int[2] pair = {5, 6}
			}`,
			ast: ast.Block{
				Expressions: []ast.Expression{
					ast.Call{
						Function: ast.Ident{
							Value: "call",
						},
						Arguments: []ast.Expression{
							ast.Integer{
								Value: 123,
							},
						},
					},
					ast.Assignment{
						Type: &ast.ArrayType{
							Type: &ast.Basic{
								Ident: ast.Ident{
									Value: "int",
								},
								Type: nil,
							},
							Length: ast.Integer{
								Value: 2,
							},
						},
						Name: ast.Ident{
							Value: "pair",
						},
						Expression: ast.List{
							Expressions: []ast.Expression{
								ast.Integer{
									Value: 5,
								},
								ast.Integer{
									Value: 6,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.source))
		parser := NewParser(lexer.Lex())
		t.Log("=== Start tokens ===")
		for _, tok := range parser.tokens {
			t.Log(tok.String())
		}
		t.Log("=== End tokens ===\n\n")
		ast := parser.Expression()
		if !reflect.DeepEqual(c.ast, ast) {

			t.Errorf("Source:\n%q\nExpected:\n%s\nGot:\n%s\n", c.source, pp.Sprint(c.ast), pp.Sprint(ast))
		}
	}

}
