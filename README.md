# go-compiler

Compiler steps
1. Parse through file and convert single characters and names to a list of tokens
2. Group some single tokens and convert names to other tokens (i.e. types, returns etc)
3. Generate an abstract syntax tree from the list of tokens
4. Infer types of assignments in abstract syntax tree
5. Validate abstract syntax tree
6. Compile abstract syntax tree to llvm

Lexer -> 