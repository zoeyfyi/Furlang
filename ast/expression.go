package ast

import (
	"github.com/bongo227/Furlang/types"

	"github.com/bongo227/Furlang/lexer"
)

type Expression interface {
	Node
	expressionNode()
}

// TypeExpression is a type
type TypeExpression struct {
	Type types.Type
}

// TODO: add start and end to this node

func (e *TypeExpression) First() lexer.Token { return lexer.Token{} }
func (e *TypeExpression) Last() lexer.Token  { return lexer.Token{} }
func (e *TypeExpression) expressionNode()    {}

// IdentExpression is any identifier
type IdentExpression struct {
	Value lexer.Token
}

func (e *IdentExpression) First() lexer.Token { return e.Value }
func (e *IdentExpression) Last() lexer.Token  { return e.Value }
func (e *IdentExpression) expressionNode()    {}

// LiteralExpression is an expression in the form: interger || float || string
type LiteralExpression struct {
	Value lexer.Token
}

func (e *LiteralExpression) First() lexer.Token { return e.Value }
func (e *LiteralExpression) Last() lexer.Token  { return e.Value }
func (e *LiteralExpression) expressionNode()    {}

// BraceLiteralExpression is an expression in the form: type{expression, expression, ...}
type BraceLiteralExpression struct {
	Type       types.Type
	LeftBrace  lexer.Token
	Elements   []Expression
	RightBrace lexer.Token
}

func (e *BraceLiteralExpression) First() lexer.Token { return e.LeftBrace }
func (e *BraceLiteralExpression) Last() lexer.Token  { return e.RightBrace }
func (e *BraceLiteralExpression) expressionNode()    {}

// ParenLiteralExpression is an expression in the form: (expression, expression, ...)
type ParenLiteralExpression struct {
	LeftParen  lexer.Token
	Elements   []Expression
	RightParen lexer.Token
}

func (e *ParenLiteralExpression) First() lexer.Token { return e.LeftParen }
func (e *ParenLiteralExpression) Last() lexer.Token  { return e.RightParen }
func (e *ParenLiteralExpression) expressionNode()    {}

// IndexExpression is an expression in the form: expression[expression]
type IndexExpression struct {
	Expression Expression
	LeftBrack  lexer.Token
	Index      Expression
	RightBrack lexer.Token
}

func (e *IndexExpression) First() lexer.Token { return e.Expression.First() }
func (e *IndexExpression) Last() lexer.Token  { return e.RightBrack }
func (e *IndexExpression) expressionNode()    {}

// SliceExpression represents an expression in the form: expression[expression:expression]
type SliceExpression struct {
	Expression Expression
	LeftBrack  lexer.Token
	Low        Expression
	Colon      lexer.Token
	High       Expression
	RightBrack lexer.Token
}

func (e *SliceExpression) First() lexer.Token { return e.Expression.First() }
func (e *SliceExpression) Last() lexer.Token  { return e.RightBrack }
func (e *SliceExpression) expressionNode()    {}

// CallExpression is an expression in the form: expression(expression, expression, ...)
type CallExpression struct {
	Function  Expression
	Arguments *ParenLiteralExpression
}

func (e *CallExpression) First() lexer.Token { return e.Function.First() }
func (e *CallExpression) Last() lexer.Token  { return e.Arguments.Last() }
func (e *CallExpression) expressionNode()    {}

// CastExpression is an expression in the form: (type)expression
type CastExpression struct {
	LeftParen  lexer.Token
	Type       types.Type
	RightParen lexer.Token
	Expression Expression
}

func (e *CastExpression) First() lexer.Token { return e.LeftParen }
func (e *CastExpression) Last() lexer.Token  { return e.Expression.Last() }
func (e *CastExpression) expressionNode()    {}

// BinaryExpression is an expression in the form: expression operator expression
type BinaryExpression struct {
	IsFp     bool
	Left     Expression
	Operator lexer.Token
	Right    Expression
}

func (e *BinaryExpression) First() lexer.Token { return e.Left.First() }
func (e *BinaryExpression) Last() lexer.Token  { return e.Right.Last() }
func (e *BinaryExpression) expressionNode()    {}

// UnaryExpression is an expression in the form: operator expression
type UnaryExpression struct {
	Operator   lexer.Token
	Expression Expression
}

func (e *UnaryExpression) First() lexer.Token { return e.Operator }
func (e *UnaryExpression) Last() lexer.Token  { return e.Expression.Last() }
func (e *UnaryExpression) expressionNode()    {}
