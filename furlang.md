#Furlang: An investigation into modern programming languages and compilers

----------
##Analysis
Until recently most game development was done in C/C++ because there were no real alternatives for game engines seeking to maximize the hardware. Most modern languages are designed with a garbage collector to prevent the problems which arise with manual memory management, and object oriented patterns which aim to encapsulate state to make the program easier to reason about. What's become apparent is that these features have excluded modern languages from serious widespread use in high performance code bases, and their is a gap for programming language with modern ideas without the heavy drawbacks from the modern alternatives. 

In this investigation the aim is to design a modern programming language and implement a subset in the form of a compiler. Due to time constraints it would be infeasible to implement all aspects of a modern programming language, standard libraries and language tooling. Instead the focus will be on implementing a subset such that a programmer has the ability to create simple algorithms such as bubble sort in Fur.

###Research
In terms of the languages design I looked at several languages with similar goals as mine and read through their specifications including: [Rust<sup>[1]</sup>](#1), [Go<sup>[2]</sup>](#2), [D<sup>[3]</sup>](#3) and [F#<sup>[4]</sup>](#4). I used my prior experience with these languages to judge whether their features would be beneficial in Fur. 

To create the compiler I looked at the source code of these same languages, especially the [Go compiler<sup>[5]</sup>](#5) as it's the language I decided to create Fur's compiler in. 

###Memory Managment
When a program needs memory to persist longer than the scope of a function, memory needs to be allocated from the heap. The heap is slower than stack but the program can choose at run-time how much memory it wants. This flexibility brings several problems such as: what if the operating system can't give you the memory you requested, what if you need more, what if the you never give it back. In languages with manual memory management the programmer must solve all these problems whenever they need to allocate memory on the heap, making the code more complex and error prone. 

One solution to this problem is a garbage collector. In languages such as Go and Java the run-time allocates blocks of memory, whenever the program requests memory a portion of that block is returned. The run-time then keeps track of where the memory is used, when the portion of memory is no longer in use its is marked and recycled ready for the next allocation. Over time these garbage collectors have been getting faster and much more sophisticated, one consequence of this is that it can be hard to understand their behavior. 

The problems arises in applications which have low latency requirements, such as games. With virtual reality and higher refresh rate monitors, games have less than 7 milliseconds to update and render the next frame. GC max pause times in Go are around [50µs<sup>[6]</sup>](#6) (with considerable CPU usage) and [50ms<sup>[7]</sup>](#7) in Java, what's worse is that they can happen at anytime causing the game to freeze and stutter. One workaround is to budget this into your frame time i.e. 5ms to update and render and 2ms for any GC pauses, this means reduced graphics, less realistic physics and simpler game mechanics. Even if you do not pause mid-frame theirs still the problem of: higher read/write latency, less CPU performance and less data locality (hence less cache utilization). For this reason Fur will not have a garbage collector.

####Objectives
 - Simple and clear procedure for allocating memory on the heap that is flexible enough to derive all other heap structures such as lists, trees, maps etc.

###Object oriented programming
Often in programming we deal with real life objects, in the context of a game this could be the humans and animals you interact with, the plants/decorations you see and tools/weapons you use. Object oriented programming is all about representing the state of these objects and things we can do with them, for example let's say we have 3 different weapons:

Weapon | State | Actions
--- | --- | ---
Normal gun | Ammunition left, firing rate | Aim, fire
Burst gun | Ammunition left, firing rate, Burst amount | Aim, fire, Burst fire
Laser gun | Cooldown time | Auto aim, Repeated fire

Each weapon has some state and some actions. You may notice that the burst gun has all the properties and actions of the normal gun, so we can say the burst gun inherits from the normal gun. In the OOP paradigm this mean we don't don't need to write two identical aim and fire methods, we create one normal gun and the burst gun inherits the state and actions of the normal gun.

Where object oriented programming (OOP) breaks down is when we change the behavior of an object. Say are game requires a laser gun, a laser gun doesn't have ammunition, it has a cool-down rate instead. A laser gun is still a gun though so we need to make a generic gun with no state or actions which both the normal and laser gun inherits from. Overtime as new ideas are implemented, inheritance trees becomes unwieldy as different behaviors are required. What started of as a way of preventing code duplication can cause exponential growth in the code base. 

Modern game engines have recognized this problem and developed the component based approach. Small bits of data and behaviors are bundled into components, and entities are just a list of components. This approach made objects in the game infinitely flexible, write the logic once and add it to any object in the game. If all the same components are stored sequentially updating them becomes extremely fast since they will almost always be in cache. 

In conclusion, Fur doesn't need the many language features required for effective OOP because it isn't a good choice for modern games development. The data orientated approach makes better use of the cache for a better performing game. 

####Objectives
 - No explicit support for object oriented programming because it would be useless complexity for the end user. 
 - Standard library should reflect this ideology with no oop behavior.

###Functional programming 
Functional programming seems like a natural fit for the composable nature of logic in video games however not all the features of functional programming are performant. There's a reason that almost all functional programming languages have a garbage collector, their design causes a lot of allocations. Their immutable structures means a manual memory allocator is forced to copy the existing memory to a new place in memory to make changes to the values instead of applying the modifications in place. This can mean massive performance penalties if the programmer does not rewrite or redesign the logic such that no references to old structures are needed and a smarter compiler/allocator can reason that a modification in place doesn't break the rules of the language. Most of the time theses transforms are non trivial and require a large amount of knowledge of the compilers behavior, slowing programmers down and reducing the hiring pool. Obviously some functional programming idea are not compatible with Fur’s performance targets. 

With that said, pure functions and function composition are features that allow programmers to compose small bits of logic to form something more complex. They can be very useful and can be implemented at no additional cost to performance. Pure functions by the nature can be ran in parallel presenting a unique opportunity for the compiler to make faster code than it otherwise could. Procedural language compilers could also use the same performance boost however with explicit language support the compiler can make more optimizations in more places resulting in better CPU usage and faster execution. 

####Objectives
 - The language should differentiate between pure functions and procedures, where pure functions are a subclass of procedures.
 - The compiler should ensure pure functions cause no side effects by only allowing calls to other pure functions and not allowing mutations to function parameters and global variables.

###Modules and packages
In non trivial code bases portions of code tend to split nicely into different modules. This code separation allows multiple programmers to work on the code base with greater ease. Additionally when code bases get so large that it's impossible for each person to memorize the whole code base, the separation of code allows the programmers to control what parts of the module external code can use. This idea of public/private constructs forms the modules interface which is hopefully easier to reason with for someone that is using the module externally. 

Most languages approach this in different ways, but with Fur’s design goals in mind a [Go style module system](#2) with modules as folders seems to be the simplest module system which is compatible for all projects. The benefit of modules as folders is that it allows anyone browsing the code base to easily build a mental map of the project's structure regardless of where they are whether it be their editor, file explorer or source control. 

####Objectives
- Modules should be defined by their folder structure where the direct children of a folder represents the code in the module and sub folders represent sub modules.
- Modules should be imported by their relative file path. Only exported constructs should be accessible from modules that import them.

###Syntax and keywords
Golang has [25 keywords<sup>[2]</sup>](#2) which helps make it's easy/quick to learn, simple to read and less keywords are reserved so it's simpler to name some constructs. The obvious drawback is the reduced expressiveness compared to languages like C# and Java which have many more keywords. Fur will be closer in go in terms of the small pool of keywords in favor of a more simpler and clearer expression of logic. 

Functional languages like F# tend to use far more symbols to express their logic. Symbols make smaller code and less visual noise however they create a steeper learning curve since they are less descriptive. Fur will feel familiar sharing most of the symbols from C based languages whilst including some of the expressive power of more functional languages for its functional inspired features. 

C++ and Java both have operator overloading which makes their source code easy to read in code bases that use things like arbitrarily wide integers, vectors and matrices. The problem is operator overloading is easily abused by hiding complex computations behind a simple looking operator. For example in the case of a two arbitrary length integers being multiplied, memory allocations and complex loops are invoked which isn't obvious from the call site. Since performance is a priority for Fur the source code should be easy to reason about its performance, hence no operator overloading.

####Objectives
- Aim for a small amount of keywords so their is no unnecessary complexity.
- Share the symbols from other C like languages so that the language feels instantly familiar.
- Disallow operator overloading so source code is easier to reason about in terms of performance.
- Pure functions should be easily composed to create more complex behavior.

###Types
####Integers
Most C like languages have the standard 8, 16, 32 and 64 bit integer types but the behave differently when going over the maximum value. In C++ there is an ongoing debate over the spec which says integer overflow is undefined behaviour. Compiler authors have been using this fact to implement several micro optimisations however most programmers expect integers to wrap (in fact some algorithms require this to be true). One of Fur’s main goals is to be predictable so integers should wrap like programmers have come to expect.

####Strings
In C strings are a sequence of chars that end with a null value. This has been the cause of many bugs in C programs because it's easy to accidentally (or maliciously) modify strings before they are outputted or forget to null terminate them causing unpredictable results and vulnerability. Most modern languages have made strings immutable, this has several advantages including constant time length look up (in C you would have to transverse the whole string making it linear), reduced vulnerability's from unintended string modifications.

####Array
Static arrays are almost the same in every programming language, so fur should feel familiar.

####Slices
Most of the time it is not known how much data a list need to hold so static arrays are no use. Some modern languages such as Go use slices which is a data type with an `index`, `length`, `capacity` and a hidden array. As long as `length < capacity` elements can be appended with no cost. As soon as more space is needed an allocation occurs expanding the hidden array's capacity. This simple structure is useful for so many different structures including queues and stacks.

####Structures
Most data is not a single type but many a collection of different types. Structures provide a simple way of grouping for ease of use.

####Objectives
- Standard integer types: 8, 16, 32 and 64 bits
- Integers must wrap around when overflowing
- Standard float types: 32 and 64 bits
- Immutable strings
- Fixed length arrays
- Variable length slices
- Structures

###Compiler executable
When developing code across many platforms often programmers find themselves in constrained environments whether it be limited resources, no graphical interface, no physical device etc. It can be difficult to install the tools required to compile a project which is why Fur should be distributed as a single executable with no install process and no dependencies. The executable would be invoked from the shell in order to compile the project.

This makes managing multiple versions of compilers trivial since you can just invoke different versions of the compiler. On a remote machine as long as you can download or transfer the compiler you can compile on that machine. Finally developing tooling on top of this such as a version manager should be trivial. 

####Objectives
- Provide a singular executable for the compiler with no install process.
- Executable should be easily obtainable from the project's website and GitHub repository.
- Provide builds for different architectures and operating systems.

----------
##Documented Design

###Overview
```markdown
┌───────────────────────┐
│        Lexer        │
└───────────────────────┘
		   ↓
┌───────────────────────┐
│       Parser        │
└───────────────────────┘
           ↓
┌───────────────────────┐
│      Analyzer       │
└───────────────────────┘
		   ↓
┌───────────────────────┐   ┌───────────────────────┐
│    IR Generation    │ ⇄ │        Goory        │
└───────────────────────┘   └───────────────────────┘
		   ↓
┌───────────────────────┐
│  LLVM Optimization  │
│     and linking     │
└───────────────────────┘
```
1. First the lexer parses over the source code and converts the characters to tokens. Tokens are the smallest usable pieces of information the compiler needs to compile the program. The reason we use tokens is because it's much easier to express logic further down in the pipeline and its alot faster to operate on tokens than it is on strings.

2. The parser runs through the list of tokens and creates an abstract syntax tree (AST). This is a representation of the source code which makes the logic of the program much easier to operate on.

3. The analizer infers any allowed omissions such as type inference and implicit casts.

4. Next the AST is turned into LLVM IR, with the help of goory, which is a much lower level language which is used by the LLVM optimizer to make the program run faster.

5. Finally the LLVM optimizer optimizes the IR and the linker links the module with any external modules before producing the final executable.

###Lexer
The lexer accepts the program's source code in the form of a slice of bytes, it then loops until it reaches the end of the of this list. At each iteration the lexer switches upon the current byte, if a number then the it must be start of a new number token so the number function is called, if its a double quote it must be the start of a new string token so the string function is called etc. Each of these functions moves the current token forward until it reaches the end for example the other double quote in a string. After the last character of a token is read the token is appended to a list and any white space after that token is skipped over, the lexer is ready to read the next token.

###Parser
The parser works in a similar way to the lexer, instead of working on characters, the parser works on tokens. Whenever the beginning of a construct is met for example an if statement the if parser is called which in turn calls the expression parser (for the condition of the if statement) and the block parser as well as any chaining else-if/else statements. Each function that parses through the tokens returns a node of the abstract syntax tree. These nodes are the combined in more complex statements to form nodes with many layers of children (such as an if statement). The resulting tree structure is called an abstract syntax tree which is very easy to compile to lower level languages or change any of the programs logic. This parser is a recursive descent parser with no backtracking allowing it to run in linear time.

####Shunting yard algorithm
For mathematical statements the shunting yard algorithm is used. Each value token (floats, ints etc) are pushed onto the out queue whereas operators are pushed on to the operator queue. If an operator precedence is less than or equal to the precedence of the the head of the operator queue the head of the operator stack is poped as well as the first two nodes on the output queue. The two nodes are then combined into a single binary node and pushed back onto the output stack.

If a parenthesis is met then then it's pushed to the operator queue. When the matching parenthesis is found all the operators up to the parenthesis are popped from the output queue the same as above.

At the end we are left with a single node on the output queue which is returned.

###Analyser
Again the analyser works similarly, recursing through the AST. Whenever an assignment node is reached it first checks if its type was specified, if not then it infers its type by recursing through the value on the right hand side. If the value was specified the it still checks if the right hand side matches this value, if not an implicit cast is inserted. The automatic cast insertion is also used for almost all other statements including call, return etc.

###IR generation
Once more the the AST is recursed through until it reaches a child with no children. We then return the value of an in memory representation of the node produced by goory (a separate library for writing llvm ir). More complex nodes use these values to return their own ir nodes until all constructs have been translated. Finally the root node is transformed into a string of llvm ir.

####Goory
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
----------
##Technical Solution
Insert code here...

----------
##Testing

###Lexer
The input represents a sample of source code and the expected output is a list of tokens the lexer is expected to produce, if the lexer tokens matches the expected output then the test passed, else it failed.

Input | Expected | Result
--- | --- | ---
`"foo"` | `Token{IDENT, "foo", 1, 1}`, `Token{SEMICOLON, "\n", 1, 4}` | Pass
`"foo"` | `Token{INT, "100", 1, 1}`, `Token{ADD, "", 1, 5}`, `Token{INT, "23", 1, 7}`, `Token{SEMICOLON, "\n", 1, 9}` | Pass


----------
##Evaluation
Insert evaluation here...

----------
##References
1. The Rust Reference <a id="1">https://doc.rust-lang.org/reference.html</a>
2. The Go Programming Language Specification <a id="2">https://golang.org/ref/spec</a>
3. Specification for the D Programming Language <a id="3">https://dlang.org/spec/spec.html</a>
4. The F# Language Specification <a id="4">http://fsharp.org/specs/language-spec/</a>
5. Go's GitHub Repository <a id="5">https://github.com/golang/go</a>
6. Go - Proposal: Eliminate STW stack re-scanning <a id="6">https://golang.org/design/17503-eliminate-rescan</a>
7. Plumber - G1 vs CMS vs Parallel GC <a id="7">https://plumbr.eu/blog/garbage-collection/g1-vs-cms-vs-parallel-gc</a>