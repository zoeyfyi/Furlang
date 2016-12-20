package parser

import (
	"fmt"
	"testing"

	"github.com/bongo227/Furlang/lexer"
	"github.com/k0kubun/pp"
)

func TestParser(t *testing.T) {
	lexer := lexer.NewLexer([]byte(`{
		print(1, 3)
		1 + 3
	}`))
	parser := NewParser(lexer.Lex())

	fmt.Println("==== Source ====")
	for _, t := range parser.tokens {
		fmt.Println(t.String())
	}
	fmt.Println("================")

	pp.Print(parser.Expression())
	t.Fail()
}
