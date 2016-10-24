package compiler

import "github.com/bongo227/goory"

type expression interface {
	compile(compileInfo) goory.Value
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

func (s *scope) push() *scope {
	return &scope{
		outerScope: s,
		values:     make(map[string]goory.Value, 1000),
	}
}

func (s *scope) pop() *scope {
	return s.outerScope
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

func gooryType(t int) goory.Type {
	switch t {
	case typeInt32:
		return goory.Int32Type
	case typeFloat32:
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
		returnType := gooryType(f.returns[0].nameType)
		argType := make([]goory.Type, len(f.args))
		for i, a := range f.args {
			argType[i] = gooryType(a.nameType)
		}

		function := module.NewFunction(f.name, returnType, argType...)
		functions[function.Name()] = function

		rootScope := scope{
			outerScope: nil,
			values:     make(map[string]goory.Value, 1000),
		}

		for _, e := range f.block.expressions {
			e.compile(compileInfo{
				functions: functions,
				block:     function.Entry(),
				scope:     &rootScope,
			})
		}
	}

	return module.LLVM()
}

func compileBlock(eBlock block, ci compileInfo) *goory.Block {
	newBlock := ci.block.Function().AddBlock()

	ci.scope.push()
	for _, e := range eBlock.expressions {
		e.compile(ci)
	}
	ci.scope.pop()

	return newBlock
}

// Blocks
func (e block) compile(ci compileInfo) goory.Value {
	return ci.block.Br(compileBlock(e, ci)).Value()
}

// Assignment
func (e assignment) compile(ci compileInfo) goory.Value {
	ci.scope.values[e.name] = e.compile(ci)
	return nil
}

// Functions
func (e function) compile(ci compileInfo) goory.Value {
	panic("Cannot embed instructions (yet)")
}

// Returns
func (e ret) compile(ci compileInfo) goory.Value {
	return ci.block.Ret(e.returns[0].compile(ci)).Value()
}

// ifBlock
func (e ifBlock) compile(ci compileInfo) goory.Value {
	return e.condition.compile(ci)
}

// ifExpression
func (e ifExpression) compile(ci compileInfo) goory.Value {
	switch len(e.blocks) {
	case 1:
		return ci.block.CondBr(
			e.blocks[0].condition.compile(ci),
			compileBlock(e.blocks[0].block, ci),
			nil).Value()
	case 2:
		return ci.block.CondBr(
			e.blocks[0].condition.compile(ci),
			compileBlock(e.blocks[0].block, ci),
			compileBlock(e.blocks[1].block, ci)).Value()
	default:
		panic("Cannot handle else if (yet)")
	}
}

// Name
func (e name) compile(ci compileInfo) goory.Value {
	return ci.scope.find(e.name)
}

// Maths
func (e maths) compile(ci compileInfo) goory.Value {
	return e.expression.compile(ci)
}

// Boolean
func (e boolean) compile(ci compileInfo) goory.Value {
	return goory.ConstBool(e.value)
}

// Addition
func (e addition) compile(ci compileInfo) goory.Value {
	return ci.block.Fadd(e.lhs.compile(ci), e.rhs.compile(ci))
}

// Subtraction
func (e subtraction) compile(ci compileInfo) goory.Value {
	return ci.block.Fsub(e.lhs.compile(ci), e.rhs.compile(ci))
}

// Multiplication
func (e multiplication) compile(ci compileInfo) goory.Value {
	return ci.block.Fmul(e.lhs.compile(ci), e.rhs.compile(ci))
}

// floatDivision
func (e floatDivision) compile(ci compileInfo) goory.Value {
	return ci.block.Fdiv(e.lhs.compile(ci), e.rhs.compile(ci))
}

// intDivision
func (e intDivision) compile(ci compileInfo) goory.Value {
	return ci.block.Div(e.lhs.compile(ci), e.rhs.compile(ci))
}

// number
func (e number) compile(ci compileInfo) goory.Value {
	return goory.ConstInt32(int32(e.value))
}

// float
func (e float) compile(ci compileInfo) goory.Value {
	return goory.ConstFloat32(float32(e.value))
}

// call
func (e call) compile(ci compileInfo) goory.Value {
	args := make([]goory.Value, len(e.args))
	for i, a := range e.args {
		args[i] = a.compile(ci)
	}

	return ci.block.Call(ci.functions[e.function], args...)
}
