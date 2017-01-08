package ast

import (
	"go/token"

	"github.com/bongo227/Furlang/lexer"
)

type Statement interface {
	statementNode()
}

type DeclareStatement struct {
	Statement Declare
}

func (e *DeclareStatement) First() lexer.Token { return e.Statement.First() }
func (e *DeclareStatement) Last() lexer.Token  { return e.Statement.Last() }
func (e *DeclareStatement) statementNode()     {}

// AssignmentStatement is a statement in the form: expression := expression
type AssignmentStatement struct {
	Left  Expression
	Token lexer.Token
	Right Expression
}

func (e *AssignmentStatement) First() lexer.Token { return e.Left.First() }
func (e *AssignmentStatement) Last() lexer.Token  { return e.Right.Last() }
func (e *AssignmentStatement) statementNode()     {}

// ReturnStatement is a statement in the form: return expression
type ReturnStatement struct {
	Return Token
	Result Expression
}

func (e *ReturnStatement) First() lexer.Token { return e.Return }
func (e *ReturnStatement) Last() lexer.Token  { return e.Result.Last() }
func (e *ReturnStatement) statementNode()     {}

// BlockStatement is a statement in the form: {statement; statement; ...}
type BlockStatement struct {
	LeftBrace  lexer.Token
	Statements []Statement
	RightBrace lexer.Token
}

func (e *BlockStatement) First() lexer.Token { return e.LeftBrace }
func (e *BlockStatement) Last() lexer.Token  { return e.RightBrace }
func (e *BlockStatement) statementNode()     {}

// IfStatment is a statement in the form: if expression {statement; ...} ...
type IfStatment struct {
	If        token.Pos
	Condition Expression
	Body      *BlockStatement
	Else      *IfStatment
}

func (e *IfStatment) First() lexer.Token { return e.If }
func (e *IfStatment) Last() lexer.Token  { return e.Else.Last() }
func (e *IfStatment) statementNode()     {}

// ForStatement is a statement in the form: for statement; expression; statement {statement; ...}
type ForStatement struct {
	For       lexer.Token
	Index     Statement
	Condition Expression
	Increment Statement
	Body      *BlockStatement
}

func (e *ForStatement) First() lexer.Token { return e.For }
func (e *ForStatement) Last() lexer.Token  { return e.Body.Last() }
func (e *ForStatement) statementNode()     {}
