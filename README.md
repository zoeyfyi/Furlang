# Fur #
---
## Design Philosophy ##

### Garbage collection ###
Their is a massive problem for performance applications such as games, with equipment enabling frame rates as high as 144 fps, games need to update and render in less than 7ms. Go's gc is around 10ms, making it almost unusable due to the frequent freezes/stuttering which would occur due to the stop-the-world gc. Garbage collection has no place in games so it has no place in Fur.

### Object Oriented Programming ###
Programs are data and transformations on data, the object oriented para-dime takes the programmed away from the data in favour of atomic encapsulation. Encapsulation try's to solve maintainability by making programs almost impossible to maintain. Large code base's end up with impossible to follow inheritance trees with surprise edge cases as a result of generalising your problems. Fur stays far away from object oriented programming because their is no case in which you gain anything from using it.
---
## Objectives ##
* Types:
⋅⋅* Integer
* Primitive types: integer, float, bool, arrays, runes/chars
* Derived types: enums, slices, strings
* Functions
1