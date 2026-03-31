package ast

import "monkey/token"

type Node interface {
	TokenLiteral() string //TokenLiteral() only for debugging/testing
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct { // Program node in root node of every AST
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// Fields for node
// Identifier, experession that produces a value in the let statement and one for the token

// LET STATEMENTS
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier // for reusability - even if identifier in statement does not produce value
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

//RETURN STATEMENTS
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
