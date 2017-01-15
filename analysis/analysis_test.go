package analysis

import (
	"testing"

	"reflect"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/parser"
	"github.com/k0kubun/pp"
)

var a = &Analysis{}

func TestFloatPromotion(t *testing.T) {
	code := "10 + 14.5"

	exp, err := parser.ParseExpression(code)
	if err != nil {
		t.Error(err)
	}

	node := exp.(*ast.BinaryExpression)
	a.binary(node)

	if _, ok := node.Left.(*ast.CastExpression); !ok {
		t.Errorf("Expected left hand side of: %s to be a cast node, got %s",
			code, pp.Sprint(node.Left))
	}

	if typ, _ := a.typ(node); typ != floatType {
		t.Errorf("Expected %s to have type float but got type: %s", code, pp.Sprint(typ))
	}
}

// func TestInferAssigment(t *testing.T) {
// 	cases := []struct {
// 		code string
// 		typ  types.Type
// 	}{
// 		{
// 			code: "a := 10",
// 			typ:  intType,
// 		},
// 		{
// 			code: "a := 14.5",
// 			typ:  floatType,
// 		},
// 		{
// 			code: "a := 10 * 13.5",
// 			typ:  floatType,
// 		},
// 		{
// 			code: "a := 10 / 2",
// 			typ:  intType,
// 		},
// 	}

// 	for _, c := range cases {
// 		node, err := parser.ParseDeclaration(c.code)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		assigmentNode, ok := node.(*ast.VaribleDeclaration)
// 		if !ok {
// 			t.Errorf("Expected node to have type assigment, got type: %s", reflect.TypeOf(assigmentNode).String())
// 		}

// 		a.varibleDcl(ass)
// 		if assigmentNode.Type != c.typ {
// 			// TODO: give types.Type a string method so we dont have to pp.Print
// 			t.Errorf("Expected %s to have infer int type but got type: %s", c.code, pp.Sprint(assigmentNode.Type))
// 		}
// 	}
// }

// func TestAssigment(t *testing.T) {
// 	code := "i8 a = 123"

// 	node, err := parser.ParseStatement(code)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	assigmentNode, ok := node.(ast.AssignmentStatement)
// 	if !ok {
// 		t.Errorf("Expected node to have type assigment, got type: %s", reflect.TypeOf(assigmentNode).String())
// 	}

// 	analizedNode := a.assigment(&assigmentNode)
// 	if _, ok := analizedNode.Expression.(ast.Cast); !ok {
// 		t.Errorf("Expected value of: %s to be a cast node, got %s",
// 			code, pp.Sprint(assigmentNode.Expression))
// 	}
// }

func TestCall(t *testing.T) {
	code := `
		proc add :: i32 a, i64 b -> i64 {
			return a + b
		}

		proc main :: -> i32 {
			return add(10, 243)
		}
	`

	lex := lexer.NewLexer([]byte(code))
	tokens, err := lex.Lex()
	if err != nil {
		t.Error(err)
	}

	parser := parser.NewParser(tokens)
	tree := parser.Parse()
	// if err != nil {
	// 	t.Error(err)
	// }

	ana := NewAnalysis(tree)

	a := ana.Analalize()

	firstSmt := a.Functions[1].Body.Statements[0]
	returnSmt, ok := firstSmt.(*ast.ReturnStatement)
	if !ok {
		t.Errorf("Expected first expression to be a return statement, got: %s", reflect.TypeOf(firstSmt).String())
	}

	call, ok := returnSmt.Result.(*ast.CallExpression)
	if !ok {
		t.Errorf("Expected return value to be of type call, got: %s", reflect.TypeOf(call).String())
	}

	if _, ok := call.Arguments.Elements[1].(*ast.LiteralExpression); !ok {
		t.Errorf("Expected parameter 1 to be a integer got: %s", pp.Sprint(call.Arguments.Elements[1]))
	}

	if _, ok := call.Arguments.Elements[0].(*ast.CastExpression); !ok {
		t.Errorf("Expected parameter 0 to be a cast got: %s", pp.Sprint(call.Arguments.Elements[0]))
	}
}
