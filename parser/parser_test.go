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
		// if ben > bob {
		// 	return ben
		// }
		{
			source: `if ben > ben {
				return ben
			}`,
			ast: &ast.If{
				Condition: ast.Binary{
					Lhs: ast.Ident{
						Value: "ben",
					},
					Op: lexer.GTR,
					Rhs: ast.Ident{
						Value: "ben",
					},
				},
				Block: ast.Block{
					Expressions: []ast.Expression{
						ast.Return{
							Value: ast.Ident{
								Value: "ben",
							},
						},
					},
				},
				Else: (*ast.If)(nil),
			},
		},
		// if bill {
		// 	bob
		// } else if bob {
		// 	bill
		// } else {
		// 	ben
		// }
		{
			source: `if bill {
				return bob
			} else if bob {
				return bill
			} else {
				return ben
			}`,
			ast: &ast.If{
				Condition: ast.Ident{
					Value: "bill",
				},
				Block: ast.Block{
					Expressions: []ast.Expression{
						ast.Return{
							Value: ast.Ident{
								Value: "bob",
							},
						},
					},
				},
				Else: &ast.If{
					Condition: ast.Ident{
						Value: "bob",
					},
					Block: ast.Block{
						Expressions: []ast.Expression{
							ast.Return{
								Value: ast.Ident{
									Value: "bill",
								},
							},
						},
					},
					Else: &ast.If{
						Condition: nil,
						Block: ast.Block{
							Expressions: []ast.Expression{
								ast.Return{
									Value: ast.Ident{
										Value: "ben",
									},
								},
							},
						},
						Else: (*ast.If)(nil),
					},
				},
			},
		},
		// for i := 0; i < 100; i = i+1 {
		// 	int ben = i
		// }
		{
			source: `for i := 0; i < 100; i = i+1 {
				int ben = i
			}`,
			ast: &ast.For{
				Index: ast.Assignment{
					Type: nil,
					Name: ast.Ident{
						Value: "i",
					},
					Expression: ast.Integer{
						Value: 0,
					},
				},
				Condition: ast.Binary{
					Lhs: ast.Ident{
						Value: "i",
					},
					Op: 39,
					Rhs: ast.Integer{
						Value: 100,
					},
				},
				Increment: ast.Assignment{
					Type: nil,
					Name: ast.Ident{
						Value: "i",
					},
					Expression: ast.Binary{
						Lhs: ast.Ident{
							Value: "i",
						},
						Op: 11,
						Rhs: ast.Integer{
							Value: 1,
						},
					},
				},
				Block: ast.Block{
					Expressions: []ast.Expression{
						ast.Assignment{
							Type: &ast.Basic{
								Ident: ast.Ident{
									Value: "int",
								},
								Type: nil,
							},
							Name: ast.Ident{
								Value: "ben",
							},
							Expression: ast.Ident{
								Value: "i",
							},
						},
					},
				},
			},
		},
		// int i = (int)0.5
		{
			source: `int i = (int)0.5`,
			ast: ast.Assignment{
				Type: &ast.Basic{
					Ident: ast.Ident{
						Value: "int",
					},
					Type: nil,
				},
				Name: ast.Ident{
					Value: "i",
				},
				Expression: ast.Cast{
					Expression: ast.Float{
						Value: 0.500000,
					},
					Type: &ast.Basic{
						Ident: ast.Ident{
							Value: "int",
						},
						Type: nil,
					},
				},
			},
		},
		// {
		// 	i++
		// 	i--
		// 	i+=10
		// 	i-=10
		// }
		{
			source: `{
				i++
				i--
				i+=10
				i-=10
			}`,
			ast: ast.Block{
				Expressions: []ast.Expression{
					ast.Assignment{
						Type: nil,
						Name: ast.Ident{
							Value: "i",
						},
						Expression: ast.Binary{
							Lhs: ast.Ident{
								Value: "i",
							},
							Op: 11,
							Rhs: ast.Integer{
								Value: 1,
							},
						},
					},
					ast.Assignment{
						Type: nil,
						Name: ast.Ident{
							Value: "i",
						},
						Expression: ast.Binary{
							Lhs: ast.Ident{
								Value: "i",
							},
							Op: 12,
							Rhs: ast.Integer{
								Value: 1,
							},
						},
					},
					ast.Assignment{
						Type: nil,
						Name: ast.Ident{
							Value: "i",
						},
						Expression: ast.Binary{
							Lhs: ast.Ident{
								Value: "i",
							},
							Op: 11,
							Rhs: ast.Integer{
								Value: 10,
							},
						},
					},
					ast.Assignment{
						Type: nil,
						Name: ast.Ident{
							Value: "i",
						},
						Expression: ast.Binary{
							Lhs: ast.Ident{
								Value: "i",
							},
							Op: 12,
							Rhs: ast.Integer{
								Value: 10,
							},
						},
					},
				},
			},
		},
		// add :: int a, int b -> int {
		// 	return a + b
		// }
		{
			source: `add :: int a, int b -> int {
				return a + b
			}`,
			ast: ast.Function{
				Type: ast.FunctionType{
					Parameters: []ast.TypedIdent{
						ast.TypedIdent{
							Type: &ast.Basic{
								Ident: ast.Ident{
									Value: "int",
								},
								Type: nil,
							},
							Ident: ast.Ident{
								Value: "a",
							},
						},
						ast.TypedIdent{
							Type: &ast.Basic{
								Ident: ast.Ident{
									Value: "int",
								},
								Type: nil,
							},
							Ident: ast.Ident{
								Value: "b",
							},
						},
					},
					Returns: []ast.Type{
						&ast.Basic{
							Ident: ast.Ident{
								Value: "int",
							},
							Type: nil,
						},
					},
				},
				Name: ast.Ident{
					Value: "add",
				},
				Body: ast.Block{
					Expressions: []ast.Expression{
						ast.Return{
							Value: ast.Binary{
								Lhs: ast.Ident{
									Value: "a",
								},
								Op: 11,
								Rhs: ast.Ident{
									Value: "b",
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
