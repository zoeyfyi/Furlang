package compiler

import (
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/goory"
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

func (s *scope) find(search string) value.Value {
	currentScope := s
	for true {
		if value := currentScope.values[search]; value != nil {
			return value
		}
		if currentScope.outerScope == nil {
			return nil
		}
		currentScope = currentScope.outerScope
	}
	return nil
}

func gooryType(tokenType lexer.TokenType) types.Type {
	switch tokenType {
	case lexer.INT:
		return goory.IntType(32)
	case lexer.INT8:
		return goory.IntType(8)
	case lexer.INT16:
		return goory.IntType(16)
	case lexer.INT32:
		return goory.IntType(32)
	case lexer.INT64:
		return goory.IntType(64)
	case lexer.FLOAT:
		return goory.FloatType()
	case lexer.FLOAT32:
		return goory.FloatType()
	case lexer.FLOAT64:
		return goory.DoubleType()
	default:
		panic("Unkown type")
	}
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

func compileBlock(eBlock block, ci *compileInfo) *goory.Block {
	newBlock := ci.block.Function().AddBlock()

	// Create new compiler infomation
	newCi := &compileInfo{
		block: newBlock,
		scope: &scope{
			outerScope: ci.scope,
			values:     make(map[string]value.Value),
		},
	}

	for _, e := range eBlock.expressions {
		e.compile(newCi)
	}

	return newBlock
}

// Blocks
func (e block) compile(ci *compileInfo) value.Value {
	newBlock := compileBlock(e, ci)
	ci.block.Br(newBlock)
	ci.block = ci.block.Function().AddBlock()
	newBlock.Br(ci.block)
	return nil
}

// Assignment
func (e assignment) compile(ci *compileInfo) value.Value {
	value := e.value.compile(ci)

	if e.nameType != lexer.ILLEGAL {
		assignmentType := gooryType(e.nameType)
		if value.Type() != assignmentType {
			ci.scope.values[e.name] = ci.block.Cast(value, assignmentType)
		} else {
			ci.scope.values[e.name] = value
		}

		return nil
	}

	ci.scope.values[e.name] = value
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
		trueBlock := compileBlock(e.blocks[0].block, ci)
		ci.block.CondBr(
			e.blocks[0].condition.compile(ci),
			trueBlock,
			nil)

		if !trueBlock.Terminated() {
			oldBlock := ci.block
			ci.block = ci.block.Function().AddBlock()
			oldBlock.Br(ci.block)
			trueBlock.Br(ci.block)
		}

	case 2:
		trueBlock := compileBlock(e.blocks[0].block, ci)
		falseBlock := compileBlock(e.blocks[1].block, ci)

		ci.block.CondBr(
			e.blocks[0].condition.compile(ci),
			trueBlock,
			falseBlock)

		if !trueBlock.Terminated() && !falseBlock.Terminated() {
			ci.block = ci.block.Function().AddBlock()
		}
		if !trueBlock.Terminated() {
			trueBlock.Br(ci.block)
		}
		if !falseBlock.Terminated() {
			falseBlock.Br(ci.block)
		}

	default:
		panic("Cannot handle else if (yet)")
	}

	return nil
}

// Name
func (e name) compile(ci *compileInfo) value.Value {
	return ci.scope.find(e.name)
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
