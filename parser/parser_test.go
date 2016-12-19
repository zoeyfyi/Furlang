package parser

import (
	"testing"

	"github.com/bongo227/Furlang/lexer"
	"github.com/davecgh/go-spew/spew"
)

func TestParser(t *testing.T) {
	lexer := lexer.NewLexer([]byte("(1 + 3) * 6"))
	parser := NewParser(lexer.Lex())

	spew.Dump(parser.Expression())
	t.Fail()
}
