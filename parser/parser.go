// Copyright 2018 solidcoredata authors.

package parser

// Parser reads in source files and builds an AST tree from it. It may also verify
// the AST.
//
// The SQL parser will need to define temp tables. Because of this
// The SQL parser will be a super set of the Schema parser. Just use the same
// parser and process the AST according to various settings.
//
// Never "as". Always require tables to declare alias. All column references must
// use table alias. No where, only "and" or "or". Top level "and" or "or" do not
// require () but nested ones do.
//
// For select, insert, update statements: "name = t.name" is the same as "t.name".
type Parser struct{}
