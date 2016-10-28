package lexer

import (
	"testing"

	"reflect"

	"github.com/bongo227/dprint"
)

func TestLex(t *testing.T) {
	cases := []struct {
		input    string
		expected []Token
	}{
		{
			input: "foo",
			expected: []Token{
				Token{IDENT, Position{1, 1, 3}, "foo"},
			},
		},

		{
			input: "123 == 321",
			expected: []Token{
				Token{INTVALUE, Position{1, 1, 3}, 123},
				Token{EQUAL, Position{1, 5, 2}, nil},
				Token{INTVALUE, Position{1, 8, 3}, 321},
			},
		},

		{
			input: "i32 ben := 23",
			expected: []Token{
				Token{TYPE, Position{1, 1, 3}, INT32},
				Token{IDENT, Position{1, 5, 3}, "ben"},
				Token{INFERASSIGN, Position{1, 9, 2}, nil},
				Token{INTVALUE, Position{1, 12, 2}, 23},
			},
		},
	}

	for _, c := range cases {
		l := NewLexer(c.input)
		if !reflect.DeepEqual(c.expected, l.Lex()) {
			t.Log("Expected: ")
			dprint.Dump(c.expected)
			t.Log("Got")
			dprint.Dump(l.Lex())
			t.Fail()
		}
	}
}
