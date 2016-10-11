# Fur #

---

## Design Philosophy ##

### Garbage collection ###
This is a massive problem for performance applications such as games. With equipment enabling frame rates as high as 144 fps, games need to update and render in less than 7ms. Go's gc is around 10ms, making it almost unusable due to the frequent freezes/stuttering which would occur due to the stop-the-world gc. Garbage collection has no place in games so it has no place in Fur.

### Object Oriented Programming ###
Programs are data and transformations on data, the object oriented para-dime takes the programmer away from the data in favour of atomic encapsulation. Encapsulation try's to solve maintainability by making programs almost impossible to maintain. Large code base's end up with impossible to follow inheritance trees with surprise edge cases as a result of generalising your problems. Fur stays far away from object oriented programming because their is no case in which you gain anything from using it.

### Usage ###
You should be able to download the compiler binary and compile strait away, no unnecessary install process or limitations. More advanced tools may be built on top of this but their should alway be the option of downloading just the compiler.

---

## Objectives ##
* Types:
    * Integer (int, i32, i64, u8, u16, u32, u64)
    * Float (float, f32, f64)
    * Bool (bool)
    * Arrays
    * Slices
    * Runes (unicode character)
    * Strings (rune slice)
* Functions:
    * Multiple return values
    * Function overloading
    * Functions in functions
    * Named returns
    * Varible argument functions
* Structs:
    * Unions and tagged unions
    * Named/nameless assignment
    * Subtyping
* Match:
    * Multiple cases
    * Optial fallthrow

## TODO's ##
* ast.go - compute the number of function arguments
* ast.go - simplify the ast structure
* ast.go - handle multiple returns
* ast.go - check openBody case, if its at a block should it push a block?
* check.go - Re add commented code
* check.go - Add token range to errors
* lexer.go - make multitokens work with a varible number of tokens
* llvm.go - Add else if support to ifExpression compililation
