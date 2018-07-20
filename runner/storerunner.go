package runner

import (
	"github.com/solidcoredata/dbc/parser"
)

type Param struct {
	Name  string
	Value interface{}
}

type Option struct {
	QueryName string

	Port  string
	Role  []string
	Param []Param
}

type StoreRunner interface {
	Run(s *parser.Store, opt Option) (parser.StreamingResultSet, error)
}
