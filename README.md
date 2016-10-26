# Fur #

## Design Philosophy ##

### Garbage collection ###
This is a massive problem for performance applications such as games. With equipment enabling frame rates as high as 144 fps, games need to update and render in less than 7ms. Go's gc is around 10ms, making it almost unusable due to the frequent freezes/stuttering which would occur due to the stop-the-world gc. Garbage collection has no place in games so it has no place in Fur.

### Object Oriented Programming ###
Programs are data and transformations on data, the object oriented para-dime takes the programmer away from the data in favour of atomic encapsulation. Encapsulation try's to solve maintainability by making programs almost impossible to maintain. Large code base's end up with impossible to follow inheritance trees with surprise edge cases as a result of generalising your problems. Fur stays far away from object oriented programming because their is no case in which you gain anything from using it.

### Usage ###
You should be able to download the compiler binary and compile strait away, no unnecessary install process or limitations. More advanced tools may be built on top of this but their should alway be the option of downloading just the compiler.

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
* parser.go - simplify the ast structure
* parser.go - handle multiple returns
* parser.go - clean up new line handling
* parser.go - Better end of file checking
* parser.go - add error handling
* parser.go - add go error handling

* all - create a more unified interface

* compiler.go - Use program name to create compiled file name

* check.go - Re add commented code
* check.go - Add token range to errors

* lexer.go - make multitokens work with a varible number of tokens
* llvm.go - Add else if support to ifExpression compililation

## Syntax ideas ##

The following are ideas for syntax not yet implemented

#### Multiple returns ####
```
swap :: i32 a, i32 b -> i32, i32 {
    return b, a
}

main :: -> i32 {
    a, b := swap(123, 432)
    return b
}
```

#### For loops ####
```
loop :: -> i32 {
    a := 0
    for i := 0; i < 123; i++ {
        a = i
    }
    return i
}

main :: -> {
    return loop()
}
```
```
values := []int{1, 2, 3, 4}

for _, v := range values {
    # v is immutable
}

for _, v := map values {
    # v is mutable
}

for _, v := filter values {
    # return true/false keeps/removes v from values
}
```
