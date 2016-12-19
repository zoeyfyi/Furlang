package compiler

import (
	"fmt"

	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/goory"
	"github.com/bongo227/goory/instructions"
	"github.com/bongo227/goory/types"
	"github.com/bongo227/goory/value"
)

type expression interface {
	compile(*compileInfo) value.Value
}

type compileInfo struct {
	block     *goory.Block
	scope     *scope
	functions map[string]*goory.Function
}

type scope struct {
	outerScope *scope
	values     map[string]value.Value
}

func (s *scope) find(search string, block *goory.Block) value.Value {
	currentScope := s
	for true {
		if value := currentScope.values[search]; value != nil {
			if allocation, ok := value.(*instructions.Alloca); ok {
				return block.Load(allocation)
			}

			return value
		}
		if currentScope.outerScope == nil {
			return nil
		}
		currentScope = currentScope.outerScope
	}
	return nil
}

func (s *scope) set(search string, block *goory.Block, value value.Value) {
	currentScope := s
	for true {
		if searchValue := currentScope.values[search]; searchValue != nil {
			block.Store(searchValue.(*instructions.Alloca), value)
			return
		}
		if currentScope.outerScope == nil {
			panic(search + " is not in scope")
		}
		currentScope = currentScope.outerScope
	}
	panic(search + " is not in scope")
}

func gooryType(typ lexer.Type) types.Type {
	switch typ := typ.(type) {
	case *lexer.Basic:
		switch typ.Type() {
		case lexer.Bool, lexer.UntypedBool:
			return goory.BoolType()
		case lexer.Int, lexer.I64, lexer.Uint, lexer.UntypedInt:
			return goory.IntType(64)
		case lexer.I8, lexer.U8, lexer.UntypedRune:
			return goory.IntType(8)
		case lexer.I16, lexer.U16:
			return goory.IntType(16)
		case lexer.I32, lexer.U32:
			return goory.IntType(32)
		case lexer.Float, lexer.F32, lexer.UntypedFloat:
			return goory.FloatType()
		case lexer.F64:
			return goory.DoubleType()
		case lexer.String, lexer.UntypedString:
			// TODO: What length of string
			return goory.ArrayType(goory.IntType(8), 100)
		case lexer.UntypedNil:
			panic("Nil")
		}
	case *lexer.Array:
		return goory.ArrayType(gooryType(typ.Type()), typ.Length())
	case *lexer.Slice:
		return goory.ArrayType(gooryType(typ.Type()), 0)
	case *lexer.Pointer:
		return types.NewPointerType(gooryType(typ.Type()))
	}

	panic("Unkown type")
}

// Llvm compiles the syntax tree to llvm ir
func Llvm(ast *syntaxTree) string {
	module := goory.NewModule("ben")
	functions := make(map[string]*goory.Function, len(ast.functions))

	for _, f := range ast.functions {
		// Create a new function in module
		returnType := gooryType(f.returns[0].nameType)
		function := module.NewFunction(f.name, returnType)

		// Create the root scope
		rootScope := scope{
			outerScope: nil,
			values:     make(map[string]value.Value, 1000),
		}

		// Add parameters to root scope
		for _, a := range f.args {
			rootScope.values[a.name] = function.AddArgument(gooryType(a.nameType), a.name)
		}

		functions[function.Name()] = function
		ci := &compileInfo{
			functions: functions,
			block:     function.Entry(),
			scope:     &rootScope,
		}

		// Compile all expressions in function
		for _, e := range f.block.expressions {
			e.compile(ci)
		}
	}

	return module.LLVM()
}

// compileBlock returns the start block to branch to and the final block (if you need to branch again)
func (e block) compileBlock(ci *compileInfo) (start *goory.Block, end *goory.Block) {
	newBlock := ci.block.Function().AddBlock()

	// Create new compiler infomation
	newCi := &compileInfo{
		functions: ci.functions,
		block:     newBlock,
		scope: &scope{
			outerScope: ci.scope,
			values:     make(map[string]value.Value),
		},
	}

	for _, e := range e.expressions {
		e.compile(newCi)
	}

	return newBlock, newCi.block
}

// Blocks
func (e block) compile(ci *compileInfo) value.Value {
	newStartBlock, newEndBlock := e.compileBlock(ci)
	ci.block.Br(newStartBlock)
	ci.block = ci.block.Function().AddBlock()
	newEndBlock.Br(ci.block)
	return nil
}

func (e forExpression) compile(ci *compileInfo) value.Value {
	// Compile for loop index varible
	e.index.compile(ci)

	fmt.Println("start of for loop", ci.block.Name())

	// Branch into for loop
	condition := e.condition.compile(ci)
	newStartBlock, newEndBlock := e.block.compileBlock(ci)
	falseBlock := ci.block.Function().AddBlock()
	ci.block.CondBr(condition, newStartBlock, falseBlock)

	// Branch to continue or exit
	ci.block = newEndBlock
	e.increment.compile(ci)
	condition = e.condition.compile(ci)
	ci.block.CondBr(condition, newStartBlock, falseBlock)

	// Continue from falseBlock
	ci.block = falseBlock

	return nil
}

func (e array) compile(ci *compileInfo) value.Value {
	arrType := goory.ArrayType(gooryType(e.baseType), e.length)

	values := make([]value.Value, e.length)
	for i, val := range e.values {
		values[i] = val.compile(ci)
	}

	return goory.Constant(arrType, values)
}

func (e arrayValue) compile(ci *compileInfo) value.Value {
	return ci.block.Extractvalue(ci.scope.find(e.name, ci.block), e.index.compile(ci))
}

// Increment
func (e increment) compile(ci *compileInfo) value.Value {
	value := ci.scope.find(e.name, ci.block)
	temp := ci.block.Add(value, goory.Constant(value.Type(), 1))
	ci.scope.set(e.name, ci.block, temp)

	return nil
}

// Decrement
func (e decrement) compile(ci *compileInfo) value.Value {
	value := ci.scope.find(e.name, ci.block)
	temp := ci.block.Sub(value, goory.Constant(value.Type(), 1))
	ci.scope.set(e.name, ci.block, temp)

	return nil
}

// Increment expression
func (e incrementExpression) compile(ci *compileInfo) value.Value {
	value := ci.scope.find(e.name, ci.block)
	temp := ci.block.Add(value, e.amount.compile(ci))
	ci.scope.set(e.name, ci.block, temp)

	return nil
}

// Decrement expression
func (e decrementExpression) compile(ci *compileInfo) value.Value {
	value := ci.scope.find(e.name, ci.block)
	temp := ci.block.Sub(value, e.amount.compile(ci))
	ci.scope.set(e.name, ci.block, temp)

	return nil
}

// Assignment
func (e assignment) compile(ci *compileInfo) value.Value {
	value := e.value.compile(ci)

	// Cast to correct type
	if e.nameType != lexer.ILLEGAL {
		assignmentType := gooryType(e.nameType)
		if value.Type() != assignmentType {
			value = ci.block.Cast(value, assignmentType)
		}
	}

	// Allocate space for varible
	allocation := ci.block.Alloca(value.Type())
	ci.block.Store(allocation, value)

	ci.scope.values[e.name] = allocation
	return nil
}

// Functions
func (e function) compile(ci *compileInfo) value.Value {
	panic("Cannot embed instructions (yet)")
}

// Returns
func (e ret) compile(ci *compileInfo) value.Value {
	value := e.returns[0].compile(ci)

	return ci.block.Ret(value)
}

// ifBlock
func (e ifBlock) compile(ci *compileInfo) value.Value {
	return e.block.compile(ci)
}

// ifExpression
func (e ifExpression) compile(ci *compileInfo) value.Value {
	switch len(e.blocks) {
	case 1:
		trueBlockStart, trueBlockEnd := e.blocks[0].block.compileBlock(ci)
		falseBlock := ci.block.Function().AddBlock()

		ci.block.CondBr(
			e.blocks[0].condition.compile(ci),
			trueBlockStart,
			falseBlock)

		if !trueBlockEnd.Terminated() {
			trueBlockEnd.Br(falseBlock)
		}

		ci.block = falseBlock

	case 2:
		trueBlockStart, trueBlockEnd := e.blocks[0].block.compileBlock(ci)
		falseBlockStart, falseBlockEnd := e.blocks[1].block.compileBlock(ci)

		ci.block.CondBr(
			e.blocks[0].condition.compile(ci),
			trueBlockStart,
			falseBlockStart)

		if !trueBlockEnd.Terminated() && !falseBlockEnd.Terminated() {
			ci.block = ci.block.Function().AddBlock()
		}
		if !trueBlockEnd.Terminated() {
			trueBlockEnd.Br(ci.block)
		}
		if !falseBlockEnd.Terminated() {
			falseBlockEnd.Br(ci.block)
		}

	default:
		panic("Cannot handle else if (yet)")
	}

	return nil
}

// Name
func (e name) compile(ci *compileInfo) value.Value {
	return ci.scope.find(e.name, ci.block)
}

// Cast
func (e cast) compile(ci *compileInfo) value.Value {
	return ci.block.Cast(e.value.compile(ci), gooryType(e.cast))
}

// Boolean
func (e boolean) compile(ci *compileInfo) value.Value {
	return goory.Constant(goory.BoolType(), e.value)
}

// Addition
func (e addition) compile(ci *compileInfo) value.Value {
	return ci.block.Add(e.lhs.compile(ci), e.rhs.compile(ci))
}

// Subtraction
func (e subtraction) compile(ci *compileInfo) value.Value {
	return ci.block.Sub(e.lhs.compile(ci), e.rhs.compile(ci))
}

// Multiplication
func (e multiplication) compile(ci *compileInfo) value.Value {
	return ci.block.Mul(e.lhs.compile(ci), e.rhs.compile(ci))
}

// floatDivision
func (e floatDivision) compile(ci *compileInfo) value.Value {
	return ci.block.Fdiv(e.lhs.compile(ci), e.rhs.compile(ci))
}

// intDivision
func (e intDivision) compile(ci *compileInfo) value.Value {
	return ci.block.Div(e.lhs.compile(ci), e.rhs.compile(ci))
}

// lessThan
func (e lessThan) compile(ci *compileInfo) value.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if types.IsInteger(lhs.Type()) {
		return ci.block.Icmp(goory.IntSlt, lhs, rhs)
	}

	return ci.block.Fcmp(goory.FloatOlt, lhs, rhs)
}

// moreThan
func (e moreThan) compile(ci *compileInfo) value.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if types.IsInteger(lhs.Type()) {
		return ci.block.Icmp(goory.IntSgt, lhs, rhs)
	}

	return ci.block.Fcmp(goory.FloatOgt, lhs, rhs)
}

// equal
func (e equal) compile(ci *compileInfo) value.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if types.IsInteger(lhs.Type()) {
		return ci.block.Icmp(goory.IntEq, lhs, rhs)
	}

	return ci.block.Fcmp(goory.FloatOeq, lhs, rhs)
}

// notEqual
func (e notEqual) compile(ci *compileInfo) value.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if types.IsInteger(lhs.Type()) {
		return ci.block.Icmp(goory.IntNe, lhs, rhs)
	}

	return ci.block.Fcmp(goory.FloatOne, lhs, rhs)
}

// mod
func (e mod) compile(ci *compileInfo) value.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	return ci.block.Srem(lhs, rhs)
}

// number
func (e number) compile(ci *compileInfo) value.Value {
	return goory.Constant(goory.IntType(32), int32(e.value))
}

// float
func (e float) compile(ci *compileInfo) value.Value {
	return goory.Constant(goory.FloatType(), float32(e.value))
}

// call
func (e call) compile(ci *compileInfo) value.Value {
	args := make([]value.Value, len(e.args))
	for i, a := range e.args {
		args[i] = a.compile(ci)
	}

	return ci.block.Call(ci.functions[e.function], args...)
}
