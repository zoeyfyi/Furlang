package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Token types
const (
	BLOCK_BEGIN = "BLOCK BEGIN"
	BLOCK_END   = "BLOCK END"

	FUNCTION   = "FUNCTION"
	ASSIGNMENT = "ASSINGMENT"

	COMMA         = "COMMA"
	BRACKET_OPEN  = "BRACKET_OPEN"
	BRACKET_CLOSE = "BRACKET_CLOSE"
	ARROW         = "ARROW"
	NEWLINE       = "NEWLINE"
	ADDITION      = "ADDITION"

	NAME = "NAME"
	TYPE = "TYPE"

	NUMBER = "NUMBER"

	UNDEFIND = "UNDEFIND"
)

type Token struct {
	ttype string
	value string
}

func tokenize(in string) []Token {
	out := make([]Token, 1000)

	i := 0
	for _, line := range strings.Split(in, "\n") {
		for _, word := range strings.Split(line, " ") {
			if word == "" {
				continue
			}

			tokenType := UNDEFIND
			switch word {
			case "{":
				tokenType = BLOCK_BEGIN
			case "}":
				tokenType = BLOCK_END
			case "::":
				tokenType = FUNCTION
			case ":=":
				tokenType = ASSIGNMENT
			case ",":
				tokenType = COMMA
			case "(":
				tokenType = BRACKET_OPEN
			case ")":
				tokenType = BRACKET_CLOSE
			case "->":
				tokenType = ARROW
			case "+":
				tokenType = ADDITION
			}

			if tokenType == UNDEFIND {
				if word == "int" || word == "float" {
					tokenType = TYPE
				} else if _, err := strconv.ParseInt(word, 10, 0); err == nil {
					tokenType = NUMBER
				} else {
					tokenType = NAME
				}
			}

			out[i] = Token{
				ttype: tokenType,
				value: word,
			}

			i++
		}

		out[i] = Token{
			ttype: NEWLINE,
			value: "\n",
		}

		i++
	}

	return out
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("No input file")
		return
	}

	in := os.Args[1]
	matched, err := regexp.MatchString("(\\w)+(\\.)+(fur)", in)
	check(err)

	if !matched {
		fmt.Printf("File '%s' is not a fur file\n", in)
		return
	}

	data, err := ioutil.ReadFile(in)
	check(err)

	f, err := os.Create("ben")
	check(err)
	defer f.Close()

	tokens := tokenize(string(data))
	for _, token := range tokens {
		if token.ttype == "" {
			return
		}
		f.WriteString(token.ttype + " -- " + strconv.Quote(token.value) + "\n")
	}

	f.Sync()
}
