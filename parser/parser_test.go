package parser

import (
	"testing"

	"github.com/bongo227/Furlang/lexer"
	"github.com/davecgh/go-spew/spew"
)

func TestParser(t *testing.T) {
	lexer := lexer.NewLexer([]byte("call(ben(1 + 5), 3)"))
	parser := NewParser(lexer.Lex())

	spew.Dump(parser.Maths())
	t.Fail()
}
