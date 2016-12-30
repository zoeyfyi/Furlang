package irgen

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

type TestCase struct {
	name string
	code string
}

type FI struct {
	name string
}

func (f *FI) Name() string { return f.name }

func TestIrgen(t *testing.T) {
	var cases []TestCase

	// Test all
	// files, err := ioutil.ReadDir("../tests/")
	// if err != nil {
	// 	t.Error(err)
	// }

	files := []FI{
		FI{"main.fur"},
	}

	for _, file := range files {
		fmt.Println(file.Name())
		c, err := ioutil.ReadFile(fmt.Sprintf("../tests/%s", file.Name()))
		if err != nil {
			t.Errorf("Error reading file: %s", err.Error())
		}
		cases = append(cases, TestCase{
			name: file.Name(),
			code: string(c),
		})
	}

	for _, c := range cases {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("File: %s\nPanic: %s", c.name, r)
			}
		}()

		lexer := lexer.NewLexer([]byte(c.code))
		parser := parser.NewParser(lexer.Lex())
		gen := NewIrgen(parser.Parse())
		llvm := gen.Generate()

		if code, msg := getReturnCode(llvm); code != 123 {
			// Make a more desciptive error message
			t.Errorf("File: %s\nReturn Code: %d\nOut: %s", c.name, code, msg)
		}
	}
}
