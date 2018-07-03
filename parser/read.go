// Copyright 2018 solidcoredata authors.

// query syntax parser
package parser

// A Position describes the position between two bytes of input.
type Position struct {
	Line     int // line in input (starting at 1)
	LineRune int // rune in line (starting at 1)
	Byte     int // byte in input (starting at 0)
}

// An Expr represents an input element.
type Expr interface {
	// Span returns the start and end position of the expression,
	// excluding leading or trailing comments.
	Span() (start, end Position)

	// Comment returns the comments attached to the expression.
	Comment() []*Comment
}

type CommentPlacement uint8

const (
	Above CommentPlacement = iota
	Below
	Left
	Right
)

type Comment struct {
	Placement CommentPlacement
	Multiline bool // Comment is a /* */ multiline comment.
	Text      string
	Next      *Comment
}

// A FileSyntax represents an entire go.mod file.
type FileSyntax struct {
	Name string
	Stmt []Expr
}

func (x *FileSyntax) Span() (start, end Position) {
	if len(x.Stmt) == 0 {
		return
	}
	start, _ = x.Stmt[0].Span()
	_, end = x.Stmt[len(x.Stmt)-1].Span()
	return start, end
}
