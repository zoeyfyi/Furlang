package compiler

import (
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/goory"
)

type expression interface {
	compile(*compileInfo) goory.Value
}

type compileInfo struct {
	block     *goory.Block
	scope     *scope
	functions map[string]*goory.Function
}

type scope struct {
	outerScope *scope
	values     map[string]goory.Value
}

func (s *scope) find(search string) goory.Value {
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

func gooryType(tokenType lexer.TokenType) goory.Type {
	switch tokenType {
	case lexer.INT32:
		return goory.Int32Type
	case lexer.FLOAT32:
		return goory.Float32Type
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
		argType := make([]goory.Type, len(f.args))
		for i, a := range f.args {
			argType[i] = gooryType(a.nameType)
		}
		function := module.NewFunction(f.name, returnType, argType...)
		functions[function.Name()] = function

		// Create the root scope
		rootScope := scope{
			outerScope: nil,
			values:     make(map[string]goory.Value, 1000),
		}

		// Added parameters to scope
		pValues := function.Parameters()
		for i, a := range f.args {
			rootScope.values[a.name] = pValues[i]
		}

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
			values:     make(map[string]goory.Value),
		},
	}

	for _, e := range eBlock.expressions {
		e.compile(newCi)
	}

	return newBlock
}

// Blocks
func (e block) compile(ci *compileInfo) goory.Value {
	newBlock := compileBlock(e, ci)
	ci.block.Br(newBlock)
	ci.block = ci.block.Function().AddBlock()
	newBlock.Br(ci.block)
	return nil
}

// Assignment
func (e assignment) compile(ci *compileInfo) goory.Value {
	ci.scope.values[e.name] = e.value.compile(ci)
	return nil
}

// Functions
func (e function) compile(ci *compileInfo) goory.Value {
	panic("Cannot embed instructions (yet)")
}

// Returns
func (e ret) compile(ci *compileInfo) goory.Value {
	return ci.block.Ret(e.returns[0].compile(ci)).Value()
}

// ifBlock
func (e ifBlock) compile(ci *compileInfo) goory.Value {
	return e.block.compile(ci)
}

// ifExpression
func (e ifExpression) compile(ci *compileInfo) goory.Value {
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
			falseBlock).Value()

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
func (e name) compile(ci *compileInfo) goory.Value {
	return ci.scope.find(e.name)
}

// Boolean
func (e boolean) compile(ci *compileInfo) goory.Value {
	return goory.ConstBool(e.value)
}

// Addition
func (e addition) compile(ci *compileInfo) goory.Value {
	return ci.block.Add(e.lhs.compile(ci), e.rhs.compile(ci)).Value()
}

// Subtraction
func (e subtraction) compile(ci *compileInfo) goory.Value {
	return ci.block.Sub(e.lhs.compile(ci), e.rhs.compile(ci)).Value()
}

// Multiplication
func (e multiplication) compile(ci *compileInfo) goory.Value {
	return ci.block.Mul(e.lhs.compile(ci), e.rhs.compile(ci)).Value()
}

// floatDivision
func (e floatDivision) compile(ci *compileInfo) goory.Value {
	return ci.block.Fdiv(e.lhs.compile(ci), e.rhs.compile(ci)).Value()
}

// intDivision
func (e intDivision) compile(ci *compileInfo) goory.Value {
	return ci.block.Div(e.lhs.compile(ci), e.rhs.compile(ci)).Value()
}

// lessThan
func (e lessThan) compile(ci *compileInfo) goory.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if lhs.Type() == goory.Int32Type || lhs.Type() == goory.Int64Type {
		return ci.block.ICmp(goory.IModeSlt(), lhs, rhs).Value()
	}

	return ci.block.FCmp(goory.FModeUlt(), lhs, rhs).Value()
}

// moreThan
func (e moreThan) compile(ci *compileInfo) goory.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if lhs.Type() == goory.Int32Type || lhs.Type() == goory.Int64Type {
		return ci.block.ICmp(goory.IModeSgt(), lhs, rhs).Value()
	}

	return ci.block.FCmp(goory.FModeUgt(), lhs, rhs).Value()
}

// equal
func (e equal) compile(ci *compileInfo) goory.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if lhs.Type() == goory.Int32Type || lhs.Type() == goory.Int64Type {
		return ci.block.ICmp(goory.IModeEq(), lhs, rhs).Value()
	}

	return ci.block.FCmp(goory.FModeUeq(), lhs, rhs).Value()
}

// notEqual
func (e notEqual) compile(ci *compileInfo) goory.Value {
	lhs := e.lhs.compile(ci)
	rhs := e.rhs.compile(ci)

	if lhs.Type() == goory.Int32Type || lhs.Type() == goory.Int64Type {
		return ci.block.ICmp(goory.IModeNe(), lhs, rhs).Value()
	}

	return ci.block.FCmp(goory.FModeUne(), lhs, rhs).Value()
}

// number
func (e number) compile(ci *compileInfo) goory.Value {
	return goory.ConstInt32(int32(e.value))
}

// float
func (e float) compile(ci *compileInfo) goory.Value {
	return goory.ConstFloat32(float32(e.value))
}

// call
func (e call) compile(ci *compileInfo) goory.Value {
	args := make([]goory.Value, len(e.args))
	for i, a := range e.args {
		args[i] = a.compile(ci)
	}

	return ci.block.Call(ci.functions[e.function], args...).Value()
}
