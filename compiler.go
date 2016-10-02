package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
	// Parse command line flags
	outputTokens := flag.Bool("tokens", false, "Create a file with the tokens")
	outputAst := flag.Bool("ast", false, "Create file with the abstract syntax tree and pretty print it out")
	noCompile := flag.Bool("nocode", false, "Stop the compiler before it generates llvm ir")
	flag.Parse()

	// Start compiler timer
	start := time.Now()

	in := flag.Arg(0)
	if in == "" {
		fmt.Println("No input file")
		return
	}

	data, err := ioutil.ReadFile(in)
	if err != nil {
		fmt.Printf("Problem reading file '%s", in)
		return
	}
	program := string(data)

	tokens := compiler.Lexer(program)
	if *outputTokens {
		f, err := os.Create("build/tokens.txt")
		if err != nil {
			fmt.Printf("Problem creating tokens file: %s\n", err.Error())
		}
		defer f.Close()

		for _, t := range tokens {
			f.WriteString(t.String() + "\n")
		}
	}

	// Create and parse the file
	parser := compiler.NewParser(tokens)
	ast := parser.Parse()

	if *outputAst {
		f, err := os.Create("build/ast.txt")
		if err != nil {
			fmt.Printf("Problem creating ast file: %s\n", err.Error())
		}
		defer f.Close()

		ast.Print()
		ast.Write(f)
	}

	if *noCompile {
		return
	}

	llvm := compiler.Llvm(&ast)
	f, err := os.Create("build/ben.ll")
	if err != nil {
		fmt.Printf("Problem creating llvm ir file: %s\n", err.Error())
	}
	defer f.Close()
	f.WriteString(llvm)

	fmt.Printf("[Compiled in: %fs]\n", time.Since(start).Seconds())
}
