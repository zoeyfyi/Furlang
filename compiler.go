package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"bitbucket.com/bongo227/furlang/compiler"
	"github.com/fatih/color"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func step(name string) {
	blue := color.New(color.FgHiBlue).SprintFunc()
	fmt.Printf("%s -> ", blue(name))
}

func main() {
	cerror := color.New(color.FgHiRed).SprintfFunc()

	// Parse command line flags
	outputTokens := flag.Bool("tokens", false, "Create a file with the tokens")
	outputAst := flag.Bool("ast", false, "Create file with the abstract syntax tree")
	flag.Parse()

	// Start compiler timer
	start := time.Now()

	// Get file from arguments
	step("Reading input file")
	in := flag.Arg(0)
	if in == "" {
		fmt.Println("\nNo input file")
		return
	}

	// Check its a fur file
	matched, err := regexp.MatchString("(\\w)+(\\.)+(fur)", in)
	if err != nil || !matched {
		fmt.Printf("\nFile '%s' is not a fur file", in)
		return
	}

	// Read file into memory
	data, err := ioutil.ReadFile(in)
	program := string(data)
	if err != nil {
		fmt.Printf("\nProblem reading file '%s", in)
		return
	}

	// Create a file to write to
	f, err := os.Create("build/ben.ll")
	if err != nil {
		fmt.Printf("\nProblem creating build file")
		return
	}
	defer f.Close()

	// Write the tokens to file
	if *outputTokens {
		step("Writing tokens to file")
		tokensFile, err := os.Create("build/tokens.txt")
		if err != nil {
			fmt.Printf("Problem creating token file")
			return
		}
		defer f.Close()

		tokens := compiler.Tokens(program)
		tokensFile.WriteString(tokens)
	}

	// Create abstract syntax tree
	if *outputAst {
		step("Printing abstract syntax tree")
		fmt.Print("\n")
		compiler.AbstractSyntaxTree(program)
	}

	// Compile
	step("Compiling")
	s, err := compiler.Compile(program)
	if err != nil {
		if err, ok := err.(compiler.Error); ok {
			line := err.Line()
			clow, chigh := err.ColumnRange()
			lines := strings.Split(program, "\n")
			fmt.Printf("\n%s\n%s\n%s%s\n", cerror("Error: %s", err.Error()), lines[line], strings.Repeat(" ", clow), cerror(strings.Repeat("^", chigh)))
		} else {
			panic("Unexpected error type")
		}

		return
	}

	step("Writing to file")
	f.WriteString(s)

	// Confirm the writes
	f.Sync()
	step("Done")

	// Print compiler statistics
	duration := time.Since(start)
	fmt.Printf("[Compiled in %fs]\n", duration.Seconds())
}
