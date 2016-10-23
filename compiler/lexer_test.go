package compiler

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bongo227/dprint"
)

type lexerTest struct {
	in   string
	want []token
}

func TestLexer(t *testing.T) {
	cases := []lexerTest{
		// Token name type
		{
			"name",
			[]token{
				token{tokenName, "name", 1, 1, 4},
			},
		},

		// Spaces test
		{
			"name1 name2 name3",
			[]token{
				token{tokenName, "name1", 1, 1, 5},
				token{tokenName, "name2", 1, 7, 5},
				token{tokenName, "name3", 1, 13, 5},
			},
		},

		// Number test
		{
			"123 41764",
			[]token{
				token{tokenNumber, 123, 1, 1, 3},
				token{tokenNumber, 41764, 1, 5, 5},
			},
		},
	}

	for _, c := range cases {
		got := lexer(c.in)

		if !reflect.DeepEqual(got, c.want) {
			fmt.Println("Expected: ")
			dprint.Dump(c.want)
			fmt.Println("Got: ")
			dprint.Dump(got)
			t.Fail()
			// t.Errorf("lexer(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
