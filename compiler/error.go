package compiler

// Error is an error realting to compilation
type Error struct {
	err        string
	tokenRange []token
}

func (e Error) Error() string {
	return e.err
}

func (e Error) Line() int {
	return e.tokenRange[0].line - 1
}

func (e Error) ColumnRange() (int, int) {
	first := e.tokenRange[0]
	last := e.tokenRange[len(e.tokenRange)-1]
	return first.column - 1, last.column - 1 + last.length
}
