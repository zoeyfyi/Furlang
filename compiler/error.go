package compiler

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Error is an error realting to compilation
type Error struct {
	err        string
	tokenRange []token
}

func (e Error) Error() string {
	return e.err
}

func (e Error) FormatedError(lines []string) string {
	if e.tokenRange != nil {
		cerror := color.New(color.FgHiRed).SprintfFunc()
		clow, chigh := e.ColumnRange()

		return fmt.Sprintf("\n%s\n%s\n%s%s\n",
			cerror("Error: %s", e.err),
			lines[e.Line()], strings.Repeat(" ", clow),
			cerror(strings.Repeat("^", chigh-clow)))
	}

}

// Line returns the line number of the error
func (e Error) Line() int {
	if len(e.tokenRange) >= 1 {
		return e.tokenRange[0].line - 1
	}

	return 0
}

// ColumnRange returns the two columns where the error lies
func (e Error) ColumnRange() (int, int) {
	if len(e.tokenRange) >= 2 {
		first := e.tokenRange[0]
		last := e.tokenRange[len(e.tokenRange)-1]
		return first.column - 1, last.column - 1 + last.length
	}

	return 0, 0
}
