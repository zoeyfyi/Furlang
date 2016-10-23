package compiler

import "github.com/bongo227/goory"

type expression interface {
	compile(*goory.Block, *scope) *goory.Instruction
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

	for _, f := range ast.functions {
		returnType := gooryType(f.returns[0].nameType)
		argType := make([]goory.Type, len(f.args))
		for i, a := range f.args {
			argType[i] = gooryType(a.nameType)
		}

		function := module.NewFunction(f.name, returnType, argType...)

		rootScope := scope{
			outerScope: nil,
			values:     make(map[string]goory.Value, 1000),
		}

		for _, e := range f.block.expressions {
			e.compile(function.Entry(), &rootScope)
		}
	}

	return module.LLVM()
}

func compileBlock(eBlock block, block *goory.Block, scope *scope) *goory.Block {
	newBlock := block.Function().AddBlock()

	scope.push()
	for _, e := range eBlock.expressions {
		e.compile(block, scope)
	}
	scope.pop()

	return newBlock
}

func (e block) compile(block *goory.Block, scope *scope) *goory.Instruction {
	return block.Br(compileBlock(e, block, scope))
}

func (e assignment) compile(block *goory.Block, scope *scope) *goory.Instruction {
	scope.values[e.name] = e.compile(block, scope).Value()
	return nil
}

func (e function) compile(block *goory.Block, scope *scope) *goory.Instruction {
	panic("Cannot embed instructions (yet)")
}

func (e ret) compile(block *goory.Block, scope *scope) *goory.Instruction {
	return block.Ret(e.returns[0].compile(block, scope).Value())
}

func (e ifBlock) compile(block *goory.Block, scope *scope) *goory.Instruction {
	return e.condition.compile(block, scope)
}

func (e ifExpression) compile(block *goory.Block, scope *scope) *goory.Instruction {
	switch len(e.blocks) {
	case 1:
		return block.CondBr(
			e.blocks[0].condition.compile(block, scope),
			compileBlock(e.blocks[0].block, block, scope),
			nil)
	case 2:
		return block.CondBr(
			e.blocks[0].condition.compile(block, scope),
			compileBlock(e.blocks[0].block, block, scope),
			compileBlock(e.blocks[1].block, block, scope))
	default:
		panic("Cannot handle else if (yet)")
	}
}
