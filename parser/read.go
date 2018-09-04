// Copyright 2018 solidcoredata authors.

// query syntax parser
package parser

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"unicode"
	"unicode/utf8"
)

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

func funcName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func run(ctx context.Context, src string, f func(Token)) error {
	l := &lexer{
		Source: src,
		Next:   f,
	}
	state := stWhitespace
	for state != nil && ctx.Err() == nil {
		state = state(ctx, l)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("%v: last state %s", ctx.Err(), funcName(state))
	}
	return nil
}

type invalid struct {
	v
	Message string
}
type newline struct {
	v
}
type ws struct {
	v
}

type sym struct {
	v
}

type str struct {
	v
	Raw       bool
	HasEscape bool // True if string has an escape sequence.
}

type number struct {
	v
}

type identifier struct {
	v
	Quoted bool
}

type comment struct {
	v
	Multiline bool
}

type Token interface {
	Value() string
}

type v string

func (v v) Value() string {
	return string(v)
}

type lexer struct {
	Source string
	At     int
	Start  int

	Next func(Token)

	runeSize int
}

type stateFn func(context.Context, *lexer) stateFn

const (
	lineComment  = "--"
	leftComment  = "/*"
	rightComment = "*/"
)

func (l *lexer) send(t Token) {
	if l.Next == nil {
		return
	}
	if len(t.Value()) == 0 {
		return
	}
	l.Next(t)
}

func (l *lexer) runeAt() (r rune) {
	r, l.runeSize = utf8.DecodeRuneInString(l.Source[l.At:])
	return
}

func (l *lexer) nextRune() {
	l.At += l.runeSize
	return
}

func (l *lexer) value() string {
	return l.Source[l.Start:l.At]
}

func (l *lexer) valueSync() v {
	s := l.Source[l.Start:l.At]
	l.Start = l.At
	return v(s)
}

func (*lexer) isIdentiferStart(r rune) bool {
	switch {
	default:
		return false
	case r == '_', unicode.IsLetter(r):
		return true
	}
}

func (*lexer) isQuoteIdentiferStart(r rune) bool {
	switch {
	default:
		return false
	case r == '"':
		return true
	}
}

func (*lexer) isIdentifer(r rune) bool {
	switch {
	default:
		return false
	case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
		return true
	}
}
func (*lexer) isWhiteSpace(r rune) bool {
	return unicode.IsSpace(r)
}
func (*lexer) isSymbol(r rune) bool {
	switch r {
	default:
		return false
	case '{', '}', '-', '/', '*', '(', ')', '+', '%', '<', '>', '=':
		return true
	}
}
func (*lexer) isNumberStart(r rune) bool {
	switch {
	default:
		return false
	case unicode.IsNumber(r):
		return true
	}
}
func (*lexer) isNumber(r rune) bool {
	switch {
	default:
		return false
	case unicode.IsNumber(r), r == '.', r == ',':
		return true
	}
}

func stIdentifier(ctx context.Context, l *lexer) stateFn {
	for ctx.Err() == nil {
		r := l.runeAt()

		switch {
		default:
			l.send(identifier{v: l.valueSync()})
			return stWhitespace
		case l.isIdentifer(r):
			l.nextRune()
		}
	}
	return nil
}
func stQuoteIdentifier(ctx context.Context, l *lexer) stateFn {
	r := l.runeAt()
	if r != '"' {
		panic("not starting a quoted identifier")
	}
	l.nextRune()
	for ctx.Err() == nil {
		r := l.runeAt()

		switch r {
		default:
			l.nextRune()
		case '"':
			l.nextRune()
			l.send(identifier{v: l.valueSync(), Quoted: true})
			return stWhitespace
		case '\n', '\r':
			l.send(invalid{v: l.valueSync(), Message: "quoted identifier not closed before newline"})
			return stWhitespace
		}
	}
	return nil
}

func stSymbol(ctx context.Context, l *lexer) stateFn {
	for ctx.Err() == nil {
		r := l.runeAt()

		switch {
		default:
			l.send(sym{v: l.valueSync()})
			return stWhitespace
		case l.isSymbol(r):
			l.nextRune()
		case l.value() == lineComment:
			l.nextRune()
			return stLineComment
		case l.value() == leftComment:
			l.nextRune()
			return stMultiComment
		}
	}
	return nil
}

func stLineComment(ctx context.Context, l *lexer) stateFn {
	for ctx.Err() == nil {
		r := l.runeAt()

		switch r {
		default:
			l.nextRune()
		case '\n', '\r':
			l.send(comment{v: l.valueSync()})
			return stWhitespace
		}
	}
	return nil
}

func stMultiComment(ctx context.Context, l *lexer) stateFn {
	for ctx.Err() == nil {
		l.runeAt()
		l.nextRune()

		end := l.Source[l.At-2 : l.At]

		if end == rightComment {
			l.send(comment{v: l.valueSync(), Multiline: true})
			return stWhitespace
		}
	}
	return nil
}

func stNumber(ctx context.Context, l *lexer) stateFn {
	for ctx.Err() == nil {
		r := l.runeAt()

		switch {
		default:
			l.send(number{v: l.valueSync()})
			return stWhitespace
		case l.isNumber(r):
			l.nextRune()
		}
	}
	return nil
}
func stString(ctx context.Context, l *lexer) stateFn {
	r := l.runeAt()
	if r != '\'' {
		panic("not starting a string")
	}
	l.nextRune()
	escapePresent := false
	for ctx.Err() == nil {
		r = l.runeAt()

		switch {
		default:
			l.nextRune()
		case r == '\'':
			l.nextRune()
			if l.runeAt() == '\'' {
				l.nextRune()
				escapePresent = true
				continue
			}
			l.send(str{v: l.valueSync(), HasEscape: escapePresent})
			return stWhitespace
		}
	}
	return nil
}

func stWhitespace(ctx context.Context, l *lexer) stateFn {
	if l.At >= len(l.Source) {
		return nil
	}

	for ctx.Err() == nil {
		r := l.runeAt()

		switch {
		default:
			l.send(ws{v: l.valueSync()})
			l.nextRune()
			l.send(invalid{v: l.valueSync(), Message: "unknown token"})
			l.nextRune()
			return nil
		case r == utf8.RuneError:
			return nil
		case r == '\n' || r == '\r':
			l.send(ws{v: l.valueSync()})
			l.nextRune()
			l.send(newline{v: l.valueSync()})
		case l.isWhiteSpace(r):
			l.nextRune()
		case l.isQuoteIdentiferStart(r):
			return stQuoteIdentifier
		case l.isIdentiferStart(r):
			l.send(ws{v: l.valueSync()})
			return stIdentifier
		case l.isSymbol(r):
			l.send(ws{v: l.valueSync()})
			return stSymbol
		case l.isNumberStart(r):
			l.send(ws{v: l.valueSync()})
			return stNumber
		case r == '\'':
			l.send(ws{v: l.valueSync()})
			return stString
		}
	}
	return nil
}
