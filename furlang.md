# Furlang: An investigation into modern programming languages and compilers

## Analysis
In this investigation the aim is to design a programming language and implement the basics of a compiler to create executable programs. Due to the time constraints it would be infeasible to implement all asspects of a modern programming language, standard library and language tooling. Instead the focus wull be on implementing a small subset such that simple algorithums like the greatest common divisor and bubble sort can be created.

### Research
In terms of the languages design I looked at several languages with similar goals as mine and read through their specifications including: [Rust<sup>[1]</sup>](#1), [Go<sup>[2]</sup>](#2), [D<sup>[3]</sup>](#3) and [F#<sup>[4]</sup>](#4). I looked at their syntax and the design disicions behind them in order the judge what Fur shoudl look like.

To create the compiler I looked at the source code of these same languages, especially the [Go compiler<sup>[5]</sup>](#5) as it's the language I decided to create Fur's compiler in.

### Syntax
Compared to C modern programming languages use a lot less characters to describe the instructions which make the final program. By using less character it becomes alot faster to read through the source code in order to understand the logic, which intern makes the language easier to use and faster to develop in. With Fur, I wanted to explore the modern ideas and how they can be implemented. 

#### Type inference
In go, most varibles dont need a type because their type can be infered:
```go
foo := 123
```
In this case foo will be an `int` type since the value on the right hand side is an integer. If you want to be more explicit you can use the longer form varible declaration:
```go
var foo int = 123
``` 
The infered type is much quicker to type and just as easy to read, healping to reduce the character count of the source code. 

#### Semicolons and parentiesis
Most languages use semi colons to tell the compiler that the current statement has ended and everyting after the semicolon is interpreted as a seperate statment. In more modern languages like python, their are no semicolons anywhere in the language. Instead python uses spaces and new lines to signify the start and end of each statement and block.

Go uses semicolons only in a few places such as for loops:
```go
for i := 0; i < 10; i++ {}
```
Here each part of the for loop is seperate by semicolons just like C, however for most statements new lines are used as in python to signify the end of a statement.

Another thing to note is the lack of parentiesis around the for loop. The lack of brackets helps to further eliminate useless syntax which only hinders the programmer. The side effect of ommiting semicolons and brackets is that the source code is much more complex to parse since the compiler can assume where statements begin and end.

#### Function definitions
Having looked at the way functions are defined in several languages I decided to create my own syntax inspired by functional programming languages. 
```
proc bar :: int foo -> float
```
First of all what whould normaly be called functions are called procedures in Fur, hence the apprevation `proc`. The double semi colon is used to provide a clear divider between the name and the arguments, this clear line of seperation helps when skimming though the source code in order to find a function with a certain name. Finaly the arrow that seperates the arguments and return type reinforces the consept of a function, to transform the input into output. 

### Memory Managment
When a program needs memory to persist longer than the scope of a function, memory needs to be allocated from the heap. The heap is slower than stack but the program can choose at run-time how much memory it wants. This flexibility brings several problems such as: what if the operating system can't give you the memory you requested, what if you need more, what if the you never give it back. In languages with manual memory management the programmer must solve all these problems whenever they need to allocate memory on the heap, making the code more complex and error prone.

One solution to this problem is a garbage collector. In languages such as Go and Java the run-time allocates blocks of memory, whenever the program requests memory a portion of that block is returned. The run-time then keeps track of where the memory is used, when the portion of memory is no longer in use its is marked and recycled ready for the next allocation. Over time these garbage collectors have been getting faster and much more sophisticated, one consequence of this is that it can be hard to understand their behavior.

The problems arises in applications which have low latency requirements, such as games. With virtual reality and higher refresh rate monitors, games have less than 7 milliseconds to update and render the next frame. GC max pause times in Go are around [50µs<sup>[6]</sup>](#6) (with considerable CPU usage) and [50ms<sup>[7]</sup>](#7) in Java, what's worse is that they can happen at anytime causing the game to freeze and stutter. One workaround is to budget this into your frame time i.e. 5ms to update and render and 2ms for any GC pauses, this means reduced graphics, less realistic physics and simpler game mechanics. Even if you do not pause mid-frame there is still the problem of: higher read/write latency, less CPU performance and less data locality (hence less cache utilization). For this reason Fur will not have a garbage collector.

#### Objectives
 - Programs should be compiled with no runtime managing the executable memory.

### Syntax and keywords
Golang has [25 keywords<sup>[2]</sup>](#2) which helps make it's easy/quick to learn, simple to read and less keywords are reserved so it's simpler to name some constructs. The obvious drawback is the reduced expressiveness compared to languages like C# and Java which have many more keywords. Fur will be closer to Go in terms of the small pool of keywords in favor of a more simpler and clearer expression of logic.

#### Objectives
- Aim for a small amount of keywords so there is no unnecessary complexity.

### Symbols
Functional languages like F# tend to use far more symbols to express their logic. Symbols make smaller code and less visual noise however they create a steeper learning curve since they are less descriptive. Fur will feel familiar sharing most of the symbols from C based languages whilst including some of the expressive power of more functional languages for its functional inspired features.

C++ and Java both have operator overloading which makes their source code easy to read in code bases that use things like arbitrarily wide integers, vectors and matrices. The problem is operator overloading is easily abused by hiding complex computations behind a simple looking operator. For example in the case of a two arbitrary length integers being multiplied, memory allocations and complex loops are invoked which isn't obvious from the call site. Source code should be easy to reason about its performance, hence no operator overloading.

#### Objectives
- Share the symbols from other C like languages so that the language feels instantly familiar.
- Disallow operator overloading so source code is easier to reason about in terms of performance.

### Types
#### Integers
Most C like languages have the standard 8, 16, 32 and 64 bit integer types but they behave differently when going over the maximum value. In C++ there is an ongoing debate over the spec which says integer overflow is undefined behaviour. Compiler authors have been using this fact to implement several micro optimisations however most programmers expect integers to wrap (in fact some algorithms require this to be true). One of Fur’s main goals is to be predictable so integers should wrap like programmers have come to expect.

#### Strings
In C, strings are a sequence of chars that end with a null value. This has been the cause of many bugs in C programs because it's easy to accidentally (or maliciously) modify strings before they are outputted. Most modern languages have made strings immutable, this has several advantages including constant time length look up (in C you would have to transverse the whole string making it linear), reduced vulnerability's from unintended string modifications.

#### Array
Static arrays are almost the same in every programming language, so fur should feel familiar.

#### Slices
Most of the time it is not known how much data a list needs to hold so static arrays are no use. Some modern languages such as Go use slices which are data types with an `index`, `length`, `capacity` and a hidden array. As long as `length < capacity` elements can be appended with no cost. As soon as more space is needed an allocation occurs, expanding the hidden array's capacity. This simple structure is useful for so many different structures including queues and stacks.

#### Structures
Most data is not a single type but a collection of different types. Structures provide a simple way of grouping for ease of use.

#### Objectives
- Standard integer types: 8, 16, 32 and 64 bits
- Integers must wrap around when overflowing
- Standard float types: 32 and 64 bits
- Immutable strings
- Fixed length arrays
- Variable length slices
- Structures

### Compiler executable
When developing code across many platforms programmers often find themselves in constrained environments whether it be limited resources, no graphical interface, no physical device etc. It can be difficult to install the tools required to compile a project which is why Fur should be distributed as a single executable with no install process and no dependencies. The executable would be invoked from the shell in order to compile the project.

This makes managing multiple versions of compilers trivial since you can just invoke different versions of the compiler. On a remote machine as long as you can download or transfer the compiler you can compile on that machine. Finally, developing tooling on top of this such as a version manager should be trivial.

#### Objectives
- Provide a singular executable for the compiler with no install process.
- Executable should be easily obtainable from the project's website and GitHub repository.
- Provide builds for different architectures and operating systems.

## Documented Design

### Overview
```markdown
┌───────────────────────┐
│         Lexer         │
└───────────────────────┘
            ↓
┌───────────────────────┐
│        Parser         │
└───────────────────────┘
            ↓
┌───────────────────────┐
│       Analyzer        │
└───────────────────────┘
            ↓
┌───────────────────────┐   ┌───────────────────────┐
│     IR Generation     │ > │         Goory         │
└───────────────────────┘   └───────────────────────┘
            ↓
┌───────────────────────┐
│   LLVM Optimization   │
│      and linking      │
└───────────────────────┘
```
1. First the lexer parses the source code and converts the characters to tokens. Tokens are the smallest usable pieces of information, like keywords, identifiers, numbers, strings etc. The reason we use tokens is because it's much easier to express logic further down in the pipeline and its alot faster than operating on strings.

2. The parser runs through the list of tokens and creates an abstract syntax tree (AST). This is a representation of the source code which makes the logic of the program much easier to operate on.

3. The analizer infers any allowed omissions such as type inference and implicit casts.

4. Next the AST is turned into LLVM IR, with the help of goory, which is a much lower level language which is used by the LLVM optimizer to make the program run faster.

5. Finally the LLVM optimizer optimizes the IR and the linker links the module with any external modules before producing the final executable.

### Lexer
The lexer accepts the program's source code in the form of a slice of bytes, it then loops over each byte until it reaches the end of the list. At each iteration the lexer switches upon the current character, if it's a number then it must be start of a new number token so the number function is called, if its a double quote it must be the start of a new string token so the string function is called etc. Each of these functions moves the current token forward until it reaches the end for example the other double quote in a string. After the last character of a token is read the token is appended to a list and any white space after that token is skipped over, the lexer is ready to read the beginning of a new character.

### Parser
The parser works in a similar way to the lexer, instead of working on characters, the parser works on tokens. Their are three types of nodes in an abstract syntax tree: expressions, statements and declarations. Expressions are bits of code which represent a value like a `a + b` (this could be an argument for a function, a returned value etc). Statements represent the logic parts of code such as if statements, for loops etc. Statements may also include expressions and declarations as children nodes. Declarations nodes represent function declarations, varible declarations etc.

Whenever the beginning of a construct is met, for example for an if statement the if parser is called which in turn calls the expression parser (for the condition of the if statement) and the block parser as well as any chaining else-if/else statements. Each function parses through the tokens returns a node of the abstract syntax tree. These nodes are the combined in more complex statements to form nodes with many layers of children such as this if statement:

```
IfStatement
  Condition: BinaryExpression
    Left: IdentifierExpression
      Value: "x"
    Operator: ">"
    Right: LiteralExpression
      Value: "3"
  Body
    Statements:
      - ReturnStatement
          Result: LiteralExpression
            Value: TRUE
  Else
    Condition: nil
    Body: BlockStatement
      Statements:
        - ReturnStatement
            Result: LiteralExpression
              Value: TRUE
```

The resulting tree structure is called an abstract syntax tree which organizes the logic of the program into somthing which is much easier to work with.

This parser is an example of a top down precedence parser which runs in linear time. Throughout the project I have implemented several diffrent parser algorithms, trying to find somthing both fast and extendable. Since the syntax changed regulaly throughout the corse of the project it was important to find an algorithm which allowed a large amount of grammer rules, but also run fast since that is one of Fur's design goals.

The idea behind top down precedence parser is that all tokens have a numerical binding power, the tokens with the larger binding power are combined first before tokens with smaller binding powers, for example given the expression `2 + 3 * 4`, we read the tokens until we reach the end. Since `*` > `+` the `3` and `4` are bound first before `2` is bound to `3 * 4`. This simple numeric value allows the parser to understand a large amount of grammer rules.

### Analyser
Again the analyser works similarly, recursing through the AST. Whenever an assignment node is reached it first checks if its type was specified, if not then it infers its type by recursing through the value on the right hand side.

The analyser also checks if the programmer adheared to the language rules, most of the syntax errors in a program will be outputed from here.

Final the analyser also handles implicit casting. For example, if a function returns a `i32` (32 bit integer), but the return statement has a `i64`, the analyser will insert a cast node which the ir generator will turn into an integer truncation. The automatic cast insertion is also used for almost all other statements including ifs, calls etc.

### IR generation
Once more the the AST is recursed through until it reaches a child with no children. We then return the value of an in memory representation of the node produced by goory (a separate library for writing LLVM IR). More complex nodes use these values to return their own IR nodes until all constructs have been translated. Finally the root node is transformed into a string of LLVM IR.

#### Goory
Goory is a library (created as part of the project) for producing LLVM IR from AST’s. It has its own internal AST which is what it uses to write the LLVM IR string after the AST is parsed. To translate between the programs AST and goory's internal tree goory exposes methods on its nodes which allows you to attach children.

For example given the following pseudo code:
```markdown
FUNCTION fpadd(a, b)
	RETURN a + b
ENDFUNCTION
```

Would be translate it to llvm ir with the following:
```go
// Create a new double type
double := goory.DoubleType()

// Create a new llvm module named test
module := goory.NewModule("test")

// Create a new function node on the module node which returns a double
function := module.NewFunction("fpadd", double)

// Add argument nodes to the function node with types double
a := function.AddArgument(double, "a")
b := function.AddArgument(double, "b")

// Get the entry block of the function
block := function.Entry()

// Append a floating point add node to the block
result := block.Fadd(a, b)

// Append a return node with the result of the floating point add instruction
block.Ret(result)

// Gets the llvm ir for the "test" module
module.LLVM()
```
## Technical Solution
Insert code here...

## Testing

### Lexer
The input represents a sample of source code and the expected output is a list of tokens the lexer is expected to produce, if the lexer tokens matches the expected output then the test passed, else it failed.

Input | Expected | Result
--- | --- | ---
`"foo"` | `Token{IDENT, "foo", 1, 1}`, `Token{SEMICOLON, "\n", 1, 4}` | Pass
`"100 + 23"` | `Token{INT, "100", 1, 1}`, `Token{ADD, "", 1, 5}`, `Token{INT, "23", 1, 7}`, `Token{SEMICOLON, "\n", 1, 9}` | Pass


## Evaluation
Insert evaluation here...

## References
1. The Rust Reference <a id="1">https://doc.rust-lang.org/reference.html</a>
2. The Go Programming Language Specification <a id="2">https://golang.org/ref/spec</a>
3. Specification for the D Programming Language <a id="3">https://dlang.org/spec/spec.html</a>
4. The F# Language Specification <a id="4">http://fsharp.org/specs/language-spec/</a>
5. Go's GitHub Repository <a id="5">https://github.com/golang/go</a>
6. Go - Proposal: Eliminate STW stack re-scanning <a id="6">https://golang.org/design/17503-eliminate-rescan</a>
7. Plumber - G1 vs CMS vs Parallel GC <a id="7">https://plumbr.eu/blog/garbage-collection/g1-vs-cms-vs-parallel-gc</a>
