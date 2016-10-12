package main

import (
	"flag"
	"fmt"

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
	buildDirectory := flag.String("builddir", "build", "Directory any files create in the compile processes should be created")
	flag.Parse()

	path := flag.Arg(0)
	comp, err := compiler.New(path)
	if err != nil {
		fmt.Println(err)
	}

	comp.OutputTokens = *outputTokens
	comp.OutputAst = *outputAst
	comp.NoCompile = *noCompile

	if err = comp.Compile(*buildDirectory); err != nil {
		fmt.Println(err)
	}
}
