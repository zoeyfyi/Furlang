package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"bitbucket.com/bongo227/furlang/compiler"
	"github.com/fatih/color"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	start := time.Now()
	blue := color.New(color.FgHiBlue).SprintFunc()

	fmt.Printf("%s -> ", blue("Reading input file"))

	// Get file from arguments
	if len(os.Args) <= 1 {
		fmt.Println("\nNo input file")
		return
	}
	in := os.Args[1]

	// Check its a fur file
	matched, err := regexp.MatchString("(\\w)+(\\.)+(fur)", in)
	check(err)
	if !matched {
		fmt.Printf("\nFile '%s' is not a fur file", in)
		return
	}

	// Read file into memory
	data, err := ioutil.ReadFile(in)
	check(err)

	// Create a file to write to
	f, err := os.Create("build/ben.ll")
	check(err)
	defer f.Close()

	// Write the tokens to file
	fmt.Printf("%s -> ", blue("Parcing token"))
	tokens := compiler.ParseTokens(string(data))
	if len(os.Args) == 3 && os.Args[2] == "-tokens" {
		fmt.Println()
		compiler.Dump(tokens, "tokens")
		tokensFile, err := os.Create("build/tokens.txt")
		check(err)
		defer f.Close()

		for _, t := range tokens {
			tokensFile.WriteString(t.String())
		}
	}

	// Create abstract syntax tree
	fmt.Printf("%s -> ", blue("Creating abstract syntax tree"))
	funcs := compiler.Ast(tokens)
	if len(os.Args) == 3 && os.Args[2] == "-ast" {
		compiler.Dump(funcs, "funcs")
	}

	// Compile
	fmt.Printf("%s -> ", blue("Compiling to LLVM"))
	s := compiler.Llvm(funcs)

	fmt.Printf("%s -> ", blue("Writing to file"))
	f.WriteString(s)

	// Confirm the writes
	f.Sync()
	fmt.Printf("%s\n", blue("Done"))
	duration := time.Since(start)
	fmt.Printf("[Compiled in %fs]\n", duration.Seconds())
}
