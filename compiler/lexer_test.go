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

func TestLexerTypeStrings(t *testing.T) {
	cases := []struct {
		tokenType       int
		tokenTypeString string
	}{
		{tokenArrow, "tokenArrow"},
		{tokenAssign, "tokenAssign"},
		{tokenCloseBody, "tokenCloseBody"},
		{tokenComma, "tokenComma"},
		{tokenDoubleColon, "tokenDoubleColon"},
		{tokenInt32, "tokenInt32"},
		{tokenFloat32, "tokenFloat32"},
		{tokenName, "tokenName"},
		{tokenNewLine, "tokenNewLine"},
		{tokenNumber, "tokenNumber"},
		{tokenFloat, "tokenFloat"},
		{tokenOpenBody, "tokenOpenBody"},
		{tokenPlus, "tokenPlus"},
		{tokenMinus, "tokenMinus"},
		{tokenMultiply, "tokenMultiply"},
		{tokenFloatDivide, "tokenFloatDivide"},
		{tokenIntDivide, "tokenIntDivide"},
		{tokenOpenBracket, "tokenOpenBracket"},
		{tokenCloseBracket, "tokenCloseBracket"},
		{tokenReturn, "tokenReturn"},
		{tokenIf, "tokenIf"},
		{tokenElse, "tokenElse"},
		{tokenTrue, "tokenTrue"},
		{tokenFalse, "tokenFalse"},
		{tokenType, "tokenType"},
	}

	for _, c := range cases {
		expected := c.tokenTypeString
		got := tokenTypeString(c.tokenType)

		if got != expected {
			t.Errorf("Expected: %s, Got: %s", expected, got)
		}
	}
}

func TestLexerTokenStrings(t *testing.T) {
	cases := []struct {
		toke     token
		expected string
	}{
		{
			token{tokenIf, "if", 12, 22, 2},
			"tokenIf, line: 12, column: 22",
		},
		{
			token{tokenNumber, "123", 8, 12, 3},
			"tokenNumber, line: 8, column: 12",
		},
	}

	for _, c := range cases {
		if c.toke.String() != c.expected {
			t.Errorf("Expected: %s, Got: %s", c.expected, c.toke.String())
		}
	}
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

		// Float test
		{
			"12.2 1214.5",
			[]token{
				token{tokenFloat, float32(12.2), 1, 1, 4},
				token{tokenFloat, float32(1214.5), 1, 6, 6},
			},
		},

		// Types
		{
			"i32 a",
			[]token{
				token{tokenType, typeInt32, 1, 1, 3},
				token{tokenName, "a", 1, 5, 1},
			},
		},

		// Control name
		{
			"if return",
			[]token{
				token{tokenIf, "if", 1, 1, 2},
				token{tokenReturn, "return", 1, 4, 6},
			},
		},

		// Symbols
		{
			"+ - / *",
			[]token{
				token{tokenPlus, "+", 1, 1, 1},
				token{tokenMinus, "-", 1, 3, 1},
				token{tokenFloatDivide, "/", 1, 5, 1},
				token{tokenMultiply, "*", 1, 7, 1},
			},
		},

		// Multi lines
		{
			"ben\nbob",
			[]token{
				token{tokenName, "ben", 1, 1, 3},
				token{tokenNewLine, "\n", 1, 4, 1},
				token{tokenName, "bob", 2, 1, 3},
			},
		},

		// Multi tokens
		{
			"++",
			[]token{
				token{tokenIncrement, nil, 1, 1, 2},
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
