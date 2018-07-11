package runner

import (
	"github.com/solidcoredata/dbc/parser"
)

type Param struct {
	Name  string
	Value interface{}
}

type StoreRunner interface {
	Run(s *parser.Store, iface, role, run string, param []Param) (parser.StreamingResultSet, error)
}
