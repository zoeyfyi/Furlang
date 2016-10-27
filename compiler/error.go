package compiler

import (
	"fmt"
	"strings"

	"github.com/bongo227/Furlang/lexer"
	"github.com/fatih/color"
)

// Error is an error realting to compilation
type Error struct {
	err        string
	tokenRange []lexer.Token
}

var cerror = color.New(color.FgHiRed).SprintfFunc()

func (e Error) Error() string {
	return e.err
}

// FormatedError returns the error message with an arrow pointing to the invalid part of the line(s)
func (e Error) FormatedError(lines []string) string {
	if e.tokenRange != nil {
		clow, chigh := e.ColumnRange()

		return fmt.Sprintf("\n%s\n%s\n%s%s\n",
			cerror("Error: %s", e.err),
			lines[e.Line()], strings.Repeat(" ", clow),
			cerror(strings.Repeat("^", chigh-clow)))
	}

	return fmt.Sprintf("\n%s\n",
		cerror("Error: %s", e.err))

}

// Line returns the line number of the error
func (e Error) Line() int {
	if len(e.tokenRange) >= 1 {
		return e.tokenRange[0].Pos.Line - 1
	}

	return 0
}

// ColumnRange returns the two columns where the error lies
func (e Error) ColumnRange() (int, int) {
	if len(e.tokenRange) >= 2 {
		first := e.tokenRange[0]
		last := e.tokenRange[len(e.tokenRange)-1]
		return first.Pos.Column - 1, last.Pos.Column - 1 + last.Pos.Width
	} else if len(e.tokenRange) >= 1 {
		toke := e.tokenRange[0]
		return toke.Pos.Column - 1, toke.Pos.Column - 1 + toke.Pos.Width
	}

	return 0, 0
}
