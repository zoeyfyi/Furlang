package types

import (
	"fmt"

	goorytypes "github.com/bongo227/goory/types"
)

// TODO: when we implement custom type defininitions create a unresolved type.
// when we run through with ananlysis this type will be substituted with the
// actual type.

// Type represents a type
type Type interface {
	// Gets the base type of the type
	Base() Type
	Llvm() goorytypes.Type
	// String() string
}

type BasicType int

const (
	Invalid BasicType = iota

	Bool
	Int
	I8
	I16
	I32
	I64
	Uint
	U8
	U16
	U32
	U64
	Float
	F32
	F64
	String

	UntypedBool
	UntypedInt
	UntypedRune
	UntypedFloat
	UntypedString
	UntypedNil

	// Aliases

	Byte = U8
	Rune = I32
)

type BasicInfo int

const (
	IsBool BasicInfo = 1 << iota
	IsInt
	IsUnsigned
	IsFloat
	IsString
	IsUntyped

	IsOrdered = IsInt | IsFloat | IsString
	IsNumeric = IsInt | IsFloat
	IsConst   = IsBool | IsNumeric | IsString
)

type Basic struct {
	typ  BasicType
	info BasicInfo
	name string
}

func IntType(bits int) *Basic {
	var typ BasicType
	var name string
	switch bits {
	case 0:
		typ = Int
		name = "int"
	case 8:
		typ = I8
		name = "i8"
	case 16:
		typ = I16
		name = "i16"
	case 32:
		typ = I32
		name = "i32"
	case 64:
		typ = I64
		name = "i64"
	default:
		panic("Invalid number of bits")
	}

	return &Basic{
		typ:  typ,
		name: name,
		info: IsInt,
	}
}

func FloatType(bits int) *Basic {
	var typ BasicType
	var name string
	switch bits {
	case 0:
		typ = Float
		name = "float"
	case 32:
		typ = F32
		name = "f32"
	case 64:
		typ = F64
		name = "f64"
	default:
		panic("Invalid number of bits")
	}

	return &Basic{
		typ:  typ,
		name: name,
		info: IsFloat,
	}
}

func IsBasic(ident string) bool {
	switch ident {
	case "int", "i8", "i16", "i32", "i64", "float", "f32", "f64":
		return true
	default:
		return false
	}
}

// Gets the type corsponding to the identifier
func GetType(ident string) *Basic {
	switch ident {
	case "int":
		return IntType(0)
	case "i8":
		return IntType(8)
	case "i16":
		return IntType(16)
	case "i32":
		return IntType(32)
	case "i64":
		return IntType(64)
	case "float":
		return FloatType(0)
	case "f32":
		return FloatType(32)
	case "f64":
		return FloatType(64)
	case "bool":
		return BasicBool
	}

	panic(fmt.Sprintf("Unrecognized basic type: %s", ident))
}

var (
	BasicBool = &Basic{
		typ:  Bool,
		info: IsBool,
		name: "bool",
	}
)

func (b *Basic) Type() BasicType {
	return b.typ
}

func (b *Basic) Info() BasicInfo {
	return b.info
}

func (b *Basic) Name() string {
	return b.name
}

type Array struct {
	typ    Type
	length int64
}

func NewArray(typ Type, length int64) *Array {
	return &Array{typ, length}
}

func (a *Array) Length() int {
	return a.Length()
}

func (a *Array) Type() Type {
	return a.typ
}

// type Slice struct {
// 	typ Type
// }

// func NewSlice(typ Type) *Slice {
// 	return &Slice{typ}
// }

// func (s *Slice) Type() Type {
// 	return s.typ
// }

type Pointer struct {
	typ Type
}

func NewPointer(typ Type) *Pointer {
	return &Pointer{typ}
}

func (p *Pointer) Type() Type {
	return p.typ
}

func (b *Basic) Base() Type { return b }

func (b *Basic) Llvm() goorytypes.Type {
	switch b.typ {
	case Bool:
		return goorytypes.NewBoolType()
	case Int:
		return goorytypes.NewIntType(64)
	case I8:
		return goorytypes.NewIntType(8)
	case I16:
		return goorytypes.NewIntType(16)
	case I32:
		return goorytypes.NewIntType(32)
	case I64:
		return goorytypes.NewIntType(64)
	case Float:
		return goorytypes.NewDoubleType()
	case F32:
		return goorytypes.NewFloatType()
	case F64:
		return goorytypes.NewDoubleType()
	default:
		panic("TODO: finish this")
	}
}

func (b *Array) Base() Type { return b.typ }

func (b *Array) Llvm() goorytypes.Type {
	return goorytypes.NewArrayType(b.typ.Llvm(), int(b.length))
}

// func (b *Slice) Base() Type   { return b.typ }

// func (b *Slice) Llvm() goorytypes.Type {
// }

func (b *Pointer) Base() Type { return b.typ }

func (b *Pointer) Llvm() goorytypes.Type {
	return goorytypes.NewPointerType(b.Base().Llvm())
}

type Function struct {
	argTypes   []Type
	returnType Type
}

func NewFunction(returnType Type, argTypes ...Type) *Function {
	return &Function{
		argTypes:   argTypes,
		returnType: returnType,
	}
}

func (b *Function) Arguments() []Type {
	return b.argTypes
}

func (b *Function) Return() Type {
	return b.returnType
}

func (b *Function) Base() Type { return b.returnType }

func (b *Function) Llvm() goorytypes.Type {
	argTypes := make([]goorytypes.Type, len(b.argTypes))
	for i, arg := range b.argTypes {
		argTypes[i] = arg.Llvm()
	}

	return goorytypes.NewFunction(b.returnType.Llvm(), argTypes...)
}
