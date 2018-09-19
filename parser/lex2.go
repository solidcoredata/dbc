// Copyright 2018 solidcoredata authors.

package parser

import (
	"context"
	"fmt"
)

type ParseError struct {
	FileName string
	Start    Position
	End      Position
	Message  string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.FileName, e.Start.Line, e.Start.LineRune, e.Message)
}

type File struct {
	Name   string
	Errors []ParseError

	DeclareOrder []string

	Table []Table
	Query []Query
}

type CommentPosition int

const (
	CommentAbove CommentPosition = iota
	CommentRight
	CommentBelow
	CommentLeft
)

type Comment struct {
	Pos       CommentPosition
	MultiLine bool
	Text      string
}

type Node interface {
	Pos() (start Position, end Position)
	Comments() []Comment
}

type Table struct {
	Name string

	Column []TableColumn
}
type TableColumn struct {
	Name string
	Type string
}
type TableIndex struct {
	Name string
}

type Query struct{}

func (f *File) err(tok Token, msg string) {
	f.Errors = append(f.Errors, ParseError{
		FileName: f.Name,
		Start:    tok.Start,
		End:      tok.End,
		Message:  msg,
	})
}

type Package struct {
	Name string
}

func Lex2(ctx context.Context, src string, f *File) error {
	tc := make(chan Token, 100)
	go func(ctx context.Context, tc chan Token) {
		type lstate1 int
		const (
			lvRoot lstate1 = iota
			lvPackage
			lvImport
			lvTable
			lvQuery
		)
		var st lstate1
		_ = st

		type lexRoute struct {
		}

		for {
			tok, ok := <-tc
			if !ok {
				return
			}
			switch tok.Type {
			default:
				panic("unknown token type")
			case TokenInvalid:
				f.err(tok, tok.Message)

			case TokenNewline:
			case TokenWS:
			case TokenSymbol:
			case TokenString:
			case TokenStringWithEscape:
			case TokenNumber:

			case TokenIdentifier:
			case TokenIdentifierQuoted:

			case TokenLineComment:
			case TokenMultiComment:
			}
		}
	}(ctx, tc)
	err := Lex1(ctx, src, tc)
	close(tc)
	return err
}
