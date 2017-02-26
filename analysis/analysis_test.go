package analysis

import (
	"log"
	"testing"

	"reflect"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/parser"
	"github.com/bongo227/Furlang/types"
	"github.com/k0kubun/pp"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

var a = &Analysis{}

func TestFloatPromotion(t *testing.T) {
	cases := []struct {
		code string
	}{
		{"10 + 14.5"},
		{"3.5 + 11"},
	}

	for _, c := range cases {
		node, err := parser.ParseExpression(c.code)
		if err != nil {
			t.Error(err)
		}

		binNode, ok := node.(*ast.BinaryExpression)
		if !ok {
			t.Errorf("Expected node to be of type \"*ast.BinaryExpression\", got %q",
				reflect.TypeOf(node).String())
		}

		anaNode, ok := a.binaryExp(binNode).(*ast.BinaryExpression)
		if !ok {
			t.Errorf("Expected analysed node to be of type \"*ast.BinaryExpression\", got: %q",
				reflect.TypeOf(anaNode).String())
		}

		var anaCastNode *ast.CastExpression
		if a.typ(binNode.Left) != floatType {
			anaCastNode, ok = anaNode.Left.(*ast.CastExpression)
			if !ok {
				t.Errorf("Expected left hand side of analysed node to be a cast node, got %q",
					reflect.TypeOf(anaNode.Left).String())
			}
		} else {
			anaCastNode, ok = anaNode.Right.(*ast.CastExpression)
			if !ok {
				t.Errorf("Expected right hand side of analysed node to be a cast node, got %q",
					reflect.TypeOf(anaNode.Right).String())
			}
		}

		if anaCastNode.Type != floatType {
			t.Errorf("Expected cast node to have type \"float\", got %q",
				reflect.TypeOf(anaCastNode.Type).String())
		}

		if typ := a.typ(anaNode); typ != floatType {
			t.Errorf("Expected type of analysed node to be \"float\", got %q",
				reflect.TypeOf(typ).String())
		}
	}
}

func TestInferAssigment(t *testing.T) {
	cases := []struct {
		code string
		typ  types.Type
	}{
		{"a := 10", intType},
		{"a := 14.5", floatType},
		{"a := 10 * 13.5", floatType},
		{"a := 10 / 2", intType},
		{"a := int[2]{0, 0}", types.NewArray(intType, 2)},
	}

	for _, c := range cases {
		node, err := parser.ParseDeclaration(c.code)
		if err != nil {
			t.Error(err)
		}

		dclNode, ok := node.(*ast.VaribleDeclaration)
		if !ok {
			t.Errorf("Expected node to have type assigment, got type %q",
				reflect.TypeOf(dclNode).String())
		}

		anaNode := a.varibleDcl(dclNode)
		anaVaribleNode, ok := anaNode.(*ast.VaribleDeclaration)
		if !ok {
			t.Errorf("Expected analysed node to have type \"*ast.VaribleDeclaration\", got %q",
				reflect.TypeOf(anaNode).String())
		}

		if !reflect.DeepEqual(anaVaribleNode.Type, c.typ) {
			// TODO: give types.Type a string method so we dont have to reflect
			t.Errorf("Expected varible node to have type %q but, got type %q",
				c.typ.String(), anaVaribleNode.Type.String())
		}
	}
}

func TestVaribleDeclare(t *testing.T) {
	code := "i8 a = 123"

	node, err := parser.ParseStatement(code)
	if err != nil {
		t.Error(err)
	}

	dclNode, ok := node.(*ast.DeclareStatement)
	if !ok {
		t.Errorf("Expected node to have type \"*ast.DeclareStatement\", got type %q",
			reflect.TypeOf(dclNode).String())
	}

	varNode, ok := dclNode.Statement.(*ast.VaribleDeclaration)
	if !ok {
		t.Errorf("Expected declaration statment to have type \"*ast.VaribleDeclaration\", got %q",
			reflect.TypeOf(dclNode.Statement).String())
	}

	anaNode, ok := a.declare(varNode).(*ast.VaribleDeclaration)
	if !ok {
		t.Errorf("Expected type of analysed declaration statment to have type \"*ast.VaribleDeclaration\", got %q",
			reflect.TypeOf(anaNode).String())
	}

	if _, ok := anaNode.Value.(*ast.CastExpression); !ok {
		t.Errorf("Expected value of analysed node to be a cast node, got %q",
			reflect.TypeOf(anaNode).String())
	}
}

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

	parser := parser.NewParser(tokens, true)
	tree := parser.Parse()

	ana := NewAnalysis(tree)
	a := ana.Analalize()

	firstSmt := a.Functions[1].Body.Statements[0]
	returnSmt, ok := firstSmt.(*ast.ReturnStatement)
	if !ok {
		t.Errorf("Expected first expression to be a return statement, got %q",
			reflect.TypeOf(firstSmt).String())
	}

	cast, ok := returnSmt.Result.(*ast.CastExpression)
	if !ok {
		t.Errorf("Expected return value to be of type \"*ast.CastExpression\", got %q",
			reflect.TypeOf(returnSmt.Result).String())
	}

	call := cast.Expression.(*ast.CallExpression)
	if !ok {
		t.Errorf("Expected casted return value to be of type \"ast.CallExpression\", got %q",
			reflect.TypeOf(cast.Expression).String())
	}

	if _, ok := call.Arguments.Elements[1].(*ast.LiteralExpression); !ok {
		t.Errorf("Expected parameter 1 to be a integer got %s",
			pp.Sprint(call.Arguments.Elements[1]))
	}

	if _, ok := call.Arguments.Elements[0].(*ast.CastExpression); !ok {
		t.Errorf("Expected parameter 0 to be a cast got %s",
			pp.Sprint(call.Arguments.Elements[0]))
	}
}

func TestCast(t *testing.T) {
	cases := []struct {
		source string
		typ    types.Type
	}{
		{`int(132)`, types.IntType(0)},
		{`i8(234)`, types.IntType(8)},
		{`i16(13)`, types.IntType(16)},
		{`i32(5)`, types.IntType(32)},
		{`i64(1415)`, types.IntType(64)},
		{`float(241)`, types.FloatType(0)},
		{`f32(1231)`, types.FloatType(32)},
		{`f64(21)`, types.FloatType(64)},
	}

	for _, c := range cases {
		exp, err := parser.ParseExpression(c.source)
		if err != nil {
			t.Error(err)
		}

		anaExp := a.expression(exp)

		castExp, ok := anaExp.(*ast.CastExpression)
		if !ok {
			t.Errorf("Expected \"*ast.CastExpression\", got %q",
				reflect.TypeOf(anaExp).String())
		}

		if !reflect.DeepEqual(castExp.Type, c.typ) {
			t.Errorf("Expected cast type to be %q, got %q", c.typ.String(), castExp.Type.String())
		}
	}
}
