package analysis

import (
	"testing"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/parser"
	"github.com/bongo227/Furlang/types"
	"github.com/k0kubun/pp"
)

var a = &Analysis{}

func TestFloatPromotion(t *testing.T) {
	code := "10 + 14.5"
	node := parser.Parse(code).(ast.Binary)
	a.binary(&node)

	if _, ok := node.Lhs.(ast.Cast); !ok {
		t.Errorf("Expected left hand side of: %s to be a cast node, got %s",
			code, pp.Sprint(node.Lhs))
	}

	if typ, _ := a.typ(node); typ != floatType {
		t.Errorf("Expected %s to have type float but got type: %s", code, pp.Sprint(typ))
	}
}

func TestInferAssigment(t *testing.T) {
	cases := []struct {
		code string
		typ  types.Type
	}{
		{
			code: "a := 10",
			typ:  intType,
		},
		{
			code: "a := 14.5",
			typ:  floatType,
		},
		{
			code: "a := 10 * 13.5",
			typ:  floatType,
		},
		{
			code: "a := 10 / 2",
			typ:  intType,
		},
	}

	for _, c := range cases {
		node := parser.Parse(c.code).(ast.Assignment)
		typ := ast.Basic{c.typ}
		newNode := a.assigment(&node).(ast.Assignment)

		if newNode.Type != typ {
			t.Errorf("Expected %s to have infer int type but got type: %s", c.code, pp.Sprint(node.Type))
		}
	}
}

func TestAssigment(t *testing.T) {
	code := "i8 a = 123"

	node := parser.Parse(code).(ast.Assignment)
	newNode := a.assigment(&node).(ast.Assignment)

	if _, ok := newNode.Expression.(ast.Cast); !ok {
		t.Errorf("Expected value of: %s to be a cast node, got %s",
			code, pp.Sprint(node.Expression))
	}
}

func TestCall(t *testing.T) {
	code := `
		add :: i32 a, i64 b -> i64 {
			return a + b
		}

		main :: -> i32 {
			return add(10, 243)
		}
	`

	lex := lexer.NewLexer([]byte(code))
	parser := parser.NewParser(lex.Lex())
	ana := NewAnalysis(parser.Parse())
	// irgen := irgen.NewIrgen(ana.Analalize())

	a := ana.Analalize()

	returnSmt := a.Functions[1].Body.Expressions[0].(ast.Return)
	exps := returnSmt.Value.(ast.Call).Arguments

	if _, ok := exps[1].(ast.Integer); !ok {
		t.Errorf("Expected parameter 1 to be a integer got: %s", pp.Sprint(exps[1]))
	}

	if _, ok := exps[0].(ast.Cast); !ok {
		t.Errorf("Expected parameter 0 to be a cast got: %s", pp.Sprint(exps[0]))
	}

}
