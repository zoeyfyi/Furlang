package irgen

import (
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/parser"
)

func getReturnCode(ir string) (int, string) {
	example := exec.Command("lli-3.8")

	var out bytes.Buffer
	example.Stdin = strings.NewReader(ir)
	example.Stdout = &out
	example.Stderr = &out

	if err := example.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			re := regexp.MustCompile("[0-9]+")
			code, _ := strconv.Atoi(re.FindAllString(exitErr.Error(), -1)[0])
			return code, out.String()
		}
		log.Fatal(err)
	}

	panic("No return code")
}

func TestIrgen(t *testing.T) {

	cases := []struct {
		code string
	}{
		{
			code: `
				main :: -> int {
					return 123
				}
			`,
		},
	}

	for _, c := range cases {
		lexer := lexer.NewLexer([]byte(c.code))
		parser := parser.NewParser(lexer.Lex())
		gen := NewIrgen(parser.Parse())
		llvm := gen.Generate()

		if code, msg := getReturnCode(llvm); code != 123 {
			t.Errorf("Expected: 123, Got: %d", code)
			t.Log(msg)
		}
	}
}
