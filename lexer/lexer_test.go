package lexer

import (
	"testing"

	"reflect"
)

func TestLex(t *testing.T) {
	cases := []struct {
		input    string
		expected []Token
	}{
		{
			input: `foo`,
			expected: []Token{
				Token{IDENT, "foo", 1, 1},
				Token{SEMICOLON, "\n", 1, 4},
			},
		},
		{
			input: `foo bar`,
			expected: []Token{
				Token{IDENT, "foo", 1, 1},
				Token{IDENT, "bar", 1, 5},
				Token{SEMICOLON, "\n", 1, 8},
			},
		},
		{
			input: `123`,
			expected: []Token{
				Token{INT, "123", 1, 1},
				Token{SEMICOLON, "\n", 1, 4},
			},
		},
		{
			input: `100 + 23`,
			expected: []Token{
				Token{INT, "100", 1, 1},
				Token{ADD, "", 1, 5},
				Token{INT, "23", 1, 7},
				Token{SEMICOLON, "\n", 1, 9},
			},
		},
		{
			input: `321.123`,
			expected: []Token{
				Token{FLOAT, "321.123", 1, 1},
				Token{SEMICOLON, "\n", 1, 8},
			},
		},
		{
			input: `int a -> int`,
			expected: []Token{
				Token{IDENT, "int", 1, 1},
				Token{IDENT, "a", 1, 5},
				Token{ARROW, "", 1, 7},
				Token{IDENT, "int", 1, 10},
				Token{SEMICOLON, "\n", 1, 13},
			},
		},
		{
			input: `1
2
3`,
			expected: []Token{
				Token{INT, "1", 1, 1},
				Token{SEMICOLON, "\n", 1, 2},
				Token{INT, "2", 2, 1},
				Token{SEMICOLON, "\n", 2, 2},
				Token{INT, "3", 3, 1},
				Token{SEMICOLON, "\n", 3, 2},
			},
		},
	}

	for _, c := range cases {
		l := NewLexer([]byte(c.input))
		got, err := l.Lex()

		if err != nil {
			t.Errorf("Lexer errored: %s", err.Error())
		}

		if !reflect.DeepEqual(c.expected, got) {
			t.Log("Expected: ")
			for _, tok := range c.expected {
				t.Log(tok.String())
			}
			t.Log("Got")
			for _, tok := range got {
				t.Log(tok.String())
			}
			t.Fail()
		}
	}
}
