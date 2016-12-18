package lexer

// Type represents a type
type Type interface {
	// Gets the base type of the type
	Base() Type
	String() string
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
	length int
}

func NewArray(typ Type, length int) *Array {
	return &Array{typ, length}
}

func (a *Array) Length() int {
	return a.Length
}

func (a *Array) Type() Type {
	return a.typ
}

type Slice struct {
	typ Type
}

func NewSlice(typ Type) *Slice {
	return &Slice{typ}
}

func (s *Slice) Type() {
	return s.typ
}

type Pointer struct {
	typ Type
}

func NewPointer(typ Type) *Pointer {
	return &Pointer{typ}
}

func (p *Pointer) Type() Type {
	return p.typ
}

func (b *Basic) Underlying() Type   { return b }
func (b *Array) Underlying() Type   { return b }
func (b *Slice) Underlying() Type   { return b }
func (b *Pointer) Underlying() Type { return b }
