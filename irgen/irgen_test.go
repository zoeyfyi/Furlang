package irgen

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	"github.com/bongo227/Furlang/analysis"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/parser"
)

func runIr(ir string) (int, string) {
	// Setup lli to run the llvm ir
	cmd := exec.Command("lli-3.8")
	cmd.Stdin = strings.NewReader(ir)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Run the command
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Extract the return code and return any error message
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

func TestIrgen(t *testing.T) {
	var cases []TestCase

	// Test all
	// files, err := ioutil.ReadDir("../tests/")
	// if err != nil {
	// 	t.Error(err)
	// }

	files := []string{
		"main.fur",
		"i8_type.fur",
		"i16_type.fur",
		"i32_type.fur",
		"i64_type.fur",
		"function.fur",
		"float.fur",
		// "if.fur",
	}

	for _, file := range files {
		c, err := ioutil.ReadFile(fmt.Sprintf("../tests/%s", file))
		if err != nil {
			t.Errorf("Error reading file: %s", err.Error())
		}
		cases = append(cases, TestCase{
			name: file,
			code: string(c),
		})
	}

	for _, c := range cases {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("File: %s\nPanic: %s", c.name, r)
				debug.PrintStack()
			}
		}()

		lexer := lexer.NewLexer([]byte(c.code))
		parser := parser.NewParser(lexer.Lex())
		analysis := analysis.NewAnalysis(parser.Parse())
		gen := NewIrgen(analysis.Analalize())
		llvm := gen.Generate()

		if code, msg := runIr(llvm); code != 123 {
			// Make a more desciptive error message
			t.Errorf("\nFile: %s\nIr:\n%s\nReturn Code: %d\nOut: %s", c.name, llvm, code, msg)
		}
	}
}
