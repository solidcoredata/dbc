// Package pf is a parser framework.
package pf

// Pass 1: lex into tokens, comments, strings, braces.
// Pass 2:

type token interface{}

type identifier struct {
	Text string
}
