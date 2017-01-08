package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/bongo227/Furlang/analysis"
	"github.com/bongo227/Furlang/irgen"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/parser"
	"github.com/davecgh/go-spew/spew"
	"github.com/k0kubun/pp"
)

// Compiler hold infomation about the file to be compiled
type Compiler struct {
	program string

	// Compiler optional flags
	OutputTokens bool
	OutputAst    bool
	NoCompile    bool
}

// New creates a new compiler for the file at filePath
func New(filePath string) (*Compiler, error) {
	if filePath == "" {
		return nil, fmt.Errorf("No input file")
	}

	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Problem reading file '%s'", filePath)
	}
	program := string(data)

	return &Compiler{
		program: program,
	}, nil
}

// Compile compiles the file and writes to the outPath
func (c *Compiler) Compile(buildDirectory string) error {
	// Start compiler timer
	start := time.Now()

	// Run lexer
	l := lexer.NewLexer([]byte(c.program))
	tokens, err := l.Lex()
	if err != nil {
		return err
	}

	// Optionaly write tokens to file
	if c.OutputTokens {
		f, err := os.Create(buildDirectory + "/tokens.txt")
		if err != nil {
			return fmt.Errorf("problem creating tokens file: %s", err.Error())
		}
		defer f.Close()

		for _, t := range tokens {
			f.WriteString(t.String() + "\n")
		}
	}

	// Run parser
	parser := parser.NewParser(tokens)
	ast := parser.Parse()

	// Run analyser
	analyser := analysis.NewAnalysis(ast)
	ast = analyser.Analalize()

	// Optionaly write ast to file (and print it)
	if c.OutputAst {
		f, err := os.Create(buildDirectory + "/ast.txt")
		if err != nil {
			return fmt.Errorf("problem creating ast file: %s", err.Error())
		}
		defer f.Close()

		pp.Print(ast)
		f.WriteString(spew.Sdump(ast))
	}

	// Compile ast to llvm
	if !c.NoCompile {
		ir := irgen.NewIrgen(ast)
		llvm := ir.Generate()
		f, err := os.Create(buildDirectory + "/ben.ll")
		if err != nil {
			return fmt.Errorf("problem creating llvm ir file: %s", err.Error())
		}
		defer f.Close()

		f.WriteString(llvm)
	}

	// Output compiler timings
	fmt.Printf("[Compiled in: %fs]\n", time.Since(start).Seconds())

	return nil
}
