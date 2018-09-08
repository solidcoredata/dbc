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

//go:generate stringer -type TokenType -trimprefix Token

type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenNewline
	TokenWS
	TokenSymbol
	TokenString
	TokenStringWithEscape
	TokenNumber
	TokenIdentifier
	TokenIdentifierQuoted
	TokenLineComment
	TokenMultiComment
)

// Position of a byte within a file.
type Position struct {
	Line     int // Line in input, starts at 1.
	LineRune int // Rune in line, starts at 1.
	Byte     int // Byte in input, starts at 0.
}

func newPos() Position {
	return Position{
		Line:     1,
		LineRune: 1,
	}
}

type Token struct {
	Type    TokenType
	Start   Position
	End     Position
	Value   string
	Message string
}

func funcName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func Lex(ctx context.Context, src string, f func(Token)) error {
	l := &lexer{
		source: src,
		next:   f,

		start: newPos(),
		end:   newPos(),
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

type lexer struct {
	source string
	next   func(Token)

	start Position
	end   Position

	runeSize     int
	currentRune  rune
	previousRune rune
}

type stateFn func(context.Context, *lexer) stateFn

const (
	lineComment  = "--"
	leftComment  = "/*"
	rightComment = "*/"
)

func (l *lexer) send(t TokenType) {
	l.sendMessage(t, "")
}

func (l *lexer) sendMessage(t TokenType, msg string) {
	if l.next == nil {
		return
	}
	start, end := l.start, l.end
	v := l.valueSync()
	if len(v) == 0 {
		return
	}
	e := Token{
		Type:    t,
		Value:   v,
		Message: msg,
		Start:   start,
		End:     end,
	}
	l.next(e)
}

func (l *lexer) runeAt() rune {
	l.previousRune = l.currentRune
	l.currentRune, l.runeSize = utf8.DecodeRuneInString(l.source[l.end.Byte:])
	return l.currentRune
}

func (l *lexer) nextRune() {
	l.end.Byte += l.runeSize
	if l.currentRune == '\n' {
		l.end.Line++
		l.end.LineRune = 1
	} else {
		l.end.LineRune++
	}
}

func (l *lexer) value() string {
	return l.source[l.start.Byte:l.end.Byte]
}

func (l *lexer) valueSync() string {
	s := l.source[l.start.Byte:l.end.Byte]
	l.start = l.end
	return s
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
	case '{', '}', '-', '/', '*', '(', ')', '+', '%', '<', '>', '=', '.', ',', ';':
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
			l.send(TokenIdentifier)
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
			l.send(TokenIdentifierQuoted)
			return stWhitespace
		case '\n', '\r':
			l.sendMessage(TokenInvalid, "quoted identifier not closed before newline")
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
			l.send(TokenSymbol)
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
			l.send(TokenLineComment)
			return stWhitespace
		}
	}
	return nil
}

func stMultiComment(ctx context.Context, l *lexer) stateFn {
	for ctx.Err() == nil {
		l.runeAt()
		l.nextRune()

		end := l.source[l.end.Byte-2 : l.end.Byte]

		if end == rightComment {
			l.send(TokenMultiComment)
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
			l.send(TokenNumber)
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
			if escapePresent {
				l.send(TokenStringWithEscape)
			} else {
				l.send(TokenString)
			}
			return stWhitespace
		}
	}
	return nil
}

func stWhitespace(ctx context.Context, l *lexer) stateFn {
	if l.end.Byte >= len(l.source) {
		return nil
	}

	for ctx.Err() == nil {
		r := l.runeAt()

		switch {
		default:
			l.send(TokenWS)
			l.nextRune()
			l.sendMessage(TokenInvalid, "unknown token")
			l.nextRune()
			return nil
		case r == utf8.RuneError:
			return nil
		case r == '\n' || r == '\r':
			l.send(TokenWS)
			l.nextRune()
			l.send(TokenNewline)
		case l.isWhiteSpace(r):
			l.nextRune()
		case l.isQuoteIdentiferStart(r):
			return stQuoteIdentifier
		case l.isIdentiferStart(r):
			l.send(TokenWS)
			return stIdentifier
		case l.isSymbol(r):
			l.send(TokenWS)
			return stSymbol
		case l.isNumberStart(r):
			l.send(TokenWS)
			return stNumber
		case r == '\'':
			l.send(TokenWS)
			return stString
		}
	}
	return nil
}
