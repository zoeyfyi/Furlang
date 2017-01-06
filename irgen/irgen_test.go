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
			numbers := re.FindAllString(exitErr.Error(), -1)
			if len(numbers) > 0 {
				code, _ := strconv.Atoi(numbers[0])
				return code, out.String()
			}
		}
		log.Fatal(err)
	}

	panic(fmt.Sprintf("No return code, err: %s", out.String()))
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
		"float_equal_to.fur",
		"float_less_than.fur",
		"float_more_than.fur",
		"float_not_equal_to.fur",
		"for_loop.fur",
		"function.fur",
		"i16_type.fur",
		"i32_type.fur",
		"i64_type.fur",
		"i8_type.fur",
		"if.fur",
		"increment_multiple.fur",
		"increment_single.fur",
		"integer_equal_to.fur",
		"integer_more_than.fur",
		"integer_not_equal_to.fur",
		"main.fur",
		"mod_operator.fur",
		"reassignment.fur",
		"returns.fur",
		"rpn.fur",
		"single_if.fur",
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
